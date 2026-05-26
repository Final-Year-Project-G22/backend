from __future__ import annotations

import uuid
from abc import ABC, abstractmethod
from datetime import datetime

from core.domain.models import AIChatMessage, AIConversationSession


class ConversationRepositoryPort(ABC):
    @abstractmethod
    async def create_session(self, session: AIConversationSession) -> AIConversationSession: ...

    @abstractmethod
    async def update_session(self, session: AIConversationSession) -> AIConversationSession: ...

    @abstractmethod
    async def get_session(self, session_id: uuid.UUID) -> AIConversationSession | None: ...

    @abstractmethod
    async def list_sessions_by_user(
        self,
        user_id: uuid.UUID,
        *,
        limit: int = 20,
        offset: int = 0,
        include_deleted: bool = False,
    ) -> list[AIConversationSession]: ...

    @abstractmethod
    async def soft_delete_session(self, session_id: uuid.UUID, *, deleted_at: datetime) -> bool: ...

    @abstractmethod
    async def update_session_title(self, session_id: uuid.UUID, title: str) -> None: ...

    @abstractmethod
    async def add_message(self, message: AIChatMessage) -> AIChatMessage: ...

    @abstractmethod
    async def update_message(self, message: AIChatMessage) -> AIChatMessage: ...

    @abstractmethod
    async def get_message(self, message_id: uuid.UUID) -> AIChatMessage | None: ...

    @abstractmethod
    async def list_messages(
        self,
        conversation_id: uuid.UUID,
        *,
        limit: int = 100,
        offset: int = 0,
    ) -> list[AIChatMessage]: ...


__all__ = ["ConversationRepositoryPort"]
