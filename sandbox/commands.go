package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/NodeOps-app/createos-go-sdk/internal/protocol"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

// RunCommand executes a command and buffers its output.
func (i *Instance) RunCommand(ctx context.Context, request structs.RunCommandRequest, options structs.ExecOptions) (structs.RunCommandResponse, error) {
	if options.StandardInput != "" {
		request.StandardInput = options.StandardInput
	}
	if options.EnvironmentVariables != nil {
		request.EnvironmentVariables = options.EnvironmentVariables
	}
	request.Stream = false
	httpOptions := transportOptions(options.RequestOptions)
	httpOptions.Body = request
	var response structs.RunCommandResponse
	if err := i.transport.Do(ctx, http.MethodPost, i.path("/exec"), httpOptions, &response); err != nil {
		return structs.RunCommandResponse{}, fmt.Errorf("run command in sandbox %q: %w", i.ID(), err)
	}
	return response, nil
}

// Shell runs a script through bash -lc and treats a nonzero exit as an error.
func (i *Instance) Shell(ctx context.Context, script string, options structs.ExecOptions) (structs.RunCommandResponse, error) {
	response, err := i.RunCommand(ctx, structs.RunCommandRequest{
		Command: "bash", Arguments: []string{"-lc", script},
	}, options)
	if err != nil {
		return structs.RunCommandResponse{}, err
	}
	if response.Result.ExitCode != 0 || response.Result.ErrorMessage != "" {
		parts := []string{fmt.Sprintf("command exited with status %d", response.Result.ExitCode)}
		if response.Result.ErrorMessage != "" {
			parts = append(parts, "error: "+response.Result.ErrorMessage)
		}
		if response.Result.StandardError != "" {
			parts = append(parts, "stderr: "+tail(response.Result.StandardError, 2000))
		}
		return response, errors.New(strings.Join(parts, "\n"))
	}
	return response, nil
}

// CommandStream receives events from a streaming command. Close it when
// stopping before io.EOF.
type CommandStream struct {
	decoder *protocol.NDJSONDecoder[structs.CommandStreamFrame]
	queued  []structs.CommandStreamEvent
}

// StreamCommand starts a command and returns its output stream.
func (i *Instance) StreamCommand(ctx context.Context, request structs.RunCommandRequest, options structs.ExecOptions) (*CommandStream, error) {
	request.Stream = true
	if options.StandardInput != "" {
		request.StandardInput = options.StandardInput
	}
	if options.EnvironmentVariables != nil {
		request.EnvironmentVariables = options.EnvironmentVariables
	}
	httpOptions := transportOptions(options.RequestOptions)
	httpOptions.Body = request
	// CommandStream owns and closes the response body.
	response, err := i.transport.Stream(ctx, http.MethodPost, i.path("/exec?stream=true"), httpOptions) //nolint:bodyclose
	if err != nil {
		return nil, fmt.Errorf("stream command in sandbox %q: %w", i.ID(), err)
	}
	return &CommandStream{decoder: protocol.NewNDJSONDecoder[structs.CommandStreamFrame](response.Body)}, nil
}

// Receive returns the next command event.
func (s *CommandStream) Receive() (structs.CommandStreamEvent, error) {
	for len(s.queued) == 0 {
		frame, err := s.decoder.Recv()
		if err != nil {
			return structs.CommandStreamEvent{}, err
		}
		s.queued = commandEvents(frame)
	}
	event := s.queued[0]
	s.queued = s.queued[1:]
	return event, nil
}

// Close closes the underlying response body.
func (s *CommandStream) Close() error { return s.decoder.Close() }

func commandEvents(frame structs.CommandStreamFrame) []structs.CommandStreamEvent {
	events := make([]structs.CommandStreamEvent, 0, 4)
	if frame.Heartbeat {
		events = append(events, structs.CommandStreamEvent{Type: structs.ExecStreamEventHeartbeat})
	}
	if frame.StandardOutput != "" {
		events = append(events, structs.CommandStreamEvent{Type: structs.ExecStreamEventStdout, Data: frame.StandardOutput})
	}
	if frame.StandardError != "" {
		events = append(events, structs.CommandStreamEvent{Type: structs.ExecStreamEventStderr, Data: frame.StandardError})
	}
	if frame.ErrorMessage != "" {
		events = append(events, structs.CommandStreamEvent{Type: structs.ExecStreamEventError, ErrorMessage: frame.ErrorMessage})
	}
	if frame.ExitCode != nil {
		events = append(events, structs.CommandStreamEvent{Type: structs.ExecStreamEventExit, ExitCode: frame.ExitCode})
	}
	return events
}

func tail(value string, maximum int) string {
	if len(value) <= maximum {
		return value
	}
	return value[len(value)-maximum:]
}

var _ io.Closer = (*CommandStream)(nil)
