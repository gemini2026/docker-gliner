# docker-gliner

A Docker Compose [provider plugin](https://docs.docker.com/compose/how-tos/provider-services/)
that runs a GLiNER server as a **native host process** so it can use the GPU —
then injects its endpoint into your other Compose services.

## Why this exists

On a Mac, the Metal GPU is not reachable from inside a container. Docker Desktop
runs Linux containers in a VM, and Apple provides no Metal passthrough, so any
PyTorch model (GLiNER included) running in a container is stuck on CPU. Docker
Model Runner works around this for llama.cpp by running it on the host and wiring
it into Compose through the provider-services extension point — but it only
serves GGUF causal-LM / embedding models, not GLiNER (a DeBERTa NER encoder).

`docker-gliner` applies the same trick to GLiNER: it's a small binary Compose
calls to start a host-side GLiNER server, wait for it to be healthy, and publish
its URL to dependent services. `docker compose up` brings up a GPU-accelerated
GLiNER on a Mac with no bespoke supervisor; on Linux/CI the same plugin runs the
server on CUDA or CPU.

It's generic, too: point the `command` option at any host process with a health
endpoint and it'll manage that instead — GLiNER is just the bundled default.

## Install

Build the binary and put it on `$PATH` (Compose resolves `provider.type: gliner`
to a `docker-gliner` binary):

```sh
go build -o docker-gliner ./cmd/docker-gliner
# move it somewhere on $PATH, keeping the bundled server beside it:
mkdir -p ~/.local/bin/server
cp docker-gliner ~/.local/bin/
cp -r server ~/.local/bin/server   # so the default server is found next to the binary
```

Install the bundled server's Python deps (use a GPU-appropriate torch build):

```sh
pip install -r server/requirements.txt
```

## Usage

In your `compose.yaml`:

```yaml
services:
  ner:
    provider:
      type: gliner
      options:
        device: auto                 # cuda -> mps -> cpu
        model: fastino/gliner2-base-v1
        port: "8100"
        endpoint_var: ENDPOINT       # consumers receive NER_ENDPOINT

  app:
    image: my-app
    depends_on:
      - ner
    environment:
      - GLINER_URL=${NER_ENDPOINT}   # http://localhost:8100
```

`docker compose up` starts GLiNER on the host (MPS on a Mac), polls
`/health` until it reports ready, and injects `NER_ENDPOINT` into `app`.
`docker compose down` stops the host process. See [`examples/compose.yaml`](examples/compose.yaml)
for a runnable demo.

### Options

| Option           | Default                      | Meaning                                              |
|------------------|------------------------------|------------------------------------------------------|
| `device`         | `auto`                       | torch device: `auto`/`cuda`/`mps`/`cpu`              |
| `model`          | `fastino/gliner2-base-v1`    | HuggingFace model id for the bundled server          |
| `port`           | `8100`                       | port the server listens on                           |
| `host`           | `localhost`                  | host used to build the published endpoint            |
| `endpoint_var`   | `ENDPOINT`                   | injected as `<SERVICE>_<endpoint_var>`               |
| `health`         | `<endpoint>/health`          | health URL polled before publishing the endpoint     |
| `health_timeout` | `120`                        | seconds to wait for health before failing            |
| `stop_grace`     | `10`                         | seconds between SIGTERM and SIGKILL on `down`         |
| `command`        | bundled GLiNER server        | run an arbitrary host process instead                |
| `server`         | `serve_gliner.py` (bundled)  | path to the server script when using the default     |
| `python`         | `python3`                    | interpreter for the bundled server                   |

## How it works

Compose invokes the binary once per lifecycle event
(`docker-gliner compose --project-name P up --device=auto "ner"`). On `up` the
plugin spawns the server as a detached host process in its own process group,
records its PID under `$XDG_STATE_HOME/docker-gliner`, polls the health URL, and
emits a `setenv` message with the endpoint. On `down` it reads the PID and stops
the process (SIGTERM, then SIGKILL after `stop_grace`). `up` is idempotent: a
second call re-publishes the endpoint of the already-running server.

The wire protocol is documented in [`docs/PROTOCOL.md`](docs/PROTOCOL.md).

## Development

```sh
go test ./...                                   # Go plugin
pip install -r server/requirements.txt
pytest server/                                  # bundled server
```

## License

Apache-2.0. See [LICENSE](LICENSE).
