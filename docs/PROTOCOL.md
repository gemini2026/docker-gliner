# Docker Compose provider protocol (captured)

Source: Docker Compose docs — provider/extension services
(`docker/compose` `docs/extension.md`, `docs.docker.com/compose/how-tos/provider-services/`).

A **provider service** is a Compose service that declares a `provider:` block
instead of an `image:`/`build:`. Compose delegates that service's lifecycle to an
external binary on `$PATH` named after `provider.type` (e.g. `type: gliner` →
binary `docker-gliner`, with bare `gliner` also accepted). This is the same
mechanism Docker Model Runner uses.

```yaml
services:
  ner:
    provider:
      type: gliner          # -> invokes the `docker-gliner` binary
      options:
        model: fastino/gliner2-base-v1
        device: auto
  app:
    image: my-app
    depends_on:
      - ner                  # receives NER_GLINER_ENDPOINT (see setenv below)
```

## 1. How Compose invokes the binary

Compose calls the binary with a `compose` super-command, the project name, the
lifecycle subcommand, each option as `--key=value`, and the **service name** as a
trailing positional argument:

```
docker-gliner compose --project-name <PROJECT> up   --<opt>=<val> ... "<service>"
docker-gliner compose --project-name <PROJECT> down "<service>"
```

(Compose may also issue `stop`; treat it like `down` if emitted.)

Example for the service above:

```
docker-gliner compose --project-name myproj up --model=fastino/gliner2-base-v1 --device=auto "ner"
```

Notes:
- `options:` map keys become `--key=value` flags. Values are always strings here.
- The service name is the **last positional arg**, not a flag.
- `up` **MUST be idempotent**: if the resource is already running, re-running
  `up` must re-emit the same `setenv` values.

## 2. stdout message protocol

The binary streams **line-delimited JSON** objects to stdout. Each object:

```json
{ "type": "<type>", "message": "<content>" }
```

| type     | purpose                                                       |
|----------|--------------------------------------------------------------|
| `info`   | progress text shown in the Compose UI                        |
| `error`  | failure reason; rendered as the service's failure message    |
| `setenv` | inject an env var into services that `depends_on` this one    |
| `debug`  | only shown when Compose runs with `--verbose`                 |

Verbatim examples from the spec:

```json
{ "type": "info", "message": "preparing mysql ..." }
{ "type": "setenv", "message": "URL=https://awesomecloud.com/db:1234" }
```

## 3. `setenv` injection convention

A `setenv` payload is `VARNAME=value`. Compose injects it into each dependent
service **prefixed with the provider service's name, uppercased**:

```
service "ner" emits  {"type":"setenv","message":"ENDPOINT=http://localhost:8100"}
        ->  dependent services get  NER_ENDPOINT=http://localhost:8100
```

So we emit `ENDPOINT=<url>` and the consumer sees `<SERVICE>_ENDPOINT`. The bare
variable name is configurable via the `endpoint_var` option (default `ENDPOINT`).

## 4. Exit codes

- `0` on success (the spec's sequence diagram shows `exit 0`).
- Non-zero on failure; emit an `error` message first so Compose can show why.
  The spec doesn't enumerate specific non-zero codes, so we use `1`.
