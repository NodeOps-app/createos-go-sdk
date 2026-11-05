package sandbox

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/NodeOps-app/createos-go-sdk/structs"
)

func TestCheckpointLifecycle(t *testing.T) {
	requests := 0
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		requests++
		switch requests {
		case 1:
			var body structs.CreateCheckpointRequest
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if request.Method != http.MethodPost || request.URL.Path != "/v1/sandboxes/sb-1/checkpoints" || body.Path != "/workspace/app" || body.Name != "before-update" {
				t.Fatalf("create request = %s %s %#v", request.Method, request.URL.Path, body)
			}
			return jsonResponse(request, `{"status":"success","data":{"id":"cp_first","path":"/workspace/app","name":"before-update","created_at":"2026-11-05T12:00:00Z"}}`), nil
		case 2:
			if request.Method != http.MethodGet || request.URL.Path != "/v1/sandboxes/sb-1/checkpoints" {
				t.Fatalf("list request = %s %s", request.Method, request.URL.Path)
			}
			return jsonResponse(request, `{"status":"success","data":[{"id":"cp_first","path":"/workspace/app","created_at":"2026-11-05T12:00:00Z"}]}`), nil
		case 3:
			if request.Method != http.MethodPost || request.URL.EscapedPath() != "/v1/sandboxes/sb-1/checkpoints/cp_first%2Fsegment/restore" || request.Body != nil {
				t.Fatalf("restore request = %s %s body=%v", request.Method, request.URL.EscapedPath(), request.Body)
			}
			return jsonResponse(request, `{"status":"success","data":{"checkpoint_id":"cp_first/segment","path":"/workspace/app"}}`), nil
		default:
			if request.Method != http.MethodDelete || request.URL.EscapedPath() != "/v1/sandboxes/sb-1/checkpoints/cp_first%2Fsegment" || request.Body != nil {
				t.Fatalf("delete request = %s %s body=%v", request.Method, request.URL.EscapedPath(), request.Body)
			}
			return jsonResponse(request, `{"status":"success","data":{}}`), nil
		}
	})
	instance := newInstance(client.transport, structs.SandboxResponse{ID: "sb-1", Status: structs.SandboxStatusRunning})

	checkpoint, err := instance.Checkpoints().Create(t.Context(), structs.CreateCheckpointRequest{Path: "/workspace/app", Name: "before-update"})
	if err != nil || checkpoint.ID != "cp_first" || checkpoint.CreatedAt.IsZero() {
		t.Fatalf("Create() = %#v, %v", checkpoint, err)
	}
	checkpoints, err := instance.Checkpoints().List(t.Context())
	if err != nil || len(checkpoints) != 1 {
		t.Fatalf("List() = %#v, %v", checkpoints, err)
	}
	restored, err := instance.Checkpoints().Restore(t.Context(), "cp_first/segment")
	if err != nil || restored.Path != "/workspace/app" {
		t.Fatalf("Restore() = %#v, %v", restored, err)
	}
	if err := instance.Checkpoints().Delete(t.Context(), "cp_first/segment"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}
