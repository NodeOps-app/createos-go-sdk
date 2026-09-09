// Command command-streaming uploads a Python program and prints its output as
// the sandbox produces it.
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
	"time"

	"github.com/NodeOps-app/createos-go-sdk/sandbox"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

const script = `import time

for number in range(1, 6):
    print(f"result {number}", flush=True)
    time.sleep(1)
`

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	err := run(ctx)
	stop()
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	client, err := sandbox.NewClient()
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	instance, err := client.CreateSandbox(ctx, structs.CreateSandboxRequest{
		Shape:  "s-1vcpu-1gb",
		RootFS: "devbox:1",
	})
	if err != nil {
		return fmt.Errorf("create sandbox: %w", err)
	}
	fmt.Printf("created: %s\n", instance.ID())
	defer destroy(ctx, instance)

	if err := instance.Files().Upload(ctx, "/tmp/script.py", strings.NewReader(script)); err != nil {
		return fmt.Errorf("upload script: %w", err)
	}

	stream, err := instance.StreamCommand(ctx, structs.RunCommandRequest{
		Command:   "python3",
		Arguments: []string{"/tmp/script.py"},
	}, structs.ExecOptions{})
	if err != nil {
		return fmt.Errorf("start command stream: %w", err)
	}
	defer stream.Close()

	fmt.Println("--- streaming output ---")
	for {
		event, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("receive command output: %w", err)
		}

		switch event.Type {
		case structs.ExecStreamEventStdout:
			_, _ = io.WriteString(os.Stdout, event.Data)
		case structs.ExecStreamEventStderr:
			_, _ = io.WriteString(os.Stderr, event.Data)
		case structs.ExecStreamEventError:
			fmt.Fprintf(os.Stderr, "agent error: %s\n", event.ErrorMessage)
		case structs.ExecStreamEventExit:
			if event.ExitCode != nil {
				fmt.Printf("(exited %d)\n", *event.ExitCode)
			}
		case structs.ExecStreamEventHeartbeat:
			// Heartbeats keep idle streams alive and carry no command output.
		}
	}
}

func destroy(parent context.Context, instance *sandbox.Instance) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), 2*time.Minute)
	defer cancel()
	if err := instance.Destroy(ctx); err != nil {
		log.Printf("cleanup: destroy sandbox %s: %v", instance.ID(), err)
		return
	}
	fmt.Println("destroyed")
}
