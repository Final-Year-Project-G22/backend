from __future__ import annotations

import uuid
from datetime import UTC, date, datetime

from core.domain.exceptions import QuotaExceededError
from core.domain.models import AIUserQuota
from core.domain.value_objects import UsageSnapshot
from core.ports.core_service import CoreServicePort
from core.ports.quota_repository import QuotaRepositoryPort


def _utc_now() -> datetime:
    return datetime.now(UTC)


class QuotaGuardUseCase:
    def __init__(
        self,
        quota_repository: QuotaRepositoryPort,
        *,
        core_service: CoreServicePort | None = None,
    ) -> None:
        self._quota_repository = quota_repository
        self._core_service = core_service

    async def get_quota(self, user_id: uuid.UUID) -> AIUserQuota:
        quota = await self._quota_repository.get_quota(user_id)
        if quota is None:
            quota = await self._quota_repository.upsert_quota(AIUserQuota(user_id=user_id))

        return await self._sync_tier(quota)

    async def enforce_limits(
        self,
        user_id: uuid.UUID,
        *,
        query_count: int,
        token_count: int,
        conversations_started: int = 0,
        at: datetime | None = None,
    ) -> AIUserQuota:
        at_dt = at or _utc_now()
        quota = await self.get_quota(user_id)

        daily_query_count, daily_token_count, daily_conversations_started = self._daily_counters(
            quota,
            on_date=at_dt.date(),
        )

        if daily_query_count + query_count > quota.daily_query_limit:
            raise QuotaExceededError(
                "daily query limit exceeded",
                details={
                    "user_id": str(user_id),
                    "current": daily_query_count,
                    "attempted": query_count,
                    "limit": quota.daily_query_limit,
                },
            )

        if daily_token_count + token_count > quota.daily_token_limit:
            raise QuotaExceededError(
                "daily token limit exceeded",
                details={
                    "user_id": str(user_id),
                    "current": daily_token_count,
                    "attempted": token_count,
                    "limit": quota.daily_token_limit,
                },
            )

        if daily_conversations_started + conversations_started > quota.max_conversations_per_day:
            raise QuotaExceededError(
                "daily conversation limit exceeded",
                details={
                    "user_id": str(user_id),
                    "current": daily_conversations_started,
                    "attempted": conversations_started,
                    "limit": quota.max_conversations_per_day,
                },
            )

        return quota

    async def consume_usage(
        self,
        user_id: uuid.UUID,
        *,
        query_count: int,
        token_count: int,
        conversations_started: int = 0,
        at: datetime | None = None,
    ) -> AIUserQuota:
        at_dt = at or _utc_now()
        await self.enforce_limits(
            user_id,
            query_count=query_count,
            token_count=token_count,
            conversations_started=conversations_started,
            at=at_dt,
        )

        return await self._quota_repository.increment_usage(
            user_id,
            query_count=query_count,
            token_count=token_count,
            conversations_started=conversations_started,
            at=at_dt,
        )

    async def get_usage_snapshot(
        self,
        user_id: uuid.UUID,
        *,
        on_date: date | None = None,
    ) -> UsageSnapshot:
        quota = await self.get_quota(user_id)

        effective_date = on_date or _utc_now().date()
        daily_query_count, daily_token_count, daily_conversations_started = self._daily_counters(
            quota,
            on_date=effective_date,
        )

        return UsageSnapshot(
            tier=quota.tier,
            daily_query_count=daily_query_count,
            daily_token_count=daily_token_count,
            daily_conversations_started=daily_conversations_started,
            daily_query_limit=quota.daily_query_limit,
            daily_token_limit=quota.daily_token_limit,
            max_conversations_per_day=quota.max_conversations_per_day,
            total_queries_used=quota.total_queries_used,
            total_tokens_used=quota.total_tokens_used,
            total_conversations=quota.total_conversations,
        )

    async def _sync_tier(self, quota: AIUserQuota) -> AIUserQuota:
        if self._core_service is None:
            return quota

        latest_tier = await self._core_service.get_user_tier(quota.user_id)
        if latest_tier is None or latest_tier == quota.tier:
            return quota

        now = _utc_now()
        updated_quota = quota.model_copy(
            update={
                "tier": latest_tier,
                "tier_updated_at": now,
                "updated_at": now,
            }
        )
        return await self._quota_repository.upsert_quota(updated_quota)

    @staticmethod
    def _daily_counters(
        quota: AIUserQuota,
        *,
        on_date: date,
    ) -> tuple[int, int, int]:
        if quota.last_query_date != on_date:
            return (0, 0, 0)
        return (
            quota.daily_query_count,
            quota.daily_token_count,
            quota.daily_conversations_started,
        )


__all__ = ["QuotaGuardUseCase"]
