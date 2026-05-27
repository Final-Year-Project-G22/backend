from __future__ import annotations

import logging
import sys
from collections.abc import AsyncGenerator
from contextlib import asynccontextmanager
from pathlib import Path
from typing import Any, cast

import httpx
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

# Ensure grpc stubs are importable at runtime
sys.path.insert(0, str(Path(__file__).parent / "grpc_stubs"))

from app.config import Settings
from app.container import Container
from infrastructure.embeddings import create_embedding_adapter
from infrastructure.llm import create_llm_adapter
from infrastructure.rpc.server import serve_rpc

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None]:
    settings: Settings = app.state.settings
    container: Container = app.state.container

    ask_ai_usecase = container.ask_ai()
    conversation_usecase = container.conversation()
    rpc_server = await serve_rpc(
        port=settings.GRPC_PORT,
        ask_ai_usecase=ask_ai_usecase,
        conversation_usecase=conversation_usecase,
        auth_token=settings.INTERNAL_GRPC_AUTH_TOKEN or None,
        ask_enabled=settings.AI_ASK_ENABLED,
    )
    app.state.rpc_server = rpc_server

    logger.info("Starting %s v%s", settings.APP_NAME, settings.APP_VERSION)
    yield

    await rpc_server.stop(grace=5)
    await rpc_server.wait_for_termination()
    logger.info("Shutting down %s", settings.APP_NAME)


def create_app() -> FastAPI:
    container = Container()
    settings = container.config()

    # Create dependencies
    http_client = httpx.AsyncClient(trust_env=settings.HTTPX_TRUST_ENV)
    embedding_adapter = create_embedding_adapter(settings, http_client=http_client)
    llm_adapter = create_llm_adapter(settings, http_client=http_client)

    # Provide dependencies
    cast(Any, container.embedding_port).override(embedding_adapter)
    cast(Any, container.llm_port).override(llm_adapter)

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
