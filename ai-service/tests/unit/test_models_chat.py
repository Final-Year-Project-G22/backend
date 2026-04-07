from __future__ import annotations

import uuid
from datetime import UTC, date, datetime

import pytest
from pydantic import ValidationError

from core.domain.enums import Language, MessageType, SessionStatus, Tier
from core.domain.models import AIChatMessage, AIConversationSession, AIUserQuota
from core.domain.value_objects import TokenUsage


def test_user_quota_accepts_valid_payload() -> None:
    quota = AIUserQuota(
        user_id=uuid.uuid4(),
        tier=Tier.PRO,
        daily_query_count=5,
        daily_token_count=500,
        daily_conversations_started=2,
        daily_query_limit=20,
        daily_token_limit=5000,
        max_conversations_per_day=5,
        total_queries_used=100,
        total_tokens_used=10000,
        total_conversations=30,
        last_query_date=date(2026, 4, 6),
        last_query_at=datetime(2026, 4, 6, 12, 0, tzinfo=UTC),
    )

    assert quota.tier is Tier.PRO


def test_user_quota_rejects_daily_query_count_over_limit() -> None:
    with pytest.raises(ValidationError):
        AIUserQuota(
            user_id=uuid.uuid4(),
            daily_query_count=11,
            daily_query_limit=10,
        )


def test_user_quota_rejects_mismatched_last_query_date_and_time() -> None:
    with pytest.raises(ValidationError):
        AIUserQuota(
            user_id=uuid.uuid4(),
            last_query_date=date(2026, 4, 5),
            last_query_at=datetime(2026, 4, 6, 9, 0, tzinfo=UTC),
        )


def test_conversation_session_requires_archived_status_when_deleted() -> None:
    with pytest.raises(ValidationError):
        AIConversationSession(
            user_id=uuid.uuid4(),
            title="Session",
            tier_at_start=Tier.BASIC,
            current_tier=Tier.BASIC,
            status=SessionStatus.ACTIVE,
            deleted_at=datetime(2026, 4, 6, tzinfo=UTC),
        )


def test_conversation_session_requires_last_message_at_when_message_count_positive() -> None:
    with pytest.raises(ValidationError):
        AIConversationSession(
            user_id=uuid.uuid4(),
            title="Session",
            tier_at_start=Tier.BASIC,
            current_tier=Tier.BASIC,
            message_count=1,
            last_message_at=None,
        )


def test_chat_message_requires_query_for_user_query_type() -> None:
    with pytest.raises(ValidationError):
        AIChatMessage(
            user_id=uuid.uuid4(),
            conversation_id=uuid.uuid4(),
            message_type=MessageType.USER_QUERY,
            user_query=None,
            message_order=1,
        )


def test_chat_message_requires_response_for_ai_response_type() -> None:
    with pytest.raises(ValidationError):
        AIChatMessage(
            user_id=uuid.uuid4(),
            conversation_id=uuid.uuid4(),
            message_type=MessageType.AI_RESPONSE,
            llm_response=None,
            message_order=1,
        )


def test_chat_message_rejects_non_finite_query_embedding() -> None:
    with pytest.raises(ValidationError):
        AIChatMessage(
            user_id=uuid.uuid4(),
            conversation_id=uuid.uuid4(),
            message_type=MessageType.USER_QUERY,
            user_query="How do I register?",
            query_language=Language.ENGLISH,
            query_embedding=[0.1, float("nan")],
            message_order=1,
        )


def test_chat_message_feedback_fields_must_be_set_together() -> None:
    with pytest.raises(ValidationError):
        AIChatMessage(
            user_id=uuid.uuid4(),
            conversation_id=uuid.uuid4(),
            message_type=MessageType.AI_RESPONSE,
            llm_response="Follow these steps...",
            user_feedback=1,
            feedback_at=None,
            message_order=2,
        )


def test_chat_message_accepts_valid_ai_response_payload() -> None:
    message = AIChatMessage(
        user_id=uuid.uuid4(),
        conversation_id=uuid.uuid4(),
        message_type=MessageType.AI_RESPONSE,
        llm_response="Follow these steps...",
        token_usage=TokenUsage(prompt_tokens=10, completion_tokens=20, total_tokens=30),
        model_used="gemini-1.5-flash",
        message_order=2,
    )

    assert message.message_type is MessageType.AI_RESPONSE
    assert message.token_usage is not None
