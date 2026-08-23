"""Judge tests with a scripted fake LLM (no network)."""

from __future__ import annotations

import json
from collections.abc import AsyncIterator
from pathlib import Path
from typing import Any

import pytest

from core.ports.llm import LLMChunk, LLMPort, LLMResult, ToolDefinition
from evals.scoring.judge import EvalJudge, _extract_json_object
from tests.unit.evals.conftest import make_item, make_trace


class ScriptedLLM(LLMPort):
    """Returns queued JSON payloads in order; records prompts."""

    def __init__(self, payloads: list[dict[str, Any]]) -> None:
        self._payloads = list(payloads)
        self.prompts: list[str] = []

    @property
    def provider(self) -> str:
        return "scripted"

    @property
    def model(self) -> str:
        return "test-model"

    async def generate(
        self,
        prompt: str,
        *,
        system_prompt: str | None = None,
        tools: list[ToolDefinition] | None = None,
        max_tokens: int = 1024,
        temperature: float = 0.2,
    ) -> LLMResult:
        self.prompts.append(prompt)
        payload = self._payloads.pop(0)
        return LLMResult(text=json.dumps(payload))

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


def test_extract_json_object_handles_fences_and_prose() -> None:
    assert _extract_json_object('{"a": 1}') == {"a": 1}
    assert _extract_json_object('```json\n{"a": 1}\n```') == {"a": 1}
    assert _extract_json_object('Verdict: {"a": 1} end') == {"a": 1}
    assert _extract_json_object("no json here") is None


@pytest.mark.asyncio
async def test_faithfulness_scores_supported_claims(tmp_path: Path) -> None:
    llm = ScriptedLLM(
        [
            {"claims": ["VAT is a consumption tax.", "VAT is 15 percent."]},
            {"supported": True, "confidence": "high", "reason": "stated"},
            {"supported": False, "confidence": "high", "reason": "not in context"},
        ]
    )
    judge = EvalJudge(llm, cache_dir=tmp_path)
    verdict = await judge.faithfulness(make_item(), make_trace())
    assert not verdict.inconclusive
    assert verdict.score == 0.5


@pytest.mark.asyncio
async def test_faithfulness_inconclusive_on_low_confidence_disagreement(tmp_path: Path) -> None:
    llm = ScriptedLLM(
        [
            {"claims": ["VAT is a consumption tax."]},
            {"supported": True, "confidence": "low", "reason": "maybe"},
            {"supported": False, "confidence": "low", "reason": "swapped view"},
        ]
    )
    judge = EvalJudge(llm, cache_dir=tmp_path)
    verdict = await judge.faithfulness(make_item(), make_trace())
    assert verdict.inconclusive


@pytest.mark.asyncio
async def test_relevancy_single_pass_when_high(tmp_path: Path) -> None:
    llm = ScriptedLLM([{"score": 0.95, "confidence": "high", "reason": "answers it"}])
    judge = EvalJudge(llm, cache_dir=tmp_path)
    verdict = await judge.answer_relevancy(make_item(), make_trace())
    assert verdict.score == 0.95
    assert not verdict.swapped_used
    assert len(llm.prompts) == 1


@pytest.mark.asyncio
async def test_relevancy_rejudges_swapped_and_averages(tmp_path: Path) -> None:
    llm = ScriptedLLM(
        [
            {"score": 0.7, "confidence": "high", "reason": "first"},
            {"score": 0.8, "confidence": "high", "reason": "second"},
        ]
    )
    judge = EvalJudge(llm, cache_dir=tmp_path)
    verdict = await judge.answer_relevancy(make_item(), make_trace())
    assert verdict.score == pytest.approx(0.75)
    assert verdict.swapped_used


@pytest.mark.asyncio
async def test_relevancy_inconclusive_on_swap_disagreement(tmp_path: Path) -> None:
    llm = ScriptedLLM(
        [
            {"score": 0.4, "confidence": "high", "reason": "first"},
            {"score": 0.9, "confidence": "high", "reason": "second"},
        ]
    )
    judge = EvalJudge(llm, cache_dir=tmp_path)
    verdict = await judge.answer_relevancy(make_item(), make_trace())
    assert verdict.inconclusive


@pytest.mark.asyncio
async def test_judge_caches_verdicts_by_input_hash(tmp_path: Path) -> None:
    llm = ScriptedLLM([{"score": 0.9, "confidence": "high", "reason": "ok"}])
    judge = EvalJudge(llm, cache_dir=tmp_path)
    first = await judge.answer_relevancy(make_item(), make_trace())
    cached_files = list(tmp_path.glob("relevancy-*.json"))
    assert len(cached_files) == 1

    empty_llm = ScriptedLLM([])
    warm = EvalJudge(empty_llm, cache_dir=tmp_path)
    second = await warm.answer_relevancy(make_item(), make_trace())
    assert second.score == first.score
    assert len(empty_llm.prompts) == 0


@pytest.mark.asyncio
async def test_amharic_items_use_amharic_rubric(tmp_path: Path) -> None:
    llm = ScriptedLLM([{"claims": ["ክሊይም"]}, {"supported": True, "confidence": "high"}])
    judge = EvalJudge(llm, cache_dir=tmp_path)
    item = make_item(id="kn-am-e1-x", locale="am")
    trace = make_trace(item_id="kn-am-e1-x", answer="መልስ")
    await judge.faithfulness(item, trace)
    assert any("CONTEXT" in prompt for prompt in llm.prompts)
