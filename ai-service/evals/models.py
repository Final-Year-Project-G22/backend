"""Typed contracts for the evaluation harness."""

from __future__ import annotations

import uuid
from datetime import datetime
from enum import StrEnum
from typing import Any

from pydantic import BaseModel, ConfigDict, Field


class Intent(StrEnum):
    KNOWLEDGE = "knowledge"
    PERSONAL = "personal"
    MIXED = "mixed"


class Locale(StrEnum):
    EN = "en"
    AM = "am"


class Difficulty(StrEnum):
    EASY = "easy"
    MEDIUM = "medium"
    HARD = "hard"


class PromptVariant(StrEnum):
    CLEAN = "clean"
    BROKEN = "broken"


class ChunkRef(BaseModel):
    """Stable reference into the frozen fixture corpus."""

    model_config = ConfigDict(frozen=True)

    document_key: str = Field(pattern=r"^(en|am):(tax_code|legal|government|guide):[a-z0-9-]+$")
    chunk_index: int = Field(ge=0)

    def ref_id(self) -> str:
        return f"{self.document_key}#{self.chunk_index}"


class GoldenItem(BaseModel):
    model_config = ConfigDict(frozen=True)

    id: str = Field(min_length=3)
    intent: Intent
    locale: Locale
    difficulty: Difficulty
    query: str = Field(min_length=1)
    expected_tools: list[str] = Field(default_factory=list)
    allowed_tools: list[str] = Field(default_factory=list)
    required_citations: list[ChunkRef] = Field(default_factory=list)
    expected_chunk_refs: list[ChunkRef] = Field(default_factory=list)
    golden_answer: str = Field(min_length=1)
    golden_claims: list[str] = Field(default_factory=list)
    unknown_expected: bool = False
    fixture_account: str = "eval-msme-01"
    dataset_version: str = "1.0.0"


class CitedSource(BaseModel):
    """A source the answer presented, mapped back to fixture coordinates."""

    model_config = ConfigDict(frozen=True)

    chunk_ref: ChunkRef | None
    title: str
    excerpt: str | None = None


class ToolCallTrace(BaseModel):
    model_config = ConfigDict(frozen=True)

    tool_name: str
    suppressed: bool = False
    success: bool = True


class ItemTrace(BaseModel):
    """Full trace captured for one golden item run."""

    item_id: str
    conversation_id: uuid.UUID
    answer: str
    context_chunks: list[ChunkRef] = Field(default_factory=list)
    context_texts: list[str] = Field(default_factory=list)
    cited_sources: list[CitedSource] = Field(default_factory=list)
    tool_calls: list[ToolCallTrace] = Field(default_factory=list)
    cache_hit: bool = False
    usage: dict[str, Any] | None = None
    model_used: str = ""
    processing_time_ms: int = 0
    error: str | None = None


class MetricValue(BaseModel):
    model_config = ConfigDict(frozen=True)

    value: float = Field(ge=0.0, le=1.0)
    applicable: bool = True
    detail: dict[str, Any] = Field(default_factory=dict)


NOT_APPLICABLE = MetricValue(value=0.0, applicable=False, detail={"reason": "not_applicable"})


class ItemScores(BaseModel):
    item_id: str
    citation_coverage: MetricValue = NOT_APPLICABLE
    grounding_precision: MetricValue = NOT_APPLICABLE
    tool_adherence: MetricValue = NOT_APPLICABLE
    context_recall: MetricValue = NOT_APPLICABLE
    faithfulness: MetricValue = NOT_APPLICABLE
    answer_relevancy: MetricValue = NOT_APPLICABLE
    unknown_handling: MetricValue = NOT_APPLICABLE
    hard_fail: bool = False
    hard_fail_reasons: list[str] = Field(default_factory=list)


class JudgeVerdict(BaseModel):
    """Outcome of one judge evaluation; inconclusive is a hard failure."""

    model_config = ConfigDict(frozen=True)

    score: float | None = Field(default=None, ge=0.0, le=1.0)
    inconclusive: bool = False
    rationale: str = ""
    swapped_used: bool = False


class ClaimVerdict(BaseModel):
    model_config = ConfigDict(frozen=True)

    claim: str
    supported: bool | None = None
    inconclusive: bool = False
    rationale: str = ""


class CellKey(BaseModel):
    model_config = ConfigDict(frozen=True)

    intent: Intent
    locale: Locale

    def key(self) -> str:
        return f"{self.intent.value}:{self.locale.value}"


class CellScore(BaseModel):
    cell: CellKey
    item_count: int
    metrics: dict[str, float]
    unknown_handling_rate: float | None = None
    hard_fail_items: list[str] = Field(default_factory=list)


class RunManifest(BaseModel):
    run_id: str
    started_at: datetime
    finished_at: datetime | None = None
    variant: PromptVariant
    dataset_version: str
    fixture_version: str
    llm_provider: str
    llm_model: str
    embedding_provider: str
    embedding_model: str
    judge_provider: str
    judge_model: str
    am_faithfulness_trusted: bool = False
    git_sha: str = ""
    config: dict[str, Any] = Field(default_factory=dict)


__all__ = [
    "NOT_APPLICABLE",
    "CellKey",
    "CellScore",
    "ChunkRef",
    "CitedSource",
    "ClaimVerdict",
    "Difficulty",
    "GoldenItem",
    "Intent",
    "ItemScores",
    "ItemTrace",
    "JudgeVerdict",
    "Locale",
    "MetricValue",
    "PromptVariant",
    "RunManifest",
    "ToolCallTrace",
]
