from __future__ import annotations

import json
import logging
from typing import Any

import aio_pika
from aio_pika.abc import AbstractChannel, AbstractExchange, AbstractRobustConnection

from core.ports.event_bus import EventBusPort

logger = logging.getLogger(__name__)


class RabbitMQEventBusAdapter(EventBusPort):
    """Publishes status events to RabbitMQ using aio-pika."""

    def __init__(
        self,
        *,
        amqp_url: str,
        exchange_name: str = "core.events",
    ) -> None:
        self._amqp_url = amqp_url
        self._exchange_name = exchange_name
        self._connection: AbstractRobustConnection | None = None
        self._channel: AbstractChannel | None = None
        self._exchange: AbstractExchange | None = None

    async def _ensure_connected(self) -> AbstractExchange:
        """Lazily connect and declare the exchange."""
        if self._exchange is not None and not self._channel.is_closed:  # type: ignore[union-attr]
            return self._exchange

        self._connection = await aio_pika.connect_robust(self._amqp_url)
        self._channel = await self._connection.channel()
        self._exchange = await self._channel.declare_exchange(
            self._exchange_name,
            aio_pika.ExchangeType.TOPIC,
            durable=True,
        )
        logger.info(
            "event bus connected exchange=%s",
            self._exchange_name,
        )
        return self._exchange

    async def publish(self, topic: str, payload: Any) -> None:
        """Publish a JSON-serialized payload to the configured exchange."""
        exchange = await self._ensure_connected()
        body = json.dumps(payload, default=str).encode("utf-8")
        message = aio_pika.Message(
            body=body,
            content_type="application/json",
            delivery_mode=aio_pika.DeliveryMode.PERSISTENT,
        )
        await exchange.publish(message, routing_key=topic)
        logger.debug("published event topic=%s", topic)

    async def subscribe(self, topic: str, handler: Any) -> None:
        """Not implemented — the AI service only publishes status events."""
        raise NotImplementedError("RabbitMQEventBusAdapter does not support subscription")

    async def close(self) -> None:
        """Gracefully close the connection."""
        if self._channel is not None and not self._channel.is_closed:
            await self._channel.close()
        if self._connection is not None and not self._connection.is_closed:
            await self._connection.close()
        self._exchange = None
