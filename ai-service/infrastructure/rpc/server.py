from __future__ import annotations

import logging
from typing import Any

from ai.conversation.v1 import service_pb2_grpc as conversation_grpc  # type: ignore
from ai.inference.v1 import service_pb2_grpc  # type: ignore

import grpc
from core.usecases.ask_ai import AskAIUseCase
from core.usecases.conversation import ConversationUseCase
from infrastructure.rpc.services.conversation_service import AIConversationService
from infrastructure.rpc.services.inference_service import AIInferenceService

logger = logging.getLogger(__name__)


class TokenAuthInterceptor:
    def __init__(self, token: str):
        self._token = token

    async def intercept_service(
        self,
        continuation: Any,
        handler_call_details: Any,
    ) -> Any:
        metadata = dict(handler_call_details.invocation_metadata)
        auth_header = metadata.get("authorization", "")

        expected_header = f"Bearer {self._token}"
        if auth_header != expected_header:
            return await _build_unauthorized_handler(continuation, handler_call_details)

        return await continuation(handler_call_details)


async def _build_unauthorized_handler(continuation: Any, handler_call_details: Any) -> Any:
    existing_handler = await continuation(handler_call_details)

    if existing_handler is None:
        return None

    if existing_handler.unary_unary is None:
        return existing_handler

    unary_handler = getattr(grpc, "unary_unary_rpc_method_handler")  # noqa: B009
    return unary_handler(
        _unauthorized_unary_unary,
        request_deserializer=existing_handler.request_deserializer,
        response_serializer=existing_handler.response_serializer,
    )


async def _unauthorized_unary_unary(request: Any, context: Any) -> Any:
    _ = request
    await context.abort(grpc.StatusCode.UNAUTHENTICATED, "Invalid or missing token")  # type: ignore


async def serve_rpc(
    port: int,
    ask_ai_usecase: AskAIUseCase,
    conversation_usecase: ConversationUseCase,
    auth_token: str | None = None,
) -> Any:
    interceptors: list[Any] = []
    if auth_token:
        interceptors.append(TokenAuthInterceptor(auth_token))

    grpc_aio: Any = getattr(grpc, "aio")  # noqa: B009
    server_ctor: Any = getattr(grpc_aio, "server")  # noqa: B009
    server = server_ctor(interceptors=interceptors)

    inference_service = AIInferenceService(ask_ai_usecase)
    service_pb2_grpc.add_AIInferenceServiceServicer_to_server(inference_service, server)  # type: ignore

    conversation_service = AIConversationService(conversation_usecase)
    conversation_grpc.add_AIConversationServiceServicer_to_server(conversation_service, server)  # type: ignore

    listen_addr = f"[::]:{port}"
    server.add_insecure_port(listen_addr)
    logger.info("Starting gRPC server on %s", listen_addr)
    await server.start()
    return server
