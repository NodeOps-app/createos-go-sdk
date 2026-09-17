# Command streaming

Runs a command and prints its output as streaming events arrive. See [main.go](main.go) for the code.

From the repository root:

```sh
export CREATEOS_API_KEY="your-api-key"
go run ./examples/command-streaming
```

This example creates live resources and cleans them up when it finishes.
