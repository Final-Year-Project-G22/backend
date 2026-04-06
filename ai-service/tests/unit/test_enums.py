from __future__ import annotations

from core.domain.enums import (
    ChunkStatus,
    DocumentSource,
    DocumentStatus,
    Language,
    MessageType,
    SessionStatus,
    Tier,
)


def test_language_values() -> None:
    assert Language.ENGLISH.value == "en"
    assert Language.AMHARIC.value == "am"


def test_document_source_values() -> None:
    assert DocumentSource.GOVERNMENT.value == "government"
    assert DocumentSource.LEGAL.value == "legal"
    assert DocumentSource.TAX_CODE.value == "tax_code"
    assert DocumentSource.GUIDE.value == "guide"
    assert DocumentSource.STEP.value == "step"
    assert DocumentSource.FAQ.value == "faq"


def test_document_status_values() -> None:
    assert DocumentStatus.PROCESSING.value == "processing"
    assert DocumentStatus.ACTIVE.value == "active"
    assert DocumentStatus.ARCHIVED.value == "archived"
    assert DocumentStatus.FAILED.value == "failed"


def test_chunk_status_values() -> None:
    assert ChunkStatus.PENDING.value == "pending"
    assert ChunkStatus.EMBEDDED.value == "embedded"
    assert ChunkStatus.FAILED.value == "failed"


def test_tier_values() -> None:
    assert Tier.BASIC.value == "basic"
    assert Tier.PRO.value == "pro"
    assert Tier.PREMIUM.value == "premium"


def test_session_status_values() -> None:
    assert SessionStatus.ACTIVE.value == "active"
    assert SessionStatus.ARCHIVED.value == "archived"


def test_message_type_values() -> None:
    assert MessageType.USER_QUERY.value == "user_query"
    assert MessageType.AI_RESPONSE.value == "ai_response"
