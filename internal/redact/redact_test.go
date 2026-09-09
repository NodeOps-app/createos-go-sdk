package redact

import (
	"net/http"
	"strings"
	"testing"
)

func TestHeaders(t *testing.T) {
	t.Parallel()
	original := http.Header{"Authorization": {"Bearer secret"}, "Cookie": {"session=secret"}, "X-Trace": {"safe"}}
	got := Headers(original)
	if got.Get("Authorization") != replacement || got.Get("Cookie") != replacement {
		t.Fatalf("Headers() = %#v", got)
	}
	if got.Get("X-Trace") != "safe" || original.Get("Authorization") != "Bearer secret" {
		t.Fatal("Headers changed a safe value or mutated its input")
	}
}

func TestURL(t *testing.T) {
	t.Parallel()
	got := URL("https://user:pass@example.com/path?token=secret&name=safe")
	if strings.Contains(got, "secret") || strings.Contains(got, "pass") {
		t.Fatalf("URL() leaked a credential: %q", got)
	}
	if !strings.Contains(got, "name=safe") {
		t.Fatalf("URL() removed safe query: %q", got)
	}
}
