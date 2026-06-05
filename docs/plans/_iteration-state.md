# Iteration State

**Source document:** user-provided ("create a standalone OSS plugin repo")
**Total scope items:** 13
**Last updated:** 2026-06-04

## Delivered Epics

| Epic | Title | Date delivered | Scope items |
|------|-------|---------------|-------------|
| 000 | Core build — Compose provider plugin + bundled GLiNER server | 2026-06-04 | REQ-1..REQ-9 (T-001..T-010 of the original build plan) |

Epic 000 shipped in commit `6c40d0b`: Go provider binary (protocol parsing,
line-delimited JSON messages, detached subprocess runner with PID-file state,
health poll), bundled GLiNER2 FastAPI server, example compose, PROTOCOL.md,
README, Apache-2.0 LICENSE. `go test ./...` green, `pytest server/` 8 passing,
`docker compose config` valid, gofmt + ruff clean.

## Current Epic

| Epic | Title | Status | Scope items |
|------|-------|--------|-------------|
| 001 | OSS readiness & publication | In Progress | REQ-10 (CI), REQ-11 (publish), REQ-12 (release), OSS-1..OSS-4 (hygiene) |

## Backlog

| Source ID | Title | Tentative Epic | Priority | Dependencies |
|-----------|-------|----------------|----------|--------------|
| AB-1 | agentboost integration: consume `provider: { type: gliner }`, wire `AGENTBOOST_GLINER_ENDPOINT` | Epic 002 | P1 | REQ-11 |
| AB-2 | agentboost `models doctor` Compose-aware command (`src/cli/commands/models.py`) | Epic 002 | P2 | AB-1 |
| AB-3 | Dedupe agentboost's own `scripts/serve_gliner.py` to consume this repo's server | Epic 002 | P2 | AB-1 |
| CL-1 | Cloud profiles + vertex/azure planner routing | Epic 003 | P3 | — |
| PUB-1 | Registry / Homebrew tap publishing | Epic 003 | P3 | REQ-12 |

## Open Blockers

| ID | Blocker | Blocks | Owner | Status |
|----|---------|--------|-------|--------|
| Q-1 | GitHub owner mismatch: module path is `github.com/amichel/...` but the only authenticated `gh` account is `gemini2026`. Determines publish target + whether `go.mod` module path changes. | REQ-11, REQ-12 | user | Pending |
| Q-2 | Add GitHub Actions CI? (global rule: confirm before adding CI/CD) | REQ-10 | user | Pending |

## Progress Summary

- Total scope items: 13
- Delivered: 9 (69%) — Epic 000 core build
- Current Epic: 001 (OSS readiness)
- Remaining: 4 items in Epic 001 + 5 backlog items across ~2 Epics
