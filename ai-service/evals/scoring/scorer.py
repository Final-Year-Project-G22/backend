"""Item-level scoring: deterministic gates + judge metrics + hard failures."""

from __future__ import annotations

from evals.dataset import FixtureIndex
from evals.models import GoldenItem, ItemScores, ItemTrace, JudgeVerdict, MetricValue
from evals.scoring import deterministic
from evals.scoring.judge import EvalJudge, JudgeError

PROVISIONAL_BARS: dict[str, float] = {
    "citation_coverage": 0.80,
    "grounding_precision": 0.90,
    "tool_adherence": 0.90,
    "context_recall": 0.70,
    "faithfulness": 0.85,
    "answer_relevancy": 0.80,
}


class ItemScorer:
    def __init__(self, fixture_index: FixtureIndex, judge: EvalJudge | None) -> None:
        self._fixture_index = fixture_index
        self._judge = judge

    async def score(self, item: GoldenItem, trace: ItemTrace) -> ItemScores:
        scores = ItemScores(item_id=item.id)
        known_refs = {ref.ref_id() for ref in self._fixture_index.id_to_ref.values()}

        scores.citation_coverage = deterministic.citation_coverage(item, trace)
        scores.grounding_precision = deterministic.grounding_precision(trace, known_refs)
        scores.context_recall = deterministic.context_recall(item, trace)
        scores.unknown_handling = deterministic.unknown_handling(item, trace.answer)

        adherence, wrong_family = deterministic.tool_adherence(item, trace)
        scores.tool_adherence = adherence

        reasons: list[str] = []
        if wrong_family:
            reasons.append("wrong-family tool use")
        if item.unknown_expected and scores.unknown_handling.value == 0.0:
            scores.unknown_handling = await self._confirm_refusal(item, trace, reasons)

        if trace.error:
            reasons.append(f"run error: {trace.error}")

        if self._judge is not None and not trace.error:
            try:
                faithfulness = await self._judge.faithfulness(item, trace)
                relevancy = await self._judge.answer_relevancy(item, trace)
            except JudgeError as exc:
                faithfulness = JudgeVerdict(inconclusive=True, rationale=str(exc))
                relevancy = JudgeVerdict(inconclusive=True, rationale=str(exc))
            scores.faithfulness = self._judge_metric(faithfulness, reasons)
            scores.answer_relevancy = self._judge_metric(relevancy, reasons)

        scores.hard_fail_reasons = reasons
        scores.hard_fail = bool(reasons)
        return scores

    async def _confirm_refusal(
        self,
        item: GoldenItem,
        trace: ItemTrace,
        reasons: list[str],
    ) -> MetricValue:
        """Markers missed — confirm honest refusal with the judge before failing."""
        if self._judge is None:
            reasons.append("fabricated instead of refusing (unknown item)")
            return MetricValue(value=0.0, detail={"marker_matched": False})
        verdict = await self._judge.honest_refusal(item, trace.answer)
        if verdict.inconclusive or verdict.score is None:
            reasons.append(f"judge inconclusive: {verdict.rationale}")
            return MetricValue(
                value=0.0,
                detail={"inconclusive": True, "rationale": verdict.rationale},
            )
        if verdict.score == 0.0:
            reasons.append("fabricated instead of refusing (unknown item)")
        return MetricValue(
            value=verdict.score,
            detail={"marker_matched": False, "judge_confirmed": True},
        )

    @staticmethod
    def _judge_metric(verdict: JudgeVerdict, reasons: list[str]) -> MetricValue:
        if verdict.inconclusive or verdict.score is None:
            reasons.append(f"judge inconclusive: {verdict.rationale}")
            return MetricValue(
                value=0.0,
                applicable=True,
                detail={"inconclusive": True, "rationale": verdict.rationale},
            )
        return MetricValue(
            value=verdict.score,
            detail={"rationale": verdict.rationale},
        )


__all__ = ["PROVISIONAL_BARS", "ItemScorer"]
