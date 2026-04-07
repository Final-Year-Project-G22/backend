from __future__ import annotations

import uuid
from datetime import date, datetime

from sqlalchemy.ext.asyncio import AsyncSession

from core.domain.models import AIUserQuota
from core.ports.quota_repository import QuotaRepositoryPort


class SqlAlchemyQuotaRepository(QuotaRepositoryPort):
    def __init__(self, session: AsyncSession) -> None:
        self._session = session

    async def get_quota(self, user_id: uuid.UUID) -> AIUserQuota | None:
        raise NotImplementedError

    async def upsert_quota(self, quota: AIUserQuota) -> AIUserQuota:
        raise NotImplementedError

    async def increment_usage(
        self,
        user_id: uuid.UUID,
        *,
        query_count: int,
        token_count: int,
        conversations_started: int = 0,
        at: datetime | None = None,
    ) -> AIUserQuota:
        raise NotImplementedError

    async def reset_daily_usage(self, user_id: uuid.UUID, *, on_date: date) -> AIUserQuota:
        raise NotImplementedError


__all__ = ["SqlAlchemyQuotaRepository"]
