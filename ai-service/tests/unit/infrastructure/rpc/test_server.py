from __future__ import annotations

from types import SimpleNamespace
from unittest.mock import AsyncMock

import pytest

import grpc
from infrastructure.rpc import server as rpc_server
from infrastructure.rpc.server import TokenAuthInterceptor


@pytest.mark.asyncio
async def test_interceptor_allows_matching_bearer_token() -> None:
    interceptor = TokenAuthInterceptor("secret-token")

    expected_handler = object()
    continuation = AsyncMock(return_value=expected_handler)
    details = SimpleNamespace(
        invocation_metadata=(("authorization", "Bearer secret-token"),),
    )

    result = await interceptor.intercept_service(continuation, details)

    continuation.assert_awaited_once_with(details)
    assert result is expected_handler


@pytest.mark.asyncio
async def test_interceptor_rejects_missing_or_invalid_token() -> None:
    interceptor = TokenAuthInterceptor("secret-token")

    handler = SimpleNamespace(
        unary_unary=object(),
        request_deserializer=object(),
        response_serializer=object(),
    )
    continuation = AsyncMock(return_value=handler)
    details = SimpleNamespace(invocation_metadata=())

    result = await interceptor.intercept_service(continuation, details)

    assert result is not handler


@pytest.mark.asyncio
async def test_interceptor_unauthorized_handler_aborts_context() -> None:
    interceptor = TokenAuthInterceptor("secret-token")

    handler = SimpleNamespace(
        unary_unary=object(),
        request_deserializer=lambda raw: raw,
        response_serializer=lambda out: out,
    )
    continuation = AsyncMock(return_value=handler)
    details = SimpleNamespace(invocation_metadata=())
    wrapped_handler = await interceptor.intercept_service(continuation, details)

    context = AsyncMock()
    await wrapped_handler.unary_unary(b"request", context)

    context.abort.assert_awaited_once_with(
        grpc.StatusCode.UNAUTHENTICATED,
        "Invalid or missing token",
    )


class _FakeServer:
    def __init__(self) -> None:
        self.listen_addr: str | None = None
        self.started = False

    def add_insecure_port(self, listen_addr: str) -> None:
        self.listen_addr = listen_addr

    async def start(self) -> None:
        self.started = True


@pytest.mark.asyncio
async def test_serve_rpc_passes_ask_enabled_to_inference_service(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    fake_server = _FakeServer()
    inference_ctor_args: dict[str, object] = {}
    server_ctor_kwargs: dict[str, object] = {}

    def _fake_server_ctor(**kwargs: object) -> _FakeServer:
        server_ctor_kwargs.update(kwargs)
        return fake_server

    class _FakeInferenceService:
        def __init__(self, ask_ai_usecase: object, *, ask_enabled: bool = True) -> None:
            inference_ctor_args["ask_ai_usecase"] = ask_ai_usecase
            inference_ctor_args["ask_enabled"] = ask_enabled

    monkeypatch.setattr(rpc_server, "AIInferenceService", _FakeInferenceService)
    monkeypatch.setattr(
        rpc_server.service_pb2_grpc,
        "add_AIInferenceServiceServicer_to_server",
        lambda _svc, _server: None,
    )
    monkeypatch.setattr(
        rpc_server.conversation_grpc,
        "add_AIConversationServiceServicer_to_server",
        lambda _svc, _server: None,
    )
    monkeypatch.setattr(
        rpc_server.grpc,
        "aio",
        SimpleNamespace(server=_fake_server_ctor),
    )

    ask_ai_usecase = object()
    conversation_usecase = object()
    server = await rpc_server.serve_rpc(
        port=50051,
        ask_ai_usecase=ask_ai_usecase,
        conversation_usecase=conversation_usecase,
        ask_enabled=False,
    )

    assert server is fake_server
    assert fake_server.listen_addr == "[::]:50051"
    assert fake_server.started is True
    assert server_ctor_kwargs["interceptors"] == []
    assert inference_ctor_args["ask_ai_usecase"] is ask_ai_usecase
    assert inference_ctor_args["ask_enabled"] is False
