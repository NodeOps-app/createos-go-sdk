// Command execution-server exposes a small HTTP API that executes commands in
// a fresh CreateOS sandbox for every request.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/NodeOps-app/createos-go-sdk/sandbox"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

const (
	defaultAddress       = "127.0.0.1:8080"
	maximumRequestBytes  = 1 << 20
	maximumConcurrency   = 4
	executionTimeout     = 2 * time.Minute
	cleanupTimeout       = 30 * time.Second
	serverShutdownPeriod = 10 * time.Second
)

type executeRequest struct {
	Command              string            `json:"command"`
	Arguments            []string          `json:"arguments,omitempty"`
	StandardInput        string            `json:"standardInput,omitempty"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`
}

type executeResponse struct {
	StandardOutput        string  `json:"stdout"`
	StandardError         string  `json:"stderr"`
	ExitCode              int     `json:"exitCode"`
	ErrorMessage          string  `json:"error,omitempty"`
	ExecutionMilliseconds float64 `json:"executionMilliseconds"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type sandboxInstance interface {
	RunCommand(context.Context, structs.RunCommandRequest, structs.ExecOptions) (structs.RunCommandResponse, error)
	Destroy(context.Context) error
}

type sandboxCreator interface {
	Create(context.Context) (sandboxInstance, error)
}

type createOSCreator struct {
	client *sandbox.Client
}

func (c createOSCreator) Create(ctx context.Context) (sandboxInstance, error) {
	return c.client.CreateSandbox(ctx, structs.CreateSandboxRequest{
		Shape:  "s-1vcpu-1gb",
		RootFS: "devbox:1",
	})
}

type executionHandler struct {
	creator sandboxCreator
	slots   chan struct{}
}

func newExecutionHandler(creator sandboxCreator, concurrency int) (*executionHandler, error) {
	if creator == nil {
		return nil, errors.New("sandbox creator must not be nil")
	}
	if concurrency <= 0 {
		return nil, errors.New("concurrency must be positive")
	}
	return &executionHandler{creator: creator, slots: make(chan struct{}, concurrency)}, nil
}

func (h *executionHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/v1/execute" {
		writeJSON(writer, http.StatusNotFound, errorResponse{Error: "not found"})
		return
	}
	if request.Method != http.MethodPost {
		writer.Header().Set("Allow", http.MethodPost)
		writeJSON(writer, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	select {
	case h.slots <- struct{}{}:
		defer func() { <-h.slots }()
	default:
		writeJSON(writer, http.StatusTooManyRequests, errorResponse{Error: "execution capacity reached"})
		return
	}

	request.Body = http.MaxBytesReader(writer, request.Body, maximumRequestBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	var input executeRequest
	if err := decoder.Decode(&input); err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: "invalid JSON request: " + err.Error()})
		return
	}
	if err := ensureJSONEnd(decoder); err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	input.Command = strings.TrimSpace(input.Command)
	if input.Command == "" {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: "command is required"})
		return
	}

	ctx, cancel := context.WithTimeout(request.Context(), executionTimeout)
	defer cancel()
	instance, err := h.creator.Create(ctx)
	if err != nil {
		writeExecutionError(ctx, writer, "create sandbox", err)
		return
	}
	defer destroySandbox(context.WithoutCancel(ctx), instance)

	result, err := instance.RunCommand(ctx, structs.RunCommandRequest{
		Command:              input.Command,
		Arguments:            input.Arguments,
		StandardInput:        input.StandardInput,
		EnvironmentVariables: input.EnvironmentVariables,
	}, structs.ExecOptions{})
	if err != nil {
		writeExecutionError(ctx, writer, "execute command", err)
		return
	}

	writeJSON(writer, http.StatusOK, executeResponse{
		StandardOutput:        result.Result.StandardOutput,
		StandardError:         result.Result.StandardError,
		ExitCode:              result.Result.ExitCode,
		ErrorMessage:          result.Result.ErrorMessage,
		ExecutionMilliseconds: result.ExecutionMilliseconds,
	})
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("invalid JSON request: %w", err)
	}
	return errors.New("request body must contain exactly one JSON object")
}

func destroySandbox(parent context.Context, instance sandboxInstance) {
	ctx, cancel := context.WithTimeout(parent, cleanupTimeout)
	defer cancel()
	if err := instance.Destroy(ctx); err != nil {
		log.Printf("destroy sandbox: %v", err)
	}
}

func writeExecutionError(ctx context.Context, writer http.ResponseWriter, operation string, err error) {
	status := http.StatusBadGateway
	if ctx.Err() != nil || errors.Is(err, context.DeadlineExceeded) {
		status = http.StatusGatewayTimeout
	}
	writeJSON(writer, status, errorResponse{Error: operation + " failed: " + err.Error()})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}

func main() {
	apiKey := strings.TrimSpace(os.Getenv("CREATEOS_SANDBOX_API_KEY"))
	if apiKey == "" {
		log.Fatal("CREATEOS_SANDBOX_API_KEY is required")
	}
	client, err := sandbox.NewClient(sandbox.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("create SDK client: %v", err)
	}
	handler, err := newExecutionHandler(createOSCreator{client: client}, maximumConcurrency)
	if err != nil {
		log.Fatal(err)
	}

	address := strings.TrimSpace(os.Getenv("EXECUTION_SERVER_ADDRESS"))
	if address == "" {
		address = defaultAddress
	}
	server := &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      executionTimeout + cleanupTimeout,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serverErrors := make(chan error, 1)
	go func() {
		log.Print("execution server listening")
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("serve HTTP: %v", err)
			return
		}
	case <-shutdownSignal.Done():
		ctx, cancel := context.WithTimeout(context.Background(), serverShutdownPeriod)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("shut down HTTP server: %v", err)
		}
	}
}
