package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NodeOps-app/createos-go-sdk/structs"
)

type fakeCreator struct {
	instance sandboxInstance
	err      error
}

func (f fakeCreator) Create(context.Context) (sandboxInstance, error) {
	return f.instance, f.err
}

type fakeInstance struct {
	result    structs.RunCommandResponse
	err       error
	destroyed bool
}

func (f *fakeInstance) RunCommand(context.Context, structs.RunCommandRequest, structs.ExecOptions) (structs.RunCommandResponse, error) {
	return f.result, f.err
}

func (f *fakeInstance) Destroy(context.Context) error {
	f.destroyed = true
	return nil
}

func TestExecutionHandlerRunsCommandAndDestroysSandbox(t *testing.T) {
	t.Parallel()
	instance := &fakeInstance{result: structs.RunCommandResponse{
		Result:                structs.CommandResult{StandardOutput: "hello\n", ExitCode: 0},
		ExecutionMilliseconds: 12.5,
	}}
	handler, err := newExecutionHandler(fakeCreator{instance: instance}, 1)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/execute", strings.NewReader(`{"command":"printf","arguments":["hello\\n"]}`))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if !instance.destroyed {
		t.Fatal("sandbox was not destroyed")
	}
	var response executeResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.StandardOutput != "hello\n" || response.ExitCode != 0 || response.ExecutionMilliseconds != 12.5 {
		t.Fatalf("response = %#v", response)
	}
}

func TestExecutionHandlerRejectsInvalidRequests(t *testing.T) {
	t.Parallel()
	handler, err := newExecutionHandler(fakeCreator{err: errors.New("must not be called")}, 1)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		method string
		body   string
		status int
	}{
		{name: "method", method: http.MethodGet, body: `{}`, status: http.StatusMethodNotAllowed},
		{name: "missing command", method: http.MethodPost, body: `{}`, status: http.StatusBadRequest},
		{name: "unknown field", method: http.MethodPost, body: `{"command":"true","unknown":1}`, status: http.StatusBadRequest},
		{name: "multiple values", method: http.MethodPost, body: `{"command":"true"}{}`, status: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(context.Background(), test.method, "/v1/execute", strings.NewReader(test.body))
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, test.status, recorder.Body.String())
			}
		})
	}
}
