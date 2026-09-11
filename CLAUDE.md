# CLAUDE.md

Agent guide for the **CreateOS Go SDK**. Contributor conventions live in
[`CONTRIBUTING.md`](CONTRIBUTING.md); this file covers what an agent needs that
the contributor guide does not.

## This repository

- Module: `github.com/NodeOps-app/createos-go-sdk`, imported as
  `sandbox` and `structs`.
- Source: `sandbox/` is the client surface (`Client`, `Instance`, commands,
  files, processes, ingress, networks, disks, templates, computer use).
  `structs/` holds the data contracts only — requests, responses, options,
  enums — and performs no network calls. `internal/` (`transport`, `protocol`,
  `polling`, `redact`) is not part of the public API.
- Go 1.25.0. `go.mod` and `.tool-versions` pin it, along with
  golangci-lint 2.12.2; `asdf install` installs both.
- Local checks: `make check` (which runs `fmt`, `lint`, `vet`, `test`), or the
  targets on their own — `make fmt`, `make lint` (`golangci-lint run ./...`),
  `make vet`, `make test` (`go test ./...`), `make test-race`. Lint config is
  `.golangci.yml`: the standard set plus `bodyclose`, `contextcheck`,
  `errorlint`, `exhaustive`, `gocritic`, `gosec`, `revive` and others, with
  `gofmt` and `goimports` as formatters.
- Hooks: `make install-hooks` points `core.hooksPath` at `.githooks/`. The
  `commit-msg` hook runs `scripts/check-commit-message.sh`, which enforces the
  Conventional Commits rules in CONTRIBUTING.md. CI (`.github/workflows/ci.yml`)
  re-checks every commit in a pull request, plus `go test -race`, `go vet`, and
  golangci-lint.
- Release: there is no publish step. Tag a semver release (`v0.0.1`, `v0.0.2`
  so far) and the module proxy and <https://pkg.go.dev> pick it up from the tag.

## The CreateOS SDK family

This is one of three clients for the **same** CreateOS Sandbox API. They are
separate repositories that are expected to stay behaviourally in sync. A change
worth making here is usually worth making in the siblings.

| Language | Repository | Package | Agent guide | Local sibling |
| --- | --- | --- | --- | --- |
| TypeScript | [createos-sandbox-sdk](https://github.com/NodeOps-app/createos-sandbox-sdk) | `@nodeops-createos/sandbox` | [`CLAUDE.md`](https://github.com/NodeOps-app/createos-sandbox-sdk/blob/main/CLAUDE.md) → [`AGENTS.md`](https://github.com/NodeOps-app/createos-sandbox-sdk/blob/main/AGENTS.md) | `../fc-sdk` |
| Go | this repo | `github.com/NodeOps-app/createos-go-sdk` | this file | — |
| Python | [createos-python-sdk](https://github.com/NodeOps-app/createos-python-sdk) | `createos-sandbox` | [`CLAUDE.md`](https://github.com/NodeOps-app/createos-python-sdk/blob/main/CLAUDE.md) | `../createos-python-sdk` |

Upstream of all three:

- **Service** — `../fc` (`nodeops-app/fc`). Source of truth for the wire
  contract: `openapi.yaml`, plus `CLAUDE.md` / `AGENT.md` for its own rules.
  If the SDKs disagree about what the API does, the service wins.
- **Public docs** — `../website-04/content/docs/Sandbox/`, published at
  <https://createos.sh/docs/Sandbox>. Language snippets are **not** written in
  Markdown: they live in `lib/docs/sdk-code-examples.ts` and render through
  `<SdkCodeTabs example="..." />`, one entry per language. A new SDK capability
  that users should see is not shipped until that file has it.

## Cross-SDK parity protocol

Run this before you call any change to this repo done. It is a read-and-report
protocol — **do not edit a sibling repository unless the user asks you to.**

1. **Classify the change.**
   - *Wire contract* (new endpoint, changed field, new request/response shape)
     → affects all three SDKs and usually the docs.
   - *Behaviour* (retry policy, timeout default, stream framing, error
     mapping) → affects all three SDKs.
   - *Bug fix* → check whether the siblings have the same bug. They were
     written from the same spec, so they usually do.
   - *Ergonomics* (a new `With...` option func, a context-aware helper that
     wraps a polling loop) → often has a natural equivalent in the siblings;
     propose it, don't assume it.
   - *Repo-local* (packaging, lint config, CI) → no parity obligation.
2. **Check the siblings.** Read the matching file under `../fc-sdk/src/` and
   `../createos-python-sdk/src/createos/`. If a sibling checkout is missing, say
   so rather than guessing.
3. **Report.** End the task with a short parity note: what ports to which SDK,
   what does not, and why. Name the file the sibling change would land in.
4. **Docs.** If the change adds or alters a user-visible capability, say
   whether `sdk-code-examples.ts` and the affected page under
   `content/docs/Sandbox/` need updating.

The same protocol runs in reverse: when the TypeScript or Python SDK gains a
feature or fix, check whether it belongs here.

### Current parity baseline

The Go and Python SDKs expose the same surface and ship the same nine examples
(`hello-world`, `command-streaming`, `files-and-snapshots`, `ingress-preview`,
`managed-process`, `network`, `custom-template`, `desktop`,
`execution-server`). The TypeScript SDK has the same core surface plus a much
larger integration-example corpus. Treat a gap against Python as a real gap;
treat a gap against a TypeScript *integration example* as optional.
</content>
