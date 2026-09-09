// Command custom-template builds a Docker-enabled root filesystem, creates a sandbox
// from it, and runs containers inside that sandbox.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/NodeOps-app/createos-go-sdk/sandbox"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

const (
	sandboxShape = "s-1vcpu-1gb"
	dockerfile   = `FROM nodeops/sandbox:debian

RUN apt-get update -qq \
 && apt-get install -y --no-install-recommends curl ca-certificates \
 && curl -fsSL https://get.docker.com | sh \
 && rm -rf /var/lib/apt/lists/*
`
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	client, err := sandbox.NewClient()
	if err != nil {
		return err
	}

	templateName := fmt.Sprintf("docker-ce-%d", time.Now().UnixMilli())
	fmt.Printf("[1/5] submitting template build: %s\n", templateName)
	template, err := client.Templates().Create(ctx, structs.TemplateCreateRequest{
		Name:       templateName,
		Dockerfile: dockerfile,
	})
	if err != nil {
		return fmt.Errorf("submit template build: %w", err)
	}
	fmt.Printf("      template id: %s  status: %s\n", template.ID, template.Status)
	defer deleteTemplate(context.WithoutCancel(ctx), client, template.ID)

	fmt.Println("[2/5] streaming build logs...")
	buildContext, cancelBuild := context.WithTimeout(ctx, 10*time.Minute)
	defer cancelBuild()
	stream, streamErr := client.Templates().FollowLogs(buildContext, template.ID, structs.TemplateLogsOptions{
		RequestOptions: structs.RequestOptions{Timeout: 10 * time.Minute},
	})
	if streamErr == nil {
		for {
			event, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					fmt.Fprintf(os.Stderr, "build log stream ended: %v\n", err)
				}
				break
			}
			if event.Line != "" {
				fmt.Println(event.Line)
			}
			if event.Final {
				fmt.Printf("      build finished: %s\n", event.Status)
				break
			}
		}
		_ = stream.Close()
	} else {
		fmt.Fprintf(os.Stderr, "build log stream unavailable: %v\n", streamErr)
	}

	if err := waitForTemplate(buildContext, client, template.ID); err != nil {
		return err
	}
	fmt.Printf("      template ready: %s\n", template.ID)

	fmt.Printf("[3/5] creating sandbox (shape=%s, rootfs=%s)...\n", sandboxShape, template.ID)
	instance, err := client.CreateSandbox(ctx, structs.CreateSandboxRequest{
		Shape:  sandboxShape,
		RootFS: template.ID,
	})
	if err != nil {
		return fmt.Errorf("create sandbox from template: %w", err)
	}
	fmt.Printf("      sandbox created: %s\n", instance.ID())
	defer destroySandbox(context.WithoutCancel(ctx), instance)

	fmt.Println("[4/5] starting dockerd...")
	if _, err := instance.Shell(ctx, "nohup setsid dockerd > /var/log/dockerd.log 2>&1 &", structs.ExecOptions{}); err != nil {
		return fmt.Errorf("start dockerd: %w", err)
	}
	if err := waitForDocker(ctx, instance); err != nil {
		return err
	}

	fmt.Println("[5/5] running containers...")
	commands := []struct {
		title   string
		request structs.RunCommandRequest
		timeout time.Duration
	}{
		{"docker run hello-world", structs.RunCommandRequest{Command: "docker", Arguments: []string{"run", "--rm", "hello-world"}}, 2 * time.Minute},
		{"docker run alpine", structs.RunCommandRequest{Command: "docker", Arguments: []string{"run", "--rm", "alpine", "sh", "-c", "echo hello from alpine && cat /etc/alpine-release"}}, time.Minute},
		{"docker images", structs.RunCommandRequest{Command: "docker", Arguments: []string{"images"}}, time.Minute},
	}
	for _, command := range commands {
		fmt.Printf("\n-- %s --\n", command.title)
		result, err := instance.RunCommand(ctx, command.request, structs.ExecOptions{
			RequestOptions: structs.RequestOptions{Timeout: command.timeout},
		})
		if err != nil {
			return fmt.Errorf("%s: %w", command.title, err)
		}
		if result.Result.ExitCode != 0 {
			return fmt.Errorf("%s exited with status %d: %s", command.title, result.Result.ExitCode, strings.TrimSpace(result.Result.StandardError))
		}
		fmt.Println(strings.TrimSpace(result.Result.StandardOutput))
	}
	return nil
}

func waitForTemplate(ctx context.Context, client *sandbox.Client, templateID string) error {
	for {
		template, err := client.Templates().Get(ctx, templateID, structs.GetTemplateOptions{})
		if err != nil {
			return fmt.Errorf("get template status: %w", err)
		}
		switch template.Status {
		case structs.TemplateStatusReady:
			return nil
		case structs.TemplateStatusPending, structs.TemplateStatusBuilding:
		case structs.TemplateStatusFailed:
			return errors.New("template build failed; see build logs above")
		default:
			return fmt.Errorf("template build entered unexpected status %q", template.Status)
		}
		if err := sleep(ctx, 2*time.Second); err != nil {
			return fmt.Errorf("wait for template build: %w", err)
		}
	}
}

func waitForDocker(ctx context.Context, instance *sandbox.Instance) error {
	for range 30 {
		result, err := instance.RunCommand(ctx, structs.RunCommandRequest{
			Command: "docker", Arguments: []string{"info", "--format", "{{.ServerVersion}}"},
		}, structs.ExecOptions{RequestOptions: structs.RequestOptions{Timeout: 5 * time.Second}})
		if err == nil && result.Result.ExitCode == 0 {
			fmt.Printf("      dockerd ready (server version: %s)\n", strings.TrimSpace(result.Result.StandardOutput))
			return nil
		}
		if err := sleep(ctx, 2*time.Second); err != nil {
			return err
		}
	}
	return errors.New("dockerd did not start within 60 seconds")
}

func sleep(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func destroySandbox(ctx context.Context, instance *sandbox.Instance) {
	cleanupContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := instance.Destroy(cleanupContext); err != nil {
		fmt.Fprintf(os.Stderr, "cleanup: destroy sandbox %s: %v\n", instance.ID(), err)
		return
	}
	fmt.Printf("destroyed sandbox: %s\n", instance.ID())
}

func deleteTemplate(ctx context.Context, client *sandbox.Client, templateID string) {
	cleanupContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := client.Templates().Delete(cleanupContext, templateID); err != nil {
		fmt.Fprintf(os.Stderr, "cleanup: delete template %s: %v\n", templateID, err)
		return
	}
	fmt.Printf("deleted template: %s\n", templateID)
}
