# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-06-04

First release.

### Added

- Docker Compose **provider plugin** (`docker-gliner`) that runs a GLiNER server
  as a native host process so it can use the GPU (the Mac Metal GPU is
  unreachable from inside a container).
- Provider protocol implementation: parses Compose's `up`/`down` invocation and
  emits line-delimited JSON messages (`info`/`error`/`setenv`/`debug`).
- Detached subprocess runner with PID-file state (`$XDG_STATE_HOME/docker-gliner`),
  graceful stop (SIGTERM → SIGKILL), idempotent `up`, and an HTTP health poll
  that publishes the endpoint via `setenv` once the server is ready.
- Bundled GLiNER2 FastAPI server with automatic device selection
  (cuda → mps → cpu) and a `/health` endpoint reporting the active device.
- `command` option to manage an arbitrary host process instead of the bundled
  server.
- Example `examples/compose.yaml`, protocol notes in `docs/PROTOCOL.md`, and an
  Apache-2.0 license.

[Unreleased]: https://github.com/gemini2026/docker-gliner/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/gemini2026/docker-gliner/releases/tag/v0.1.0
