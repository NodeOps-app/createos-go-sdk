# Checkpoints

Capture a directory, modify it, restore its earlier contents, and remove the
checkpoint.

```sh
export CREATEOS_API_KEY="your-api-key"
go run ./examples/checkpoints
```

The example creates live resources and cleans them up when it finishes. In real
applications, stop all writers before creating or restoring a checkpoint.
