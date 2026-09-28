from __future__ import annotations

import asyncio
import logging
import sys
from contextlib import AsyncExitStack
from pathlib import Path
from typing import Any, cast

import httpx

# Ensure grpc stubs are importable at runtime
sys.path.insert(0, str(Path(__file__).parent.parent / "grpc_stubs"))

from dependency_injector import providers

from app.config import Settings, get_settings
from app.container import Container
from core.domain.enums import DocumentStatus
from core.ports.ingestion_event_ledger import RecordIngestionEventResult
from core.usecases.ingestion_orchestrator import IngestionOrchestratorUseCase, PreparedIndex
from infrastructure.database.connection import Database
from infrastructure.embeddings import create_embedding_adapter
from infrastructure.messagebus import MessageHandlingRejectError
from infrastructure.messagebus.ingestion_consumer import IngestionRequestedConsumer
from infrastructure.messagebus.rabbitmq_event_bus import RabbitMQEventBusAdapter
from workers.tasks.ingestion_requested import (
    IngestionRequestedTaskHandler,
    ValidatedIngestionRequest,
)


def _reject_if_previously_handled(result: RecordIngestionEventResult) -> None:
    if result is RecordIngestionEventResult.DUPLICATE_EVENT:
        raise MessageHandlingRejectError("duplicate event id")
    if result is RecordIngestionEventResult.DUPLICATE_IDEMPOTENCY:
        raise MessageHandlingRejectError("duplicate idempotency key")
    if result is RecordIngestionEventResult.STALE_LINEAGE:
        raise MessageHandlingRejectError("stale ingestion lineage event")


async def handle_ingestion_message(
    payload: dict[str, Any],
    *,
    database: Database,
    validator: IngestionRequestedTaskHandler,
    orchestrator: IngestionOrchestratorUseCase,
) -> None:
    request: ValidatedIngestionRequest = validator.validate(payload)
    # Read without claiming the event during remote download, parse, and embedding.
    async with database.command() as command:
        previous = await command.ledger.classify_existing(
            event_id=request.event_id,
            idempotency_key=request.idempotency_key,
            account_id=request.account_id,
            document_id=request.document_id,
            occurred_at=request.occurred_at,
        )
        indexed = False
        if previous in (
            RecordIngestionEventResult.DUPLICATE_EVENT,
            RecordIngestionEventResult.DUPLICATE_IDEMPOTENCY,
        ):
            matches = await command.ledger.matches_record(
                event_id=request.event_id,
                idempotency_key=request.idempotency_key,
                account_id=request.account_id,
                document_id=request.document_id,
            )
            if matches:
                document = await command.knowledge.get_document(request.document_id)
                indexed = document is not None and document.status is DocumentStatus.ACTIVE

    if indexed:
        await orchestrator.announce_completed(
            event_id=request.event_id,
            document_id=request.document_id,
            account_id=request.account_id,
        )
        return
    _reject_if_previously_handled(previous)

    async def persist_index(index: PreparedIndex) -> None:
        async with database.command() as command:
            _reject_if_previously_handled(
                await command.ledger.record_or_classify(
                    event_id=request.event_id,
                    idempotency_key=request.idempotency_key,
                    account_id=request.account_id,
                    document_id=request.document_id,
                    occurred_at=request.occurred_at,
                )
            )
            await orchestrator.index_chunks(index, knowledge_repository=command.knowledge)

    result = await orchestrator.start_ingestion(payload, persist_index=persist_index)
    if result.status == "failed":
        # A deleted source is terminal: record it so redelivery does not download it again.
        async with database.command() as command:
            _reject_if_previously_handled(
                await command.ledger.record_or_classify(
                    event_id=request.event_id,
                    idempotency_key=request.idempotency_key,
                    account_id=request.account_id,
                    document_id=request.document_id,
                    occurred_at=request.occurred_at,
                )
            )


async def run_worker(settings: Settings | None = None) -> None:
    settings = settings or get_settings()
    container = Container(config=providers.Object(settings))

    logging.basicConfig(
        level=getattr(logging, settings.LOG_LEVEL.upper(), logging.INFO),
        format="%(asctime)s %(levelname)s [%(name)s] %(message)s",
    )

    async with AsyncExitStack() as resources:
        database = container.database()
        resources.push_async_callback(database.engine.dispose)
        http_client = await resources.enter_async_context(
            httpx.AsyncClient(trust_env=settings.HTTPX_TRUST_ENV)
        )
        cast(Any, container.http_client).override(http_client)
        cast(Any, container.embedding_port).override(
            create_embedding_adapter(settings, http_client=http_client)
        )

        core_client = container.core_grpc_client()
        resources.push_async_callback(core_client.close)
        document_channel = container.document_fetch_channel()
        resources.push_async_callback(document_channel.close)
        event_bus = container.event_bus_port()
        if isinstance(event_bus, RabbitMQEventBusAdapter):
            resources.push_async_callback(event_bus.close)
        if container.core_service_port() is None:
            raise RuntimeError("core_service_port is not available: check CORE_GRPC_ENDPOINT")

        orchestrator = container.ingestion_orchestrator()
        validator = container.ingestion_requested_task_handler()
        consumer: IngestionRequestedConsumer = container.ingestion_consumer()
        resources.push_async_callback(consumer.stop)

        async def handle_message(payload: dict[str, Any]) -> None:
            await handle_ingestion_message(
                payload, database=database, validator=validator, orchestrator=orchestrator
            )

        await consumer.start(handler=handle_message)
        await asyncio.Event().wait()


def main() -> None:
    asyncio.run(run_worker())


if __name__ == "__main__":
    main()
