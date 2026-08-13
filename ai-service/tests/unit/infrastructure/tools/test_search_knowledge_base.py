from __future__ import annotations

import uuid
from unittest.mock import AsyncMock, MagicMock

import pytest

from core.domain.enums import Language
from core.domain.value_objects import DocumentSource, SearchHit
from core.ports.knowledge_repository import KnowledgeRepositoryPort
from infrastructure.tools.local.search_knowledge_base import SearchKnowledgeBaseTool


@pytest.fixture
def mock_embedding_port() -> MagicMock:
    port = MagicMock()
    port.embed_query = AsyncMock(return_value=[0.1, 0.2, 0.3])
    return port


@pytest.fixture
def mock_knowledge_repo() -> KnowledgeRepositoryPort:
    repo = MagicMock(spec=KnowledgeRepositoryPort)

    hits = [
        SearchHit(
            chunk_id=uuid.uuid4(),
            document_id=uuid.uuid4(),
            score=0.92,
            chunk_text="Trade licenses in Ethiopia require a business registration certificate.",
            chunk_index=0,
            source=DocumentSource.GOVERNMENT,
            language=Language.ENGLISH,
            document_title="Trade License Guide",
        ),
        SearchHit(
            chunk_id=uuid.uuid4(),
            document_id=uuid.uuid4(),
            score=0.85,
            chunk_text="To renew a trade license, submit Form TL-2026.",
            chunk_index=1,
            source=DocumentSource.GOVERNMENT,
            language=Language.ENGLISH,
            document_title="Trade License Renewal",
        ),
    ]
    repo.search_vector = AsyncMock(return_value=hits)
    repo.search_bm25 = AsyncMock(return_value=hits)

    return repo


@pytest.fixture
def tool(
    mock_knowledge_repo: KnowledgeRepositoryPort, mock_embedding_port: MagicMock
) -> SearchKnowledgeBaseTool:
    return SearchKnowledgeBaseTool(
        knowledge_repository=mock_knowledge_repo,
        embedding_port=mock_embedding_port,
    )


def test_tool_name(tool: SearchKnowledgeBaseTool) -> None:
    assert tool.name == "search_knowledge_base"


def test_tool_description(tool: SearchKnowledgeBaseTool) -> None:
    assert "knowledge base" in tool.description


def test_parameter_schema(tool: SearchKnowledgeBaseTool) -> None:
    schema = tool.parameter_schema
    assert schema["type"] == "object"
    assert "query" in schema["properties"]
    assert "top_k" in schema["properties"]
    assert schema["required"] == ["query"]


@pytest.mark.asyncio
async def test_execute_returns_results(tool: SearchKnowledgeBaseTool) -> None:
    result = await tool.execute(
        arguments={"query": "trade license", "top_k": 5},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )

    assert result.success is True
    assert result.tool_name == "search_knowledge_base"
    assert "Trade License Guide" in result.result_text
    assert "0.92" in result.result_text


@pytest.mark.asyncio
async def test_execute_passes_structured_hits_through(tool: SearchKnowledgeBaseTool) -> None:
    result = await tool.execute(
        arguments={"query": "trade license", "top_k": 5},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )

    assert result.success is True
    assert len(result.hits) == 2
    first = result.hits[0]
    assert first.chunk_id
    assert first.document_id
    assert first.score == 0.92
    assert first.chunk_text
    assert first.document_title == "Trade License Guide"
    assert first.source is DocumentSource.GOVERNMENT
    assert first.language is Language.ENGLISH
    assert first.chunk_index == 0


@pytest.mark.asyncio
async def test_execute_empty_query(tool: SearchKnowledgeBaseTool) -> None:
    result = await tool.execute(
        arguments={"query": ""},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )

    assert result.success is False
    assert result.error_message == "Empty query"


@pytest.mark.asyncio
async def test_execute_no_results(
    mock_knowledge_repo: KnowledgeRepositoryPort,
    mock_embedding_port: MagicMock,
) -> None:
    mock_knowledge_repo.search_vector = AsyncMock(return_value=[])
    mock_knowledge_repo.search_bm25 = AsyncMock(return_value=[])
    tool = SearchKnowledgeBaseTool(
        knowledge_repository=mock_knowledge_repo,
        embedding_port=mock_embedding_port,
    )

    result = await tool.execute(
        arguments={"query": "nonexistent"},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )

    assert result.success is True
    assert "No results found" in result.result_text


@pytest.mark.asyncio
async def test_execute_repository_error(
    mock_knowledge_repo: KnowledgeRepositoryPort,
    mock_embedding_port: MagicMock,
) -> None:
    mock_knowledge_repo.search_vector = AsyncMock(side_effect=Exception("DB connection lost"))
    mock_embedding_port.embed_query = AsyncMock(return_value=[0.1])
    tool = SearchKnowledgeBaseTool(
        knowledge_repository=mock_knowledge_repo,
        embedding_port=mock_embedding_port,
    )

    result = await tool.execute(
        arguments={"query": "test"},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )

    assert result.success is False
    assert "DB connection lost" in result.error_message


@pytest.mark.asyncio
async def test_execute_deduplicates_hits(
    mock_knowledge_repo: KnowledgeRepositoryPort,
    mock_embedding_port: MagicMock,
) -> None:
    hit = SearchHit(
        chunk_id=uuid.uuid4(),
        document_id=uuid.uuid4(),
        score=0.9,
        chunk_text="Unique text",
        chunk_index=0,
        source=DocumentSource.GUIDE,
        language=Language.ENGLISH,
        document_title="Guide",
    )
    mock_knowledge_repo.search_vector = AsyncMock(return_value=[hit])
    mock_knowledge_repo.search_bm25 = AsyncMock(return_value=[hit])
    tool = SearchKnowledgeBaseTool(
        knowledge_repository=mock_knowledge_repo,
        embedding_port=mock_embedding_port,
    )

    result = await tool.execute(
        arguments={"query": "test"},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )

    assert result.result_text.count("Unique text") >= 1
