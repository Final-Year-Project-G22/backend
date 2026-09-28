from __future__ import annotations

import logging
import sys
from collections.abc import AsyncGenerator
from contextlib import AsyncExitStack, asynccontextmanager
from pathlib import Path
from typing import Any, cast

import httpx
from dependency_injector import providers
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

# Ensure grpc stubs are importable at runtime
sys.path.insert(0, str(Path(__file__).parent / "grpc_stubs"))

from app.config import Settings, get_settings
from app.container import Container
from infrastructure.embeddings import create_embedding_adapter
from infrastructure.llm import create_llm_adapter
from infrastructure.messagebus.rabbitmq_event_bus import RabbitMQEventBusAdapter
from infrastructure.rpc.server import serve_rpc

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None]:
    settings: Settings = app.state.settings
    container: Container = app.state.container

    async with AsyncExitStack() as resources:
        database = container.database()
        resources.push_async_callback(database.engine.dispose)
        http_client = await resources.enter_async_context(
            httpx.AsyncClient(trust_env=settings.HTTPX_TRUST_ENV)
        )
        cast(Any, container.http_client).override(http_client)
        if not cast(Any, container.embedding_port).overridden:
            cast(Any, container.embedding_port).override(
                create_embedding_adapter(settings, http_client=http_client)
            )
        if not cast(Any, container.llm_port).overridden:
            cast(Any, container.llm_port).override(
                create_llm_adapter(settings, http_client=http_client)
            )

        core_client = container.core_grpc_client()
        resources.push_async_callback(core_client.close)
        tool_client = container.ai_tool_client()
        resources.push_async_callback(tool_client.close)
        document_channel = container.document_fetch_channel()
        resources.push_async_callback(document_channel.close)
        event_bus = container.event_bus_port()
        if isinstance(event_bus, RabbitMQEventBusAdapter):
            resources.push_async_callback(event_bus.close)

        tool_registry = container.tool_registry()
        await tool_registry.initialize()
        logger.info(
            "Tool registry initialized with %d tools",
            len(await tool_registry.get_tool_definitions()),
        )
        await container.intent_classifier().initialize()

        rpc_server = await serve_rpc(
            port=settings.GRPC_PORT,
            ask_factory=container.ask_ai,
            conversation_factory=container.conversation,
            auth_token=settings.INTERNAL_GRPC_AUTH_TOKEN or None,
            ask_enabled=settings.AI_ASK_ENABLED,
        )
        app.state.rpc_server = rpc_server

        async def stop_rpc() -> None:
            await rpc_server.stop(grace=5)
            await rpc_server.wait_for_termination()

        resources.push_async_callback(stop_rpc)
        logger.info("Starting %s v%s", settings.APP_NAME, settings.APP_VERSION)
        yield
        logger.info("Shutting down %s", settings.APP_NAME)


def create_app(settings: Settings | None = None) -> FastAPI:
    settings = settings or get_settings()
    container = Container(config=providers.Object(settings))

    app = FastAPI(
        title=settings.APP_NAME,
        version=settings.APP_VERSION,
        debug=settings.DEBUG,
        lifespan=lifespan,
    )
    app.state.settings = settings
    app.state.container = container
    container.wire(modules=["main"])

    app.add_middleware(
        CORSMiddleware,
        allow_origins=settings.CORS_ORIGINS,
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    @app.get("/")
    async def root() -> dict[str, str]:
        return {"service": settings.APP_NAME, "version": settings.APP_VERSION}

    _ = root

    @app.get("/health")
    async def health() -> dict[str, str]:
        return {"status": "healthy", "service": settings.APP_NAME, "version": settings.APP_VERSION}

    _ = health

    return app


app = create_app()

if __name__ == "__main__":
    import uvicorn

    from app.config import get_settings

    settings = get_settings()
    uvicorn.run("main:app", host="0.0.0.0", port=settings.HTTP_PORT, reload=settings.DEBUG)  # noqa: S104
