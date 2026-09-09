package sandbox

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/NodeOps-app/createos-go-sdk/internal/protocol"
	"github.com/NodeOps-app/createos-go-sdk/internal/transport"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// ProcessesService manages persistent processes and PTYs in one sandbox.
type ProcessesService struct{ instance *Instance }

func (s *ProcessesService) path(processID, suffix string) string {
	path := s.instance.path("/processes")
	if processID != "" {
		path += "/" + transport.EncodePath(processID)
	}
	return path + suffix
}

// Create starts a managed process or PTY.
func (s *ProcessesService) Create(ctx context.Context, request structs.ManagedProcessCreateRequest) (structs.ManagedProcess, error) {
	var process structs.ManagedProcess
	err := s.instance.transport.Do(ctx, http.MethodPost, s.path("", ""), transport.RequestOptions{Body: request}, &process)
	if err != nil {
		return structs.ManagedProcess{}, fmt.Errorf("create process in sandbox %q: %w", s.instance.ID(), err)
	}
	return process, nil
}

// List returns retained managed processes and PTYs.
func (s *ProcessesService) List(ctx context.Context) ([]structs.ManagedProcess, error) {
	var response structs.ManagedProcessListResponse
	if err := s.instance.transport.Do(ctx, http.MethodGet, s.path("", ""), transport.RequestOptions{}, &response); err != nil {
		return nil, fmt.Errorf("list processes in sandbox %q: %w", s.instance.ID(), err)
	}
	return response.Processes, nil
}

// Get returns one managed process.
func (s *ProcessesService) Get(ctx context.Context, processID string) (structs.ManagedProcess, error) {
	var process structs.ManagedProcess
	if err := s.instance.transport.Do(ctx, http.MethodGet, s.path(processID, ""), transport.RequestOptions{}, &process); err != nil {
		return structs.ManagedProcess{}, fmt.Errorf("get process %q: %w", processID, err)
	}
	return process, nil
}

// ProcessStream replays and follows managed process output.
type ProcessStream struct {
	decoder *protocol.NDJSONDecoder[structs.ManagedProcessConnectFrame]
}

// Connect opens a replayable process output stream.
func (s *ProcessesService) Connect(ctx context.Context, processID string, options structs.ManagedProcessConnectOptions) (*ProcessStream, error) {
	httpOptions := transportOptions(options.RequestOptions)
	httpOptions.Query = url.Values{"after": {strconv.FormatInt(options.AfterSequence, 10)}}
	// ProcessStream owns and closes the response body.
	response, err := s.instance.transport.Stream(ctx, http.MethodGet, s.path(processID, "/connect"), httpOptions) //nolint:bodyclose
	if err != nil {
		return nil, fmt.Errorf("connect to process %q: %w", processID, err)
	}
	return &ProcessStream{decoder: protocol.NewNDJSONDecoder[structs.ManagedProcessConnectFrame](response.Body)}, nil
}

// Receive returns the next decoded process event.
func (s *ProcessStream) Receive() (structs.ManagedProcessConnectEvent, error) {
	frame, err := s.decoder.Recv()
	if err != nil {
		return structs.ManagedProcessConnectEvent{}, err
	}
	event := structs.ManagedProcessConnectEvent{
		Type: frame.Type, Sequence: frame.Sequence, Stream: frame.Stream,
		ExitCode: frame.ExitCode, Signal: frame.Signal, ErrorMessage: frame.ErrorMessage,
		OldestAvailableSequence: frame.OldestAvailableSequence,
	}
	if frame.DataBase64 != "" {
		event.Data, err = base64.StdEncoding.DecodeString(frame.DataBase64)
		if err != nil {
			return structs.ManagedProcessConnectEvent{}, fmt.Errorf("decode process output: %w", err)
		}
	}
	return event, nil
}

// Close closes the process stream.
func (s *ProcessStream) Close() error { return s.decoder.Close() }

// Input writes UTF-8 input to a process or PTY.
func (s *ProcessesService) Input(ctx context.Context, processID, data string) (int64, error) {
	return s.InputBytes(ctx, processID, []byte(data))
}

// InputBytes writes binary input to a process or PTY.
func (s *ProcessesService) InputBytes(ctx context.Context, processID string, data []byte) (int64, error) {
	request := structs.ManagedProcessInputRequest{DataBase64: base64.StdEncoding.EncodeToString(data)}
	var response structs.ManagedProcessInputResponse
	err := s.instance.transport.Do(ctx, http.MethodPost, s.path(processID, "/input"), transport.RequestOptions{Body: request}, &response)
	if err != nil {
		return 0, fmt.Errorf("write input to process %q: %w", processID, err)
	}
	return response.InputSequence, nil
}

// CloseStandardInput closes a pipe process's standard input.
func (s *ProcessesService) CloseStandardInput(ctx context.Context, processID string) error {
	return s.doOK(ctx, http.MethodPost, processID, "/stdin/close", nil)
}

// Resize changes PTY dimensions.
func (s *ProcessesService) Resize(ctx context.Context, processID string, size structs.PTYSize) error {
	return s.doOK(ctx, http.MethodPost, processID, "/resize", size)
}

// Signal sends a signal to a process.
func (s *ProcessesService) Signal(ctx context.Context, processID string, signal structs.ManagedProcessSignal) error {
	return s.doOK(ctx, http.MethodPost, processID, "/signal", structs.ManagedProcessSignalRequest{Signal: signal})
}

// Wait long-polls until a process leader or tree exits.
func (s *ProcessesService) Wait(ctx context.Context, processID string, options structs.ManagedProcessWaitOptions) (structs.ManagedProcess, error) {
	httpOptions := transportOptions(options.RequestOptions)
	httpOptions.Query = url.Values{"scope": {string(options.Scope)}}
	if options.WaitTimeout > 0 {
		httpOptions.Query.Set("timeout_ms", strconv.FormatInt(options.WaitTimeout.Milliseconds(), 10))
	}
	var process structs.ManagedProcess
	if err := s.instance.transport.Do(ctx, http.MethodGet, s.path(processID, "/wait"), httpOptions, &process); err != nil {
		return structs.ManagedProcess{}, fmt.Errorf("wait for process %q: %w", processID, err)
	}
	return process, nil
}

// Delete terminates a managed process tree.
func (s *ProcessesService) Delete(ctx context.Context, processID string, options structs.ManagedProcessDeleteOptions) (structs.ManagedProcess, error) {
	httpOptions := transportOptions(options.RequestOptions)
	if options.GracePeriod > 0 {
		httpOptions.Query = url.Values{"grace_ms": {strconv.FormatInt(options.GracePeriod.Milliseconds(), 10)}}
	}
	var process structs.ManagedProcess
	if err := s.instance.transport.Do(ctx, http.MethodDelete, s.path(processID, ""), httpOptions, &process); err != nil {
		return structs.ManagedProcess{}, fmt.Errorf("delete process %q: %w", processID, err)
	}
	return process, nil
}

func (s *ProcessesService) doOK(ctx context.Context, method, processID, suffix string, body any) error {
	var response structs.OKResponse
	if err := s.instance.transport.Do(ctx, method, s.path(processID, suffix), transport.RequestOptions{Body: body}, &response); err != nil {
		return fmt.Errorf("control process %q: %w", processID, err)
	}
	return nil
}

var _ io.Closer = (*ProcessStream)(nil)
