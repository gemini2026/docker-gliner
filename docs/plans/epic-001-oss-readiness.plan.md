# Epic 001: OSS readiness & publication

<!--
METADATA (agent-parseable):
epic_id: 001
date: 2026-06-04
status: In Progress
source_document: user-provided
scope_items: REQ-10, REQ-11, REQ-12, OSS-1, OSS-2, OSS-3, OSS-4
total_tasks: 8
estimated_effort: 0.5 day
iteration: 2 of ~4
-->

**Date:** 2026-06-04
**Status:** In Progress
**Source:** User-provided requirements ("plan and execute standalone OSS plugin repo")
**Iteration:** 2 of ~4 estimated

---

## 1. Epic Summary

The core plugin (Epic 000) is built, tested, and committed locally. This Epic
turns that local repo into a publishable open-source project: standard
contributor hygiene (CONTRIBUTING, CHANGELOG, Makefile, issue/PR templates),
continuous integration, a first tagged release, and publication to GitHub. After
this Epic the repo is a self-standing OSS project others can clone, build, and
contribute to. No application code changes — only repo metadata, CI, and release.

### Scope items included

| Source ID | Title | Type | Priority |
|-----------|-------|------|----------|
| OSS-1 | CONTRIBUTING.md + CHANGELOG.md | docs | P1 |
| OSS-2 | Makefile (build/test/install/fmt convenience) | infra | P2 |
| OSS-3 | Issue + PR templates | docs | P2 |
| OSS-4 | README badges + repo metadata polish | docs | P3 |
| REQ-10 | GitHub Actions CI (go + server) | infra | P1 |
| REQ-11 | Publish to GitHub (create repo, push) | infra | P0 |
| REQ-12 | Tag v0.1.0 release | infra | P2 |

### Scope items deferred (next Epics)

| Source ID | Title | Tentative Epic | Reason deferred |
|-----------|-------|----------------|-----------------|
| AB-1..AB-3 | agentboost integration | Epic 002 | Separate codebase; depends on REQ-11 |
| CL-1 | Cloud profiles + vertex/azure routing | Epic 003 | Lower priority, separate concern |
| PUB-1 | Registry / Homebrew publishing | Epic 003 | Depends on REQ-12 |

---

## 2. Prerequisites

- [x] Epic 000 delivered and committed (`6c40d0b`).
- [x] `go test ./...` and `pytest server/` green.
- [ ] **Q-1 resolved:** GitHub owner — module path is `github.com/amichel/...`
      but authenticated `gh` account is `gemini2026`. Pick the publish target
      (and whether `go.mod` module path changes to match) before REQ-11/REQ-12.
- [ ] **Q-2 resolved:** confirm adding GitHub Actions CI (global rule requires
      explicit confirmation before adding CI/CD).

---

## 3. Task Registry

| ID | Title | Type | Files | Depends on | Effort | Source |
|----|-------|------|-------|------------|--------|--------|
| T-001 | Add CONTRIBUTING.md | docs | `CONTRIBUTING.md` | — | XS | OSS-1 |
| T-002 | Add CHANGELOG.md (Keep a Changelog) | docs | `CHANGELOG.md` | — | XS | OSS-1 |
| T-003 | Add Makefile | infra | `Makefile` | — | S | OSS-2 |
| T-004 | Add issue + PR templates | docs | `.github/ISSUE_TEMPLATE/*`, `.github/PULL_REQUEST_TEMPLATE.md` | — | XS | OSS-3 |
| T-005 | Resolve owner (Q-1) + align `go.mod` module path | config | `go.mod`, imports | — | S | REQ-11 |
| T-006 | Add CI workflow (gated on Q-2) | infra | `.github/workflows/ci.yml` | T-003 | S | REQ-10 |
| T-007 | Create GitHub repo + push (gated on Q-1) | infra | remote | T-005 | S | REQ-11 |
| T-008 | Tag + release v0.1.0 | infra | tag | T-007 | XS | REQ-12 |

---

## 4. Task Details

### T-001: Add CONTRIBUTING.md
- **Type:** docs · **Source:** OSS-1 · **Dependencies:** None
- **Files:** `CONTRIBUTING.md` — dev setup, how to run both test suites, code style (gofmt, ruff), PR expectations.
- **Acceptance criteria:**
  - [ ] Documents `go test ./...` and `pytest server/`.
  - [ ] States gofmt + ruff must pass; no AI-attribution trailers in commits.
- **Test command:** none (manual review)

### T-002: Add CHANGELOG.md
- **Type:** docs · **Source:** OSS-1 · **Dependencies:** None
- **Files:** `CHANGELOG.md` — Keep-a-Changelog format; `0.1.0` Unreleased→first release with the Epic 000 feature set.
- **Acceptance criteria:**
  - [ ] Lists the provider plugin, bundled server, example, protocol docs under 0.1.0.
- **Test command:** none

### T-003: Add Makefile
- **Type:** infra · **Source:** OSS-2 · **Dependencies:** None
- **Files:** `Makefile` — `build`, `test`, `test-server`, `fmt`, `vet`, `install`, `clean` targets.
- **Implementation notes:** `install` copies the binary + `server/` dir together (the binary locates the default server beside itself, per `defaultServerPath()` in `cmd/docker-gliner/main.go`).
- **Acceptance criteria:**
  - [ ] `make build` produces `./docker-gliner`.
  - [ ] `make test` runs `go test ./...`; `make test-server` runs pytest.
- **Test command:** `make build && make test`

### T-004: Add issue + PR templates
- **Type:** docs · **Source:** OSS-3 · **Dependencies:** None
- **Files:** `.github/ISSUE_TEMPLATE/bug_report.md`, `.github/ISSUE_TEMPLATE/feature_request.md`, `.github/PULL_REQUEST_TEMPLATE.md`.
- **Implementation notes:** Templates only — NOT workflows. These are not CI/CD and are safe to add without the Q-2 gate.
- **Acceptance criteria:**
  - [ ] Bug template asks for host OS, device (mps/cuda/cpu), compose snippet, logs path.
- **Test command:** none

### T-005: Resolve owner (Q-1) + align go.mod module path
- **Type:** config · **Source:** REQ-11 · **Dependencies:** None (needs Q-1 answer)
- **Files:** `go.mod` module line + all `github.com/<owner>/docker-gliner` imports in `cmd/` if owner changes.
- **Implementation notes:** If publishing under `gemini2026`, change module to `github.com/gemini2026/docker-gliner` and update imports via a single `sed`, then `go build ./...`.
- **Acceptance criteria:**
  - [ ] `go build ./...` green after any module rename.
  - [ ] Module path matches the chosen GitHub owner.
- **Test command:** `go build ./... && go test ./...`

### T-006: Add CI workflow (GATED on Q-2)
- **Type:** infra · **Source:** REQ-10 · **Dependencies:** T-003
- **Files:** `.github/workflows/ci.yml` — matrix: Go job (`go vet`, `go build`, `go test ./...`) + Python job (`ruff check server/`, `pytest server/`).
- **Implementation notes:** Do NOT create until the user explicitly confirms (global rule: never add CI/CD without confirmation). Pin action versions.
- **Acceptance criteria:**
  - [ ] Workflow runs both suites on push + PR.
  - [ ] User explicitly approved adding CI.
- **Test command:** `act` dry-run or first push CI run

### T-007: Create GitHub repo + push (GATED on Q-1)
- **Type:** infra · **Source:** REQ-11 · **Dependencies:** T-005
- **Files:** git remote.
- **Implementation notes:** `gh repo create <owner>/docker-gliner --<public|private> --source=. --remote=origin --push`. Outward-facing — confirm visibility + owner first.
- **Acceptance criteria:**
  - [ ] Repo exists at the chosen owner; `main` pushed.
- **Test command:** `git ls-remote origin`

### T-008: Tag + release v0.1.0
- **Type:** infra · **Source:** REQ-12 · **Dependencies:** T-007
- **Files:** annotated tag `v0.1.0`.
- **Implementation notes:** `git tag -a v0.1.0`; optional `gh release create v0.1.0` with notes from CHANGELOG.
- **Acceptance criteria:**
  - [ ] `v0.1.0` tag pushed; CHANGELOG `0.1.0` dated.
- **Test command:** `git tag -l v0.1.0`

---

## 5. Verification Checklist

```bash
go build ./... && go vet ./... && go test ./... && gofmt -l .
( cd server && python -m pytest -q && ruff check . )
docker compose -f examples/compose.yaml config -q
```

- [ ] All tests pass; gofmt + ruff clean.
- [ ] Module path matches publish owner.
- [ ] CI green on first run (if T-006 approved).
- [ ] Repo public/visible at chosen owner (if T-007 approved).

---

## 6. Commit Guidance

| Commit | Tasks | Message |
|--------|-------|---------|
| 1 | T-001..T-004 | `chore: add OSS hygiene (contributing, changelog, makefile, templates)` |
| 2 | T-005 | `chore: align go module path to <owner>` |
| 3 | T-006 | `ci: add go + server workflow` |

---

## 7. Next Iteration Preview

**Next Epic:** Epic 002 — agentboost integration
**Scope:** consume `provider: { type: gliner }` in agentboost's compose, wire `AGENTBOOST_GLINER_ENDPOINT`, add a Compose-aware `models doctor`, dedupe agentboost's own `serve_gliner.py`.
**Blockers:** REQ-11 (this repo must be published/consumable first).

---

*Plan generated on 2026-06-04 from user-provided requirements.*
