from __future__ import annotations

import uuid
from datetime import UTC, datetime
from unittest.mock import AsyncMock, Mock

import pytest
from sqlalchemy.exc import IntegrityError

from core.ports.ingestion_event_ledger import RecordIngestionEventResult
from infrastructure.database.repositories.ingestion_event_ledger_repository import (
    SqlAlchemyIngestionEventLedgerRepository,
)


class _ScalarResult:
    def __init__(self, value: object | None) -> None:
        self._value = value

    def scalar_one_or_none(self) -> object | None:
        return self._value


@pytest.mark.asyncio
async def test_record_or_classify_records_new_event() -> None:
    session = AsyncMock()
    session.add = Mock()
    session.execute.return_value = _ScalarResult(None)
    repo = SqlAlchemyIngestionEventLedgerRepository(session)

    result = await repo.record_or_classify(
        event_id="evt-1",
        idempotency_key="idem-1",
        account_id=uuid.uuid4(),
        document_id=uuid.uuid4(),
        occurred_at=datetime(2026, 4, 15, 9, 0, tzinfo=UTC),
    )

    assert result is RecordIngestionEventResult.RECORDED
    session.add.assert_called_once()
    session.flush.assert_awaited_once()


@pytest.mark.asyncio
async def test_record_or_classify_marks_stale_lineage_when_older_occurrence() -> None:
    session = AsyncMock()
    session.add = Mock()
    session.execute.return_value = _ScalarResult(
        datetime(
            2026,
            4,
            15,
            10,
            0,
            tzinfo=UTC,
        )
    )
    repo = SqlAlchemyIngestionEventLedgerRepository(session)

    result = await repo.record_or_classify(
        event_id="evt-2",
        idempotency_key="idem-2",
        account_id=uuid.uuid4(),
        document_id=uuid.uuid4(),
        occurred_at=datetime(2026, 4, 15, 9, 0, tzinfo=UTC),
    )

    assert result is RecordIngestionEventResult.STALE_LINEAGE
    session.add.assert_not_called()


@pytest.mark.asyncio
async def test_record_or_classify_classifies_duplicate_event() -> None:
    session = AsyncMock()
    session.add = Mock()
    session.execute.side_effect = [
        _scalar_result(None),
        _scalar_result(uuid.uuid4()),
    ]
    session.flush.side_effect = _integrity_error()
    repo = SqlAlchemyIngestionEventLedgerRepository(session)

    result = await repo.record_or_classify(
        event_id="evt-3",
        idempotency_key="idem-3",
        account_id=uuid.uuid4(),
        document_id=uuid.uuid4(),
        occurred_at=datetime(2026, 4, 15, 11, 0, tzinfo=UTC),
    )

    assert result is RecordIngestionEventResult.DUPLICATE_EVENT


@pytest.mark.asyncio
async def test_record_or_classify_classifies_duplicate_idempotency() -> None:
    session = AsyncMock()
    session.add = Mock()
    session.execute.side_effect = [
        _scalar_result(None),
        _scalar_result(None),
        _scalar_result(uuid.uuid4()),
    ]
    session.flush.side_effect = _integrity_error()
    repo = SqlAlchemyIngestionEventLedgerRepository(session)

    result = await repo.record_or_classify(
        event_id="evt-4",
        idempotency_key="idem-4",
        account_id=uuid.uuid4(),
        document_id=uuid.uuid4(),
        occurred_at=datetime(2026, 4, 15, 11, 30, tzinfo=UTC),
    )

    assert result is RecordIngestionEventResult.DUPLICATE_IDEMPOTENCY


def _scalar_result(value: object | None) -> object:
    return _ScalarResult(value)


def _integrity_error() -> IntegrityError:
    return IntegrityError("duplicate", params=None, orig=Exception("duplicate"))
