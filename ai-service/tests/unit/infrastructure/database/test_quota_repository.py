from __future__ import annotations

import uuid
from datetime import UTC, date, datetime
from unittest.mock import AsyncMock, Mock

import pytest
from sqlalchemy.exc import SQLAlchemyError
from sqlalchemy.ext.asyncio import AsyncSession

from core.domain.enums import Tier
from core.domain.exceptions import RepositoryError
from core.domain.models import AIUserQuota
from infrastructure.database import models_sqlalchemy as sa_models
from infrastructure.database.repositories.mappers import to_orm_quota
from infrastructure.database.repositories.quota_repository import SqlAlchemyQuotaRepository


class _ScalarResult:
    def __init__(self, value: sa_models.AIUserQuota | None) -> None:
        self._value = value

    def scalar_one_or_none(self) -> sa_models.AIUserQuota | None:
        return self._value


def _build_session() -> AsyncMock:
    session = AsyncMock(spec=AsyncSession)
    session.add = Mock()
    return session


def _build_quota_model(*, user_id: uuid.UUID | None = None) -> sa_models.AIUserQuota:
    now = datetime(2026, 4, 7, 10, 0, tzinfo=UTC)
    domain = AIUserQuota(
        user_id=user_id or uuid.uuid4(),
        tier=Tier.BASIC,
        tier_updated_at=now,
        daily_query_count=0,
        daily_token_count=0,
        daily_conversations_started=0,
        daily_query_limit=10,
        daily_token_limit=50000,
        max_conversations_per_day=5,
        total_queries_used=0,
        total_tokens_used=0,
        total_conversations=0,
        created_at=now,
        updated_at=now,
    )
    return to_orm_quota(domain)


@pytest.mark.asyncio
async def test_get_quota_returns_domain_when_found() -> None:
    user_id = uuid.uuid4()
    model = _build_quota_model(user_id=user_id)
    session = _build_session()
    session.execute.return_value = _ScalarResult(model)

    repo = SqlAlchemyQuotaRepository(session)
    result = await repo.get_quota(user_id)

    assert result is not None
    assert result.user_id == user_id
    assert result.tier is Tier.BASIC


@pytest.mark.asyncio
async def test_get_quota_returns_none_when_missing() -> None:
    session = _build_session()
    session.execute.return_value = _ScalarResult(None)

    repo = SqlAlchemyQuotaRepository(session)
    result = await repo.get_quota(uuid.uuid4())

    assert result is None


@pytest.mark.asyncio
async def test_upsert_quota_creates_when_missing() -> None:
    session = _build_session()
    session.execute.return_value = _ScalarResult(None)
    repo = SqlAlchemyQuotaRepository(session)
    quota = AIUserQuota(user_id=uuid.uuid4(), daily_query_limit=20)

    result = await repo.upsert_quota(quota)

    assert result.user_id == quota.user_id
    session.add.assert_called_once()
    session.flush.assert_awaited_once()


@pytest.mark.asyncio
async def test_increment_usage_resets_when_day_changes() -> None:
    query_count = 2
    token_count = 50
    conversations_started = 1

    user_id = uuid.uuid4()
    model = _build_quota_model(user_id=user_id)
    model.daily_query_count = 9
    model.daily_token_count = 100
    model.daily_conversations_started = 3
    model.last_query_date = date(2026, 4, 6)

    session = _build_session()
    session.execute.return_value = _ScalarResult(model)
    repo = SqlAlchemyQuotaRepository(session)

    at = datetime(2026, 4, 7, 12, 0, tzinfo=UTC)
    result = await repo.increment_usage(
        user_id,
        query_count=query_count,
        token_count=token_count,
        conversations_started=conversations_started,
        at=at,
    )

    assert result.daily_query_count == query_count
    assert result.daily_token_count == token_count
    assert result.daily_conversations_started == conversations_started
    assert result.total_queries_used == query_count
    assert result.total_tokens_used == token_count
    assert result.total_conversations == conversations_started
    assert result.last_query_date == at.date()
    assert result.last_query_at == at
    assert result.last_conversation_started == at


@pytest.mark.asyncio
async def test_increment_usage_creates_default_quota_if_missing() -> None:
    query_count = 1
    token_count = 20

    user_id = uuid.uuid4()
    session = _build_session()
    session.execute.return_value = _ScalarResult(None)
    repo = SqlAlchemyQuotaRepository(session)

    at = datetime(2026, 4, 7, 13, 0, tzinfo=UTC)
    result = await repo.increment_usage(
        user_id,
        query_count=query_count,
        token_count=token_count,
        conversations_started=0,
        at=at,
    )

    assert result.user_id == user_id
    assert result.daily_query_count == query_count
    assert result.daily_token_count == token_count
    assert result.total_queries_used == query_count
    assert result.total_tokens_used == token_count
    session.add.assert_called_once()


@pytest.mark.asyncio
async def test_reset_daily_usage_resets_counts_and_last_query_at() -> None:
    user_id = uuid.uuid4()
    model = _build_quota_model(user_id=user_id)
    model.daily_query_count = 7
    model.daily_token_count = 700
    model.daily_conversations_started = 2
    model.last_query_at = datetime(2026, 4, 7, 9, 0, tzinfo=UTC)

    session = _build_session()
    session.execute.return_value = _ScalarResult(model)
    repo = SqlAlchemyQuotaRepository(session)

    on_date = date(2026, 4, 8)
    result = await repo.reset_daily_usage(user_id, on_date=on_date)

    assert result.daily_query_count == 0
    assert result.daily_token_count == 0
    assert result.daily_conversations_started == 0
    assert result.last_query_date == on_date
    assert result.last_query_at is None


@pytest.mark.asyncio
async def test_get_quota_wraps_sqlalchemy_errors() -> None:
    session = _build_session()
    session.execute.side_effect = SQLAlchemyError("boom")
    repo = SqlAlchemyQuotaRepository(session)

    with pytest.raises(RepositoryError, match="failed to fetch quota"):
        await repo.get_quota(uuid.uuid4())
