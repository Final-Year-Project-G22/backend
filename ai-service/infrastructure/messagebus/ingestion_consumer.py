from __future__ import annotations

import json
import logging
import random
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

from core.domain.retry_policy import BackoffType, RetryPolicy

IngestionMessageHandler = Callable[[dict[str, Any]], Awaitable[None]]

logger = logging.getLogger(__name__)


class MessageHandlingRejectError(Exception):
    pass


class MessageHandlingRetryError(Exception):
    pass


def calculate_retry_delay_ms(policy: RetryPolicy, retry_count: int) -> int:
    attempt = retry_count + 1
    base = policy.base_delay_ms
    max_delay = policy.max_delay_ms
    jitter_factor = policy.jitter_factor

    if policy.backoff_type is BackoffType.EXPONENTIAL:
        delay = base * (2 ** (attempt - 1))
    else:
        delay = base * attempt

    delay = min(delay, max_delay)

    if jitter_factor > 0:
        jitter = int(delay * jitter_factor)
        lower = max(0, delay - jitter)
        upper = delay + jitter
        delay = random.randint(lower, upper)  # noqa: S311

    return delay


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
        max_retries: int = 3,
        retry_base_delay_ms: int = 1000,
        retry_max_delay_ms: int = 30000,
        retry_jitter_factor: float = 0.1,
        dlq_enabled: bool = True,
        dlq_exchange: str = "ai.ingestion.dlq",
        dlq_routing_key: str = "document.ingestion.dlq.v1",
    ) -> None:
        self._rabbitmq_url = rabbitmq_url
        self._queue_name = queue_name
        self._exchange_name = exchange_name
        self._routing_key = routing_key
        self._prefetch_count = max(1, prefetch_count)
        self._requeue_on_failure = requeue_on_failure
        self._max_retries = max(0, max_retries)
        self._retry_policy = RetryPolicy(
            max_retries=max_retries,
            base_delay_ms=retry_base_delay_ms,
            max_delay_ms=retry_max_delay_ms,
            jitter_factor=retry_jitter_factor,
        )
        self._dlq_enabled = dlq_enabled
        self._dlq_exchange = dlq_exchange
        self._dlq_routing_key = dlq_routing_key

        self._connection: AbstractRobustConnection | None = None
        self._channel: AbstractChannel | None = None
        self._queue: AbstractQueue | None = None
        self._dlq_exchange_channel: AbstractExchange | None = None
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

        if self._dlq_enabled:
            self._dlq_exchange_channel = await self._channel.declare_exchange(
                self._dlq_exchange,
                aio_pika.ExchangeType.TOPIC,
                durable=True,
            )

        self._consumer_tag = await queue.consume(
            lambda message: self.on_message(message=message, handler=handler)
        )

        logger.info(
            "ingestion consumer started queue=%s exchange=%s routing_key=%s prefetch=%s max_retries=%s dlq_enabled=%s",
            self._queue_name,
            self._exchange_name,
            self._routing_key,
            self._prefetch_count,
            self._max_retries,
            self._dlq_enabled,
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
        retry_count = self._get_retry_count(message)
        error_message = ""

        logger.info(
            "message received delivery_tag=%s content_type=%s body_len=%d",
            message.delivery_tag,
            message.content_type,
            len(message.body),
        )

        try:
            payload = _decode_payload(message.body)
        except (TypeError, ValueError):
            await message.nack(requeue=False)
            logger.warning("invalid ingestion message payload rejected")
            return

        logger.info("message decoded event_type=%s", payload.get("event_type", "unknown"))

        try:
            await handler(payload)
        except MessageHandlingRejectError as exc:
            error_message = str(exc)
            await message.nack(requeue=False)
            logger.warning("ingestion message rejected by handler: %s", error_message)
            if self._dlq_enabled and self._dlq_exchange_channel:
                await self._send_to_dlq(message, error_message, retry_count)
            return
        except MessageHandlingRetryError as exc:
            error_message = str(exc)
            await message.nack(requeue=True)
            logger.warning(
                "ingestion message scheduled for retry (attempt %s/%s): %s",
                retry_count + 1,
                self._max_retries,
                error_message,
            )
            return
        except Exception as exc:
            error_message = str(exc)
            if not self._requeue_on_failure:
                if self._dlq_enabled and self._dlq_exchange_channel:
                    await self._send_to_dlq(message, error_message, 0)
                await message.nack(requeue=False)
                return
            retry_count += 1
            if retry_count >= self._max_retries:
                logger.exception(
                    "ingestion message exhausted retries (max=%s), sending to DLQ: %s",
                    self._max_retries,
                    error_message,
                )
                if self._dlq_enabled and self._dlq_exchange_channel:
                    await self._send_to_dlq(message, error_message, retry_count)
                await message.nack(requeue=False)
            else:
                logger.warning(
                    "ingestion message failed (attempt %s/%s), requeueing: %s",
                    retry_count,
                    self._max_retries,
                    error_message,
                )
                await message.nack(requeue=True)
            return

        await message.ack()

    def _get_retry_count(self, message: AbstractIncomingMessage) -> int:
        header = message.headers
        if header and "x-retry-count" in header:
            raw = header["x-retry-count"]
            if raw is not None:
                try:
                    return int(cast(int, raw))
                except (TypeError, ValueError):
                    pass
        return 0

    async def _send_to_dlq(
        self,
        message: AbstractIncomingMessage,
        error_message: str,
        retry_count: int,
    ) -> None:
        if self._dlq_exchange_channel is None or self._channel is None:
            return

        try:
            dlq_payload = {
                "original_payload": _decode_payload(message.body),
                "error_message": error_message,
                "retry_count": retry_count,
            }
            timestamp = message.timestamp
            if timestamp is not None:
                dlq_payload["dlq_routed_at"] = timestamp.isoformat()

            dlq_body = json.dumps(dlq_payload).encode("utf-8")
            dlq_message = aio_pika.Message(
                body=dlq_body,
                content_type="application/json",
            )
            await self._dlq_exchange_channel.publish(
                dlq_message,
                routing_key=self._dlq_routing_key,
            )
            logger.info(
                "message routed to DLQ after %s retry attempts",
                retry_count,
            )
        except Exception:
            logger.exception("failed to route message to DLQ")


def _decode_payload(raw: bytes) -> dict[str, Any]:
    try:
        decoded: Any = json.loads(raw.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise ValueError("message body is not valid json") from exc

    if not isinstance(decoded, dict):
        raise TypeError("message payload must be a json object")

    return cast(dict[str, Any], decoded)


__all__ = [
    "IngestionMessageHandler",
    "IngestionRequestedConsumer",
    "MessageHandlingRejectError",
    "MessageHandlingRetryError",
]
