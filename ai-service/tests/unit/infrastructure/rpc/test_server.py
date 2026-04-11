from __future__ import annotations

from types import SimpleNamespace
from unittest.mock import AsyncMock

import grpc
import pytest

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
