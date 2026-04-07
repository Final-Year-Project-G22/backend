from __future__ import annotations

import uuid
from datetime import datetime

from sqlalchemy.ext.asyncio import AsyncSession

from core.domain.models import AIChatMessage, AIConversationSession
from core.ports.conversation_repository import ConversationRepositoryPort


class SqlAlchemyConversationRepository(ConversationRepositoryPort):
    def __init__(self, session: AsyncSession) -> None:
        self._session = session

    async def create_session(self, session: AIConversationSession) -> AIConversationSession:
        raise NotImplementedError

    async def update_session(self, session: AIConversationSession) -> AIConversationSession:
        raise NotImplementedError

    async def get_session(self, session_id: uuid.UUID) -> AIConversationSession | None:
        raise NotImplementedError

    async def list_sessions_by_user(
        self,
        user_id: uuid.UUID,
        *,
        limit: int = 20,
        offset: int = 0,
        include_deleted: bool = False,
    ) -> list[AIConversationSession]:
        raise NotImplementedError

    async def soft_delete_session(self, session_id: uuid.UUID, *, deleted_at: datetime) -> bool:
        raise NotImplementedError

    async def add_message(self, message: AIChatMessage) -> AIChatMessage:
        raise NotImplementedError

    async def get_message(self, message_id: uuid.UUID) -> AIChatMessage | None:
        raise NotImplementedError

    async def list_messages(
        self,
        conversation_id: uuid.UUID,
        *,
        limit: int = 100,
        offset: int = 0,
    ) -> list[AIChatMessage]:
        raise NotImplementedError


__all__ = ["SqlAlchemyConversationRepository"]
