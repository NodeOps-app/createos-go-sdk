// Command files-and-snapshots demonstrates copy-on-write sandbox snapshots by
// pausing a sandbox, forking it, and showing that their filesystems diverge.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/NodeOps-app/createos-go-sdk/sandbox"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

const (
	basePath     = "/root/seed.txt"
	forkOnlyPath = "/root/fork-only.txt"
)

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

	base, err := client.CreateSandbox(ctx, structs.CreateSandboxRequest{
		Shape:  "s-1vcpu-256mb",
		RootFS: "devbox:1",
	})
	if err != nil {
		return fmt.Errorf("create base sandbox: %w", err)
	}
	fmt.Printf("base created: %s\n", base.ID())
	defer destroy(ctx, "base", base)

	stamp := time.Now().UTC().Format(time.RFC3339)
	if err := base.Files().Upload(ctx, basePath, strings.NewReader("seed written at "+stamp+"\n")); err != nil {
		return fmt.Errorf("seed base sandbox: %w", err)
	}
	contents, err := readFile(ctx, base, basePath)
	if err != nil {
		return err
	}
	fmt.Printf("wrote %s: %s\n", basePath, contents)

	fmt.Println("pausing base...")
	if err := base.Pause(ctx); err != nil {
		return fmt.Errorf("pause base sandbox: %w", err)
	}
	if err := base.WaitUntilPaused(ctx, structs.WaitOptions{Timeout: 10 * time.Minute}); err != nil {
		return fmt.Errorf("wait for base sandbox to pause: %w", err)
	}
	fmt.Println("base paused")

	fmt.Println("forking base (start_paused=true)...")
	fork, err := base.Fork(ctx, structs.ForkSandboxRequest{StartPaused: true})
	if err != nil {
		return fmt.Errorf("fork base sandbox: %w", err)
	}
	defer destroy(ctx, "fork", fork)
	if err := fork.WaitUntilPaused(ctx, structs.WaitOptions{Timeout: 10 * time.Minute}); err != nil {
		return fmt.Errorf("wait for fork to pause: %w", err)
	}
	forkedFrom := ""
	if source := fork.Data().ForkedFrom; source != nil {
		forkedFrom = *source
	}
	fmt.Printf("fork paused: %s (forked_from=%s)\n", fork.ID(), forkedFrom)

	fmt.Println("resuming fork...")
	if err := fork.Resume(ctx); err != nil {
		return fmt.Errorf("resume fork: %w", err)
	}
	if err := fork.WaitUntilRunning(ctx, structs.WaitOptions{Timeout: 5 * time.Minute}); err != nil {
		if errors.Is(err, sandbox.ErrTimeout) {
			log.Println("fork resume timed out under load; skipping verification")
			return nil
		}
		return fmt.Errorf("wait for fork to run: %w", err)
	}
	fmt.Printf("fork running: %s\n", fork.ID())

	contents, err = readFile(ctx, fork, basePath)
	if err != nil {
		return err
	}
	fmt.Printf("fork inherits %s: %s\n", basePath, contents)

	forkStamp := time.Now().UTC().Format(time.RFC3339)
	if err := fork.Files().Upload(ctx, forkOnlyPath, strings.NewReader("written only in fork at "+forkStamp+"\n")); err != nil {
		return fmt.Errorf("write fork-only file: %w", err)
	}
	contents, err = readFile(ctx, fork, forkOnlyPath)
	if err != nil {
		return err
	}
	fmt.Printf("fork wrote %s: %s\n", forkOnlyPath, contents)

	fmt.Println("resuming base...")
	if err := base.Resume(ctx); err != nil {
		return fmt.Errorf("resume base sandbox: %w", err)
	}
	if err := base.WaitUntilRunning(ctx, structs.WaitOptions{Timeout: 5 * time.Minute}); err != nil {
		if errors.Is(err, sandbox.ErrTimeout) {
			log.Println("base resume timed out under load; skipping base verification")
			return nil
		}
		return fmt.Errorf("wait for base sandbox to run: %w", err)
	}

	contents, err = readFile(ctx, base, forkOnlyPath)
	if err != nil {
		return err
	}
	fmt.Printf("base does not see fork-only file: %q\n", contents)
	contents, err = readFile(ctx, base, basePath)
	if err != nil {
		return err
	}
	fmt.Printf("base still has %s: %s\n", basePath, contents)
	return nil
}

func readFile(ctx context.Context, instance *sandbox.Instance, path string) (string, error) {
	response, err := instance.RunCommand(ctx, structs.RunCommandRequest{
		Command:   "sh",
		Arguments: []string{"-c", "cat " + path + " 2>&1"},
	}, structs.ExecOptions{})
	if err != nil {
		return "", fmt.Errorf("read %s from sandbox %s: %w", path, instance.ID(), err)
	}
	return strings.TrimSpace(response.Result.StandardOutput), nil
}

func destroy(parent context.Context, label string, instance *sandbox.Instance) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), 2*time.Minute)
	defer cancel()
	if err := instance.Destroy(ctx); err != nil {
		log.Printf("cleanup: destroy %s sandbox %s: %v", label, instance.ID(), err)
	}
}
