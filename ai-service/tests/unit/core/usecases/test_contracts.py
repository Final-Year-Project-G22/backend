from __future__ import annotations

import uuid

import pytest
from pydantic import ValidationError

from core.domain.enums import Language
from core.usecases.contracts import AskAICommand, CreateSessionCommand, ListSessionsQuery
from core.usecases.defaults import (
    DEFAULT_BM25_TOP_K,
    DEFAULT_SESSION_LIST_LIMIT,
    DEFAULT_VECTOR_TOP_K,
)


def test_create_session_command_strips_title() -> None:
    command = CreateSessionCommand(
        user_id=uuid.uuid4(),
        title="  Registration support  ",
        language=Language.AMHARIC,
    )

    assert command.title == "Registration support"
    assert command.language is Language.AMHARIC


def test_ask_ai_command_uses_default_retrieval_values() -> None:
    command = AskAICommand(
        user_id=uuid.uuid4(),
        prompt="How can I start a business?",
    )

    assert command.vector_top_k == DEFAULT_VECTOR_TOP_K
    assert command.bm25_top_k == DEFAULT_BM25_TOP_K


def test_ask_ai_command_rejects_blank_prompt() -> None:
    with pytest.raises(ValidationError, match="prompt cannot be empty"):
        AskAICommand(
            user_id=uuid.uuid4(),
            prompt="   ",
        )


def test_list_sessions_query_defaults() -> None:
    query = ListSessionsQuery(user_id=uuid.uuid4())

    assert query.limit == DEFAULT_SESSION_LIST_LIMIT
    assert query.offset == 0
    assert query.include_deleted is False
