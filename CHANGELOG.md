# Changelog

Notable changes to the CreateOS Go SDK are recorded here. Versions follow the
repository's Git tags.

## [Unreleased]

### Added

- Sandbox access token create, inspect, rotate, and disable methods on
  `Instance`, with separate plaintext and metadata response types. A
  `WithAccessToken` handle runs sandbox operations using the delegated token.

## [0.0.5] - 2026-09-16

This release publishes the current SDK under a new immutable module version.
There are no API or behavior changes from v0.0.4.

## [0.0.4] - 2026-09-16

### Added

- Release changelog for the Go SDK, linked from the README and contributing
  guide.

## [0.0.3] - 2026-09-16

### Changed

- `NewClient()` now reads `CREATEOS_API_KEY` as its only API key environment
  variable. Set that variable in existing deployments or pass `WithAPIKey`.
- Updated the README with the SDK list, authentication details, and a command
  for running the hello-world example.

## [0.0.2] - 2026-09-09

### Added

- Per-request timeouts for file uploads and downloads. Download timeouts remain
  active while the response body is read.

## [0.0.1] - 2026-09-09

### Added

- Initial Go SDK for sandbox lifecycle, command execution, file transfer,
  ingress, networks, templates, managed processes, and desktop control.
- Runnable examples and development checks.

[Unreleased]: https://github.com/NodeOps-app/createos-go-sdk/compare/v0.0.5...HEAD
[0.0.5]: https://github.com/NodeOps-app/createos-go-sdk/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/NodeOps-app/createos-go-sdk/compare/v0.0.3...v0.0.4
[0.0.3]: https://github.com/NodeOps-app/createos-go-sdk/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/NodeOps-app/createos-go-sdk/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/NodeOps-app/createos-go-sdk/tree/v0.0.1
