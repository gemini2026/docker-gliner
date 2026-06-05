"""Standalone FastAPI server for GLiNER2, bundled as the default host process
for the docker-gliner Compose provider plugin.

Usage:
    python serve_gliner.py [--port 8100] [--model fastino/gliner2-base-v1] [--device auto]

Device selection (best accelerator first):
    explicit GLINER_DEVICE env / --device flag -> cuda -> mps (Apple Silicon) -> cpu

On macOS this native host server is the only path to GPU acceleration: Docker
Desktop runs Linux containers in a VM and cannot pass the Metal/MPS GPU through,
so GLiNER inside Docker on a Mac is always CPU-only. The provider plugin launches
this server on the host so it can reach the Metal GPU, then injects its endpoint
into the dependent Compose services.

Endpoints:
    GET  /health           - Health check (reports active device)
    POST /extract          - Single extraction
    POST /extract/batch    - Batched extraction
"""

import logging
import os


# Enable CPU fallback for MPS ops that lack a Metal kernel. Must be set before
# torch is imported (gliner2 imports it lazily at model-load time), so extraction
# degrades to CPU for unsupported ops instead of hard-failing on Apple Silicon.
os.environ.setdefault("PYTORCH_ENABLE_MPS_FALLBACK", "1")

from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field
import uvicorn


app = FastAPI(title="GLiNER2 Model Server")
logger = logging.getLogger("serve_gliner")

# Global model instance (loaded once at startup)
_model = None
_model_name = os.getenv("GLINER_MODEL", "fastino/gliner2-base-v1")
_requested_device = os.getenv("GLINER_DEVICE")  # None => auto-select
_device = "cpu"


class ExtractRequest(BaseModel):
    text: str
    schema: dict  # GLiNER2 schema format: {"tool_args": ["field::type::desc", ...]}
    threshold: float | None = None


class BatchExtractRequest(BaseModel):
    items: list[ExtractRequest] = Field(max_length=100)


def resolve_device(requested: str | None = None) -> str:
    """Resolve the torch device to run GLiNER2 on.

    Priority: an explicit ``requested`` device (honored only if available) ->
    ``cuda`` -> ``mps`` (Apple Silicon) -> ``cpu``. An invalid or unavailable
    request logs a warning and falls back to ``cpu``; this never raises.
    """
    import torch

    def cuda_ok() -> bool:
        return torch.cuda.is_available()

    def mps_ok() -> bool:
        backend = getattr(torch.backends, "mps", None)
        return bool(backend) and backend.is_available() and backend.is_built()

    if requested:
        req = requested.strip().lower()
        if req == "auto":
            pass  # fall through to auto-selection below
        elif req == "cpu":
            return "cpu"
        elif req == "cuda" and cuda_ok():
            return "cuda"
        elif req == "mps" and mps_ok():
            return "mps"
        else:
            logger.warning(
                "Requested device %r is invalid or unavailable; falling back to cpu",
                requested,
            )
            return "cpu"

    if cuda_ok():
        return "cuda"
    if mps_ok():
        return "mps"
    return "cpu"


@app.on_event("startup")
def load_model():
    global _model, _device
    from gliner2 import GLiNER2

    _device = resolve_device(_requested_device)
    logger.info("Loading GLiNER2 model: %s on device=%s", _model_name, _device)
    _model = GLiNER2.from_pretrained(_model_name)
    if _device != "cpu":
        try:
            _model.to(_device)
        except Exception as exc:
            logger.warning(
                "Failed to move model to device=%s (%s); falling back to cpu",
                _device,
                exc,
            )
            _device = "cpu"
            _model.to("cpu")
    logger.info("GLiNER2 model loaded successfully on device=%s", _device)


@app.get("/health")
def health():
    if _model is None:
        return JSONResponse(
            status_code=503,
            content={"status": "loading", "model": _model_name, "device": _device},
        )
    return {"status": "ok", "model": _model_name, "device": _device}


@app.post("/extract")
def extract(req: ExtractRequest):
    if _model is None:
        raise HTTPException(503, "Model not loaded")
    result = _model.extract_json(req.text, req.schema, threshold=req.threshold)
    return result or {}


@app.post("/extract/batch")
def extract_batch(req: BatchExtractRequest):
    if _model is None:
        raise HTTPException(503, "Model not loaded")
    results = []
    for item in req.items:
        result = _model.extract_json(item.text, item.schema, threshold=item.threshold)
        results.append(result or {})
    return {"results": results}


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="GLiNER2 Model Server")
    parser.add_argument("--port", type=int, default=8100)
    parser.add_argument("--model", type=str, default=_model_name)
    parser.add_argument(
        "--device",
        type=str,
        default=_requested_device,
        help="Torch device: auto (default), cuda, mps, or cpu. "
        "Overrides the GLINER_DEVICE env var.",
    )
    args = parser.parse_args()
    _model_name = args.model
    _requested_device = args.device
    uvicorn.run(app, host="0.0.0.0", port=args.port)  # noqa: S104
