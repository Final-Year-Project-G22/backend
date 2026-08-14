from __future__ import annotations

import logging
import uuid

import grpc
import grpc.aio
from ai.conversation.v1 import service_pb2, service_pb2_grpc

from core.domain.enums import MessageType
from core.domain.models import AIChatMessage
from core.usecases.contracts import ListSessionsQuery
from core.usecases.conversation import ConversationUseCase

logger = logging.getLogger(__name__)

_ListContext = grpc.aio.ServicerContext[
    service_pb2.ListConversationsRequest, service_pb2.ListConversationsResponse
]
_GetContext = grpc.aio.ServicerContext[
    service_pb2.GetConversationRequest, service_pb2.GetConversationResponse
]
_ArchiveContext = grpc.aio.ServicerContext[
    service_pb2.ArchiveConversationRequest, service_pb2.ArchiveConversationResponse
]


class AIConversationService(service_pb2_grpc.AIConversationServiceServicer):
    def __init__(self, conversation_usecase: ConversationUseCase) -> None:
        self._conversation_usecase = conversation_usecase

    async def ListConversations(  # noqa: N802
        self,
        request: service_pb2.ListConversationsRequest,
        context: _ListContext,
    ) -> service_pb2.ListConversationsResponse:
        try:
            user_id = uuid.UUID(request.user_id)
            account_id = uuid.UUID(request.account_id)
            limit = max(1, request.limit)
            offset = max(0, request.offset)
        except ValueError as e:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"Invalid UUID or enum: {e}")

        query = ListSessionsQuery(
            user_id=user_id,
            limit=limit,
            offset=offset,
            include_deleted=False,
        )

        sessions = await self._conversation_usecase.list_sessions(query)

        session_summaries = [
            service_pb2.SessionSummary(
                id=str(s.id),
                account_id=str(account_id),
                title=s.title,
                language=s.language.value,
                created_at=s.created_at.isoformat(),
                updated_at=s.updated_at.isoformat(),
            )
            for s in sessions
        ]

        return service_pb2.ListConversationsResponse(
            sessions=session_summaries,
            total=len(session_summaries),
        )

    async def GetConversation(  # noqa: N802
        self,
        request: service_pb2.GetConversationRequest,
        context: _GetContext,
    ) -> service_pb2.GetConversationResponse:
        try:
            session_id = uuid.UUID(request.session_id)
            account_id = uuid.UUID(request.account_id)
            message_limit = max(1, request.message_limit)
            message_offset = max(0, request.message_offset)
            include_deleted = request.include_deleted
        except ValueError as e:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"Invalid UUID: {e}")

        session = await self._conversation_usecase.get_session(
            session_id,
            include_deleted=include_deleted,
        )
        if session is None:
            await context.abort(grpc.StatusCode.NOT_FOUND, "Session not found")

        messages = await self._conversation_usecase.list_messages(
            session_id,
            limit=message_limit,
            offset=message_offset,
        )

        session_detail = service_pb2.SessionDetail(
            id=str(session.id),
            account_id=str(account_id),
            title=session.title,
            language=session.language.value,
            status=session.status.value,
            created_at=session.created_at.isoformat(),
            updated_at=session.updated_at.isoformat(),
        )

        message_details = [self._map_message_to_detail(m) for m in messages]

        return service_pb2.GetConversationResponse(
            session=session_detail,
            messages=message_details,
            total_messages=session.message_count,
        )

    async def ArchiveConversation(  # noqa: N802
        self,
        request: service_pb2.ArchiveConversationRequest,
        context: _ArchiveContext,
    ) -> service_pb2.ArchiveConversationResponse:
        try:
            session_id = uuid.UUID(request.session_id)
            _ = uuid.UUID(request.account_id)
        except ValueError as e:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"Invalid UUID: {e}")

        session = await self._conversation_usecase.get_session(session_id)
        if session is None:
            await context.abort(grpc.StatusCode.NOT_FOUND, "Session not found")

        success = await self._conversation_usecase.archive_session(session_id)

        refreshed_session = await self._conversation_usecase.get_session(
            session_id, include_deleted=True
        )
        updated_at = ""
        if refreshed_session is not None:
            updated_at = refreshed_session.updated_at.isoformat()

        return service_pb2.ArchiveConversationResponse(success=success, updated_at=updated_at)

    def _map_message_to_detail(self, message: AIChatMessage) -> service_pb2.MessageDetail:
        content = message.user_query or message.llm_response or ""

        citations: list[service_pb2.Citation] = [
            service_pb2.Citation(
                chunk_id=str(src.chunk_id) if src.chunk_id else "",
                document_id=str(src.document_id),
                source_type=src.source.value,
                title=src.title,
                score=src.score or 0.0,
                excerpt=src.excerpt,
            )
            for src in message.response_sources
        ]

        usage: service_pb2.Usage | None = None
        if message.token_usage:
            usage = service_pb2.Usage(
                prompt_tokens=message.token_usage.prompt_tokens,
                completion_tokens=message.token_usage.completion_tokens,
                total_tokens=message.token_usage.total_tokens,
            )

        role = "user" if message.message_type is MessageType.USER_QUERY else "assistant"

        return service_pb2.MessageDetail(
            id=str(message.id),
            role=role,
            content=content,
            citations=citations,
            usage=usage,
            created_at=message.created_at.isoformat(),
        )
