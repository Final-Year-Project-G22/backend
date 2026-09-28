from __future__ import annotations

import asyncio
import os
import uuid

import aio_pika
import pytest

from app.config import Settings
from workers.ingestion_worker import run_worker


@pytest.mark.integration
async def test_worker_cancels_its_broker_subscription() -> None:
    broker_url = os.environ.get("AI_TEST_RABBITMQ_URL")
    if broker_url is None:
        pytest.skip("AI_TEST_RABBITMQ_URL requires RabbitMQ")

    queue_name = f"ai247.worker.{uuid.uuid4().hex}"
    settings = Settings(RABBITMQ_URL=broker_url, INGESTION_WORKER_QUEUE=queue_name)
    connection = await aio_pika.connect(broker_url)
    channel = await connection.channel()
    await channel.declare_queue(queue_name, durable=True)
    task = asyncio.create_task(run_worker(settings))
    try:

        async def wait_for_consumer() -> None:
            while True:
                queue = await channel.declare_queue(queue_name, passive=True)
                if queue.declaration_result and queue.declaration_result.consumer_count == 1:
                    return
                await asyncio.sleep(0.05)

        await asyncio.wait_for(wait_for_consumer(), 10)
        task.cancel()
        with pytest.raises(asyncio.CancelledError):
            await task
        queue = await channel.declare_queue(queue_name, passive=True)
        assert queue.declaration_result is not None
        assert queue.declaration_result.consumer_count == 0
        await channel.queue_delete(queue_name)
    finally:
        task.cancel()
        await asyncio.gather(task, return_exceptions=True)
        await connection.close()
