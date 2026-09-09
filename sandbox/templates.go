package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/NodeOps-app/createos-go-sdk/internal/protocol"
	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// TemplatesService manages custom root filesystem templates.
type TemplatesService struct{ transport *transport.Client }

// List returns templates owned by the caller.
func (s *TemplatesService) List(ctx context.Context, options structs.PaginationOptions) ([]structs.TemplateView, error) {
	query := make(url.Values)
	if options.Offset > 0 {
		query.Set("offset", strconv.Itoa(options.Offset))
	}
	templates, err := fetchAll[structs.TemplateView](ctx, s.transport, "/v1/templates", query, options.Limit, "templates")
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	return templates, nil
}

// Create submits a Dockerfile to build into a reusable root filesystem.
func (s *TemplatesService) Create(ctx context.Context, request structs.TemplateCreateRequest) (*structs.TemplateView, error) {
	var result structs.TemplateView
	if err := s.transport.Do(ctx, http.MethodPost, "/v1/templates", transport.RequestOptions{Body: request}, &result); err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	return &result, nil
}

// Get returns a template by ID.
func (s *TemplatesService) Get(ctx context.Context, templateID string, options structs.GetTemplateOptions) (*structs.TemplateView, error) {
	query := make(url.Values)
	if options.Include != "" {
		query.Set("include", string(options.Include))
	}
	var result structs.TemplateView
	path := "/v1/templates/" + transport.EncodePath(templateID)
	requestOptions := transportOptions(options.RequestOptions)
	requestOptions.Query = query
	if err := s.transport.Do(ctx, http.MethodGet, path, requestOptions, &result); err != nil {
		return nil, fmt.Errorf("get template %q: %w", templateID, err)
	}
	return &result, nil
}

// Delete removes a template. Existing sandboxes remain unaffected.
func (s *TemplatesService) Delete(ctx context.Context, templateID string) error {
	path := "/v1/templates/" + transport.EncodePath(templateID)
	var result structs.OKResponse
	if err := s.transport.Do(ctx, http.MethodDelete, path, transport.RequestOptions{}, &result); err != nil {
		return fmt.Errorf("delete template %q: %w", templateID, err)
	}
	return nil
}

// Logs returns the build log collected so far as plain text.
func (s *TemplatesService) Logs(ctx context.Context, templateID string, options structs.TemplateLogsOptions) (string, error) {
	query := make(url.Values)
	if options.Attempt > 0 {
		query.Set("attempt", strconv.Itoa(options.Attempt))
	}
	path := "/v1/templates/" + transport.EncodePath(templateID) + "/logs"
	requestOptions := transportOptions(options.RequestOptions)
	requestOptions.Query = query
	response, err := s.transport.DoRaw(ctx, http.MethodGet, path, requestOptions)
	if err != nil {
		return "", fmt.Errorf("get template %q logs: %w", templateID, err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read template %q logs: %w", templateID, err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", responseAPIError(response, http.MethodGet, path, body)
	}
	return string(body), nil
}

// TemplateLogStream incrementally reads template build-log events.
type TemplateLogStream struct {
	decoder *protocol.NDJSONDecoder[structs.TemplateLogEvent]
}

// Recv returns the next event or io.EOF when the stream ends.
func (s *TemplateLogStream) Recv() (structs.TemplateLogEvent, error) { return s.decoder.Recv() }

// Close releases the underlying HTTP response body.
func (s *TemplateLogStream) Close() error { return s.decoder.Close() }

// FollowLogs follows a template build as newline-delimited log events.
func (s *TemplatesService) FollowLogs(ctx context.Context, templateID string, options structs.TemplateLogsOptions) (*TemplateLogStream, error) {
	query := make(url.Values)
	query.Set("follow", "true")
	if options.Attempt > 0 {
		query.Set("attempt", strconv.Itoa(options.Attempt))
	}
	path := "/v1/templates/" + transport.EncodePath(templateID) + "/logs"
	requestOptions := transportOptions(options.RequestOptions)
	requestOptions.Query = query
	// TemplateLogStream owns and closes the response body.
	response, err := s.transport.Stream(ctx, http.MethodGet, path, requestOptions) //nolint:bodyclose
	if err != nil {
		return nil, fmt.Errorf("follow template %q logs: %w", templateID, err)
	}
	return &TemplateLogStream{decoder: protocol.NewNDJSONDecoder[structs.TemplateLogEvent](response.Body)}, nil
}

func responseAPIError(response *http.Response, method, path string, body []byte) error {
	var envelope structs.JSendEnvelope[json.RawMessage]
	_ = json.Unmarshal(body, &envelope)
	return &APIError{
		StatusCode: response.StatusCode,
		Body:       body,
		RequestID:  response.Header.Get("X-Request-ID"),
		Code:       envelope.Code,
		Endpoint:   path,
		Method:     method,
		Header:     response.Header.Clone(),
	}
}
