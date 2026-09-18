package sandbox

import (
	"net/http"
	"testing"
	"time"

	"github.com/NodeOps-app/createos-go-sdk/structs"
)

func TestInstanceAccessTokenLifecycle(t *testing.T) {
	requests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/v1/sandboxes/sb-1/access-token", `{"status":"success","data":{"token":"skp_sb_first","enabled":true,"created_at":"2026-09-18T10:00:00Z"}}`},
		{http.MethodGet, "/v1/sandboxes/sb-1/access-token", `{"status":"success","data":{"enabled":true,"token_hint":"skp_sb_fi...irst","created_at":"2026-09-18T10:00:00Z"}}`},
		{http.MethodPost, "/v1/sandboxes/sb-1/access-token/rotate", `{"status":"success","data":{"token":"skp_sb_second","enabled":true,"created_at":"2026-09-18T10:00:00Z","rotated_at":"2026-09-18T11:00:00Z"}}`},
		{http.MethodDelete, "/v1/sandboxes/sb-1/access-token", `{"status":"success","data":{"enabled":false}}`},
		{http.MethodGet, "/v1/sandboxes/sb-1/access-token", `{"status":"success","data":{"enabled":false}}`},
	}
	next := 0
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		if next >= len(requests) {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		want := requests[next]
		next++
		if request.Method != want.method || request.URL.Path != want.path {
			t.Errorf("request = %s %s, want %s %s", request.Method, request.URL.Path, want.method, want.path)
		}
		if got := request.Header.Get("X-Api-Key"); got != "test-key" {
			t.Errorf("X-Api-Key = %q, want owner key", got)
		}
		return jsonResponse(request, want.body), nil
	})
	instance := newInstance(client.transport, structs.SandboxResponse{ID: "sb-1"})

	created, err := instance.CreateAccessToken(t.Context())
	if err != nil || created.Token != "skp_sb_first" || !created.Enabled || created.CreatedAt.IsZero() || created.RotatedAt != nil {
		t.Fatalf("CreateAccessToken() = %+v, %v", created, err)
	}
	metadata, err := instance.GetAccessToken(t.Context())
	if err != nil || !metadata.Enabled || metadata.TokenHint != "skp_sb_fi...irst" || metadata.CreatedAt == nil {
		t.Fatalf("GetAccessToken() = %+v, %v", metadata, err)
	}
	rotated, err := instance.RotateAccessToken(t.Context())
	if err != nil || rotated.Token != "skp_sb_second" || rotated.RotatedAt == nil || rotated.RotatedAt.Hour() != 11 {
		t.Fatalf("RotateAccessToken() = %+v, %v", rotated, err)
	}
	disabled, err := instance.DisableAccessToken(t.Context())
	if err != nil || disabled.Enabled || disabled.TokenHint != "" {
		t.Fatalf("DisableAccessToken() = %+v, %v", disabled, err)
	}
	metadata, err = instance.GetAccessToken(t.Context())
	if err != nil || metadata.Enabled || metadata.CreatedAt != nil {
		t.Fatalf("GetAccessToken() after disable = %+v, %v", metadata, err)
	}
	if next != len(requests) {
		t.Errorf("sent %d requests, want %d", next, len(requests))
	}
	if created.CreatedAt != time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC) {
		t.Errorf("created_at = %s", created.CreatedAt)
	}
}

func TestWithAccessTokenUsesDelegatedCredentialOnlyOnNewHandle(t *testing.T) {
	requests := 0
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		requests++
		switch requests {
		case 1:
			if request.Method != http.MethodPost || request.URL.Path != "/v1/sandboxes/sb-1/exec" || request.Header.Get("X-Api-Key") != "skp_sb_worker" {
				t.Errorf("delegated request = %s %s, key %q", request.Method, request.URL.Path, request.Header.Get("X-Api-Key"))
			}
			return jsonResponse(request, `{"status":"success","data":{"result":{"stdout":"hello\n","stderr":"","exit_code":0},"exec_ms":1}}`), nil
		case 2:
			if request.Method != http.MethodGet || request.URL.Path != "/v1/sandboxes/sb-1/access-token" || request.Header.Get("X-Api-Key") != "test-key" {
				t.Errorf("owner request = %s %s, key %q", request.Method, request.URL.Path, request.Header.Get("X-Api-Key"))
			}
			return jsonResponse(request, `{"status":"success","data":{"enabled":true}}`), nil
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
			return nil, nil
		}
	})
	owner := newInstance(client.transport, structs.SandboxResponse{ID: "sb-1"})
	worker, err := owner.WithAccessToken("skp_sb_worker")
	if err != nil {
		t.Fatalf("WithAccessToken() error = %v", err)
	}
	if worker == owner || worker.Files() == nil || worker.Processes() == nil || worker.Computer() == nil {
		t.Fatal("WithAccessToken() did not return a separate initialized handle")
	}
	result, err := worker.RunCommand(t.Context(), structs.RunCommandRequest{Command: "echo", Arguments: []string{"hello"}}, structs.ExecOptions{})
	if err != nil || result.Result.StandardOutput != "hello\n" {
		t.Fatalf("delegated RunCommand() = %+v, %v", result, err)
	}
	if _, err := owner.GetAccessToken(t.Context()); err != nil {
		t.Fatalf("owner GetAccessToken() error = %v", err)
	}
	if requests != 2 {
		t.Errorf("requests = %d, want 2", requests)
	}
}

func TestWithAccessTokenRejectsEmptyToken(t *testing.T) {
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		return nil, nil
	})
	owner := newInstance(client.transport, structs.SandboxResponse{ID: "sb-1"})
	if _, err := owner.WithAccessToken(" \t "); err == nil {
		t.Fatal("WithAccessToken() accepted an empty token")
	}
}
