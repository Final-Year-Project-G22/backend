from __future__ import annotations

import uuid

import pytest
from pydantic import ValidationError

from core.domain.enums import MessageType
from core.domain.models import AIChatMessage, ToolCallRecord


def test_tool_call_record_accepts_valid_payload() -> None:
    tc = ToolCallRecord(
        tool_name="search_knowledge_base",
        arguments={"query": "trade license renewal"},
        result_summary="Found 3 relevant documents",
        success=True,
        execution_ms=150,
        iteration=2,
    )

    assert tc.tool_name == "search_knowledge_base"
    assert tc.arguments["query"] == "trade license renewal"
    assert tc.result_summary == "Found 3 relevant documents"
    assert tc.success is True
    assert tc.error_message is None
    assert tc.execution_ms == 150
    assert tc.iteration == 2


def test_tool_call_record_defaults() -> None:
    tc = ToolCallRecord(tool_name="get_user_profile")

    assert tc.arguments == {}
    assert tc.result_summary == ""
    assert tc.success is True
    assert tc.error_message is None
    assert tc.execution_ms == 0
    assert tc.iteration == 1


def test_tool_call_record_rejects_empty_tool_name() -> None:
    with pytest.raises(ValidationError):
        ToolCallRecord(tool_name="")


def test_tool_call_record_rejects_negative_execution_ms() -> None:
    with pytest.raises(ValidationError):
        ToolCallRecord(tool_name="search_knowledge_base", execution_ms=-1)


def test_tool_call_record_rejects_zero_iteration() -> None:
    with pytest.raises(ValidationError):
        ToolCallRecord(tool_name="search_knowledge_base", iteration=0)


def test_chat_message_accepts_tool_calls_and_strategy() -> None:
    tool_call = ToolCallRecord(
        tool_name="search_knowledge_base",
        arguments={"query": "license"},
        result_summary="3 results",
    )
    message = AIChatMessage(
        user_id=uuid.uuid4(),
        conversation_id=uuid.uuid4(),
        message_type=MessageType.AI_RESPONSE,
        llm_response="Here is what I found.",
        tool_calls=[tool_call],
        agent_strategy="agentic",
        message_order=2,
    )

    assert message.tool_calls is not None
    assert len(message.tool_calls) == 1
    assert message.tool_calls[0].tool_name == "search_knowledge_base"
    assert message.agent_strategy == "agentic"


def test_chat_message_defaults_strategy_to_simple() -> None:
    message = AIChatMessage(
        user_id=uuid.uuid4(),
        conversation_id=uuid.uuid4(),
        message_type=MessageType.AI_RESPONSE,
        llm_response="Hello!",
        message_order=2,
    )

    assert message.agent_strategy == "simple"
    assert message.tool_calls is None


def test_chat_message_tool_calls_none_handling() -> None:
    message = AIChatMessage(
        user_id=uuid.uuid4(),
        conversation_id=uuid.uuid4(),
        message_type=MessageType.AI_RESPONSE,
        llm_response="No tools used.",
        tool_calls=None,
        agent_strategy="simple",
        message_order=2,
    )

    assert message.tool_calls is None


def test_tool_call_record_is_frozen() -> None:
    tc = ToolCallRecord(tool_name="test")

    with pytest.raises(ValidationError):
        tc.tool_name = "changed"
