# Reusable snapshots

Capture a reusable disk snapshot, wait for it to become ready, and restore a
new sandbox from it.

```sh
export CREATEOS_API_KEY="your-api-key"
go run ./examples/reusable-snapshots
```

The example creates live resources and cleans them up when it finishes.
