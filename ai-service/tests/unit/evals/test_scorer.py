"""Item scorer and cell aggregation tests."""

from __future__ import annotations

from collections.abc import AsyncIterator
from typing import Any

import pytest

from core.ports.llm import LLMChunk, LLMPort, LLMResult, ToolDefinition
from evals.models import CellKey, ItemScores, JudgeVerdict, MetricValue, ToolCallTrace
from evals.scoring import EvalJudge, ItemScorer, aggregate_cell
from evals.scoring.judge import JudgeError
from tests.unit.evals.conftest import make_item, make_trace


class _NoopLLM(LLMPort):
    @property
    def provider(self) -> str:
        return "noop"

    @property
    def model(self) -> str:
        return "noop"

    async def generate(
        self,
        prompt: str,
        *,
        system_prompt: str | None = None,
        tools: list[ToolDefinition] | None = None,
        max_tokens: int = 1024,
        temperature: float = 0.2,
    ) -> LLMResult:
        return LLMResult(text="{}")

    def generate_stream(
        self,
        prompt: str,
        *,
        system_prompt: str | None = None,
        tools: list[ToolDefinition] | None = None,
        max_tokens: int = 1024,
        temperature: float = 0.2,
    ) -> AsyncIterator[LLMChunk]:
        raise NotImplementedError
        yield LLMChunk()


class ExplodingJudge(EvalJudge):
    """Judge whose calls always fail — inconclusive must become a hard failure."""

    async def faithfulness(self, item: Any, trace: Any) -> JudgeVerdict:
        raise JudgeError("provider down")

    async def answer_relevancy(self, item: Any, trace: Any) -> JudgeVerdict:
        raise JudgeError("provider down")


@pytest.mark.asyncio
async def test_scorer_marks_judge_inconclusive_as_hard_fail(fixture_index) -> None:
    scorer = ItemScorer(fixture_index, ExplodingJudge(_NoopLLM(), cache_dir=None))
    scores = await scorer.score(make_item(), make_trace())
    assert scores.hard_fail
    assert any("judge inconclusive" in reason for reason in scores.hard_fail_reasons)
    assert scores.faithfulness.detail["inconclusive"] is True


@pytest.mark.asyncio
async def test_scorer_wrong_family_tool_hard_fails(fixture_index) -> None:
    scorer = ItemScorer(fixture_index, judge=None)
    trace = make_trace(tool_calls=[ToolCallTrace(tool_name="search_trusted_web")])
    scores = await scorer.score(make_item(), trace)
    assert scores.hard_fail
    assert "wrong-family tool use" in scores.hard_fail_reasons


@pytest.mark.asyncio
async def test_scorer_unknown_fabrication_hard_fails(fixture_index) -> None:
    scorer = ItemScorer(fixture_index, judge=None)
    item = make_item(id="kn-en-m2-u", unknown_expected=True)
    scores = await scorer.score(
        item, make_trace(item_id=item.id, answer="The VAT rate is 15 percent.")
    )
    assert scores.hard_fail
    assert any("fabricated" in reason for reason in scores.hard_fail_reasons)


def test_aggregate_excludes_not_applicable_metrics() -> None:
    items = [make_item(id="kn-en-e1-test")]
    scores = [
        ItemScores(
            item_id="kn-en-e1-test",
            citation_coverage=MetricValue(value=0.5),
            context_recall=MetricValue(value=1.0),
        )
    ]
    cell = aggregate_cell(CellKey(intent=items[0].intent, locale=items[0].locale), items, scores)
    assert cell.metrics["citation_coverage"] == 0.5
    assert cell.metrics["context_recall"] == 1.0
    assert "grounding_precision" not in cell.metrics
    assert cell.unknown_handling_rate is None


def test_aggregate_collects_hard_fails() -> None:
    items = [make_item(id="aaa"), make_item(id="bbb")]
    failing = ItemScores(item_id="aaa", hard_fail=True, hard_fail_reasons=["boom"])
    passing = ItemScores(item_id="bbb")
    cell = aggregate_cell(
        CellKey(intent=items[0].intent, locale=items[0].locale), items, [failing, passing]
    )
    assert cell.hard_fail_items == ["aaa"]
