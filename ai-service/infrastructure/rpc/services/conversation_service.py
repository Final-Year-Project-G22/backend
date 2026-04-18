from __future__ import annotations

import logging
import uuid
from typing import Any

from ai.conversation.v1 import service_pb2, service_pb2_grpc

import grpc
from core.domain.models import AIChatMessage
from core.usecases.contracts import ListSessionsQuery
from core.usecases.conversation import ConversationUseCase

logger = logging.getLogger(__name__)


class AIConversationService(service_pb2_grpc.AIConversationServiceServicer):
    def __init__(self, conversation_usecase: ConversationUseCase):
        self._conversation_usecase = conversation_usecase

    async def ListConversations(  # noqa: N802
        self,
        request: Any,
        context: Any,
    ) -> Any:
        try:
            user_id = uuid.UUID(request.user_id)
            account_id = uuid.UUID(request.account_id)
            limit = max(1, request.limit)
            offset = max(0, request.offset)
        except ValueError as e:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"Invalid UUID or enum: {e}")
            return service_pb2.ListConversationsResponse()

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
        request: Any,
        context: Any,
    ) -> Any:
        try:
            session_id = uuid.UUID(request.session_id)
            account_id = uuid.UUID(request.account_id)
            message_limit = max(1, request.message_limit)
            message_offset = max(0, request.message_offset)
            include_deleted = request.include_deleted
        except ValueError as e:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"Invalid UUID: {e}")
            return service_pb2.GetConversationResponse()

        session = await self._conversation_usecase.get_session(
            session_id,
            include_deleted=include_deleted,
        )
        if session is None:
            await context.abort(grpc.StatusCode.NOT_FOUND, "Session not found")
            return service_pb2.GetConversationResponse()

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
        request: Any,
        context: Any,
    ) -> Any:
        try:
            session_id = uuid.UUID(request.session_id)
            _ = uuid.UUID(request.account_id)
        except ValueError as e:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"Invalid UUID: {e}")
            return service_pb2.ArchiveConversationResponse()

        session = await self._conversation_usecase.get_session(session_id)
        if session is None:
            await context.abort(grpc.StatusCode.NOT_FOUND, "Session not found")
            return service_pb2.ArchiveConversationResponse()

        success = await self._conversation_usecase.archive_session(session_id)

        return service_pb2.ArchiveConversationResponse(success=success)

    def _map_message_to_detail(self, message: AIChatMessage) -> Any:
        content = message.user_query or message.llm_response or ""

        citations = []
        if message.retrieved_chunk_ids:
            chunk_ids = message.retrieved_chunk_ids
            citations.extend(
                service_pb2.Citation(
                    chunk_id=str(cid),
                    document_id="",
                    source_type="chunk",
                )
                for cid in chunk_ids
            )

        usage = None
        if message.token_usage:
            usage = service_pb2.Usage(
                prompt_tokens=message.token_usage.prompt_tokens,
                completion_tokens=message.token_usage.completion_tokens,
                total_tokens=message.token_usage.total_tokens,
            )

        role = "user" if message.message_type.value == "user_query" else "assistant"

        return service_pb2.MessageDetail(
            id=str(message.id),
            role=role,
            content=content,
            citations=citations,
            usage=usage,
            created_at=message.created_at.isoformat(),
        )
