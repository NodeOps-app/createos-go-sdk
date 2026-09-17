# Custom template

Builds a template, creates a sandbox from it, and runs Docker. See [main.go](main.go) for the code.

From the repository root:

```sh
export CREATEOS_API_KEY="your-api-key"
go run ./examples/custom-template
```

This example creates live resources and cleans them up when it finishes.
