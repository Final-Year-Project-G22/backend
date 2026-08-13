from __future__ import annotations

import logging
import math
from dataclasses import dataclass
from enum import StrEnum

from core.ports.embedding import EmbeddingPort

logger = logging.getLogger(__name__)

DUP_COSINE_THRESHOLD = 0.92
AMBIGUOUS_BAND_MIN = 0.80
JACCARD_THRESHOLD = 0.6
DRIFT_COSINE_THRESHOLD = 0.40


class SuppressionReason(StrEnum):
    DUPLICATE_OF_PROMPT = "duplicate_of_prompt"
    DUPLICATE_OF_PRIOR_SEARCH = "duplicate_of_prior_search"
    DRIFT = "drift"


@dataclass(frozen=True)
class KBQueryDecision:
    allowed: bool
    reason: SuppressionReason | None = None
    matched_query: str | None = None
    embedding: list[float] | None = None


@dataclass(frozen=True)
class _ExecutedQuery:
    raw: str
    norm: str
    embedding: list[float] | None


def _normalize_query(text: str) -> str:
    return " ".join(text.split()).casefold()


def _cosine_similarity(a: list[float], b: list[float]) -> float:
    if not a or not b or len(a) != len(b):
        return 0.0
    dot = 0.0
    norm_a = 0.0
    norm_b = 0.0
    for x, y in zip(a, b, strict=True):
        dot += x * y
        norm_a += x * x
        norm_b += y * y
    denom = math.sqrt(norm_a) * math.sqrt(norm_b)
    if denom == 0.0:
        return 0.0
    return dot / denom


def _jaccard_overlap(a_norm: str, b_norm: str) -> float:
    a_tokens = set(a_norm.split())
    b_tokens = set(b_norm.split())
    if not a_tokens or not b_tokens:
        return 0.0
    return len(a_tokens & b_tokens) / len(a_tokens | b_tokens)


class KBQueryGuard:
    """Deterministic per-turn guards for search_knowledge_base tool calls.

    Implements the FIN-75 follow-up query policy: duplicate detection against
    the original turn prompt and every KB query already executed this turn,
    with a cosine band + token-overlap Jaccard tiebreak, plus drift detection.
    """

    def __init__(
        self,
        embedding_port: EmbeddingPort,
        *,
        dup_cosine_threshold: float = DUP_COSINE_THRESHOLD,
        ambiguous_band_min: float = AMBIGUOUS_BAND_MIN,
        jaccard_threshold: float = JACCARD_THRESHOLD,
        drift_cosine_threshold: float = DRIFT_COSINE_THRESHOLD,
    ) -> None:
        self._embedding_port = embedding_port
        self._dup_cosine_threshold = dup_cosine_threshold
        self._ambiguous_band_min = ambiguous_band_min
        self._jaccard_threshold = jaccard_threshold
        self._drift_cosine_threshold = drift_cosine_threshold

        self._prompt_raw = ""
        self._prompt_norm = ""
        self._prompt_embedding: list[float] = []
        self._prompt_covered = False
        self._executed: list[_ExecutedQuery] = []

    def start_turn(
        self,
        prompt: str,
        prompt_embedding: list[float],
        *,
        prompt_covered: bool,
    ) -> None:
        self._prompt_raw = prompt
        self._prompt_norm = _normalize_query(prompt)
        self._prompt_embedding = prompt_embedding
        self._prompt_covered = prompt_covered
        self._executed = []

    async def check(self, query: str) -> KBQueryDecision:  # noqa: PLR0911
        norm = _normalize_query(query)

        # Cheap deterministic exact-match checks first (case/whitespace variants).
        if self._prompt_covered and norm == self._prompt_norm:
            return KBQueryDecision(
                allowed=False,
                reason=SuppressionReason.DUPLICATE_OF_PROMPT,
                matched_query=self._prompt_raw,
            )
        for executed in reversed(self._executed):
            if norm == executed.norm:
                return KBQueryDecision(
                    allowed=False,
                    reason=SuppressionReason.DUPLICATE_OF_PRIOR_SEARCH,
                    matched_query=executed.raw,
                )

        embedding = await self._embed(query)
        if embedding is None:
            # Conservative by design: when unsure, run the search.
            return KBQueryDecision(allowed=True)

        # Prompt comparisons apply only when the turn's pre-fetched search
        # actually retrieved context (prompt_covered). When it found nothing,
        # the first tool search with the prompt query is the primary fetch
        # and must run; it registers as executed and repeats are suppressed.
        similarity_to_prompt = _cosine_similarity(embedding, self._prompt_embedding)

        if self._prompt_covered and similarity_to_prompt >= self._dup_cosine_threshold:
            return KBQueryDecision(
                allowed=False,
                reason=SuppressionReason.DUPLICATE_OF_PROMPT,
                matched_query=self._prompt_raw,
            )

        # Duplicate check against every KB query already executed this turn:
        # cosine >= threshold, or ambiguous band + token-overlap Jaccard.
        matched_executed: _ExecutedQuery | None = None
        for executed in self._executed:
            if executed.embedding is None:
                continue
            similarity = _cosine_similarity(embedding, executed.embedding)
            if similarity >= self._dup_cosine_threshold or (
                self._in_ambiguous_band(similarity)
                and _jaccard_overlap(norm, executed.norm) >= self._jaccard_threshold
            ):
                matched_executed = executed
                break
        if matched_executed is not None:
            return KBQueryDecision(
                allowed=False,
                reason=SuppressionReason.DUPLICATE_OF_PRIOR_SEARCH,
                matched_query=matched_executed.raw,
            )

        if similarity_to_prompt < self._drift_cosine_threshold:
            return KBQueryDecision(
                allowed=False,
                reason=SuppressionReason.DRIFT,
                matched_query=self._prompt_raw,
            )

        # Ambiguous band tiebreak via token-overlap Jaccard against the prompt.
        if (
            self._prompt_covered
            and self._in_ambiguous_band(similarity_to_prompt)
            and _jaccard_overlap(norm, self._prompt_norm) >= self._jaccard_threshold
        ):
            return KBQueryDecision(
                allowed=False,
                reason=SuppressionReason.DUPLICATE_OF_PROMPT,
                matched_query=self._prompt_raw,
            )

        return KBQueryDecision(allowed=True, embedding=embedding)

    def register_executed(self, query: str, embedding: list[float] | None = None) -> None:
        self._executed.append(
            _ExecutedQuery(raw=query, norm=_normalize_query(query), embedding=embedding)
        )

    def _in_ambiguous_band(self, similarity: float) -> bool:
        return self._ambiguous_band_min <= similarity < self._dup_cosine_threshold

    async def _embed(self, query: str) -> list[float] | None:
        try:
            return await self._embedding_port.embed_query(query)
        except Exception:
            logger.warning("Failed to embed KB query for duplicate guard; allowing search")
            return None
