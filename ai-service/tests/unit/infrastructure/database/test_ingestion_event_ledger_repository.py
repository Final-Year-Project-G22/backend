from __future__ import annotations

import uuid
from datetime import UTC, datetime
from unittest.mock import AsyncMock

import pytest

from core.ports.ingestion_event_ledger import RecordIngestionEventResult
from infrastructure.database.repositories.ingestion_event_ledger_repository import (
    SqlAlchemyIngestionEventLedgerRepository,
)


class _ScalarResult:
    def __init__(self, value: object | None) -> None:
        self._value = value

    def scalar_one_or_none(self) -> object | None:
        return self._value


async def _classify(*values: object | None) -> RecordIngestionEventResult:
    session = AsyncMock()
    session.execute.side_effect = [_ScalarResult(value) for value in values]
    repo = SqlAlchemyIngestionEventLedgerRepository(session)
    return await repo.record_or_classify(
        event_id="evt-1",
        idempotency_key="idem-1",
        account_id=uuid.uuid4(),
        document_id=uuid.uuid4(),
        occurred_at=datetime(2026, 4, 15, 9, 0, tzinfo=UTC),
    )


@pytest.mark.asyncio
async def test_record_or_classify_records_new_event() -> None:
    assert await _classify(None, None, None, uuid.uuid4()) is RecordIngestionEventResult.RECORDED


@pytest.mark.asyncio
async def test_record_or_classify_marks_stale_lineage_when_older_occurrence() -> None:
    assert (
        await _classify(None, None, datetime(2026, 4, 15, 10, 0, tzinfo=UTC))
        is RecordIngestionEventResult.STALE_LINEAGE
    )


@pytest.mark.asyncio
async def test_record_or_classify_classifies_duplicate_event() -> None:
    assert await _classify(uuid.uuid4()) is RecordIngestionEventResult.DUPLICATE_EVENT


@pytest.mark.asyncio
async def test_record_or_classify_classifies_duplicate_idempotency() -> None:
    assert await _classify(None, uuid.uuid4()) is RecordIngestionEventResult.DUPLICATE_IDEMPOTENCY
