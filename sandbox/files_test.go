package sandbox

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/NodeOps-app/createos-go-sdk/structs"
)

const fileTransferTestTimeout = 5 * time.Minute

func TestFileUploadUsesRequestTimeout(t *testing.T) {
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		assertFileRequest(t, request, http.MethodPut)
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read upload body: %v", err)
		}
		if string(body) != "contents" {
			t.Errorf("upload body = %q, want contents", body)
		}
		return jsonResponse(request, `{"status":"success","data":{"bytes":8,"path":"/workspace/file.txt"}}`), nil
	})
	instance := newInstance(client.transport, structs.SandboxResponse{ID: "sb-1"})

	err := instance.Files().Upload(t.Context(), "/workspace/file.txt", bytes.NewBufferString("contents"), structs.RequestOptions{
		Timeout: fileTransferTestTimeout,
	})
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
}

func TestFileDownloadUsesRequestTimeout(t *testing.T) {
	client := testClient(t, func(request *http.Request) (*http.Response, error) {
		assertFileRequest(t, request, http.MethodGet)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString("contents")),
			Request:    request,
		}, nil
	})
	instance := newInstance(client.transport, structs.SandboxResponse{ID: "sb-1"})

	file, err := instance.Files().Download(t.Context(), "/workspace/file.txt", structs.RequestOptions{
		Timeout: fileTransferTestTimeout,
	})
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	defer file.Close()
	contents, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("read download: %v", err)
	}
	if string(contents) != "contents" {
		t.Errorf("download body = %q, want contents", contents)
	}
}

func assertFileRequest(t *testing.T, request *http.Request, method string) {
	t.Helper()
	if request.Method != method || request.URL.Path != "/v1/sandboxes/sb-1/files" {
		t.Errorf("request = %s %s, want %s /v1/sandboxes/sb-1/files", request.Method, request.URL.Path, method)
	}
	if path := request.URL.Query().Get("path"); path != "/workspace/file.txt" {
		t.Errorf("path query = %q, want /workspace/file.txt", path)
	}
	deadline, ok := request.Context().Deadline()
	if !ok {
		t.Fatal("request context has no deadline")
	}
	remaining := time.Until(deadline)
	if remaining < fileTransferTestTimeout-time.Second || remaining > fileTransferTestTimeout {
		t.Errorf("request timeout remaining = %v, want approximately %v", remaining, fileTransferTestTimeout)
	}
}
