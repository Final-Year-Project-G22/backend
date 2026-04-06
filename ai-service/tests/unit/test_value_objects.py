from __future__ import annotations

import uuid
from datetime import date

import pytest
from pydantic import ValidationError

from core.domain.enums import DocumentSource, Language, Tier
from core.domain.value_objects import (
    ResponseSource,
    SearchFilters,
    SearchHit,
    TokenUsage,
    UsageSnapshot,
)


def test_token_usage_accepts_valid_totals() -> None:
    expected_total = 20
    usage = TokenUsage(prompt_tokens=12, completion_tokens=8, total_tokens=expected_total)

    assert usage.total_tokens == expected_total


def test_token_usage_rejects_invalid_totals() -> None:
    with pytest.raises(ValidationError):
        TokenUsage(prompt_tokens=12, completion_tokens=8, total_tokens=19)


def test_response_source_requires_positive_score_when_provided() -> None:
    doc_id = uuid.uuid4()
    expected_score = 0.83

    source = ResponseSource(
        source=DocumentSource.GUIDE,
        document_id=doc_id,
        title="Tax registration guide",
        score=expected_score,
    )

    assert source.document_id == doc_id
    assert source.score == expected_score


def test_search_filters_rejects_empty_sources() -> None:
    with pytest.raises(ValidationError):
        SearchFilters(sources=[])


def test_search_filters_deduplicates_sources() -> None:
    filters = SearchFilters(
        sources=[DocumentSource.GUIDE, DocumentSource.GUIDE, DocumentSource.FAQ],
        language=Language.ENGLISH,
        effective_on=date(2026, 1, 1),
    )

    assert filters.sources == [DocumentSource.GUIDE, DocumentSource.FAQ]


def test_search_hit_validates_non_negative_chunk_index() -> None:
    with pytest.raises(ValidationError):
        SearchHit(
            chunk_id=uuid.uuid4(),
            document_id=uuid.uuid4(),
            score=0.9,
            chunk_text="valid text",
            chunk_index=-1,
            source=DocumentSource.LEGAL,
            language=Language.AMHARIC,
        )


def test_usage_snapshot_rejects_daily_count_over_limit() -> None:
    with pytest.raises(ValidationError):
        UsageSnapshot(
            tier=Tier.BASIC,
            daily_query_count=11,
            daily_token_count=200,
            daily_conversations_started=1,
            daily_query_limit=10,
            daily_token_limit=500,
            max_conversations_per_day=2,
            total_queries_used=100,
            total_tokens_used=5000,
            total_conversations=20,
        )


def test_usage_snapshot_accepts_valid_limits() -> None:
    snapshot = UsageSnapshot(
        tier=Tier.PRO,
        daily_query_count=3,
        daily_token_count=300,
        daily_conversations_started=2,
        daily_query_limit=20,
        daily_token_limit=1000,
        max_conversations_per_day=4,
        total_queries_used=200,
        total_tokens_used=10000,
        total_conversations=40,
    )

    assert snapshot.tier == Tier.PRO
