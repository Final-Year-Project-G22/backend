from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from core.ports.embedding import EmbeddingPort
from core.usecases.strategies.kb_query_guard import KBQueryGuard, SuppressionReason

PROMPT = "plc registration process"
PROMPT_EMBEDDING = [1.0, 0.0]


def _make_guard(
    embedding_map: dict[str, list[float] | None],
    *,
    prompt_covered: bool = True,
    **kwargs: object,
) -> KBQueryGuard:
    embedding_port = MagicMock(spec=EmbeddingPort)

    async def fake_embed(query: str) -> list[float] | None:
        return embedding_map.get(query)

    embedding_port.embed_query = AsyncMock(side_effect=fake_embed)
    guard = KBQueryGuard(embedding_port, **kwargs)
    guard.start_turn(PROMPT, PROMPT_EMBEDDING, prompt_covered=prompt_covered)
    return guard


@pytest.mark.asyncio
async def test_exact_normalized_duplicate_of_prompt_is_suppressed() -> None:
    guard = _make_guard({})

    decision = await guard.check("  PLC   Registration PROCESS \n")

    assert decision.allowed is False
    assert decision.reason is SuppressionReason.DUPLICATE_OF_PROMPT
    assert decision.matched_query == PROMPT


@pytest.mark.asyncio
async def test_exact_prompt_query_runs_when_prompt_not_covered() -> None:
    guard = _make_guard({}, prompt_covered=False)

    decision = await guard.check("PLC Registration Process")

    assert decision.allowed is True
    assert decision.reason is None


@pytest.mark.asyncio
async def test_case_and_whitespace_variant_of_executed_query_is_suppressed() -> None:
    guard = _make_guard({"Ethiopian government business registration process": [0.6, 0.8]})

    first = await guard.check("Ethiopian government business registration process")
    assert first.allowed is True
    guard.register_executed("Ethiopian government business registration process", first.embedding)

    second = await guard.check("ETHIOPIAN GOVERNMENT  BUSINESS Registration process")
    assert second.allowed is False
    assert second.reason is SuppressionReason.DUPLICATE_OF_PRIOR_SEARCH
    assert second.matched_query == "Ethiopian government business registration process"


@pytest.mark.asyncio
async def test_cosine_duplicate_of_prompt_is_suppressed() -> None:
    guard = _make_guard({"plc registration steps": [0.95, 0.3122]})

    decision = await guard.check("plc registration steps")

    assert decision.allowed is False
    assert decision.reason is SuppressionReason.DUPLICATE_OF_PROMPT
    assert decision.matched_query == PROMPT


@pytest.mark.asyncio
async def test_ambiguous_band_tiebreak_by_jaccard_suppresses() -> None:
    guard = _make_guard({"PLC registration": [0.85, 0.5268]})

    decision = await guard.check("PLC registration")

    assert decision.allowed is False
    assert decision.reason is SuppressionReason.DUPLICATE_OF_PROMPT


@pytest.mark.asyncio
async def test_ambiguous_band_with_low_overlap_runs() -> None:
    guard = _make_guard({"value added tax filing": [0.85, 0.5268]})

    decision = await guard.check("value added tax filing")

    assert decision.allowed is True
    assert decision.reason is None


@pytest.mark.asyncio
async def test_jaccard_boundary_is_inclusive() -> None:
    guard = _make_guard({"a b c": [0.85, 0.5268]})
    guard.start_turn("a b c d e", [1.0, 0.0], prompt_covered=True)

    decision = await guard.check("a b c")

    assert decision.allowed is False
    assert decision.reason is SuppressionReason.DUPLICATE_OF_PROMPT


@pytest.mark.asyncio
async def test_drifted_query_is_suppressed_even_when_prompt_not_covered() -> None:
    guard = _make_guard({"how to bake bread": [0.0, 1.0]}, prompt_covered=False)

    decision = await guard.check("how to bake bread")

    assert decision.allowed is False
    assert decision.reason is SuppressionReason.DRIFT
    assert decision.matched_query == PROMPT


@pytest.mark.asyncio
async def test_distinct_follow_up_below_ambiguous_band_runs() -> None:
    guard = _make_guard({"Ethiopian government business registration process": [0.6, 0.8]})

    decision = await guard.check("Ethiopian government business registration process")

    assert decision.allowed is True
    assert decision.reason is None
    assert decision.embedding == [0.6, 0.8]


@pytest.mark.asyncio
async def test_cosine_duplicate_of_executed_query_is_suppressed() -> None:
    guard = _make_guard(
        {
            "broadened registration query": [0.6, 0.8],
            "reworded registration lookup": [0.6, 0.8],
        }
    )

    first = await guard.check("broadened registration query")
    guard.register_executed("broadened registration query", first.embedding)

    second = await guard.check("reworded registration lookup")
    assert second.allowed is False
    assert second.reason is SuppressionReason.DUPLICATE_OF_PRIOR_SEARCH
    assert second.matched_query == "broadened registration query"


@pytest.mark.asyncio
async def test_embedding_failure_is_conservative_and_runs() -> None:
    guard = _make_guard({"unknown query": None})

    decision = await guard.check("unknown query")

    assert decision.allowed is True
    assert decision.embedding is None


@pytest.mark.asyncio
async def test_reworded_prior_search_is_duplicate_not_drift() -> None:
    guard = _make_guard(
        {
            "ethiopian registration": [0.6, 0.8],
            "ethiopian registration reworded": [0.0886, 0.996],
        }
    )

    first = await guard.check("ethiopian registration")
    guard.register_executed("ethiopian registration", first.embedding)

    second = await guard.check("ethiopian registration reworded")

    assert second.allowed is False
    assert second.reason is SuppressionReason.DUPLICATE_OF_PRIOR_SEARCH
    assert second.matched_query == "ethiopian registration"


@pytest.mark.asyncio
async def test_start_turn_resets_executed_queries() -> None:
    guard = _make_guard({"first turn query": [0.6, 0.8]})

    first = await guard.check("first turn query")
    guard.register_executed("first turn query", first.embedding)

    guard.start_turn(PROMPT, PROMPT_EMBEDDING, prompt_covered=True)

    second = await guard.check("first turn query")
    assert second.allowed is True
