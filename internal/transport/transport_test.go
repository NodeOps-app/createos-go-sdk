package transport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func newTestClient(t *testing.T, fn roundTripFunc, retry RetryConfig) *Client {
	t.Helper()
	client, err := New(Config{
		BaseURL:    "https://sandbox.example.test/v1/",
		APIKey:     "secret",
		HTTPClient: &http.Client{Transport: fn},
		Retry:      retry,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return client
}

func TestDoBuildsRequestAndUnwrapsJSend(t *testing.T) {
	t.Parallel()
	client := newTestClient(t, func(request *http.Request) (*http.Response, error) {
		if request.URL.String() != "https://sandbox.example.test/v1/sandboxes?limit=10" {
			t.Fatalf("URL = %q", request.URL)
		}
		if request.Header.Get("X-Api-Key") != "secret" {
			t.Fatalf("X-Api-Key = %q", request.Header.Get("X-Api-Key"))
		}
		if request.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("Content-Type = %q", request.Header.Get("Content-Type"))
		}
		return response(http.StatusOK, `{"status":"success","data":{"id":"box-1"}}`), nil
	}, RetryConfig{Disabled: true})

	var got struct {
		ID string `json:"id"`
	}
	err := client.Do(context.Background(), http.MethodPost, "sandboxes", RequestOptions{
		Query: url.Values{"limit": {"10"}},
		Body:  map[string]string{"shape": "small"},
	}, &got)
	if err != nil || got.ID != "box-1" {
		t.Fatalf("Do() = %#v, %v", got, err)
	}
}

func TestLeadingSlashPreservesBasePath(t *testing.T) {
	t.Parallel()
	client := newTestClient(t, func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != "/v1/healthz" {
			t.Fatalf("path = %q, want /v1/healthz", request.URL.Path)
		}
		return response(http.StatusNoContent, ""), nil
	}, RetryConfig{Disabled: true})
	if err := client.Do(context.Background(), http.MethodGet, "/healthz", RequestOptions{SkipAuth: true}, nil); err != nil {
		t.Fatalf("Do() error = %v", err)
	}
}

func TestDoReturnsResponseError(t *testing.T) {
	t.Parallel()
	client := newTestClient(t, func(*http.Request) (*http.Response, error) {
		result := response(http.StatusInternalServerError, `{"status":"error","message":"internal error","code":500}`)
		result.Header.Set("X-Request-ID", "request-1")
		return result, nil
	}, RetryConfig{Disabled: true})

	err := client.Do(context.Background(), http.MethodPost, "sandboxes", RequestOptions{}, nil)
	var responseErr *ResponseError
	if !errors.As(err, &responseErr) {
		t.Fatalf("error = %T, want *ResponseError", err)
	}
	if responseErr.StatusCode != 500 || responseErr.Code != 500 || responseErr.RequestID != "request-1" {
		t.Fatalf("ResponseError = %#v", responseErr)
	}
}

func TestResponseErrorIncludesFailData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "string", body: `{"status":"fail","data":"sandbox not found"}`, want: "sandbox not found"},
		{name: "fields", body: `{"status":"fail","data":{"shape":"required"}}`, want: "shape: required"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := (&ResponseError{StatusCode: http.StatusBadRequest, Method: http.MethodPost, Endpoint: "/v1/sandboxes", Body: []byte(test.body)}).Error()
			if !strings.Contains(err, test.want) {
				t.Fatalf("Error() = %q, want it to contain %q", err, test.want)
			}
		})
	}
}

func TestCrossOriginRequestIsRejected(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	client := newTestClient(t, func(*http.Request) (*http.Response, error) {
		calls.Add(1)
		return response(http.StatusOK, ""), nil
	}, RetryConfig{Disabled: true})
	_, err := client.DoRaw(context.Background(), http.MethodGet, "https://attacker.example/steal", RequestOptions{}) //nolint:bodyclose // Rejected before a response exists.
	if err == nil || calls.Load() != 0 {
		t.Fatalf("DoRaw() error = %v, calls = %d", err, calls.Load())
	}
}

func TestCrossOriginRedirectIsRejected(t *testing.T) {
	t.Parallel()
	client := newTestClient(t, func(*http.Request) (*http.Response, error) {
		return response(http.StatusOK, ""), nil
	}, RetryConfig{Disabled: true})
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://attacker.example/steal", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.httpClient.CheckRedirect(request, nil); err == nil {
		t.Fatal("CheckRedirect() error = nil, want cross-origin rejection")
	}
}

func TestSkipAuthRemovesCredentials(t *testing.T) {
	t.Parallel()
	client := newTestClient(t, func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Authorization") != "" || request.Header.Get("Cookie") != "" {
			t.Fatalf("credentials were sent: %#v", request.Header)
		}
		return response(http.StatusNoContent, ""), nil
	}, RetryConfig{Disabled: true})
	err := client.Do(context.Background(), http.MethodGet, "healthz", RequestOptions{
		SkipAuth: true,
		Headers:  http.Header{"Authorization": {"Basic secret"}, "Cookie": {"secret"}},
	}, nil)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
}

func TestRetryPolicy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		method    string
		status    int
		wantCalls int32
	}{
		{name: "GET 500", method: http.MethodGet, status: 500, wantCalls: 2},
		{name: "POST 500", method: http.MethodPost, status: 500, wantCalls: 1},
		{name: "POST 503", method: http.MethodPost, status: 503, wantCalls: 2},
		{name: "PATCH 429", method: http.MethodPatch, status: 429, wantCalls: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var calls atomic.Int32
			client := newTestClient(t, func(*http.Request) (*http.Response, error) {
				if calls.Add(1) == 1 {
					result := response(test.status, "failed")
					result.Header.Set("Retry-After", "0")
					return result, nil
				}
				return response(http.StatusOK, `{"status":"success","data":null}`), nil
			}, RetryConfig{MaxRetries: 1, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond})
			result, err := client.DoRaw(context.Background(), test.method, "sandboxes", RequestOptions{})
			if err != nil {
				t.Fatalf("DoRaw() error = %v", err)
			}
			result.Body.Close()
			if calls.Load() != test.wantCalls {
				t.Fatalf("calls = %d, want %d", calls.Load(), test.wantCalls)
			}
		})
	}
}

func TestNetworkRetryOnlyForIdempotentMethod(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	client := newTestClient(t, func(*http.Request) (*http.Response, error) {
		if calls.Add(1) == 1 {
			return nil, errors.New("temporary network failure")
		}
		return response(http.StatusOK, "ok"), nil
	}, RetryConfig{MaxRetries: 1, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond})
	result, err := client.DoRaw(context.Background(), http.MethodGet, "sandboxes", RequestOptions{})
	if err != nil {
		t.Fatalf("DoRaw() error = %v", err)
	}
	result.Body.Close()
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
}

func TestStreamDoesNotRetry(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	client := newTestClient(t, func(*http.Request) (*http.Response, error) {
		calls.Add(1)
		return response(http.StatusServiceUnavailable, `{"status":"error","message":"busy"}`), nil
	}, RetryConfig{MaxRetries: 2, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond})
	_, err := client.Stream(context.Background(), http.MethodGet, "commands/stream", RequestOptions{}) //nolint:bodyclose // The error response is closed by Stream.
	if err == nil || calls.Load() != 1 {
		t.Fatalf("Stream() error = %v, calls = %d", err, calls.Load())
	}
}

func TestHookMetadataIsRedacted(t *testing.T) {
	t.Parallel()
	var metadata HookMetadata
	client := newTestClient(t, func(*http.Request) (*http.Response, error) {
		return response(http.StatusOK, "ok"), nil
	}, RetryConfig{Disabled: true})
	client.hooks.OnRequest = func(value HookMetadata) { metadata = value }
	result, err := client.DoRaw(context.Background(), http.MethodGet, "sandboxes?token=secret", RequestOptions{})
	if err != nil {
		t.Fatalf("DoRaw() error = %v", err)
	}
	result.Body.Close()
	if strings.Contains(metadata.URL, "secret") || metadata.Headers.Get("X-Api-Key") != "<redacted>" {
		t.Fatalf("hook metadata leaked credentials: %#v", metadata)
	}
}
