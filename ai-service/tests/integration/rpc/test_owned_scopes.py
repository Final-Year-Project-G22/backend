from __future__ import annotations

import asyncio
import uuid
from datetime import UTC, datetime

import pytest

from core.domain.enums import Language, Tier
from core.domain.models import AIConversationSession, AIUserQuota
from core.ports.ingestion_event_ledger import RecordIngestionEventResult
from infrastructure.database.connection import Database


@pytest.mark.integration
async def test_command_commits_related_writes_and_new_connection_sees_them(
    database: Database,
) -> None:
    user_id = uuid.uuid4()
    conversation = AIConversationSession(
        user_id=user_id,
        title="First turn",
        language=Language.ENGLISH,
        tier_at_start=Tier.BASIC,
        current_tier=Tier.BASIC,
    )
    async with database.command() as command:
        await command.quota.upsert_quota(AIUserQuota(user_id=user_id))
        await command.conversation.create_session(conversation)

    async with database.command() as reader:
        assert await reader.quota.get_quota(user_id) is not None
        assert await reader.conversation.get_session(conversation.id) is not None


@pytest.mark.integration
async def test_failed_command_rolls_back_related_writes(database: Database) -> None:
    user_id = uuid.uuid4()
    conversation = AIConversationSession(
        user_id=user_id,
        title="Failed turn",
        language=Language.ENGLISH,
        tier_at_start=Tier.BASIC,
        current_tier=Tier.BASIC,
    )

    async def fail() -> None:
        async with database.command() as command:
            await command.quota.upsert_quota(AIUserQuota(user_id=user_id))
            await command.conversation.create_session(conversation)
            raise RuntimeError("stop")

    with pytest.raises(RuntimeError, match="stop"):
        await fail()

    async with database.command() as reader:
        assert await reader.quota.get_quota(user_id) is None
        assert await reader.conversation.get_session(conversation.id) is None


@pytest.mark.integration
async def test_cancelled_command_releases_connection(database: Database) -> None:
    started = asyncio.Event()

    async def cancelled() -> None:
        async with database.command() as command:
            await command.quota.upsert_quota(AIUserQuota(user_id=uuid.uuid4()))
            started.set()
            await asyncio.Event().wait()

    task = asyncio.create_task(cancelled())
    await asyncio.wait_for(started.wait(), 5)
    task.cancel()
    with pytest.raises(asyncio.CancelledError):
        await task
    async with asyncio.timeout(5):
        async with database.command() as command:
            await command.quota.upsert_quota(AIUserQuota(user_id=uuid.uuid4()))


@pytest.mark.integration
async def test_duplicate_ledger_event_does_not_poison_command(database: Database) -> None:
    event = {
        "event_id": "event-1",
        "idempotency_key": "idempotent-1",
        "account_id": uuid.uuid4(),
        "document_id": uuid.uuid4(),
        "occurred_at": datetime.now(UTC),
    }
    async with database.command() as command:
        assert (
            await command.ledger.record_or_classify(**event) is RecordIngestionEventResult.RECORDED
        )
    async with database.command() as command:
        assert (
            await command.ledger.classify_existing(**event)
            is RecordIngestionEventResult.DUPLICATE_EVENT
        )
        assert (
            await command.ledger.record_or_classify(**event)
            is RecordIngestionEventResult.DUPLICATE_EVENT
        )
