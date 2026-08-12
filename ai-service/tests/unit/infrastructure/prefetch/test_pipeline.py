from __future__ import annotations

import uuid
from unittest.mock import AsyncMock, MagicMock

import pytest

from core.domain.tools import ToolResult
from core.ports.intent_classifier import IntentClass
from core.ports.tool_registry import ToolRegistryPort
from infrastructure.prefetch.pipeline import PreFetchPipeline


def _make_registry(
    result: ToolResult | None = None, side_effect: Exception | None = None
) -> MagicMock:
    registry = MagicMock(spec=ToolRegistryPort)
    if side_effect:
        registry.execute_tool = AsyncMock(side_effect=side_effect)
    else:
        registry.execute_tool = AsyncMock(
            return_value=result
            or ToolResult(
                tool_name="tool",
                arguments={},
                result_text="Result data",
                success=True,
            )
        )
    return registry


@pytest.fixture
def pipeline() -> PreFetchPipeline:
    return PreFetchPipeline(tool_registry=_make_registry())


@pytest.mark.asyncio
async def test_pre_fetch_knowledge_only() -> None:
    reg = _make_registry(
        ToolResult(
            tool_name="search_knowledge_base",
            arguments={},
            result_text="Trade license requires registration",
            success=True,
        )
    )
    p = PreFetchPipeline(tool_registry=reg)

    results = await p.pre_fetch(
        intent=IntentClass.KNOWLEDGE,
        query="business license",
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )

    assert "kb" in results
    assert results["kb"] == "Trade license requires registration"
    assert "profile" not in results


@pytest.mark.asyncio
async def test_pre_fetch_can_skip_knowledge_base() -> None:
    reg = _make_registry()
    p = PreFetchPipeline(tool_registry=reg)

    results = await p.pre_fetch(
        intent=IntentClass.KNOWLEDGE,
        query="business license",
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
        include_kb=False,
    )

    assert results == {}
    reg.execute_tool.assert_not_awaited()


@pytest.mark.asyncio
async def test_pre_fetch_personal_only() -> None:
    reg = _make_registry(
        ToolResult(
            tool_name="get_user_profile",
            arguments={},
            result_text="Restaurant, Addis Ababa",
            success=True,
        )
    )
    p = PreFetchPipeline(tool_registry=reg)

    results = await p.pre_fetch(
        intent=IntentClass.PERSONAL,
        query="my profile",
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )

    assert "kb" not in results


@pytest.mark.asyncio
async def test_pre_fetch_mixed() -> None:
    reg = _make_registry(
        ToolResult(
            tool_name="search_knowledge_base",
            arguments={},
            result_text="Knowledge result",
            success=True,
        )
    )
    p = PreFetchPipeline(tool_registry=reg)

    results = await p.pre_fetch(
        intent=IntentClass.MIXED,
        query="everything",
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )

    assert isinstance(results, dict)


@pytest.mark.asyncio
async def test_pre_fetch_handles_tool_failure() -> None:
    reg = _make_registry(side_effect=RuntimeError("Tool failed"))
    p = PreFetchPipeline(tool_registry=reg)

    results = await p.pre_fetch(
        intent=IntentClass.KNOWLEDGE,
        query="test",
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )

    assert results.get("kb") is None


@pytest.mark.asyncio
async def test_pre_fetch_personal_calls_multiple_tools() -> None:
    calls: list[str] = []

    async def tracking_execute(
        name: str, arguments: dict, account_id: str, user_id: str
    ) -> ToolResult:
        calls.append(name)
        return ToolResult(tool_name=name, arguments={}, result_text="data", success=True)

    registry = MagicMock(spec=ToolRegistryPort)
    registry.execute_tool = tracking_execute
    p = PreFetchPipeline(tool_registry=registry)

    await p.pre_fetch(
        intent=IntentClass.PERSONAL,
        query="my status",
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )

    assert "get_user_profile" in calls
