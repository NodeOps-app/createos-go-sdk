package sandbox

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/NodeOps-app/createos-go-sdk/structs"
)

func TestSnapshotLifecycleRequests(t *testing.T) {
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		switch request.Method + " " + request.URL.Path {
		case "POST /v1/sandboxes/sb-1/snapshots":
			var body structs.CreateSnapshotRequest
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode create snapshot request: %v", err)
			}
			if body.Type != structs.SnapshotTypeFull || body.Name != "checkpoint" {
				t.Errorf("create snapshot body = %#v", body)
			}
			return jsonResponse(request, snapshotResponse("creating")), nil
		case "GET /v1/snapshots/snap-1":
			return jsonResponse(request, snapshotResponse("ready")), nil
		case "GET /v1/snapshots":
			return jsonResponse(request, `{"status":"success","data":{"data":[`+snapshotData("ready")+`],"pagination":{"total":1,"limit":500,"offset":0,"count":1}}}`), nil
		case "DELETE /v1/snapshots/snap-1":
			return jsonResponse(request, `{"status":"success","data":{"id":"snap-1","status":"deleting"}}`), nil
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.String())
			return nil, nil
		}
	})

	instance := newInstance(client.transport, structs.SandboxResponse{ID: "sb-1"})
	created, err := instance.CreateSnapshot(t.Context(), structs.CreateSnapshotRequest{Type: structs.SnapshotTypeFull, Name: "checkpoint"})
	if err != nil || created.Status != structs.SnapshotStatusCreating {
		t.Fatalf("CreateSnapshot() = %#v, %v", created, err)
	}
	got, err := client.Snapshots().Get(t.Context(), "snap-1")
	if err != nil || got.Status != structs.SnapshotStatusReady {
		t.Fatalf("Snapshots().Get() = %#v, %v", got, err)
	}
	listed, err := client.Snapshots().List(t.Context(), structs.PaginationOptions{})
	if err != nil || len(listed) != 1 || listed[0].ID != "snap-1" {
		t.Fatalf("Snapshots().List() = %#v, %v", listed, err)
	}
	deleted, err := client.Snapshots().Delete(t.Context(), "snap-1")
	if err != nil || deleted.Status != structs.SnapshotStatusDeleting {
		t.Fatalf("Snapshots().Delete() = %#v, %v", deleted, err)
	}
}

func snapshotResponse(status string) string {
	return `{"status":"success","data":` + snapshotData(status) + `}`
}

func snapshotData(status string) string {
	return strings.ReplaceAll(`{"id":"snap-1","sandbox_id":"sb-1","name":"checkpoint","type":"full","status":"STATUS","created_at":"2026-11-03T00:00:00Z","shape":"s-1vcpu-1gb","vcpu":1,"mem_mib":1024,"disk_mib":10240,"size_bytes":2048}`, "STATUS", status)
}

func TestCreateSandboxRestoresSnapshot(t *testing.T) {
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		var body structs.CreateSandboxRequest
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode restore request: %v", err)
		}
		if body.SnapshotID != "snap-1" {
			t.Errorf("snapshot_id = %q", body.SnapshotID)
		}
		return jsonResponse(request, `{"status":"success","data":{"id":"sb-2","snapshot_id":"snap-1","status":"running","ip":"10.0.0.2","shape":"s-1vcpu-1gb","vcpu":1,"mem_mib":1024,"disk_mib":10240,"spawn_ms":5}}`), nil
	})

	instance, err := client.CreateSandbox(t.Context(), structs.CreateSandboxRequest{SnapshotID: "snap-1"})
	if err != nil || instance.Data().SnapshotID == nil || *instance.Data().SnapshotID != "snap-1" {
		t.Fatalf("CreateSandbox(snapshot) = %#v, %v", instance, err)
	}
}
