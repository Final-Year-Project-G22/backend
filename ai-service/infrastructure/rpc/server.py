from __future__ import annotations

import logging
from collections.abc import Awaitable, Callable
from typing import TYPE_CHECKING, Any, TypeAlias, cast

import grpc
import grpc.aio
from ai.conversation.v1 import service_pb2_grpc as conversation_grpc
from ai.inference.v1 import service_pb2_grpc

from core.usecases.ask_ai import AskAIUseCase
from core.usecases.conversation import ConversationUseCase
from infrastructure.rpc.services.conversation_service import AIConversationService
from infrastructure.rpc.services.inference_service import AIInferenceService

logger = logging.getLogger(__name__)

if TYPE_CHECKING:
    # The grpc-stubs .pyi declares these as Generic, but the runtime classes
    # are not subscriptable, so parameterization is typechecking-only.
    _Handler = grpc.RpcMethodHandler[Any, Any]
    _ServerInterceptorBase = grpc.aio.ServerInterceptor[Any, Any]
    _ServerCtor = Callable[..., grpc.aio.Server]
else:
    _Handler = grpc.RpcMethodHandler
    _ServerInterceptorBase = grpc.aio.ServerInterceptor
    _ServerCtor = Callable[..., Any]

_HandlerContinuation: TypeAlias = "Callable[[grpc.HandlerCallDetails], Awaitable[_Handler | None]]"


def _aio_server_ctor() -> _ServerCtor:
    # grpc.aio.server carries unbound generic types in its stub signature,
    # which makes the symbol "partially unknown"; getattr + cast keeps the
    # call site strictly typed instead of cascading Any.
    return cast(_ServerCtor, getattr(grpc.aio, "server"))  # noqa: B009


class TokenAuthInterceptor(_ServerInterceptorBase):
    def __init__(self, token: str) -> None:
        self._token = token

    async def intercept_service(  # type: ignore[override]  # stub omits Optional return; grpc allows None
        self,
        continuation: _HandlerContinuation,
        handler_call_details: grpc.HandlerCallDetails,
    ) -> _Handler | None:
        metadata = dict(handler_call_details.invocation_metadata)
        auth_header = metadata.get("authorization", "")

        expected_header = f"Bearer {self._token}"
        if auth_header != expected_header:
            return await _build_unauthorized_handler(continuation, handler_call_details)

        return await continuation(handler_call_details)


async def _build_unauthorized_handler(
    continuation: _HandlerContinuation,
    handler_call_details: grpc.HandlerCallDetails,
) -> _Handler | None:
    existing_handler = await continuation(handler_call_details)

    if existing_handler is None:
        return None

    if existing_handler.unary_unary is None:
        return existing_handler

    unary_handler = cast(
        Callable[..., _Handler],
        getattr(grpc, "unary_unary_rpc_method_handler"),  # noqa: B009
    )
    # grpc-stubs declares the serializer fields as bare Callable, so direct
    # attribute access would be "partially unknown"; getattr keeps the call
    # site strictly typed.
    request_deserializer = getattr(existing_handler, "request_deserializer")  # noqa: B009
    response_serializer = getattr(existing_handler, "response_serializer")  # noqa: B009
    return unary_handler(
        _unauthorized_unary_unary,
        request_deserializer=request_deserializer,
        response_serializer=response_serializer,
    )


async def _unauthorized_unary_unary(request: Any, context: Any) -> None:
    _ = request
    await context.abort(grpc.StatusCode.UNAUTHENTICATED, "Invalid or missing token")


async def serve_rpc(
    port: int,
    ask_ai_usecase: AskAIUseCase,
    conversation_usecase: ConversationUseCase,
    *,
    auth_token: str | None = None,
    ask_enabled: bool = True,
) -> grpc.aio.Server:
    interceptors: list[grpc.aio.ServerInterceptor[Any, Any]] = []
    if auth_token:
        interceptors.append(TokenAuthInterceptor(auth_token))

    server = _aio_server_ctor()(interceptors=interceptors)

    inference_service = AIInferenceService(ask_ai_usecase, ask_enabled=ask_enabled)
    service_pb2_grpc.add_AIInferenceServiceServicer_to_server(inference_service, server)

    conversation_service = AIConversationService(conversation_usecase)
    conversation_grpc.add_AIConversationServiceServicer_to_server(conversation_service, server)

    listen_addr = f"[::]:{port}"
    server.add_insecure_port(listen_addr)
    logger.info("Starting gRPC server on %s", listen_addr)
    await server.start()
    return server
