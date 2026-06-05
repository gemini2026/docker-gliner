---
name: Bug report
about: Something isn't working
labels: bug
---

**What happened**

A clear description of the bug.

**Environment**

- Host OS / arch (e.g. macOS 26 arm64, Ubuntu 24.04 x86_64):
- Docker Desktop / Compose version (`docker compose version`):
- Device GLiNER landed on (from `/health`: `mps` / `cuda` / `cpu`):
- Go version (`go version`) if building from source:

**Compose snippet**

```yaml
# the provider service block from your compose.yaml
```

**Logs**

The host process log printed by the plugin:
`$XDG_STATE_HOME/docker-gliner/<project>__<service>.log` (or
`~/.local/state/docker-gliner/...`). Paste the relevant lines.

**Expected vs actual**

What you expected to happen, and what happened instead.
