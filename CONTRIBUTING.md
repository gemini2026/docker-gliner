# Contributing

Thanks for your interest in docker-gliner. It's a small repo — a Go Compose
provider plugin plus a bundled Python GLiNER server — so contributing is
straightforward.

## Dev setup

You need Go 1.25+ and Python 3.12+.

```sh
# Go side
go build ./...

# Python server side
cd server
python -m venv .venv && . .venv/bin/activate
pip install -r requirements.txt          # runtime (gliner2, torch, fastapi, ...)
pip install -r requirements-dev.txt      # test deps (no torch/gliner2 — they're faked)
```

A `Makefile` wraps the common commands: `make build`, `make test`,
`make test-server`, `make fmt`, `make vet`, `make install`.

## Tests

Both suites must pass before a PR is merged:

```sh
go test ./...          # provider parsing, message encoding, runner lifecycle, health-wait
cd server && python -m pytest -q   # device resolution + /health contract (fake torch/gliner2)
```

The server tests inject fake `torch`/`gliner2` modules, so they run on any
machine including CPU-only CI — no GPU or model download required.

## Style

- Go: `gofmt` must be clean (`gofmt -l .` prints nothing) and `go vet ./...` passes.
- Python: `ruff check server/` must pass.
- Keep the provider protocol behavior in sync with `docs/PROTOCOL.md`. If you
  change how Compose invokes the binary or the message shapes, update that doc.

## Commits & PRs

- Use feature branches and PRs; keep the description focused on what changed.
- Do not add `Co-Authored-By`, `Signed-off-by`, or other attribution trailers.
- Small, reviewable PRs are preferred over large ones.

## Reporting bugs

Open an issue with your host OS, the device GLiNER landed on (`mps`/`cuda`/`cpu`,
visible at `/health`), the relevant `compose.yaml` snippet, and the server log
path printed by the plugin (`$XDG_STATE_HOME/docker-gliner/<project>__<service>.log`).
