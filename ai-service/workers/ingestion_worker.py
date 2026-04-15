from __future__ import annotations

import asyncio
import logging

from app.container import Container
from infrastructure.messagebus.ingestion_consumer import IngestionRequestedConsumer
from workers.tasks.ingestion_requested import IngestionRequestedTaskHandler


async def run_worker() -> None:
    container = Container()
    settings = container.config()

    logging.basicConfig(
        level=getattr(logging, settings.LOG_LEVEL.upper(), logging.INFO),
        format="%(asctime)s %(levelname)s [%(name)s] %(message)s",
    )

    consumer: IngestionRequestedConsumer = container.ingestion_consumer()
    task_handler: IngestionRequestedTaskHandler = container.ingestion_requested_task_handler()

    await consumer.start(handler=task_handler.handle)
    try:
        await asyncio.Event().wait()
    finally:
        await consumer.stop()


def main() -> None:
    asyncio.run(run_worker())


if __name__ == "__main__":
    main()
