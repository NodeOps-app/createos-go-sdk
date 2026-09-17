# Ingress preview

Serves a page from a sandbox and opens it through a preview URL. See [main.go](main.go) for the code.

From the repository root:

```sh
export CREATEOS_API_KEY="your-api-key"
go run ./examples/ingress-preview
```

This example creates live resources and cleans them up when it finishes.
