"""AI evaluation harness (FIN-69/FIN-87).

Canonical runner for the 36-item Agentic Ask golden set. The CLI drives the
in-process ``AskAIUseCase.execute`` path, captures full traces, scores items
with deterministic gates plus LLM-judge metrics, and aggregates per-cell
scores for calibration (FIN-77).
"""

from evals.models import CellKey, CellScore, GoldenItem, ItemScores, ItemTrace, RunManifest

__all__ = [
    "CellKey",
    "CellScore",
    "GoldenItem",
    "ItemScores",
    "ItemTrace",
    "RunManifest",
]
