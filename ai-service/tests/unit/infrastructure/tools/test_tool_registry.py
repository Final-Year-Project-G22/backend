from __future__ import annotations

import json
import uuid
from unittest.mock import AsyncMock, MagicMock

import pytest

from core.domain.tools import ToolResult
from core.ports.llm import ToolDefinition
from infrastructure.rpc.ai_tool_client import AIToolGrpcClient
from infrastructure.tools.local.search_knowledge_base import SearchKnowledgeBaseTool
from infrastructure.tools.local.search_trusted_web import SearchTrustedWebTool
from infrastructure.tools.tool_registry import ToolRegistry


@pytest.fixture
def mock_local_tools() -> tuple[MagicMock, MagicMock]:
    kb = MagicMock(spec=SearchKnowledgeBaseTool)
    kb.name = "search_knowledge_base"
    kb.description = "Search the knowledge base"
    kb.parameter_schema = {
        "type": "object",
        "properties": {"query": {"type": "string"}},
        "required": ["query"],
    }
    kb.execute = AsyncMock(
        return_value=ToolResult(
            tool_name="search_knowledge_base",
            arguments={"query": "test"},
            result_text="Result from KB",
            success=True,
            execution_ms=50,
        )
    )

    web = MagicMock(spec=SearchTrustedWebTool)
    web.name = "search_trusted_web"
    web.description = "Search trusted web"
    web.parameter_schema = {
        "type": "object",
        "properties": {"url": {"type": "string"}},
        "required": ["url"],
    }
    web.execute = AsyncMock(
        return_value=ToolResult(
            tool_name="search_trusted_web",
            arguments={"url": "https://example.com"},
            result_text="Result from web",
            success=True,
            execution_ms=100,
        )
    )

    return kb, web


@pytest.fixture
def mock_grpc_client() -> MagicMock:
    client = MagicMock(spec=AIToolGrpcClient)
    client.list_tools = AsyncMock(
        return_value=[
            ToolDefinition(
                name="search_guides",
                description="Search business guides",
                parameter_schema_json='{"type":"object","properties":{"keyword":{"type":"string"}},"required":[]}',
            ),
            ToolDefinition(
                name="check_compliance_status",
                description="Check compliance",
                parameter_schema_json='{"type":"object","properties":{},"required":[]}',
            ),
        ]
    )
    client.execute_tool = AsyncMock(return_value={"result": "Guide search results"})
    return client


@pytest.fixture
async def registry(
    mock_local_tools: tuple[MagicMock, MagicMock],
    mock_grpc_client: MagicMock,
) -> ToolRegistry:
    kb, web = mock_local_tools
    reg = ToolRegistry(
        ai_tool_client=mock_grpc_client,
        expected_remote_tools=["search_guides", "check_compliance_status"],
        search_knowledge_base=kb,
        search_trusted_web=web,
    )
    await reg.initialize()
    return reg


@pytest.mark.asyncio
async def test_initialize_fetches_remote_tools(registry: ToolRegistry) -> None:
    defs = await registry.get_tool_definitions()
    names = [t.name for t in defs]
    assert "search_knowledge_base" in names
    assert "search_trusted_web" in names
    assert "search_guides" in names
    assert "check_compliance_status" in names


@pytest.mark.asyncio
async def test_initialize_filters_unexpected_remote_tools(
    mock_local_tools: tuple[MagicMock, MagicMock],
    mock_grpc_client: MagicMock,
) -> None:
    mock_grpc_client.list_tools = AsyncMock(
        return_value=[
            ToolDefinition(name="search_guides", description="", parameter_schema_json="{}"),
            ToolDefinition(name="list_sectors", description="", parameter_schema_json="{}"),
        ]
    )
    kb, web = mock_local_tools
    reg = ToolRegistry(
        ai_tool_client=mock_grpc_client,
        expected_remote_tools=["search_guides"],
        search_knowledge_base=kb,
        search_trusted_web=web,
    )
    await reg.initialize()
    defs = await reg.get_tool_definitions()
    names = [t.name for t in defs]
    assert "search_knowledge_base" in names
    assert "search_trusted_web" in names
    assert "search_guides" in names
    assert "list_sectors" not in names


@pytest.mark.asyncio
async def test_execute_local_tool(registry: ToolRegistry) -> None:
    result = await registry.execute_tool(
        name="search_knowledge_base",
        arguments={"query": "test"},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )
    assert result.success is True
    assert result.result_text == "Result from KB"


@pytest.mark.asyncio
async def test_execute_remote_tool(registry: ToolRegistry) -> None:
    result = await registry.execute_tool(
        name="search_guides",
        arguments={"keyword": "guide"},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )
    assert result.success is True
    assert "Guide search" in result.result_text


@pytest.mark.asyncio
async def test_execute_unknown_tool(registry: ToolRegistry) -> None:
    result = await registry.execute_tool(
        name="nonexistent_tool",
        arguments={},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )
    assert result.success is False
    assert "Unknown tool" in result.result_text


@pytest.mark.asyncio
async def test_execute_local_tool_failure(
    mock_local_tools: tuple[MagicMock, MagicMock],
    mock_grpc_client: MagicMock,
) -> None:
    kb, web = mock_local_tools
    kb.execute = AsyncMock(side_effect=RuntimeError("Internal error"))

    reg = ToolRegistry(
        ai_tool_client=mock_grpc_client,
        expected_remote_tools=[],
        search_knowledge_base=kb,
        search_trusted_web=web,
    )
    await reg.initialize()

    result = await reg.execute_tool(
        name="search_knowledge_base",
        arguments={"query": "test"},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )
    assert result.success is False
    assert "Internal error" in result.result_text


@pytest.mark.asyncio
async def test_execute_remote_tool_failure(
    mock_local_tools: tuple[MagicMock, MagicMock],
    mock_grpc_client: MagicMock,
) -> None:
    mock_grpc_client.execute_tool = AsyncMock(return_value={"error": "Guide search unavailable"})
    mock_grpc_client.list_tools = AsyncMock(
        return_value=[
            ToolDefinition(name="search_guides", description="", parameter_schema_json="{}"),
        ]
    )

    kb, web = mock_local_tools
    reg = ToolRegistry(
        ai_tool_client=mock_grpc_client,
        expected_remote_tools=["search_guides"],
        search_knowledge_base=kb,
        search_trusted_web=web,
    )
    await reg.initialize()

    result = await reg.execute_tool(
        name="search_guides",
        arguments={},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )
    assert result.success is False


@pytest.mark.asyncio
async def test_tool_definitions_provide_json_schema(
    mock_local_tools: tuple[MagicMock, MagicMock],
    mock_grpc_client: MagicMock,
) -> None:
    kb, web = mock_local_tools
    reg = ToolRegistry(
        ai_tool_client=mock_grpc_client,
        expected_remote_tools=[],
        search_knowledge_base=kb,
        search_trusted_web=web,
    )
    await reg.initialize()

    defs = await reg.get_tool_definitions()
    for d in defs:
        assert d.name
        assert d.description
        assert d.parameter_schema_json
        parsed = json.loads(d.parameter_schema_json)
        assert "type" in parsed
        assert "properties" in parsed
