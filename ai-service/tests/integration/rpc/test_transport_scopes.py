from __future__ import annotations

import asyncio
import socket
import uuid
from collections.abc import AsyncIterator
from unittest.mock import AsyncMock

import grpc.aio
import pytest
from ai.conversation.v1 import service_pb2 as conversation_pb2
from ai.conversation.v1 import service_pb2_grpc as conversation_grpc
from ai.inference.v1 import service_pb2 as inference_pb2
from ai.inference.v1 import service_pb2_grpc as inference_grpc
from dependency_injector import providers

from app.config import Settings
from app.container import Container
from core.ports.embedding import EmbeddingPort
from core.ports.llm import LLMPort, LLMResult
from infrastructure.database.connection import Database
from infrastructure.rpc.server import serve_rpc
from main import create_app


@pytest.fixture
async def rpc_graph(database: Database) -> AsyncIterator[tuple[Container, grpc.aio.Channel]]:
    settings = Settings(CORE_GRPC_ENDPOINT="localhost:50052", AI_AGENTIC_ENABLED=False)
    container = Container(config=providers.Object(settings))
    container.database.override(database)
    embedding = AsyncMock(spec=EmbeddingPort)
    embedding.embed_query.return_value = [0.0] * 1024
    llm = AsyncMock(spec=LLMPort)
    llm.generate.return_value = LLMResult(text="Register the business first")
    llm.model = "fixture"
    container.embedding_port.override(embedding)
    container.llm_port.override(llm)
    container.core_service_port.override(None)
    container.event_bus_port.override(None)

    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        port = sock.getsockname()[1]
    server = await serve_rpc(
        port=port,
        ask_factory=container.ask_ai,
        conversation_factory=container.conversation,
    )
    async with grpc.aio.insecure_channel(f"127.0.0.1:{port}") as channel:
        try:
            yield container, channel
        finally:
            await server.stop(grace=0)
            await server.wait_for_termination()


@pytest.mark.integration
async def test_two_concurrent_ask_requests_persist_ordered_turns(
    rpc_graph: tuple[Container, grpc.aio.Channel],
) -> None:
    _, channel = rpc_graph
    inference = inference_grpc.AIInferenceServiceStub(channel)
    conversation = conversation_grpc.AIConversationServiceStub(channel)
    user_id, account_id = str(uuid.uuid4()), str(uuid.uuid4())

    def ask(session_id: str = "") -> inference_pb2.AskRequest:
        return inference_pb2.AskRequest(
            request_id=str(uuid.uuid4()),
            user_id=user_id,
            account_id=account_id,
            query="How do I register?",
            language="en",
            session_id=session_id,
            strategy="simple",
        )

    first = await inference.Ask(ask())
    a, b = await asyncio.wait_for(
        asyncio.gather(inference.Ask(ask(first.session_id)), inference.Ask(ask(first.session_id))),
        10,
    )
    assert a.answer == b.answer == "Register the business first"
    result = await conversation.GetConversation(
        conversation_pb2.GetConversationRequest(
            account_id=account_id,
            session_id=first.session_id,
            message_limit=20,
        )
    )
    assert [message.role for message in result.messages[:2]] == ["user", "assistant"]
    assert [message.role for message in result.messages].count("user") == 3
    assert [message.role for message in result.messages].count("assistant") == 3
    assert len({message.id for message in result.messages}) == 6


@pytest.mark.integration
async def test_running_app_uses_its_database_and_closes_it(database: Database) -> None:
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        port = sock.getsockname()[1]
    app = create_app(Settings(GRPC_PORT=port, AI_AGENTIC_ENABLED=False))
    container = app.state.container
    container.database.override(database)
    embedding = AsyncMock(spec=EmbeddingPort)
    embedding.embed_query.return_value = [0.0] * 1024
    llm = AsyncMock(spec=LLMPort)
    llm.generate.return_value = LLMResult(text="Answered")
    llm.model = "fixture"
    container.embedding_port.override(embedding)
    container.llm_port.override(llm)
    container.core_service_port.override(None)
    container.event_bus_port.override(None)
    remote = AsyncMock()
    remote.list_tools.return_value = []
    container.ai_tool_client.override(remote)

    async with (
        app.router.lifespan_context(app),
        grpc.aio.insecure_channel(f"127.0.0.1:{port}") as channel,
    ):
        stub = inference_grpc.AIInferenceServiceStub(channel)
        result = await stub.Ask(
            inference_pb2.AskRequest(
                request_id=str(uuid.uuid4()),
                user_id=str(uuid.uuid4()),
                account_id=str(uuid.uuid4()),
                query="What next?",
                strategy="simple",
            )
        )
        assert result.answer == "Answered"
    assert database.engine.pool.checkedout() == 0
