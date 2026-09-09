package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

const maximumPageSize = 500

// Client is the authenticated entry point to the CreateOS Sandbox API.
type Client struct {
	transport *transport.Client
	baseURL   string
	templates *TemplatesService
	networks  *NetworksService
	disks     *DisksService
}

// NewClient creates a fully initialized client. The API key and base URL can
// also be read from CREATEOS_SANDBOX_API_KEY and CREATEOS_SANDBOX_BASE_URL.
func NewClient(options ...ClientOption) (*Client, error) {
	configuration, err := resolveConfig(options...)
	if err != nil {
		return nil, fmt.Errorf("create sandbox client: %w", err)
	}

	httpTransport, err := transport.New(transport.Config{
		BaseURL:    configuration.baseURL,
		APIKey:     configuration.apiKey,
		HTTPClient: configuration.httpClient,
		UserAgent:  configuration.userAgent,
		Timeout:    configuration.timeout,
		Retry: transport.RetryConfig{
			MaxRetries: configuration.retry.maxRetries,
			BaseDelay:  configuration.retry.baseDelay,
			MaxDelay:   configuration.retry.maxDelay,
			Disabled:   configuration.retry.maxRetries == 0,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create HTTP transport: %w", err)
	}

	client := &Client{transport: httpTransport, baseURL: configuration.baseURL}
	client.templates = &TemplatesService{transport: httpTransport}
	client.networks = &NetworksService{transport: httpTransport}
	client.disks = &DisksService{transport: httpTransport}
	return client, nil
}

// BaseURL returns the configured control-plane base URL.
func (c *Client) BaseURL() string { return c.baseURL }

// Templates returns the automatically initialized template service.
func (c *Client) Templates() *TemplatesService { return c.templates }

// Networks returns the automatically initialized network service.
func (c *Client) Networks() *NetworksService { return c.networks }

// Disks returns the automatically initialized disk service.
func (c *Client) Disks() *DisksService { return c.disks }

// Healthz returns the unauthenticated liveness state of the control plane.
func (c *Client) Healthz(ctx context.Context) (*structs.HealthzResponse, error) {
	var result structs.HealthzResponse
	if err := c.transport.Do(ctx, http.MethodGet, "/healthz", transport.RequestOptions{SkipAuth: true}, &result); err != nil {
		return nil, fmt.Errorf("check control-plane health: %w", err)
	}
	return &result, nil
}

// Readyz returns the unauthenticated readiness state. HTTP 503 is returned as
// a normal Ready=false result because it is a readiness signal.
func (c *Client) Readyz(ctx context.Context) (*structs.ReadyzResponse, error) {
	const path = "/readyz"
	response, err := c.transport.DoRaw(ctx, http.MethodGet, path, transport.RequestOptions{
		SkipAuth:     true,
		DisableRetry: true,
	})
	if err != nil {
		return nil, fmt.Errorf("check control-plane readiness: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read control-plane readiness: %w", err)
	}
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusServiceUnavailable {
		return nil, responseAPIError(response, http.MethodGet, path, body)
	}

	var envelope structs.JSendEnvelope[*structs.ReadyzResponse]
	if err := json.Unmarshal(body, &envelope); err != nil || envelope.Data == nil {
		// Preserve the readiness meaning for empty or non-envelope 200/503
		// responses, matching the control-plane health contract.
		return &structs.ReadyzResponse{Ready: response.StatusCode == http.StatusOK}, nil
	}
	return envelope.Data, nil
}

// WhoAmI returns the identity associated with the configured API key.
func (c *Client) WhoAmI(ctx context.Context) (*structs.WhoAmIView, error) {
	var result structs.WhoAmIView
	if err := c.transport.Do(ctx, http.MethodGet, "/v1/whoami", transport.RequestOptions{}, &result); err != nil {
		return nil, fmt.Errorf("get caller identity: %w", err)
	}
	return &result, nil
}

// CreateSandbox creates a running sandbox and returns its connected handle.
func (c *Client) CreateSandbox(ctx context.Context, request structs.CreateSandboxRequest, options ...structs.CreateSandboxOptions) (*Instance, error) {
	requestOptions := transport.RequestOptions{Body: request}
	if len(options) > 0 {
		requestOptions = transportOptions(options[0].RequestOptions)
		requestOptions.Body = request
	}

	var created structs.CreateSandboxResponse
	if err := c.transport.Do(ctx, http.MethodPost, "/v1/sandboxes", requestOptions, &created); err != nil {
		return nil, fmt.Errorf("create sandbox: %w", err)
	}
	view := &structs.SandboxResponse{
		ID:                 created.ID,
		Status:             created.Status,
		IPAddress:          &created.IPAddress,
		VCPU:               created.VCPU,
		MemoryMiB:          created.MemoryMiB,
		DiskMiB:            created.DiskMiB,
		IngressEnabled:     request.IngressEnabled,
		IngressURLTemplate: created.IngressURLTemplate,
		Name:               created.Name,
		SpawnMilliseconds:  created.SpawnMilliseconds,
		Shape:              created.Shape,
		RootFS:             created.RootFS,
		EgressRules:        created.EgressRules,
	}
	return newInstance(c.transport, *view), nil
}

// GetSandbox returns a connected handle for a sandbox ID.
func (c *Client) GetSandbox(ctx context.Context, sandboxID string) (*Instance, error) {
	return c.getSandbox(ctx, "/v1/sandboxes/"+transport.EncodePath(sandboxID), "get sandbox")
}

// GetSandboxByIP returns a connected handle for a sandbox private IP address.
func (c *Client) GetSandboxByIP(ctx context.Context, ipAddress string) (*Instance, error) {
	return c.getSandbox(ctx, "/v1/sandboxes/by-ip/"+transport.EncodePath(ipAddress), "get sandbox by IP")
}

func (c *Client) getSandbox(ctx context.Context, path, operation string) (*Instance, error) {
	var view structs.SandboxResponse
	if err := c.transport.Do(ctx, http.MethodGet, path, transport.RequestOptions{}, &view); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	return newInstance(c.transport, view), nil
}

// ListSandboxes returns connected handles for sandboxes visible to the caller.
func (c *Client) ListSandboxes(ctx context.Context, options structs.ListSandboxesOptions) ([]*Instance, error) {
	query := make(url.Values)
	if options.Status != "" {
		query.Set("status", string(options.Status))
	}
	requestOptions := transportOptions(options.RequestOptions)
	views, err := fetchAll[structs.SandboxResponse](ctx, c.transport, "/v1/sandboxes", query, options.Limit, "", requestOptions)
	if err != nil {
		return nil, fmt.Errorf("list sandboxes: %w", err)
	}
	instances := make([]*Instance, 0, len(views))
	for index := range views {
		instances = append(instances, newInstance(c.transport, views[index]))
	}
	return instances, nil
}

type pagePayload[T any] struct {
	Items []T
	Total *int
}

func fetchAll[T any](ctx context.Context, client *transport.Client, path string, query url.Values, resultLimit int, legacyKey string, baseOptions ...transport.RequestOptions) ([]T, error) {
	if query == nil {
		query = make(url.Values)
	} else {
		query = cloneValues(query)
	}
	items := make([]T, 0)
	offset, _ := strconv.Atoi(query.Get("offset"))
	for {
		pageSize := maximumPageSize
		if resultLimit > 0 && resultLimit-len(items) < pageSize {
			pageSize = resultLimit - len(items)
		}
		if pageSize <= 0 {
			break
		}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))
		requestOptions := transport.RequestOptions{}
		if len(baseOptions) != 0 {
			requestOptions = baseOptions[0]
		}
		requestOptions.Query = query
		var raw json.RawMessage
		if err := client.Do(ctx, http.MethodGet, path, requestOptions, &raw); err != nil {
			return nil, err
		}
		page, err := decodePage[T](raw, legacyKey)
		if err != nil {
			return nil, fmt.Errorf("decode collection response: %w", err)
		}
		items = append(items, page.Items...)
		if len(page.Items) == 0 || page.Total == nil || offset+len(page.Items) >= *page.Total {
			break
		}
		offset += len(page.Items)
	}
	if resultLimit > 0 && len(items) > resultLimit {
		items = items[:resultLimit]
	}
	return items, nil
}

func paginationQuery(offset int) url.Values {
	query := make(url.Values)
	if offset > 0 {
		query.Set("offset", strconv.Itoa(offset))
	}
	return query
}

func cloneValues(values url.Values) url.Values {
	clone := make(url.Values, len(values))
	for key, entries := range values {
		clone[key] = append([]string(nil), entries...)
	}
	return clone
}

func decodePage[T any](raw json.RawMessage, legacyKey string) (pagePayload[T], error) {
	var direct []T
	if err := json.Unmarshal(raw, &direct); err == nil {
		return pagePayload[T]{Items: direct}, nil
	}
	var object structs.PaginatedResponse[T]
	if err := json.Unmarshal(raw, &object); err != nil {
		return pagePayload[T]{}, err
	}
	if object.Data != nil {
		total := object.Pagination.Total
		return pagePayload[T]{Items: object.Data, Total: &total}, nil
	}
	if legacyKey != "" {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			return pagePayload[T]{}, err
		}
		if field, ok := fields[legacyKey]; ok {
			if err := json.Unmarshal(field, &object.Data); err != nil {
				return pagePayload[T]{}, err
			}
			return pagePayload[T]{Items: object.Data}, nil
		}
	}
	return pagePayload[T]{Items: []T{}}, nil
}
