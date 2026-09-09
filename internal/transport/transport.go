// Package transport implements the private HTTP transport used by the SDK.
package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/NodeOps-app/createos-go-sdk/internal/polling"
	"github.com/NodeOps-app/createos-go-sdk/internal/protocol"
	"github.com/NodeOps-app/createos-go-sdk/internal/redact"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

const (
	defaultTimeout   = 60 * time.Second
	defaultUserAgent = "createos-go-sdk"
	maxErrorBody     = 4 << 20
)

// RetryConfig controls exponential-backoff retries.
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Disabled   bool
}

// HookMetadata is redacted request metadata passed to observability hooks.
type HookMetadata struct {
	URL       string
	Method    string
	Headers   http.Header
	Attempt   int
	Status    int
	RequestID string
	Duration  time.Duration
	Delay     time.Duration
	Reason    string
}

// Hooks contains best-effort observability callbacks. Hook errors are ignored
// so instrumentation cannot make a request fail.
type Hooks struct {
	OnRequest  func(HookMetadata)
	OnResponse func(HookMetadata)
	OnRetry    func(HookMetadata)
}

// Config configures a Client.
type Config struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	UserAgent  string
	Timeout    time.Duration
	Retry      RetryConfig
	Hooks      Hooks
}

// RequestOptions configures one request.
type RequestOptions struct {
	Query        url.Values
	Headers      http.Header
	Body         any
	RawBody      io.Reader
	ContentType  string
	SkipAuth     bool
	Timeout      time.Duration
	DisableRetry bool
	Retry        *RetryConfig
}

// ResponseError describes a non-successful HTTP response.
type ResponseError struct {
	StatusCode int
	Body       []byte
	RequestID  string
	Code       int
	Endpoint   string
	Method     string
	Header     http.Header
}

func (e *ResponseError) Error() string {
	message := http.StatusText(e.StatusCode)
	var env structs.JSendEnvelope[json.RawMessage]
	if json.Unmarshal(e.Body, &env) == nil {
		if env.Message != "" {
			message = env.Message
		} else if dataMessage := failDataMessage(env.Data); dataMessage != "" {
			message = dataMessage
		}
	}
	if message == "" {
		message = "request failed"
	}
	return fmt.Sprintf("%s %s: HTTP %d: %s", e.Method, e.Endpoint, e.StatusCode, message)
}

func failDataMessage(data json.RawMessage) string {
	if len(data) == 0 || string(data) == "null" {
		return ""
	}
	var message string
	if json.Unmarshal(data, &message) == nil {
		return message
	}
	var fields map[string]string
	if json.Unmarshal(data, &fields) != nil || len(fields) == 0 {
		return ""
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+": "+fields[key])
	}
	return strings.Join(parts, "; ")
}

// Client executes requests against one control-plane origin.
type Client struct {
	baseURL    *url.URL
	baseOrigin string
	apiKey     string
	httpClient *http.Client
	userAgent  string
	timeout    time.Duration
	retry      RetryConfig
	hooks      Hooks
}

// New validates config and creates a transport client.
func New(config Config) (*Client, error) {
	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, errors.New("base URL scheme must be http or https")
	}
	if baseURL.Host == "" {
		return nil, errors.New("base URL must contain a host")
	}
	if baseURL.User != nil {
		return nil, errors.New("base URL must not contain user information")
	}
	baseURL.RawQuery = ""
	baseURL.Fragment = ""
	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + "/"

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{}
	}
	// Clone the caller's client before installing a strict redirect guard. Go's
	// default policy may forward sensitive headers to subdomains; SDK
	// credentials must never leave the exact configured origin.
	httpClient := *config.HTTPClient
	previousRedirectPolicy := httpClient.CheckRedirect
	baseOrigin := origin(baseURL)
	httpClient.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if origin(request.URL) != baseOrigin {
			return fmt.Errorf("refusing redirect to non-base origin %q", origin(request.URL))
		}
		if previousRedirectPolicy != nil {
			return previousRedirectPolicy(request, via)
		}
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}
	if config.UserAgent == "" {
		config.UserAgent = defaultUserAgent
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.Timeout < 0 {
		return nil, errors.New("timeout cannot be negative")
	}
	if config.Retry.MaxRetries < 0 {
		return nil, errors.New("maximum retries cannot be negative")
	}
	if !config.Retry.Disabled && config.Retry.MaxRetries == 0 && config.Retry.BaseDelay == 0 && config.Retry.MaxDelay == 0 {
		config.Retry.MaxRetries = 2
		config.Retry.BaseDelay = 500 * time.Millisecond
		config.Retry.MaxDelay = 30 * time.Second
	}
	if config.Retry.BaseDelay == 0 {
		config.Retry.BaseDelay = 500 * time.Millisecond
	}
	if config.Retry.MaxDelay == 0 {
		config.Retry.MaxDelay = 30 * time.Second
	}
	if config.Retry.BaseDelay < 0 || config.Retry.MaxDelay < config.Retry.BaseDelay {
		return nil, errors.New("retry delays are invalid")
	}

	return &Client{
		baseURL:    baseURL,
		baseOrigin: baseOrigin,
		apiKey:     config.APIKey,
		httpClient: &httpClient,
		userAgent:  config.UserAgent,
		timeout:    config.Timeout,
		retry:      config.Retry,
		hooks:      config.Hooks,
	}, nil
}

// EncodePath escapes one path segment.
func EncodePath(value string) string { return url.PathEscape(value) }

// Do executes a request, requires a successful status, and unwraps a JSend
// success envelope into out. A 204 or 205 response may have an empty body.
func (c *Client) Do(ctx context.Context, method, path string, opts RequestOptions, out any) error {
	response, err := c.DoRaw(ctx, method, path, opts)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return responseError(response, method, path)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read %s %s response: %w", strings.ToUpper(method), path, err)
	}
	if len(bytes.TrimSpace(body)) == 0 && (response.StatusCode == http.StatusNoContent || response.StatusCode == http.StatusResetContent) {
		return nil
	}
	data, err := protocol.UnmarshalJSend[json.RawMessage](body)
	if err != nil {
		return fmt.Errorf("decode %s %s response: %w", strings.ToUpper(method), path, err)
	}
	if out != nil && len(data) != 0 && string(data) != "null" {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decode %s %s response data: %w", strings.ToUpper(method), path, err)
		}
	}
	return nil
}

// DoRaw executes a request with retries and returns its final response without
// converting a non-2xx status to an error. The caller must close response.Body.
func (c *Client) DoRaw(ctx context.Context, method, path string, opts RequestOptions) (*http.Response, error) {
	prepared, err := c.prepare(method, path, opts)
	if err != nil {
		return nil, err
	}
	retry := c.requestRetry(opts.Retry)
	if retry.MaxRetries < 0 || retry.BaseDelay <= 0 || retry.MaxDelay < retry.BaseDelay {
		return nil, errors.New("request retry configuration is invalid")
	}
	maxRetries := retry.MaxRetries
	if retry.Disabled || opts.DisableRetry || opts.RawBody != nil {
		maxRetries = 0
	}

	for attempt := 0; ; attempt++ {
		meta := prepared.meta(attempt + 1)
		c.fire(c.hooks.OnRequest, meta)
		started := time.Now()
		response, requestErr := c.dispatch(ctx, prepared, opts.Timeout)
		if requestErr != nil {
			if attempt < maxRetries && isIdempotent(prepared.method) && retryableNetworkError(ctx, requestErr) {
				delay := backoff(attempt, retry)
				meta.Duration, meta.Delay, meta.Reason = time.Since(started), delay, "network"
				c.fire(c.hooks.OnRetry, meta)
				if err := polling.Sleep(ctx, delay); err != nil {
					return nil, err
				}
				continue
			}
			return nil, requestErr
		}

		meta.Status = response.StatusCode
		meta.RequestID = response.Header.Get("X-Request-ID")
		meta.Duration = time.Since(started)
		c.fire(c.hooks.OnResponse, meta)
		if attempt >= maxRetries || !isRetryableStatus(prepared.method, response.StatusCode) {
			return response, nil
		}

		delay, honored := retryAfter(response.Header.Get("Retry-After"), time.Now())
		if !honored {
			delay = backoff(attempt, retry)
		}
		meta.Delay = delay
		if honored {
			meta.Reason = "rate-limit"
		} else {
			meta.Reason = "status"
		}
		c.fire(c.hooks.OnRetry, meta)
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
		if err := polling.Sleep(ctx, delay); err != nil {
			return nil, err
		}
	}
}

// Stream executes one request without retries and requires a successful HTTP
// status. The caller owns the returned response body.
func (c *Client) Stream(ctx context.Context, method, path string, opts RequestOptions) (*http.Response, error) {
	opts.DisableRetry = true
	response, err := c.DoRaw(ctx, method, path, opts)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, responseError(response, method, path)
	}
	return response, nil
}

type preparedRequest struct {
	method  string
	path    string
	url     string
	headers http.Header
	body    []byte
	rawBody io.Reader
}

func (p preparedRequest) meta(attempt int) HookMetadata {
	return HookMetadata{URL: redact.URL(p.url), Method: p.method, Headers: redact.Headers(p.headers), Attempt: attempt}
}

func (c *Client) prepare(method, path string, opts RequestOptions) (preparedRequest, error) {
	if opts.Timeout < 0 {
		return preparedRequest{}, errors.New("request timeout cannot be negative")
	}
	requestURL, err := c.resolveURL(path, opts.Query)
	if err != nil {
		return preparedRequest{}, err
	}
	headers := opts.Headers.Clone()
	if headers == nil {
		headers = make(http.Header)
	}
	if headers.Get("Accept") == "" {
		headers.Set("Accept", "application/json")
	}
	if headers.Get("User-Agent") == "" {
		headers.Set("User-Agent", c.userAgent)
	}
	// Caller-supplied credentials must neither shadow SDK authentication nor
	// survive an unauthenticated health request.
	removeCredentials(headers)
	if !opts.SkipAuth && c.apiKey != "" {
		headers.Set("X-Api-Key", c.apiKey)
	} else if !opts.SkipAuth {
		return preparedRequest{}, errors.New("authentication is required: configure an API key")
	}

	var body []byte
	if opts.RawBody != nil && opts.Body != nil {
		return preparedRequest{}, errors.New("request cannot have both Body and RawBody")
	}
	if opts.Body != nil {
		body, err = json.Marshal(opts.Body)
		if err != nil {
			return preparedRequest{}, fmt.Errorf("encode request body: %w", err)
		}
		headers.Set("Content-Type", "application/json")
	} else if opts.RawBody != nil && opts.ContentType != "" {
		headers.Set("Content-Type", opts.ContentType)
	}
	return preparedRequest{method: strings.ToUpper(method), path: path, url: requestURL.String(), headers: headers, body: body, rawBody: opts.RawBody}, nil
}

func (c *Client) resolveURL(path string, query url.Values) (*url.URL, error) {
	reference, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("parse request path: %w", err)
	}
	// Endpoint paths are relative to the configured base path even when callers
	// use the conventional leading slash (for example, "/v1/sandboxes").
	if !reference.IsAbs() && reference.Host == "" {
		reference.Path = strings.TrimLeft(reference.Path, "/")
	}
	resolved := c.baseURL.ResolveReference(reference)
	if origin(resolved) != c.baseOrigin {
		return nil, fmt.Errorf("refusing request to non-base origin %q", origin(resolved))
	}
	if resolved.User != nil {
		return nil, errors.New("request URL must not contain user information")
	}
	values := resolved.Query()
	for key, entries := range query {
		values.Del(key)
		for _, value := range entries {
			values.Add(key, value)
		}
	}
	resolved.RawQuery = values.Encode()
	return resolved, nil
}

func (c *Client) dispatch(ctx context.Context, prepared preparedRequest, timeout time.Duration) (*http.Response, error) {
	if timeout == 0 {
		timeout = c.timeout
	}
	requestCtx := ctx
	cancel := func() {}
	if timeout > 0 {
		requestCtx, cancel = context.WithTimeout(ctx, timeout)
	}

	var body io.Reader
	if prepared.rawBody != nil {
		body = prepared.rawBody
	} else if prepared.body != nil {
		body = bytes.NewReader(prepared.body)
	}
	request, err := http.NewRequestWithContext(requestCtx, prepared.method, prepared.url, body)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create HTTP request: %w", err)
	}
	request.Header = prepared.headers.Clone()
	response, err := c.httpClient.Do(request)
	if err != nil {
		cancel()
		return nil, err
	}
	response.Body = &cancelReadCloser{ReadCloser: response.Body, cancel: cancel}
	return response, nil
}

type cancelReadCloser struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (r *cancelReadCloser) Close() error {
	err := r.ReadCloser.Close()
	r.cancel()
	return err
}

func responseError(response *http.Response, method, endpoint string) error {
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, maxErrorBody))
	var env structs.JSendEnvelope[json.RawMessage]
	_ = json.Unmarshal(body, &env)
	return &ResponseError{
		StatusCode: response.StatusCode,
		Body:       body,
		RequestID:  response.Header.Get("X-Request-ID"),
		Code:       env.Code,
		Endpoint:   endpoint,
		Method:     strings.ToUpper(method),
		Header:     response.Header.Clone(),
	}
}

func origin(value *url.URL) string {
	return strings.ToLower(value.Scheme) + "://" + strings.ToLower(value.Host)
}

func removeCredentials(headers http.Header) {
	for name := range headers {
		if redact.IsSensitiveHeader(name) {
			headers.Del(name)
		}
	}
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

func isRetryableStatus(method string, status int) bool {
	if status == http.StatusTooManyRequests || status == http.StatusServiceUnavailable {
		return true
	}
	if !isIdempotent(method) {
		return false
	}
	switch status {
	case http.StatusRequestTimeout, http.StatusInternalServerError, http.StatusBadGateway, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func retryableNetworkError(ctx context.Context, err error) bool {
	return ctx.Err() == nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
}

func (c *Client) requestRetry(override *RetryConfig) RetryConfig {
	if override == nil {
		return c.retry
	}
	retry := *override
	if retry.BaseDelay == 0 {
		retry.BaseDelay = c.retry.BaseDelay
	}
	if retry.MaxDelay == 0 {
		retry.MaxDelay = c.retry.MaxDelay
	}
	return retry
}

func backoff(attempt int, retry RetryConfig) time.Duration {
	exponential := retry.BaseDelay * time.Duration(1<<min(attempt, 30))
	// Jitter does not protect a secret; a fast pseudorandom source is correct.
	jitter := time.Duration(rand.Int64N(max(int64(retry.BaseDelay), 1))) //nolint:gosec
	return min(exponential+jitter, retry.MaxDelay)
}

func retryAfter(value string, now time.Time) (time.Duration, bool) {
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second, true
	}
	when, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}
	return max(when.Sub(now), 0), true
}

func (c *Client) fire(hook func(HookMetadata), metadata HookMetadata) {
	if hook == nil {
		return
	}
	func() {
		defer func() { _ = recover() }()
		hook(metadata)
	}()
}
