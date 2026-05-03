from __future__ import annotations

import asyncio
import logging
import sys
from pathlib import Path
from typing import Any, cast

import httpx

# Ensure grpc stubs are importable at runtime
sys.path.insert(0, str(Path(__file__).parent.parent / "grpc_stubs"))

from app.config import Settings
from app.container import Container
from app.security import build_ingestion_envelope_verifier
from core.usecases.ingestion_orchestrator import IngestionOrchestratorUseCase
from infrastructure.database.connection import session_scope
from infrastructure.database.repositories import (
    SqlAlchemyIngestionEventLedgerRepository,
    SqlAlchemyKnowledgeRepository,
)
from infrastructure.embeddings import create_embedding_adapter
from infrastructure.llm import create_llm_adapter
from infrastructure.messagebus.ingestion_consumer import IngestionRequestedConsumer
from workers.tasks.ingestion_requested import IngestionRequestedTaskHandler


async def run_worker() -> None:
    container = Container()
    settings: Settings = container.config()

    logging.basicConfig(
        level=getattr(logging, settings.LOG_LEVEL.upper(), logging.INFO),
        format="%(asctime)s %(levelname)s [%(name)s] %(message)s",
    )

    http_client = httpx.AsyncClient(trust_env=settings.HTTPX_TRUST_ENV)
    embedding_adapter = create_embedding_adapter(settings, http_client=http_client)
    llm_adapter = create_llm_adapter(settings, http_client=http_client)
    cast(Any, container.embedding_port).override(embedding_adapter)
    cast(Any, container.llm_port).override(llm_adapter)

    consumer: IngestionRequestedConsumer = container.ingestion_consumer()
    envelope_verifier = build_ingestion_envelope_verifier(settings)
    event_bus = container.event_bus_port()
    parser_registry = container.parser_registry()
    chunking_registry = container.chunking_registry()
    core_service = container.core_service_port()
    if core_service is None:
        raise RuntimeError("core_service_port is not available — check CORE_GRPC_ENDPOINT")

    async def handle_message(payload: dict[str, Any]) -> None:
        """Process a single ingestion message inside a transactional scope."""
        async with session_scope() as session:
            ledger_repo = SqlAlchemyIngestionEventLedgerRepository(session=session)
            knowledge_repo = SqlAlchemyKnowledgeRepository(session=session)

            orchestrator = IngestionOrchestratorUseCase(
                event_bus=event_bus,
                parser_registry=parser_registry,
                chunking_registry=chunking_registry,
                embedding_port=embedding_adapter,
                knowledge_repository=knowledge_repo,
                core_service_port=core_service,
                seaweedfs_filer_url=settings.SEAWEEDFS_FILER_URL,
            )

            handler = IngestionRequestedTaskHandler(
                envelope_verifier=envelope_verifier,
                ingestion_event_ledger_repository=ledger_repo,
                ingestion_orchestrator=orchestrator,
            )
            await handler.handle(payload)

    await consumer.start(handler=handle_message)
    try:
        await asyncio.Event().wait()
    finally:
        await consumer.stop()


def main() -> None:
    asyncio.run(run_worker())


if __name__ == "__main__":
    main()
