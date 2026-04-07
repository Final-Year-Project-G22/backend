from __future__ import annotations

import uuid
from datetime import UTC, date, datetime

from sqlalchemy import select
from sqlalchemy.exc import SQLAlchemyError
from sqlalchemy.ext.asyncio import AsyncSession

from core.domain.exceptions import RepositoryError
from core.domain.models import AIUserQuota
from core.ports.quota_repository import QuotaRepositoryPort
from infrastructure.database import models_sqlalchemy as sa_models
from infrastructure.database.repositories.mappers import to_domain_quota, to_orm_quota


class SqlAlchemyQuotaRepository(QuotaRepositoryPort):
    def __init__(self, session: AsyncSession) -> None:
        self._session = session

    @property
    def session(self) -> AsyncSession:
        return self._session

    async def get_quota(self, user_id: uuid.UUID) -> AIUserQuota | None:
        try:
            model = await self._get_model_by_user_id(user_id)
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to fetch quota",
                details={"user_id": str(user_id)},
            ) from exc

        if model is None:
            return None
        return to_domain_quota(model)

    async def upsert_quota(self, quota: AIUserQuota) -> AIUserQuota:
        try:
            model = await self._get_model_by_user_id(quota.user_id)
            if model is None:
                model = to_orm_quota(quota)
                self._session.add(model)
            else:
                self._apply_quota_domain_values(model, quota)
            await self._session.flush()
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to upsert quota",
                details={"user_id": str(quota.user_id)},
            ) from exc

        return to_domain_quota(model)

    async def increment_usage(
        self,
        user_id: uuid.UUID,
        *,
        query_count: int,
        token_count: int,
        conversations_started: int = 0,
        at: datetime | None = None,
    ) -> AIUserQuota:
        at_dt = at or datetime.now(UTC)

        try:
            model = await self._get_model_by_user_id(user_id)
            if model is None:
                model = to_orm_quota(AIUserQuota(user_id=user_id))
                self._session.add(model)

            if model.last_query_date != at_dt.date():
                model.daily_query_count = 0
                model.daily_token_count = 0
                model.daily_conversations_started = 0

            model.daily_query_count += query_count
            model.daily_token_count += token_count
            model.daily_conversations_started += conversations_started
            model.total_queries_used += query_count
            model.total_tokens_used += token_count
            model.total_conversations += conversations_started
            model.last_query_date = at_dt.date()
            model.last_query_at = at_dt
            model.updated_at = at_dt
            if conversations_started > 0:
                model.last_conversation_started = at_dt

            await self._session.flush()
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to increment quota usage",
                details={"user_id": str(user_id)},
            ) from exc

        return to_domain_quota(model)

    async def reset_daily_usage(self, user_id: uuid.UUID, *, on_date: date) -> AIUserQuota:
        try:
            model = await self._get_model_by_user_id(user_id)
            if model is None:
                model = to_orm_quota(AIUserQuota(user_id=user_id, last_query_date=on_date))
                self._session.add(model)
            else:
                model.daily_query_count = 0
                model.daily_token_count = 0
                model.daily_conversations_started = 0
                model.last_query_date = on_date
                model.last_query_at = None
                model.updated_at = datetime.now(UTC)

            await self._session.flush()
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to reset daily quota usage",
                details={"user_id": str(user_id), "on_date": on_date.isoformat()},
            ) from exc

        return to_domain_quota(model)

    async def _get_model_by_user_id(self, user_id: uuid.UUID) -> sa_models.AIUserQuota | None:
        result = await self._session.execute(
            select(sa_models.AIUserQuota).where(sa_models.AIUserQuota.user_id == user_id)
        )
        return result.scalar_one_or_none()

    def _apply_quota_domain_values(
        self,
        model: sa_models.AIUserQuota,
        quota: AIUserQuota,
    ) -> None:
        model.tier = quota.tier.value
        model.tier_updated_at = quota.tier_updated_at
        model.daily_query_count = quota.daily_query_count
        model.daily_token_count = quota.daily_token_count
        model.daily_conversations_started = quota.daily_conversations_started
        model.daily_query_limit = quota.daily_query_limit
        model.daily_token_limit = quota.daily_token_limit
        model.max_conversations_per_day = quota.max_conversations_per_day
        model.total_queries_used = quota.total_queries_used
        model.total_tokens_used = quota.total_tokens_used
        model.total_conversations = quota.total_conversations
        model.last_query_date = quota.last_query_date
        model.last_query_at = quota.last_query_at
        model.last_conversation_started = quota.last_conversation_started
        model.created_at = quota.created_at
        model.updated_at = quota.updated_at


__all__ = ["SqlAlchemyQuotaRepository"]
