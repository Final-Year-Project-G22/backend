from __future__ import annotations

import uuid
from abc import ABC, abstractmethod
from datetime import date, datetime

from core.domain.models import AIUserQuota


class QuotaRepositoryPort(ABC):
    @abstractmethod
    async def get_quota(self, user_id: uuid.UUID) -> AIUserQuota | None: ...

    @abstractmethod
    async def upsert_quota(self, quota: AIUserQuota) -> AIUserQuota: ...

    @abstractmethod
    async def increment_usage(
        self,
        user_id: uuid.UUID,
        *,
        query_count: int,
        token_count: int,
        conversations_started: int = 0,
        at: datetime | None = None,
    ) -> AIUserQuota: ...

    @abstractmethod
    async def reset_daily_usage(self, user_id: uuid.UUID, *, on_date: date) -> AIUserQuota: ...


__all__ = ["QuotaRepositoryPort"]
