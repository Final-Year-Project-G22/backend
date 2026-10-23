"""Per-cell aggregation of item scores."""

from __future__ import annotations

from evals.models import CellKey, CellScore, GoldenItem, ItemScores

METRIC_FIELDS = (
    "citation_coverage",
    "grounding_precision",
    "tool_adherence",
    "context_recall",
    "faithfulness",
    "answer_relevancy",
)


def aggregate_cell(
    cell: CellKey,
    items: list[GoldenItem],
    scores: list[ItemScores],
) -> CellScore:
    by_id = {item.id: item for item in items}
    metrics: dict[str, float] = {}
    for field in METRIC_FIELDS:
        values = [
            getattr(score, field).value
            for score in scores
            if getattr(score, field).applicable and score.item_id in by_id
        ]
        if values:
            metrics[field] = round(sum(values) / len(values), 4)

    unknown_scores = [
        score.unknown_handling.value
        for score in scores
        if score.unknown_handling.applicable and score.item_id in by_id
    ]
    unknown_rate: float | None = None
    if unknown_scores:
        unknown_rate = round(sum(unknown_scores) / len(unknown_scores), 4)

    hard_fail_items = [score.item_id for score in scores if score.hard_fail]

    return CellScore(
        cell=cell,
        item_count=len(scores),
        metrics=metrics,
        unknown_handling_rate=unknown_rate,
        hard_fail_items=hard_fail_items,
    )


__all__ = ["METRIC_FIELDS", "aggregate_cell"]
