from __future__ import annotations

import uuid
from abc import ABC, abstractmethod
from datetime import datetime
from enum import StrEnum


class RecordIngestionEventResult(StrEnum):
    RECORDED = "recorded"
    DUPLICATE_EVENT = "duplicate_event"
    DUPLICATE_IDEMPOTENCY = "duplicate_idempotency"
    STALE_LINEAGE = "stale_lineage"


class IngestionEventOutcome(StrEnum):
    COMPLETED = "completed"
    FAILED = "failed"


class IngestionEventLedgerPort(ABC):
    @abstractmethod
    async def record_or_classify(
        self,
        *,
        event_id: str,
        idempotency_key: str,
        account_id: uuid.UUID,
        document_id: uuid.UUID,
        occurred_at: datetime,
        outcome: IngestionEventOutcome | None = None,
    ) -> RecordIngestionEventResult: ...


__all__ = ["IngestionEventLedgerPort", "IngestionEventOutcome", "RecordIngestionEventResult"]
