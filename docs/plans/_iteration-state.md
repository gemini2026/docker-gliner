# Iteration State

**Source document:** user-provided ("create a standalone OSS plugin repo")
**Total scope items:** 13
**Last updated:** 2026-06-04

## Delivered Epics

| Epic | Title | Date delivered | Scope items |
|------|-------|---------------|-------------|
| 000 | Core build — Compose provider plugin + bundled GLiNER server | 2026-06-04 | REQ-1..REQ-9 (T-001..T-010 of the original build plan) |
| 001 | OSS readiness & publication | 2026-06-04 | REQ-10, REQ-11, REQ-12, OSS-1..OSS-4 |

Epic 000 shipped in commit `6c40d0b`: Go provider binary (protocol parsing,
line-delimited JSON messages, detached subprocess runner with PID-file state,
health poll), bundled GLiNER2 FastAPI server, example compose, PROTOCOL.md,
README, Apache-2.0 LICENSE. `go test ./...` green, `pytest server/` 8 passing,
`docker compose config` valid, gofmt + ruff clean.

Epic 001: published to https://github.com/gemini2026/docker-gliner (public),
module path aligned to `github.com/gemini2026/docker-gliner`, GitHub Actions CI
green (go + server jobs), tagged + released `v0.1.0`, plus OSS hygiene
(CONTRIBUTING, CHANGELOG, Makefile, issue/PR templates).

## Current Epic

None active. Next up is Epic 002 (agentboost integration) from the backlog.

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
| Q-1 | GitHub owner mismatch. | REQ-11, REQ-12 | user | Resolved: publish under `gemini2026`, module renamed to match. |
| Q-2 | Add GitHub Actions CI? | REQ-10 | user | Resolved: yes — `.github/workflows/ci.yml` added, green. |

## Progress Summary

- Total scope items: 13
- Delivered: 13 (100% of scoped) — Epics 000 + 001
- Current Epic: none active
- Remaining: 5 backlog items across ~2 Epics (agentboost integration, cloud profiles, registry/Homebrew)
