package sandbox

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func jsonResponse(request *http.Request, body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Request:    request,
	}
}

func TestNewClientInitializesServices(t *testing.T) {
	client, err := NewClient(WithAPIKey("test-key"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client.Templates() == nil || client.Networks() == nil || client.Disks() == nil {
		t.Fatal("NewClient() did not initialize every service")
	}
}

func TestClientHealthzDoesNotSendAuthentication(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if got := request.Header.Get("X-Api-Key"); got != "" {
			t.Errorf("X-Api-Key = %q, want empty", got)
		}
		return jsonResponse(request, `{"status":"success","data":{"up":true}}`), nil
	})}

	client, err := NewClient(WithAPIKey("test-key"), WithHTTPClient(httpClient), WithoutRetry())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	health, err := client.Healthz(context.Background())
	if err != nil {
		t.Fatalf("Healthz() error = %v", err)
	}
	if !health.Up {
		t.Error("Healthz().Up = false, want true")
	}
}

func TestClientWhoAmISendsAPIKey(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if got := request.Header.Get("X-Api-Key"); got != "test-key" {
			t.Errorf("X-Api-Key = %q, want test-key", got)
		}
		return jsonResponse(request, `{"status":"success","data":{"user_id":"user-1","stats":{"total":3}}}`), nil
	})}

	client, err := NewClient(WithAPIKey("test-key"), WithHTTPClient(httpClient), WithoutRetry())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	identity, err := client.WhoAmI(context.Background())
	if err != nil {
		t.Fatalf("WhoAmI() error = %v", err)
	}
	if identity.UserID != "user-1" || identity.Stats.Total != 3 {
		t.Errorf("WhoAmI() = %+v, want user-1 with total 3", identity)
	}
}
