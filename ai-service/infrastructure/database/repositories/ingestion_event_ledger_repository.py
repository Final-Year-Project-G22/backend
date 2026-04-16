from __future__ import annotations

import uuid
from datetime import datetime

from sqlalchemy import func, select
from sqlalchemy.exc import IntegrityError, SQLAlchemyError
from sqlalchemy.ext.asyncio import AsyncSession

from core.domain.exceptions import RepositoryError
from core.ports.ingestion_event_ledger import IngestionEventLedgerPort, RecordIngestionEventResult
from infrastructure.database import models_sqlalchemy as sa_models


class SqlAlchemyIngestionEventLedgerRepository(IngestionEventLedgerPort):
    def __init__(self, session: AsyncSession) -> None:
        self._session = session

    @property
    def session(self) -> AsyncSession:
        return self._session

    async def record_or_classify(
        self,
        *,
        event_id: str,
        idempotency_key: str,
        account_id: uuid.UUID,
        document_id: uuid.UUID,
        occurred_at: datetime,
    ) -> RecordIngestionEventResult:
        latest_occurred_at = await self._latest_occurred_at(
            account_id=account_id,
            document_id=document_id,
        )
        if latest_occurred_at is not None and occurred_at <= latest_occurred_at:
            return RecordIngestionEventResult.STALE_LINEAGE

        model = sa_models.IngestionEventLedger(
            event_id=event_id,
            idempotency_key=idempotency_key,
            account_id=account_id,
            document_id=document_id,
            occurred_at=occurred_at,
        )

        self._session.add(model)
        try:
            await self._session.flush()
        except IntegrityError:
            return await self._classify_duplicate(
                event_id=event_id, idempotency_key=idempotency_key
            )
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to persist ingestion ledger event",
                details={
                    "event_id": event_id,
                    "idempotency_key": idempotency_key,
                    "account_id": str(account_id),
                    "document_id": str(document_id),
                },
            ) from exc

        return RecordIngestionEventResult.RECORDED

    async def _latest_occurred_at(
        self,
        *,
        account_id: uuid.UUID,
        document_id: uuid.UUID,
    ) -> datetime | None:
        try:
            result = await self._session.execute(
                select(func.max(sa_models.IngestionEventLedger.occurred_at)).where(
                    sa_models.IngestionEventLedger.account_id == account_id,
                    sa_models.IngestionEventLedger.document_id == document_id,
                )
            )
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to query ingestion ledger lineage",
                details={
                    "account_id": str(account_id),
                    "document_id": str(document_id),
                },
            ) from exc

        return result.scalar_one_or_none()

    async def _classify_duplicate(
        self,
        *,
        event_id: str,
        idempotency_key: str,
    ) -> RecordIngestionEventResult:
        try:
            event_result = await self._session.execute(
                select(sa_models.IngestionEventLedger.id).where(
                    sa_models.IngestionEventLedger.event_id == event_id
                )
            )
            if event_result.scalar_one_or_none() is not None:
                return RecordIngestionEventResult.DUPLICATE_EVENT

            idempotency_result = await self._session.execute(
                select(sa_models.IngestionEventLedger.id).where(
                    sa_models.IngestionEventLedger.idempotency_key == idempotency_key
                )
            )
            if idempotency_result.scalar_one_or_none() is not None:
                return RecordIngestionEventResult.DUPLICATE_IDEMPOTENCY
        except SQLAlchemyError as exc:
            raise RepositoryError(
                "failed to classify duplicate ingestion event",
                details={
                    "event_id": event_id,
                    "idempotency_key": idempotency_key,
                },
            ) from exc

        return RecordIngestionEventResult.DUPLICATE_EVENT


__all__ = ["SqlAlchemyIngestionEventLedgerRepository"]
