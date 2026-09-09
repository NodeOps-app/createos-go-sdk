package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/NodeOps-app/createos-go-sdk/structs"
)

func testClient(t *testing.T, roundTrip roundTripFunc) *Client {
	t.Helper()
	client, err := NewClient(
		WithAPIKey("test-key"),
		WithBaseURL("https://sandbox.test"),
		WithHTTPClient(&http.Client{Transport: roundTrip}),
		WithoutRetry(),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}

func TestCreateSandboxInitializesInstanceServices(t *testing.T) {
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost || request.URL.Path != "/v1/sandboxes" {
			t.Errorf("request = %s %s, want POST /v1/sandboxes", request.Method, request.URL.Path)
		}
		return jsonResponse(request, `{"status":"success","data":{"id":"sb-1","status":"running","name":"worker","ip":"10.0.0.2","shape":"small","rootfs":"devbox:1","vcpu":1,"mem_mib":256,"disk_mib":1024,"spawn_ms":5.25,"egress":[],"bandwidth_quota_bytes":-1}}`), nil
	})

	instance, err := client.CreateSandbox(t.Context(), structs.CreateSandboxRequest{Shape: "small"})
	if err != nil {
		t.Fatalf("CreateSandbox() error = %v", err)
	}
	if instance.ID() != "sb-1" || instance.Status() != structs.SandboxStatusRunning {
		t.Errorf("instance = %#v, want sb-1 running", instance.Data())
	}
	if instance.Name() != "worker" || instance.IPAddress() != "10.0.0.2" || instance.Data().SpawnMilliseconds != 5.25 {
		t.Errorf("instance = %#v, want nullable fields and fractional spawn timing preserved", instance.Data())
	}
	if instance.Files() == nil || instance.Processes() == nil || instance.Computer() == nil {
		t.Fatal("instance services were not initialized")
	}
}

func TestInstancePauseUpdatesCachedState(t *testing.T) {
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != "/v1/sandboxes/sb-1/pause" {
			t.Errorf("path = %q, want pause endpoint", request.URL.Path)
		}
		return jsonResponse(request, `{"status":"success","data":{"id":"sb-1","status":"paused","vcpu":1,"mem_mib":256,"disk_mib":1024,"created_at":"2026-09-08T00:00:00Z","ingress_enabled":false}}`), nil
	})
	instance := newInstance(client.transport, structs.SandboxResponse{ID: "sb-1", Status: structs.SandboxStatusRunning})

	if err := instance.Pause(t.Context()); err != nil {
		t.Fatalf("Pause() error = %v", err)
	}
	if got := instance.Status(); got != structs.SandboxStatusPaused {
		t.Errorf("Status() = %q, want paused", got)
	}
}

func TestRunCommandUsesClearRequestFields(t *testing.T) {
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		var body structs.RunCommandRequest
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body.Command != "printf" || len(body.Arguments) != 1 || body.EnvironmentVariables["MODE"] != "test" {
			t.Errorf("request body = %#v", body)
		}
		return jsonResponse(request, `{"status":"success","data":{"result":{"stdout":"ok","stderr":"","exit_code":0},"exec_ms":2.5}}`), nil
	})
	instance := newInstance(client.transport, structs.SandboxResponse{ID: "sb-1"})

	response, err := instance.RunCommand(t.Context(), structs.RunCommandRequest{
		Command:              "printf",
		Arguments:            []string{"ok"},
		EnvironmentVariables: map[string]string{"MODE": "test"},
	}, structs.ExecOptions{})
	if err != nil {
		t.Fatalf("RunCommand() error = %v", err)
	}
	if response.Result.StandardOutput != "ok" {
		t.Errorf("stdout = %q, want ok", response.Result.StandardOutput)
	}
	if response.ExecutionMilliseconds != 2.5 {
		t.Errorf("exec milliseconds = %v, want 2.5", response.ExecutionMilliseconds)
	}
}

func TestForkCanExplicitlyDisableIngress(t *testing.T) {
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if ingress, ok := body["ingress_enabled"]; !ok || ingress != false {
			t.Errorf("ingress_enabled = %#v, present=%v; want explicit false", ingress, ok)
		}
		if _, ok := body["bandwidth_quota_bytes"]; ok {
			t.Error("fork request sent forbidden bandwidth_quota_bytes")
		}
		return jsonResponse(request, `{"status":"success","data":{"id":"sb-fork","status":"forking","vcpu":1,"mem_mib":256,"disk_mib":1024,"created_at":"2026-09-08T00:00:00Z","ingress_enabled":false}}`), nil
	})
	instance := newInstance(client.transport, structs.SandboxResponse{ID: "sb-source"})
	disabled := false

	fork, err := instance.Fork(t.Context(), structs.ForkSandboxRequest{IngressEnabled: &disabled})
	if err != nil {
		t.Fatalf("Fork() error = %v", err)
	}
	if fork.ID() != "sb-fork" {
		t.Errorf("fork ID = %q, want sb-fork", fork.ID())
	}
}

func TestAttachDiskDecodesIDAcknowledgement(t *testing.T) {
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		return jsonResponse(request, `{"status":"success","data":{"id":"sb-1"}}`), nil
	})
	instance := newInstance(client.transport, structs.SandboxResponse{ID: "sb-1"})

	err := instance.AttachDisk(t.Context(), structs.AttachDiskOptions{DiskID: "disk-1", MountPath: "/mnt/data"})
	if err != nil {
		t.Fatalf("AttachDisk() error = %v", err)
	}
}

func TestCommandStreamProjectsFrames(t *testing.T) {
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		body := "{\"stdout\":\"hello\"}\n{\"exit_code\":0}\n"
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(bytes.NewBufferString(body)), Request: request}, nil
	})
	instance := newInstance(client.transport, structs.SandboxResponse{ID: "sb-1"})
	stream, err := instance.StreamCommand(t.Context(), structs.RunCommandRequest{Command: "echo"}, structs.ExecOptions{})
	if err != nil {
		t.Fatalf("StreamCommand() error = %v", err)
	}
	defer stream.Close()

	event, err := stream.Receive()
	if err != nil || event.Type != structs.ExecStreamEventStdout || event.Data != "hello" {
		t.Fatalf("first Receive() = %#v, %v", event, err)
	}
	event, err = stream.Receive()
	if err != nil || event.Type != structs.ExecStreamEventExit || event.ExitCode == nil || *event.ExitCode != 0 {
		t.Fatalf("second Receive() = %#v, %v", event, err)
	}
}

func TestListSandboxesWalksNestedPagination(t *testing.T) {
	requests := 0
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		requests++
		offset := request.URL.Query().Get("offset")
		body := `{"status":"success","data":{"data":[{"id":"sb-1","status":"running","vcpu":1,"mem_mib":256,"disk_mib":1024,"created_at":"2026-09-08T00:00:00Z","ingress_enabled":false}],"pagination":{"total":2,"limit":1,"offset":0,"count":1}}}`
		if offset == "1" {
			body = `{"status":"success","data":{"data":[{"id":"sb-2","status":"paused","vcpu":1,"mem_mib":256,"disk_mib":1024,"created_at":"2026-09-08T00:00:00Z","ingress_enabled":false}],"pagination":{"total":2,"limit":1,"offset":1,"count":1}}}`
		}
		return jsonResponse(request, body), nil
	})

	instances, err := client.ListSandboxes(context.Background(), structs.ListSandboxesOptions{})
	if err != nil {
		t.Fatalf("ListSandboxes() error = %v", err)
	}
	if len(instances) != 2 || requests != 2 {
		t.Errorf("got %d instances in %d requests, want 2 in 2", len(instances), requests)
	}
}
