from __future__ import annotations

import uuid
from datetime import UTC, datetime

from core.domain.enums import Tier
from core.domain.models import AIChatMessage, AIConversationSession
from core.ports.conversation_repository import ConversationRepositoryPort
from core.usecases.contracts import CreateSessionCommand, ListSessionsQuery
from core.usecases.quota_guard import QuotaGuardUseCase


def _utc_now() -> datetime:
    return datetime.now(UTC)


class ConversationUseCase:
    def __init__(
        self,
        conversation_repository: ConversationRepositoryPort,
        *,
        quota_guard: QuotaGuardUseCase | None = None,
    ) -> None:
        self._conversation_repository = conversation_repository
        self._quota_guard = quota_guard

    async def create_session(
        self,
        command: CreateSessionCommand,
        *,
        at: datetime | None = None,
    ) -> AIConversationSession:
        now = at or _utc_now()
        tier = await self._resolve_tier(command.user_id)
        session = AIConversationSession(
            user_id=command.user_id,
            title=command.title,
            language=command.language,
            tier_at_start=tier,
            current_tier=tier,
            created_at=now,
            updated_at=now,
        )
        return await self._conversation_repository.create_session(session)

    async def get_session(
        self,
        session_id: uuid.UUID,
        *,
        include_deleted: bool = False,
    ) -> AIConversationSession | None:
        session = await self._conversation_repository.get_session(session_id)
        if session is None:
            return None
        if not include_deleted and session.deleted_at is not None:
            return None
        return session

    async def list_sessions(self, query: ListSessionsQuery) -> list[AIConversationSession]:
        return await self._conversation_repository.list_sessions_by_user(
            query.user_id,
            limit=query.limit,
            offset=query.offset,
            include_deleted=query.include_deleted,
        )

    async def archive_session(
        self,
        session_id: uuid.UUID,
        *,
        deleted_at: datetime | None = None,
    ) -> bool:
        return await self._conversation_repository.soft_delete_session(
            session_id,
            deleted_at=deleted_at or _utc_now(),
        )

    async def get_message(self, message_id: uuid.UUID) -> AIChatMessage | None:
        return await self._conversation_repository.get_message(message_id)

    async def list_messages(
        self,
        conversation_id: uuid.UUID,
        *,
        limit: int = 100,
        offset: int = 0,
    ) -> list[AIChatMessage]:
        return await self._conversation_repository.list_messages(
            conversation_id,
            limit=limit,
            offset=offset,
        )

    async def update_session_title(
        self,
        session_id: uuid.UUID,
        title: str,
    ) -> None:
        await self._conversation_repository.update_session_title(session_id, title)

    async def _resolve_tier(self, user_id: uuid.UUID) -> Tier:
        if self._quota_guard is None:
            return Tier.BASIC
        quota = await self._quota_guard.get_quota(user_id)
        return quota.tier

    @property
    def conversation_repository(self) -> ConversationRepositoryPort:
        return self._conversation_repository


__all__ = ["ConversationUseCase"]
