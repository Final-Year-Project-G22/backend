from __future__ import annotations

import asyncio
import socket
import uuid
from collections.abc import AsyncIterator
from unittest.mock import AsyncMock

import grpc
import grpc.aio
import pytest
from ai.conversation.v1 import service_pb2 as conversation_pb2
from ai.conversation.v1 import service_pb2_grpc as conversation_grpc
from ai.inference.v1 import service_pb2 as inference_pb2
from ai.inference.v1 import service_pb2_grpc as inference_grpc
from dependency_injector import providers
from sqlalchemy import text

from app.config import Settings
from app.container import Container
from core.ports.embedding import EmbeddingPort
from core.ports.llm import LLMChunk, LLMPort, LLMResult
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
async def test_failed_conversation_creation_rolls_back_quota_over_transport(
    rpc_graph: tuple[Container, grpc.aio.Channel],
) -> None:
    container, channel = rpc_graph
    async with container.database().engine.begin() as connection:
        await connection.execute(
            text(
                "ALTER TABLE ai_conversation_sessions "
                "ADD CONSTRAINT reject_test_title CHECK (title <> 'rollback')"
            )
        )
    user_id, account_id = str(uuid.uuid4()), str(uuid.uuid4())
    inference = inference_grpc.AIInferenceServiceStub(channel)
    conversation = conversation_grpc.AIConversationServiceStub(channel)
    with pytest.raises(grpc.aio.AioRpcError) as failure:
        await inference.Ask(
            inference_pb2.AskRequest(
                request_id=str(uuid.uuid4()),
                user_id=user_id,
                account_id=account_id,
                query="Any query",
                title="rollback",
                strategy="simple",
            )
        )
    assert failure.value.code() is grpc.StatusCode.INTERNAL
    result = await conversation.ListConversations(
        conversation_pb2.ListConversationsRequest(
            user_id=user_id,
            account_id=account_id,
        )
    )
    assert not result.sessions
    async with container.database().command() as reader:
        assert await reader.quota.get_quota(uuid.UUID(user_id)) is None


@pytest.mark.integration
async def test_failed_ask_keeps_submitted_message_but_can_retry(
    rpc_graph: tuple[Container, grpc.aio.Channel],
) -> None:
    container, channel = rpc_graph
    llm = container.llm_port()
    llm.generate.side_effect = RuntimeError("provider unavailable")
    inference = inference_grpc.AIInferenceServiceStub(channel)
    conversation = conversation_grpc.AIConversationServiceStub(channel)
    user_id, account_id = str(uuid.uuid4()), str(uuid.uuid4())
    request = inference_pb2.AskRequest(
        request_id=str(uuid.uuid4()),
        user_id=user_id,
        account_id=account_id,
        query="How do I register?",
        strategy="simple",
    )
    with pytest.raises(grpc.aio.AioRpcError) as failure:
        await inference.Ask(request)
    assert failure.value.code() is grpc.StatusCode.INTERNAL
    sessions = await conversation.ListConversations(
        conversation_pb2.ListConversationsRequest(
            user_id=user_id,
            account_id=account_id,
        )
    )
    assert len(sessions.sessions) == 1
    session_id = sessions.sessions[0].id
    before = await conversation.GetConversation(
        conversation_pb2.GetConversationRequest(
            session_id=session_id,
            account_id=account_id,
        )
    )
    assert [message.role for message in before.messages] == ["user"]
    llm.generate.side_effect = None
    response = await inference.Ask(
        inference_pb2.AskRequest(
            request_id=str(uuid.uuid4()),
            user_id=user_id,
            account_id=account_id,
            session_id=session_id,
            query="Try again",
            strategy="simple",
        )
    )
    assert response.answer == "Register the business first"


@pytest.mark.integration
async def test_parallel_local_tools_do_not_share_a_session(
    rpc_graph: tuple[Container, grpc.aio.Channel],
) -> None:
    container, channel = rpc_graph
    registry = container.tool_registry()
    stub = inference_grpc.AIInferenceServiceStub(channel)
    user_id, account_id = str(uuid.uuid4()), str(uuid.uuid4())

    async def ask() -> inference_pb2.AskResponse:
        return await stub.Ask(
            inference_pb2.AskRequest(
                request_id=str(uuid.uuid4()),
                user_id=user_id,
                account_id=account_id,
                query="How do I register?",
                strategy="simple",
            )
        )

    first, second, tool_one, tool_two = await asyncio.wait_for(
        asyncio.gather(
            ask(),
            ask(),
            registry.execute_tool(
                "search_knowledge_base", {"query": "registration"}, account_id, user_id
            ),
            registry.execute_tool(
                "search_knowledge_base", {"query": "licensing"}, account_id, user_id
            ),
        ),
        10,
    )
    assert first.answer == second.answer == "Register the business first"
    assert tool_one.success
    assert tool_two.success


@pytest.mark.integration
async def test_cancelling_after_first_chunk_closes_provider_stream(
    rpc_graph: tuple[Container, grpc.aio.Channel],
) -> None:
    container, channel = rpc_graph
    closed = asyncio.Event()
    llm = container.llm_port()

    async def provider_stream(*_args: object, **_kwargs: object) -> AsyncIterator[LLMChunk]:
        try:
            yield LLMChunk(text="first")
            await asyncio.Event().wait()
        finally:
            closed.set()

    llm.generate_stream = provider_stream
    stub = inference_grpc.AIInferenceServiceStub(channel)
    call = stub.AskStream(
        inference_pb2.AskRequest(
            user_id=str(uuid.uuid4()),
            account_id=str(uuid.uuid4()),
            query="What next?",
            strategy="simple",
        )
    )
    chunk = await asyncio.wait_for(call.read(), 5)
    assert chunk.text.text == "first"
    call.cancel()
    await asyncio.wait_for(closed.wait(), 5)
    assert container.database().engine.pool.checkedout() == 0


@pytest.mark.integration
async def test_cancelled_stream_releases_database_before_model_finishes(
    rpc_graph: tuple[Container, grpc.aio.Channel],
) -> None:
    container, channel = rpc_graph
    started, resume = asyncio.Event(), asyncio.Event()
    llm = container.llm_port()

    async def slow_stream(*_args: object, **_kwargs: object) -> AsyncIterator[LLMChunk]:
        started.set()
        await resume.wait()
        yield LLMChunk(text="Too late")

    llm.generate_stream = slow_stream
    inference = inference_grpc.AIInferenceServiceStub(channel)
    conversation = conversation_grpc.AIConversationServiceStub(channel)
    user_id, account_id = str(uuid.uuid4()), str(uuid.uuid4())
    call = inference.AskStream(
        inference_pb2.AskRequest(
            user_id=user_id,
            account_id=account_id,
            query="How?",
            strategy="simple",
        )
    )
    pending = asyncio.create_task(call.read())
    try:
        await asyncio.wait_for(started.wait(), 5)
        assert container.database().engine.pool.checkedout() == 0
        call.cancel()
        with pytest.raises(asyncio.CancelledError):
            await pending
        sessions = await conversation.ListConversations(
            conversation_pb2.ListConversationsRequest(
                user_id=user_id,
                account_id=account_id,
            )
        )
        assert len(sessions.sessions) == 1
    finally:
        resume.set()
        if not pending.done():
            pending.cancel()
            await asyncio.gather(pending, return_exceptions=True)


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


@pytest.mark.integration
async def test_startup_failure_disposes_every_created_resource(database: Database) -> None:
    app = create_app(Settings())
    container = app.state.container
    container.database.override(database)
    embedding = AsyncMock(spec=EmbeddingPort)
    llm = AsyncMock(spec=LLMPort)
    container.embedding_port.override(embedding)
    container.llm_port.override(llm)
    client = AsyncMock()
    client.list_tools.return_value = []
    embedding.embed_documents.side_effect = RuntimeError("unavailable")
    container.ai_tool_client.override(client)
    with pytest.raises(RuntimeError, match="unavailable"):
        async with app.router.lifespan_context(app):
            pass
    client.close.assert_awaited_once()
    assert app.state.http_client.is_closed
    assert database.engine.pool.checkedout() == 0
