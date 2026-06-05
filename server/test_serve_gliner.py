"""Tests for the bundled GLiNER server.

torch and gliner2 are heavy and device-dependent, so we inject fakes into
sys.modules before importing the server. That lets us exercise device
resolution and the /health contract on any machine, including CPU-only CI.
"""

import importlib
import sys
import types

import pytest


def _fake_torch(*, cuda: bool, mps: bool) -> types.ModuleType:
    """Build a stand-in torch module advertising the given accelerators."""
    mod = types.ModuleType("torch")

    cuda_ns = types.SimpleNamespace(is_available=lambda: cuda)
    mps_ns = types.SimpleNamespace(is_available=lambda: mps, is_built=lambda: mps)
    backends = types.SimpleNamespace(mps=mps_ns)

    mod.cuda = cuda_ns
    mod.backends = backends
    return mod


@pytest.fixture
def server(monkeypatch):
    """Import serve_gliner with a CPU-only fake torch; reload to reset globals."""
    monkeypatch.setitem(sys.modules, "torch", _fake_torch(cuda=False, mps=False))
    mod = importlib.import_module("serve_gliner")
    importlib.reload(mod)
    return mod


def test_resolve_device_auto_prefers_cuda(server, monkeypatch):
    monkeypatch.setitem(sys.modules, "torch", _fake_torch(cuda=True, mps=True))
    assert server.resolve_device(None) == "cuda"
    assert server.resolve_device("auto") == "cuda"


def test_resolve_device_auto_prefers_mps_when_no_cuda(server, monkeypatch):
    monkeypatch.setitem(sys.modules, "torch", _fake_torch(cuda=False, mps=True))
    assert server.resolve_device(None) == "mps"


def test_resolve_device_falls_back_to_cpu(server, monkeypatch):
    monkeypatch.setitem(sys.modules, "torch", _fake_torch(cuda=False, mps=False))
    assert server.resolve_device(None) == "cpu"


def test_resolve_device_explicit_cpu(server):
    assert server.resolve_device("cpu") == "cpu"


def test_resolve_device_unavailable_request_falls_back(server, monkeypatch):
    # Requesting mps on a machine without it must not raise; it degrades to cpu.
    monkeypatch.setitem(sys.modules, "torch", _fake_torch(cuda=False, mps=False))
    assert server.resolve_device("mps") == "cpu"
    assert server.resolve_device("cuda") == "cpu"


def test_resolve_device_honors_available_request(server, monkeypatch):
    monkeypatch.setitem(sys.modules, "torch", _fake_torch(cuda=True, mps=True))
    assert server.resolve_device("mps") == "mps"
    assert server.resolve_device("cuda") == "cuda"


def test_health_reports_loading_before_startup(server):
    # With no model loaded yet, the handler returns a 503 "loading" payload.
    resp = server.health()
    assert resp.status_code == 503
    import json

    body = json.loads(resp.body)
    assert body["status"] == "loading"


def test_health_ok_after_startup(server, monkeypatch):
    from fastapi.testclient import TestClient

    # Stub gliner2 so startup can "load" a model without the real dependency.
    class FakeModel:
        def to(self, device):
            return self

    fake_gliner2 = types.ModuleType("gliner2")
    fake_gliner2.GLiNER2 = types.SimpleNamespace(from_pretrained=lambda name: FakeModel())
    monkeypatch.setitem(sys.modules, "gliner2", fake_gliner2)
    monkeypatch.setitem(sys.modules, "torch", _fake_torch(cuda=False, mps=False))

    with TestClient(server.app) as client:
        resp = client.get("/health")
        assert resp.status_code == 200
        body = resp.json()
        assert body["status"] == "ok"
        assert body["device"] == "cpu"
        assert "model" in body
