# Execution server

This example turns the CreateOS Go SDK into a small HTTP execution service. A
`POST /v1/execute` request creates a fresh sandbox, runs one command, captures
its output, destroys the sandbox, and returns JSON.

Run the server:

```sh
export CREATEOS_SANDBOX_API_KEY="your-api-key"
go run ./examples/execution-server
```

It listens on `127.0.0.1:8080` by default. Send a command without shell
interpolation:

```sh
curl --fail-with-body http://127.0.0.1:8080/v1/execute \
  --header 'Content-Type: application/json' \
  --data '{"command":"python3","arguments":["-c","print(sum(range(10)))"]}'
```

The response includes command output even when the command exits nonzero:

```json
{
  "stdout": "45\n",
  "stderr": "",
  "exitCode": 0,
  "executionMilliseconds": 154.48
}
```

The example limits request bodies to 1 MiB and permits four concurrent
executions. Excess requests receive HTTP 429. Invalid input receives HTTP 400,
control-plane failures receive HTTP 502, and timeouts receive HTTP 504.

> [!WARNING]
> This endpoint runs arbitrary commands and intentionally binds only to
> localhost. Add authentication, authorization, rate limiting, audit logging,
> and workload policy before exposing a similar service to any network.

Set `EXECUTION_SERVER_ADDRESS` to override the listen address.
