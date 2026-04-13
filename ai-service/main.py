from __future__ import annotations

import logging
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.config import Settings
from app.container import Container
from infrastructure.rpc.server import serve_rpc

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    settings: Settings = app.state.settings
    container: Container = app.state.container

    ask_ai_usecase = container.ask_ai()
    rpc_server = await serve_rpc(
        port=settings.GRPC_PORT,
        ask_ai_usecase=ask_ai_usecase,
        auth_token=settings.INTERNAL_GRPC_AUTH_TOKEN or None,
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
