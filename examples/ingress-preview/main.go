// Command ingress-preview starts a web server in a sandbox and reaches it through the
// sandbox's public ingress URL.
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/NodeOps-app/createos-go-sdk/sandbox"
	"github.com/NodeOps-app/createos-go-sdk/structs"
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

	instance, err := client.CreateSandbox(ctx, structs.CreateSandboxRequest{
		Shape:          "s-1vcpu-256mb",
		RootFS:         "devbox:1",
		IngressEnabled: true,
	})
	if err != nil {
		return fmt.Errorf("create sandbox: %w", err)
	}
	fmt.Printf("created: %s\n", instance.ID())
	defer destroy(context.WithoutCancel(ctx), instance)

	if _, err := instance.RunCommand(ctx, structs.RunCommandRequest{
		Command: "mkdir", Arguments: []string{"-p", "/srv"},
	}, structs.ExecOptions{}); err != nil {
		return fmt.Errorf("prepare document root: %w", err)
	}
	if err := instance.Files().Upload(ctx, "/srv/index.html",
		strings.NewReader("<h1>hello from CreateOS Sandbox preview URL</h1>")); err != nil {
		return fmt.Errorf("upload index page: %w", err)
	}
	if _, err := instance.Processes().Create(ctx, structs.ManagedProcessCreateRequest{
		Command:          "python3",
		Arguments:        []string{"-m", "http.server", "8080", "--bind", "0.0.0.0"},
		WorkingDirectory: "/srv",
	}); err != nil {
		return fmt.Errorf("start HTTP server: %w", err)
	}
	if err := instance.WaitForPort(ctx, "127.0.0.1", 8080, 10*time.Second); err != nil {
		return err
	}

	previewURL, err := instance.PreviewURL(8080)
	if err != nil {
		return err
	}
	fmt.Printf("URL: %s\n", previewURL)

	// The ingress currently uses a self-signed certificate. This relaxed TLS
	// configuration is deliberately scoped to this example's single client.
	tlsConfig := &tls.Config{InsecureSkipVerify: true} //nolint:gosec // preview endpoint uses a self-signed certificate
	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, previewURL.String(), nil)
	if err != nil {
		return fmt.Errorf("create preview request: %w", err)
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("fetch preview URL: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("fetch preview URL: HTTP %d", response.StatusCode)
	}

	fmt.Println("--- response ---")
	if _, err := io.Copy(os.Stdout, response.Body); err != nil {
		return fmt.Errorf("print preview response: %w", err)
	}
	return nil
}

func destroy(ctx context.Context, instance *sandbox.Instance) {
	cleanupContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := instance.Destroy(cleanupContext); err != nil {
		fmt.Fprintf(os.Stderr, "cleanup: destroy sandbox %s: %v\n", instance.ID(), err)
		return
	}
	fmt.Printf("destroyed: %s\n", instance.ID())
}
