# CreateOS Go SDK

Launch an isolated cloud sandbox, run real commands, stream output, move files,
open a preview URL, and tear everything down from Go.

## Your first sandbox

```sh
go get github.com/NodeOps-app/createos-go-sdk
```

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/NodeOps-app/createos-go-sdk/sandbox"
	"github.com/NodeOps-app/createos-go-sdk/structs"
)

func main() {
	ctx := context.Background()
	client, err := sandbox.NewClient(
		sandbox.WithAPIKey("your-api-key"),
	)
	if err != nil {
		log.Fatal(err)
	}

	instance, err := client.CreateSandbox(ctx, structs.CreateSandboxRequest{
		Name:   "hello-go",
		Shape:  "s-4vcpu-4gb",
		RootFS: "devbox:1",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := instance.Destroy(context.Background()); err != nil {
			log.Printf("destroy sandbox: %v", err)
		}
	}()

	response, err := instance.RunCommand(ctx, structs.RunCommandRequest{
		Command:   "sh",
		Arguments: []string{"-c", `printf "Go says hello from $(uname -m)\n"`},
	}, structs.ExecOptions{})
	if err != nil {
		log.Printf("run command: %v", err)
		return
	}

	fmt.Print(response.Result.StandardOutput)
}
```

```text
Go says hello from x86_64
```

`WithAPIKey` configures authentication explicitly. Do not commit a real API key
to source control; inject the value through your application's secret manager.
Additional client options can configure the endpoint and default timeout:

```go
client, err := sandbox.NewClient(
	sandbox.WithAPIKey(apiKey),
	sandbox.WithBaseURL("http://localhost:8080"),
	sandbox.WithTimeout(30*time.Second),
)
```

As an optional alternative, `NewClient()` reads `CREATEOS_SANDBOX_API_KEY`
when `WithAPIKey` is not provided. Explicit options always take precedence.

## Documentation

- [CreateOS Sandbox overview](https://createos.sh/docs/Sandbox/Overview)
  explains the sandbox model, lifecycle, networking, storage, and isolation.
- [CreateOS Sandbox documentation](https://createos.sh/docs)
  contains the REST API reference and product guides.
- [Go API reference](https://pkg.go.dev/github.com/NodeOps-app/createos-go-sdk)
  is generated from the SDK's public GoDoc after a tagged release.
- [CreateOS TypeScript SDK](https://github.com/NodeOps-app/createos-sandbox-sdk)
  provides the same sandbox capabilities for JavaScript and TypeScript
  applications.
- [CreateOS Python SDK](https://github.com/NodeOps-app/createos-python-sdk)
  provides the same sandbox capabilities for Python applications.
- [Runnable examples](#examples) cover command execution, files, streaming,
  ingress, snapshots, networking, templates, managed processes, and desktop use.
- [Contributing guide](CONTRIBUTING.md) documents development checks and commit
  conventions.
- [`CLAUDE.md`](CLAUDE.md) is the agent guide, covering repository conventions,
  the sibling-SDK map, and the cross-SDK parity protocol.

## Stream output as it happens

Long-running commands do not need to disappear behind a buffered HTTP call:

```go
stream, err := instance.StreamCommand(ctx, structs.RunCommandRequest{
	Command:   "sh",
	Arguments: []string{"-c", `for n in 1 2 3; do echo "step $n"; sleep 1; done`},
}, structs.ExecOptions{})
if err != nil {
	return err
}
defer func() {
	if err := stream.Close(); err != nil {
		log.Printf("close command stream: %v", err)
	}
}()

for {
	event, err := stream.Receive()
	if errors.Is(err, io.EOF) {
		break
	}
	if err != nil {
		return err
	}

	switch event.Type {
	case structs.ExecStreamEventStdout:
		fmt.Print(event.Data)
	case structs.ExecStreamEventStderr:
		fmt.Fprint(os.Stderr, event.Data)
	case structs.ExecStreamEventExit:
		fmt.Printf("exit code: %d\n", *event.ExitCode)
	}
}
```

Stopping early is safe: closing the stream closes the response body and
cancels the underlying request.

## Move files without shell escaping

```go
err := instance.Files().Upload(
	ctx,
	"/workspace/config.json",
	strings.NewReader(`{"mode":"production"}`),
)
if err != nil {
	return err
}

file, err := instance.Files().Download(ctx, "/workspace/config.json")
if err != nil {
	return err
}
defer func() {
	if err := file.Close(); err != nil {
		log.Printf("close downloaded file: %v", err)
	}
}()

contents, err := io.ReadAll(file)
if err != nil {
	return err
}
```

For large transfers, override the timeout for that operation without changing
the client's default timeout:

```go
transferOptions := structs.RequestOptions{Timeout: 30 * time.Minute}

if err := instance.Files().Upload(ctx, remotePath, source, transferOptions); err != nil {
	return err
}
file, err := instance.Files().Download(ctx, remotePath, transferOptions)
if err != nil {
	return err
}
defer func() {
	if err := file.Close(); err != nil {
		log.Printf("close downloaded file: %v", err)
	}
}()
```

The timeout covers the complete transfer, including reading the downloaded
body. Uploads are not retried because an arbitrary `io.Reader` may not be safe
to replay after a partial write.

## Keep a process alive after disconnecting

Managed processes are resources rather than fragile terminal sessions. Start
one, reconnect from its output sequence, send input or signals, and wait for
either the leader or its complete process tree:

```go
process, err := instance.Processes().Create(ctx, structs.ManagedProcessCreateRequest{
	Command:   "python3",
	Arguments: []string{"-m", "http.server", "8080"},
})
if err != nil {
	return err
}

process, err = instance.Processes().Wait(ctx, process.ProcessID,
	structs.ManagedProcessWaitOptions{
		Scope:       structs.ManagedProcessWaitScopeTree,
		WaitTimeout: 30 * time.Second,
	},
)
if err != nil {
	return err
}
```

## Turn a service into a URL

Create with ingress enabled, wait for the server to listen, then ask the
instance for its public URL:

```go
instance, err := client.CreateSandbox(ctx, structs.CreateSandboxRequest{
	Shape:          "s-4vcpu-4gb",
	RootFS:         "devbox:1",
	IngressEnabled: true,
})
if err != nil {
	return err
}

_, err = instance.Processes().Create(ctx, structs.ManagedProcessCreateRequest{
	Command:   "python3",
	Arguments: []string{"-m", "http.server", "8080", "--bind", "0.0.0.0"},
})
if err != nil {
	return err
}

if err := instance.WaitForPort(ctx, "127.0.0.1", 8080, 15*time.Second); err != nil {
	return err
}

previewURL, err := instance.PreviewURL(8080)
if err != nil {
	return err
}
fmt.Println(previewURL)
```

## Everything is already connected

Account-level services are initialized by `NewClient`:

```go
templates := client.Templates()
networks := client.Networks()
disks := client.Disks()

customTemplates, err := templates.List(ctx, structs.PaginationOptions{})
if err != nil {
	return err
}
fmt.Printf("%d templates ready; networks=%T disks=%T\n", len(customTemplates), networks, disks)
```

Instance-level services are initialized when a sandbox handle is created or
retrieved:

```go
instance.Files()
instance.Processes()
instance.Computer().Mouse()
instance.Computer().Keyboard()
instance.Computer().Windows()
instance.Computer().Screens()
```

## Connect sandboxes on a private network

Create an overlay network, attach a running sandbox, and inspect the resulting
membership. Cleanup runs in reverse order, so the sandbox detaches before the
network is deleted:

```go
network, err := client.Networks().Create(ctx, structs.NetworkCreateRequest{
	Name: "agent-mesh",
})
if err != nil {
	return err
}
defer func() {
	if err := client.Networks().Delete(context.Background(), network.ID); err != nil {
		log.Printf("delete network: %v", err)
	}
}()

if err := instance.AttachNetwork(ctx, network.ID); err != nil {
	return err
}
defer func() {
	if err := instance.DetachNetwork(context.Background(), network.ID); err != nil {
		log.Printf("detach network: %v", err)
	}
}()

connected, err := client.Networks().Get(ctx, network.ID)
if err != nil {
	return err
}

for _, member := range connected.Members {
	fmt.Printf("sandbox=%s private-ip=%s status=%s\n",
		member.SandboxID,
		member.IPAddress,
		member.Status,
	)
}
```

## Lifecycle reads like the domain

```go
if err := instance.Pause(ctx); err != nil {
	return err
}
if err := instance.WaitUntilPaused(ctx, structs.WaitOptions{}); err != nil {
	return err
}

clone, err := instance.Fork(ctx, structs.ForkSandboxRequest{})
if err != nil {
	return err
}
defer func() {
	if err := clone.Destroy(context.Background()); err != nil {
		log.Printf("destroy clone: %v", err)
	}
}()

if err := instance.Resume(ctx); err != nil {
	return err
}
if err := instance.Destroy(ctx); err != nil {
	return err
}
```

The `Instance` handle caches the latest server projection safely. Lifecycle
mutations and `Refresh` update it, while `ID`, `Name`, `Status`, `IPAddress`,
and `Data` provide concurrent-safe reads.

## Errors stay inspectable

```go
var apiError *sandbox.APIError
if errors.As(err, &apiError) {
	fmt.Printf("HTTP %d, code=%d, request=%s\n",
		apiError.StatusCode,
		apiError.Code,
		apiError.RequestID,
	)
}

if errors.Is(err, sandbox.ErrTimeout) {
	// A lifecycle or readiness wait exhausted its budget.
}
```

## Examples

Runnable examples live under [`examples/`](examples/):

- [Hello world](examples/hello-world/main.go)
- [HTTP execution server](examples/execution-server/README.md)
- [Command streaming](examples/command-streaming/main.go)
- [Files and snapshots](examples/files-and-snapshots/main.go)
- [Ingress preview](examples/ingress-preview/main.go)
- [Private overlay network](examples/network/main.go)
- [Custom template](examples/custom-template/main.go)
- [Managed process lifecycle](examples/managed-process/main.go)
- [Desktop and noVNC](examples/desktop/main.go)

## Development

Tool versions are pinned in `.tool-versions` for asdf:

```sh
asdf install
make install-hooks
make check
make test-race
```

Commits follow Conventional Commits and are validated locally and in pull
requests. See [CONTRIBUTING.md](CONTRIBUTING.md) for accepted types and examples.

The CI pipeline runs race-enabled tests, `go vet`, and golangci-lint. The
golangci-lint v2 configuration includes error, context, body-closing, security,
static-analysis, and formatting checks.

## Package layout

```text
sandbox/   client behavior and stateful resource handles
structs/   public request, response, option, and enum contracts
internal/  transport, JSend/NDJSON, polling, and redaction
examples/  independently compilable programs
```

## About CreateOS

[CreateOS](https://createos.sh) is an execution and governance platform for AI
agents and applications. Learn more about isolated Firecracker-based workloads
on the [CreateOS Sandbox product page](https://createos.sh/products/sandbox).

## License

This SDK is available under the [MIT License](LICENSE).
