from __future__ import annotations

import uuid
from datetime import datetime

from sqlalchemy import select
from sqlalchemy.exc import SQLAlchemyError
from sqlalchemy.ext.asyncio import AsyncSession

from core.domain.enums import SessionStatus
from core.domain.exceptions import RepositoryError
from core.domain.models import AIChatMessage, AIConversationSession
from core.ports.conversation_repository import ConversationRepositoryPort
from infrastructure.database import models_sqlalchemy as sa_models
from infrastructure.database.repositories.mappers import (
    serialize_response_sources,
    serialize_search_hits,
    serialize_token_usage,
    to_domain_message,
    to_domain_session,
    to_orm_message,
    to_orm_session,
)


class SqlAlchemyConversationRepository(ConversationRepositoryPort):
    def __init__(self, session: AsyncSession) -> None:
        self._session = session

    @property
    def session(self) -> AsyncSession:
        return self._session

    async def create_session(self, session: AIConversationSession) -> AIConversationSession:
        model = to_orm_session(session)
        try:
            self._session.add(model)
            await self._session.flush()
            await self._session.commit()
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to create conversation session",
                details={"session_id": str(session.id)},
            ) from exc
        return to_domain_session(model)

    async def update_session(self, session: AIConversationSession) -> AIConversationSession:
        try:
            model = await self._get_session_model(session.id)
            if model is None:
                raise RepositoryError(
                    "conversation session not found",
                    details={"session_id": str(session.id)},
                )

            self._apply_session_domain_values(model, session)
            await self._session.flush()
            await self._session.commit()
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to update conversation session",
                details={"session_id": str(session.id)},
            ) from exc

        return to_domain_session(model)

    async def get_session(self, session_id: uuid.UUID) -> AIConversationSession | None:
        try:
            model = await self._get_session_model(session_id)
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to fetch conversation session",
                details={"session_id": str(session_id)},
            ) from exc

        if model is None:
            return None
        return to_domain_session(model)

    async def list_sessions_by_user(
        self,
        user_id: uuid.UUID,
        *,
        limit: int = 20,
        offset: int = 0,
        include_deleted: bool = False,
    ) -> list[AIConversationSession]:
        statement = select(sa_models.AIConversationSession).where(
            sa_models.AIConversationSession.user_id == user_id
        )
        if not include_deleted:
            statement = statement.where(sa_models.AIConversationSession.deleted_at.is_(None))

        statement = statement.order_by(
            sa_models.AIConversationSession.created_at.desc(),
            sa_models.AIConversationSession.id.desc(),
        ).offset(offset)
        if limit > 0:
            statement = statement.limit(limit)

        try:
            result = await self._session.execute(statement)
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to list conversation sessions",
                details={"user_id": str(user_id)},
            ) from exc

        models = result.scalars().all()
        return [to_domain_session(model) for model in models]

    async def soft_delete_session(self, session_id: uuid.UUID, *, deleted_at: datetime) -> bool:
        try:
            model = await self._get_session_model(session_id)
            if model is None:
                return False

            model.status = SessionStatus.ARCHIVED.value
            model.deleted_at = deleted_at
            model.updated_at = deleted_at
            await self._session.flush()
            await self._session.commit()
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to soft-delete conversation session",
                details={"session_id": str(session_id)},
            ) from exc

        return True

    async def add_message(self, message: AIChatMessage) -> AIChatMessage:
        model = to_orm_message(message)
        try:
            self._session.add(model)
            await self._session.flush()
            await self._session.commit()
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to add chat message",
                details={"message_id": str(message.id)},
            ) from exc
        return to_domain_message(model)

    async def update_message(self, message: AIChatMessage) -> AIChatMessage:
        try:
            model = await self._get_message_model(message.id)
            if model is None:
                raise RepositoryError(
                    "chat message not found",
                    details={"message_id": str(message.id)},
                )
            self._apply_message_domain_values(model, message)
            await self._session.flush()
            await self._session.commit()
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to update chat message",
                details={
                    "message_id": str(message.id),
                    "conversation_id": str(message.conversation_id),
                },
            ) from exc

        await self._session.refresh(model)
        return to_domain_message(model)

    async def get_message(self, message_id: uuid.UUID) -> AIChatMessage | None:
        try:
            model = await self._get_message_model(message_id)
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to fetch chat message",
                details={"message_id": str(message_id)},
            ) from exc

        if model is None:
            return None
        return to_domain_message(model)

    async def list_messages(
        self,
        conversation_id: uuid.UUID,
        *,
        limit: int = 100,
        offset: int = 0,
    ) -> list[AIChatMessage]:
        statement = (
            select(sa_models.AIChatMessage)
            .where(sa_models.AIChatMessage.conversation_id == conversation_id)
            .order_by(
                sa_models.AIChatMessage.message_order.asc(),
                sa_models.AIChatMessage.created_at.asc(),
                sa_models.AIChatMessage.id.asc(),
            )
            .offset(offset)
        )
        if limit > 0:
            statement = statement.limit(limit)

        try:
            result = await self._session.execute(statement)
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to list chat messages",
                details={"conversation_id": str(conversation_id)},
            ) from exc

        models = result.scalars().all()
        return [to_domain_message(model) for model in models]

    async def _get_session_model(
        self,
        session_id: uuid.UUID,
    ) -> sa_models.AIConversationSession | None:
        result = await self._session.execute(
            select(sa_models.AIConversationSession).where(
                sa_models.AIConversationSession.id == session_id
            )
        )
        return result.scalar_one_or_none()

    async def _get_message_model(self, message_id: uuid.UUID) -> sa_models.AIChatMessage | None:
        result = await self._session.execute(
            select(sa_models.AIChatMessage).where(sa_models.AIChatMessage.id == message_id)
        )
        return result.scalar_one_or_none()

    def _apply_session_domain_values(
        self,
        model: sa_models.AIConversationSession,
        session: AIConversationSession,
    ) -> None:
        model.user_id = session.user_id
        model.title = session.title
        model.language = session.language.value
        model.tier_at_start = session.tier_at_start.value
        model.current_tier = session.current_tier.value
        model.status = session.status.value
        model.context_summary = session.context_summary
        model.last_message_preview = session.last_message_preview
        model.message_count = session.message_count
        model.total_tokens_used = session.total_tokens_used
        model.last_message_at = session.last_message_at
        model.created_at = session.created_at
        model.updated_at = session.updated_at
        model.deleted_at = session.deleted_at

    def _apply_message_domain_values(
        self,
        model: sa_models.AIChatMessage,
        message: AIChatMessage,
    ) -> None:
        model.user_id = message.user_id
        model.conversation_id = message.conversation_id
        model.message_type = message.message_type.value
        model.user_query = message.user_query
        model.query_language = message.query_language.value
        model.query_embedding = message.query_embedding
        model.retrieved_chunk_ids = message.retrieved_chunk_ids
        model.context_chunks = serialize_search_hits(message.context_chunks)
        model.llm_response = message.llm_response
        model.response_sources = serialize_response_sources(message.response_sources)
        model.processing_time_ms = message.processing_time_ms
        model.token_usage = serialize_token_usage(message.token_usage)
        model.model_used = message.model_used
        model.prompt_version = message.prompt_version
        model.trace_id = message.trace_id
        model.cache_hit = message.cache_hit
        model.user_feedback = message.user_feedback
        model.feedback_at = message.feedback_at
        model.is_context_cleared = message.is_context_cleared
        model.message_order = message.message_order
        model.created_at = message.created_at
        model.updated_at = message.updated_at


__all__ = ["SqlAlchemyConversationRepository"]
