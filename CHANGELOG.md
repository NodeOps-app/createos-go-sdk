# Changelog

Notable changes to the CreateOS Go SDK are recorded here. Versions follow the
repository's Git tags.

## [Unreleased]

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

[Unreleased]: https://github.com/NodeOps-app/createos-go-sdk/compare/v0.0.3...HEAD
[0.0.3]: https://github.com/NodeOps-app/createos-go-sdk/compare/v0.0.2...v0.0.3
[0.0.2]: https://github.com/NodeOps-app/createos-go-sdk/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/NodeOps-app/createos-go-sdk/tree/v0.0.1
