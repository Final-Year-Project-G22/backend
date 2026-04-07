from __future__ import annotations

import uuid
from datetime import UTC, date, datetime
from unittest.mock import AsyncMock

import pytest

from core.domain.enums import Tier
from core.domain.exceptions import QuotaExceededError
from core.domain.models import AIUserQuota
from core.usecases.quota_guard import QuotaGuardUseCase


def _build_quota(*, user_id: uuid.UUID) -> AIUserQuota:
    now = datetime(2026, 4, 7, 9, 0, tzinfo=UTC)
    return AIUserQuota(
        user_id=user_id,
        tier=Tier.BASIC,
        daily_query_count=2,
        daily_token_count=100,
        daily_conversations_started=1,
        daily_query_limit=10,
        daily_token_limit=500,
        max_conversations_per_day=3,
        total_queries_used=20,
        total_tokens_used=2000,
        total_conversations=7,
        last_query_date=date(2026, 4, 7),
        last_query_at=now,
        created_at=now,
        updated_at=now,
    )


@pytest.mark.asyncio
async def test_get_quota_creates_default_when_missing() -> None:
    user_id = uuid.uuid4()
    quota_repository = AsyncMock()
    quota_repository.get_quota.return_value = None
    created_quota = AIUserQuota(user_id=user_id)
    quota_repository.upsert_quota.return_value = created_quota

    usecase = QuotaGuardUseCase(quota_repository)
    result = await usecase.get_quota(user_id)

    assert result == created_quota
    quota_repository.upsert_quota.assert_awaited_once()


@pytest.mark.asyncio
async def test_get_quota_syncs_tier_when_core_service_differs() -> None:
    user_id = uuid.uuid4()
    quota = _build_quota(user_id=user_id)

    quota_repository = AsyncMock()
    quota_repository.get_quota.return_value = quota
    updated_quota = quota.model_copy(update={"tier": Tier.PRO})
    quota_repository.upsert_quota.return_value = updated_quota

    core_service = AsyncMock()
    core_service.get_user_tier.return_value = Tier.PRO

    usecase = QuotaGuardUseCase(quota_repository, core_service=core_service)
    result = await usecase.get_quota(user_id)

    assert result.tier is Tier.PRO
    core_service.get_user_tier.assert_awaited_once_with(user_id)
    quota_repository.upsert_quota.assert_awaited_once()


@pytest.mark.asyncio
async def test_enforce_limits_raises_when_query_limit_exceeded() -> None:
    user_id = uuid.uuid4()
    quota = _build_quota(user_id=user_id).model_copy(
        update={
            "daily_query_count": 10,
            "daily_query_limit": 10,
        }
    )

    quota_repository = AsyncMock()
    quota_repository.get_quota.return_value = quota

    usecase = QuotaGuardUseCase(quota_repository)

    with pytest.raises(QuotaExceededError, match="daily query limit exceeded"):
        await usecase.enforce_limits(
            user_id,
            query_count=1,
            token_count=0,
            conversations_started=0,
            at=datetime(2026, 4, 7, 11, 0, tzinfo=UTC),
        )


@pytest.mark.asyncio
async def test_enforce_limits_resets_daily_counts_for_new_day() -> None:
    user_id = uuid.uuid4()
    quota = _build_quota(user_id=user_id).model_copy(
        update={
            "daily_query_count": 10,
            "daily_token_count": 500,
            "daily_conversations_started": 3,
            "last_query_date": date(2026, 4, 6),
        }
    )

    quota_repository = AsyncMock()
    quota_repository.get_quota.return_value = quota

    usecase = QuotaGuardUseCase(quota_repository)

    checked_quota = await usecase.enforce_limits(
        user_id,
        query_count=1,
        token_count=1,
        conversations_started=1,
        at=datetime(2026, 4, 7, 12, 0, tzinfo=UTC),
    )

    assert checked_quota == quota


@pytest.mark.asyncio
async def test_consume_usage_calls_increment_after_enforcement() -> None:
    user_id = uuid.uuid4()
    quota = _build_quota(user_id=user_id)

    quota_repository = AsyncMock()
    quota_repository.get_quota.return_value = quota
    consumed_quota = quota.model_copy(update={"daily_query_count": 3, "daily_token_count": 130})
    quota_repository.increment_usage.return_value = consumed_quota

    usecase = QuotaGuardUseCase(quota_repository)
    at = datetime(2026, 4, 7, 13, 0, tzinfo=UTC)

    result = await usecase.consume_usage(
        user_id,
        query_count=1,
        token_count=30,
        conversations_started=0,
        at=at,
    )

    assert result == consumed_quota
    quota_repository.increment_usage.assert_awaited_once_with(
        user_id,
        query_count=1,
        token_count=30,
        conversations_started=0,
        at=at,
    )


@pytest.mark.asyncio
async def test_get_usage_snapshot_resets_daily_values_for_new_day() -> None:
    user_id = uuid.uuid4()
    quota = _build_quota(user_id=user_id).model_copy(update={"last_query_date": date(2026, 4, 6)})

    quota_repository = AsyncMock()
    quota_repository.get_quota.return_value = quota

    usecase = QuotaGuardUseCase(quota_repository)
    snapshot = await usecase.get_usage_snapshot(user_id, on_date=date(2026, 4, 7))

    assert snapshot.daily_query_count == 0
    assert snapshot.daily_token_count == 0
    assert snapshot.daily_conversations_started == 0
    assert snapshot.total_queries_used == quota.total_queries_used
