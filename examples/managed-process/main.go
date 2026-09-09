// Command managed-process demonstrates persistent pipe and PTY process
// lifecycles, including input, output replay, resize, wait, and termination.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/NodeOps-app/createos-go-sdk/sandbox"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx)
	stop()
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) (runErr error) {
	client, err := sandbox.NewClient()
	if err != nil {
		return err
	}
	instance, err := client.CreateSandbox(ctx, structs.CreateSandboxRequest{
		Shape:  "s-1vcpu-1gb",
		RootFS: "devbox:1",
		EnvironmentVariables: map[string]string{
			"PROCESS_DEMO_BASE":     "from-sandbox-env",
			"PROCESS_DEMO_OVERRIDE": "declared-at-create",
		},
	})
	if err != nil {
		return err
	}
	fmt.Println("created:", instance.ID())
	defer func(cleanupParent context.Context) {
		cleanupCtx, cancel := context.WithTimeout(cleanupParent, 30*time.Second)
		defer cancel()
		if err := instance.Destroy(cleanupCtx); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("destroy sandbox: %w", err))
			return
		}
		fmt.Println("destroyed")
	}(context.WithoutCancel(ctx))

	processes := instance.Processes()
	fmt.Println("\n[1/4] pipe process with stdin/stdout/stderr...")
	pipeProcess, err := processes.Create(ctx, structs.ManagedProcessCreateRequest{
		Command: "/bin/sh",
		Arguments: []string{"-c", strings.Join([]string{
			`printf 'base:%s\n' "$PROCESS_DEMO_BASE"`,
			`printf 'override:%s\n' "$PROCESS_DEMO_OVERRIDE"`,
			`printf 'stderr:ready\n' >&2`,
			`IFS= read -r line`,
			`printf 'stdin:%s\n' "$line"`,
		}, "; ")},
		WorkingDirectory: "/root",
		EnvironmentVariables: map[string]string{
			"PROCESS_DEMO_OVERRIDE": "from-process-env",
		},
	})
	if err != nil {
		return err
	}
	fmt.Println("      pipe:", pipeProcess.ProcessID)

	listed, err := processes.List(ctx)
	if err != nil {
		return err
	}
	fmt.Println("      listed processes:", len(listed))
	if _, err := processes.Input(ctx, pipeProcess.ProcessID, "hello managed process\n"); err != nil {
		return err
	}
	if err := processes.CloseStandardInput(ctx, pipeProcess.ProcessID); err != nil {
		return err
	}
	pipeDone, err := processes.Wait(ctx, pipeProcess.ProcessID, structs.ManagedProcessWaitOptions{
		Scope:       structs.ManagedProcessWaitScopeTree,
		WaitTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	fmt.Printf("      pipe exit: %v\n", pointerValue(pipeDone.ExitCode))

	pipeStream, err := processes.Connect(ctx, pipeProcess.ProcessID, structs.ManagedProcessConnectOptions{})
	if err != nil {
		return err
	}
	pipeOutput, err := collectOutput(pipeStream)
	if err != nil {
		return err
	}
	fmt.Println("      stdout:")
	fmt.Print(indent(strings.TrimSpace(pipeOutput.standardOutput)))
	fmt.Println("      stderr:")
	fmt.Print(indent(strings.TrimSpace(pipeOutput.standardError)))

	fmt.Println("\n[2/4] reconnect from output offset...")
	afterSequence := max(pipeOutput.lastSequence-1, 0)
	replayStream, err := processes.Connect(ctx, pipeProcess.ProcessID, structs.ManagedProcessConnectOptions{AfterSequence: afterSequence})
	if err != nil {
		return err
	}
	replay, err := collectOutput(replayStream)
	if err != nil {
		return err
	}
	fmt.Printf("      replayed data frames after sequence %d: %d\n", afterSequence, replay.dataFrames)

	fmt.Println("\n[3/4] interactive PTY shell...")
	ptyProcess, err := processes.Create(ctx, structs.ManagedProcessCreateRequest{
		WorkingDirectory: "/root",
		PTY:              &structs.PTYSize{Rows: 24, Cols: 80},
	})
	if err != nil {
		return err
	}
	fmt.Println("      PTY:", ptyProcess.ProcessID)
	if _, err := processes.Input(ctx, ptyProcess.ProcessID, "echo terminal-ready; stty size\n"); err != nil {
		return err
	}
	if err := processes.Resize(ctx, ptyProcess.ProcessID, structs.PTYSize{Rows: 32, Cols: 100}); err != nil {
		return err
	}
	if _, err := processes.Input(ctx, ptyProcess.ProcessID, "echo after-resize; stty size; exit\n"); err != nil {
		return err
	}
	ptyDone, err := processes.Wait(ctx, ptyProcess.ProcessID, structs.ManagedProcessWaitOptions{
		Scope:       structs.ManagedProcessWaitScopeTree,
		WaitTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	fmt.Printf("      PTY exit: %v\n", pointerValue(ptyDone.ExitCode))
	ptyStream, err := processes.Connect(ctx, ptyProcess.ProcessID, structs.ManagedProcessConnectOptions{})
	if err != nil {
		return err
	}
	ptyOutput, err := collectOutput(ptyStream)
	if err != nil {
		return err
	}
	fmt.Print(indent(strings.TrimSpace(ptyOutput.pty)))

	fmt.Println("\n[4/4] terminate a long-running process tree...")
	longRunning, err := processes.Create(ctx, structs.ManagedProcessCreateRequest{
		Command:   "/bin/sh",
		Arguments: []string{"-c", "trap '' TERM; sleep 300 & wait"},
	})
	if err != nil {
		return err
	}
	terminated, err := processes.Delete(ctx, longRunning.ProcessID, structs.ManagedProcessDeleteOptions{GracePeriod: 100 * time.Millisecond})
	if err != nil {
		return err
	}
	fmt.Printf("      terminated: leader_exited=%t tree_exited=%t\n", terminated.LeaderExited, terminated.TreeExited)

	if !strings.Contains(pipeOutput.standardOutput, "stdin:hello managed process") {
		return errors.New("pipe stdout did not include stdin echo")
	}
	if !strings.Contains(ptyOutput.pty, "terminal-ready") || !strings.Contains(ptyOutput.pty, "after-resize") {
		return errors.New("PTY output did not include command markers")
	}
	if !terminated.TreeExited {
		return errors.New("terminated process tree did not exit")
	}
	fmt.Println("\nverified end-to-end: managed pipe process and PTY lifecycle")
	return nil
}

type collectedOutput struct {
	standardOutput string
	standardError  string
	pty            string
	lastSequence   int64
	dataFrames     int
}

func collectOutput(stream *sandbox.ProcessStream) (output collectedOutput, resultErr error) {
	defer func() {
		if err := stream.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close process stream: %w", err))
		}
	}()
	for {
		event, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			return output, nil
		}
		if err != nil {
			return output, err
		}
		switch event.Type {
		case structs.ManagedProcessConnectEventData:
			value := string(event.Data)
			switch event.Stream {
			case structs.ManagedProcessStreamStdout:
				output.standardOutput += value
			case structs.ManagedProcessStreamStderr:
				output.standardError += value
			case structs.ManagedProcessStreamPTY:
				output.pty += value
			}
			output.lastSequence = max(output.lastSequence, event.Sequence)
			output.dataFrames++
		case structs.ManagedProcessConnectEventExit:
			return output, nil
		case structs.ManagedProcessConnectEventHeartbeat:
			continue
		case structs.ManagedProcessConnectEventError:
			if event.ErrorMessage == "" {
				return output, errors.New("process stream reported an error")
			}
			return output, fmt.Errorf("process stream: %s", event.ErrorMessage)
		}
	}
}

func pointerValue(value *int) any {
	if value == nil {
		return "unknown"
	}
	return *value
}

func indent(value string) string {
	if value == "" {
		return "        (empty)\n"
	}
	return "        " + strings.ReplaceAll(value, "\n", "\n        ") + "\n"
}
