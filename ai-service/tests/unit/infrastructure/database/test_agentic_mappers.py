from __future__ import annotations

import uuid
from datetime import UTC, datetime

from core.domain.enums import MessageType
from core.domain.models import AIChatMessage, ToolCallRecord
from core.domain.value_objects import TokenUsage
from infrastructure.database.repositories.mappers import (
    _deserialize_tool_calls,
    serialize_tool_calls,
    to_domain_message,
    to_orm_message,
)


def test_serialize_tool_calls_none() -> None:
    assert serialize_tool_calls(None) is None


def test_serialize_tool_calls_empty_list() -> None:
    result = serialize_tool_calls([])
    assert result == []


def test_serialize_tool_calls() -> None:
    tc = ToolCallRecord(
        tool_name="search_knowledge_base",
        arguments={"query": "license"},
        result_summary="3 results",
        success=True,
        execution_ms=150,
        iteration=1,
    )

    result = serialize_tool_calls([tc])
    assert isinstance(result, list)
    assert len(result) == 1
    assert isinstance(result[0], dict)
    assert result[0]["tool_name"] == "search_knowledge_base"
    assert result[0]["arguments"] == {"query": "license"}


def test_deserialize_tool_calls_none() -> None:
    assert _deserialize_tool_calls(None) is None


def test_deserialize_tool_calls_empty_list() -> None:
    result = _deserialize_tool_calls([])
    assert result == []


def test_tool_calls_round_trip() -> None:
    tc = ToolCallRecord(
        tool_name="find_template",
        arguments={"query": "tax form"},
        result_summary="Found 2 templates",
        success=True,
        error_message=None,
        execution_ms=200,
        iteration=3,
    )

    serialized = serialize_tool_calls([tc])
    deserialized = _deserialize_tool_calls(serialized)

    assert deserialized is not None
    assert len(deserialized) == 1
    assert deserialized[0] == tc


def test_tool_calls_round_trip_with_error() -> None:
    tc = ToolCallRecord(
        tool_name="search_guides",
        arguments={"keyword": "restaurant"},
        result_summary="",
        success=False,
        error_message="Timeout",
        execution_ms=5000,
        iteration=1,
    )

    serialized = serialize_tool_calls([tc])
    deserialized = _deserialize_tool_calls(serialized)

    assert deserialized is not None
    assert len(deserialized) == 1
    assert deserialized[0].success is False
    assert deserialized[0].error_message == "Timeout"


def test_message_mapper_round_trip_with_tool_calls() -> None:
    now = datetime(2026, 4, 7, 10, 30, tzinfo=UTC)
    tool_call = ToolCallRecord(
        tool_name="search_knowledge_base",
        arguments={"query": "trade license"},
        result_summary="Found 3 relevant documents",
        success=True,
        execution_ms=120,
        iteration=1,
    )
    domain_message = AIChatMessage(
        user_id=uuid.uuid4(),
        conversation_id=uuid.uuid4(),
        message_type=MessageType.AI_RESPONSE,
        llm_response="Here is the result.",
        token_usage=TokenUsage(prompt_tokens=10, completion_tokens=5, total_tokens=15),
        model_used="command-a-03-2025",
        tool_calls=[tool_call],
        agent_strategy="agentic",
        message_order=3,
        created_at=now,
        updated_at=now,
    )

    orm_message = to_orm_message(domain_message)
    mapped_back = to_domain_message(orm_message)

    assert mapped_back == domain_message
    assert mapped_back.tool_calls is not None
    assert len(mapped_back.tool_calls) == 1
    assert mapped_back.tool_calls[0].tool_name == "search_knowledge_base"
    assert mapped_back.agent_strategy == "agentic"


def test_message_mapper_defaults_tool_calls_to_none() -> None:
    domain_message = AIChatMessage(
        user_id=uuid.uuid4(),
        conversation_id=uuid.uuid4(),
        message_type=MessageType.AI_RESPONSE,
        llm_response="Hello!",
        message_order=1,
    )

    orm_message = to_orm_message(domain_message)
    mapped_back = to_domain_message(orm_message)

    assert mapped_back.tool_calls is None
    assert mapped_back.agent_strategy == "simple"
