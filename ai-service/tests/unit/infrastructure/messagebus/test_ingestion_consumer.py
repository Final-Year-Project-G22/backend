from __future__ import annotations

import json
from unittest.mock import AsyncMock

import pytest

from infrastructure.messagebus.ingestion_consumer import IngestionRequestedConsumer


def _build_consumer(*, requeue_on_failure: bool = True) -> IngestionRequestedConsumer:
    return IngestionRequestedConsumer(
        rabbitmq_url="amqp://guest:guest@localhost:5672/",
        queue_name="ai.ingestion.requested.v1",
        exchange_name="ai.ingestion.events",
        routing_key="document.ingestion.requested.v1",
        prefetch_count=8,
        requeue_on_failure=requeue_on_failure,
    )


@pytest.mark.asyncio
async def test_on_message_acks_when_handler_succeeds() -> None:
    consumer = _build_consumer()
    message = AsyncMock()
    message.body = json.dumps({"event_id": "e-1"}).encode("utf-8")
    handler = AsyncMock()

    await consumer.on_message(message=message, handler=handler)

    handler.assert_awaited_once_with({"event_id": "e-1"})
    message.ack.assert_awaited_once()
    message.nack.assert_not_awaited()


@pytest.mark.asyncio
async def test_on_message_rejects_invalid_json_without_requeue() -> None:
    consumer = _build_consumer()
    message = AsyncMock()
    message.body = b"not-json"
    handler = AsyncMock()

    await consumer.on_message(message=message, handler=handler)

    handler.assert_not_awaited()
    message.ack.assert_not_awaited()
    message.nack.assert_awaited_once_with(requeue=False)


@pytest.mark.asyncio
async def test_on_message_nacks_with_configured_requeue_on_handler_error() -> None:
    consumer = _build_consumer(requeue_on_failure=False)
    message = AsyncMock()
    message.body = json.dumps({"event_id": "e-2"}).encode("utf-8")
    handler = AsyncMock(side_effect=RuntimeError("boom"))

    await consumer.on_message(message=message, handler=handler)

    message.ack.assert_not_awaited()
    message.nack.assert_awaited_once_with(requeue=False)
