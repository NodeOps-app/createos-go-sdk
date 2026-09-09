package sandbox

import (
	"net/http"
	"testing"
	"time"
)

func TestResolveConfigPrecedence(t *testing.T) {
	t.Setenv("CREATEOS_SANDBOX_API_KEY", "environment-key")
	t.Setenv("CREATEOS_SANDBOX_BASE_URL", "https://environment.example")

	configuration, err := resolveConfig(
		WithAPIKey("option-key"),
		WithBaseURL("https://option.example/"),
		WithTimeout(12*time.Second),
	)
	if err != nil {
		t.Fatalf("resolveConfig() error = %v", err)
	}
	if configuration.apiKey != "option-key" {
		t.Errorf("apiKey = %q, want option-key", configuration.apiKey)
	}
	if configuration.baseURL != "https://option.example" {
		t.Errorf("baseURL = %q, want https://option.example", configuration.baseURL)
	}
	if configuration.timeout != 12*time.Second {
		t.Errorf("request timeout = %v, want 12s", configuration.timeout)
	}
	if configuration.httpClient.Timeout != 0 {
		t.Errorf("HTTP client timeout = %v, want request contexts to own the timeout", configuration.httpClient.Timeout)
	}
}

func TestResolveConfigPreservesCustomHTTPClient(t *testing.T) {
	t.Setenv("CREATEOS_SANDBOX_BASE_URL", "")
	custom := &http.Client{Timeout: 7 * time.Second}
	configuration, err := resolveConfig(WithHTTPClient(custom))
	if err != nil {
		t.Fatalf("resolveConfig() error = %v", err)
	}
	if configuration.httpClient != custom {
		t.Error("resolveConfig() replaced the custom HTTP client")
	}
	if custom.Timeout != 7*time.Second {
		t.Errorf("custom HTTP client timeout = %v, want 7s", custom.Timeout)
	}
}

func TestResolveConfigRejectsInvalidBaseURL(t *testing.T) {
	t.Setenv("CREATEOS_SANDBOX_BASE_URL", "")
	_, err := resolveConfig(WithBaseURL("https://example.com?token=secret"))
	if err == nil {
		t.Fatal("resolveConfig() error = nil, want an invalid base URL error")
	}
}
