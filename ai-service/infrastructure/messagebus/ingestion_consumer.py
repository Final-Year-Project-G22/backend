from __future__ import annotations

import json
import logging
from collections.abc import Awaitable, Callable
from typing import Any, cast

import aio_pika
from aio_pika.abc import (
    AbstractChannel,
    AbstractExchange,
    AbstractIncomingMessage,
    AbstractQueue,
    AbstractRobustConnection,
)

IngestionMessageHandler = Callable[[dict[str, Any]], Awaitable[None]]

logger = logging.getLogger(__name__)


class IngestionRequestedConsumer:
    def __init__(
        self,
        *,
        rabbitmq_url: str,
        queue_name: str,
        exchange_name: str,
        routing_key: str,
        prefetch_count: int,
        requeue_on_failure: bool = True,
    ) -> None:
        self._rabbitmq_url = rabbitmq_url
        self._queue_name = queue_name
        self._exchange_name = exchange_name
        self._routing_key = routing_key
        self._prefetch_count = max(1, prefetch_count)
        self._requeue_on_failure = requeue_on_failure

        self._connection: AbstractRobustConnection | None = None
        self._channel: AbstractChannel | None = None
        self._queue: AbstractQueue | None = None
        self._consumer_tag: str | None = None

    async def start(self, *, handler: IngestionMessageHandler) -> None:
        if self._consumer_tag is not None:
            return

        self._connection = await aio_pika.connect_robust(self._rabbitmq_url)
        self._channel = await self._connection.channel()
        await self._channel.set_qos(prefetch_count=self._prefetch_count)

        exchange = await self._declare_exchange(self._channel)
        queue = await self._channel.declare_queue(self._queue_name, durable=True)
        await queue.bind(exchange, routing_key=self._routing_key)
        self._queue = queue

        self._consumer_tag = await queue.consume(
            lambda message: self.on_message(message=message, handler=handler)
        )

        logger.info(
            "ingestion consumer started queue=%s exchange=%s routing_key=%s prefetch=%s",
            self._queue_name,
            self._exchange_name,
            self._routing_key,
            self._prefetch_count,
        )

    async def stop(self) -> None:
        if self._queue is not None and self._consumer_tag is not None:
            await self._queue.cancel(self._consumer_tag)

        self._consumer_tag = None

        if self._channel is not None and not self._channel.is_closed:
            await self._channel.close()
        self._channel = None

        if self._connection is not None and not self._connection.is_closed:
            await self._connection.close()
        self._connection = None

        self._queue = None
        logger.info("ingestion consumer stopped")

    async def _declare_exchange(self, channel: AbstractChannel) -> AbstractExchange:
        return await channel.declare_exchange(
            self._exchange_name,
            aio_pika.ExchangeType.TOPIC,
            durable=True,
        )

    async def on_message(
        self,
        *,
        message: AbstractIncomingMessage,
        handler: IngestionMessageHandler,
    ) -> None:
        try:
            payload = _decode_payload(message.body)
        except (TypeError, ValueError):
            await message.nack(requeue=False)
            logger.warning("invalid ingestion message payload rejected")
            return

        try:
            await handler(payload)
        except Exception:
            await message.nack(requeue=self._requeue_on_failure)
            logger.exception("ingestion handler failed")
            return

        await message.ack()


def _decode_payload(raw: bytes) -> dict[str, Any]:
    try:
        decoded: Any = json.loads(raw.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise ValueError("message body is not valid json") from exc

    if not isinstance(decoded, dict):
        raise TypeError("message payload must be a json object")

    return cast(dict[str, Any], decoded)


__all__ = ["IngestionMessageHandler", "IngestionRequestedConsumer"]
