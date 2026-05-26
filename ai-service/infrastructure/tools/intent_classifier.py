from __future__ import annotations

import logging
import math

from core.ports.embedding import EmbeddingPort
from core.ports.intent_classifier import IntentClass, IntentClassifierPort

logger = logging.getLogger(__name__)


class EmbeddingIntentClassifier(IntentClassifierPort):
    def __init__(
        self,
        embedding_port: EmbeddingPort,
        knowledge_seeds: list[str],
        personal_seeds: list[str],
        threshold: float = 0.6,
    ) -> None:
        self._embedding_port = embedding_port
        self._knowledge_seeds = knowledge_seeds
        self._personal_seeds = personal_seeds
        self._threshold = threshold
        self._knowledge_centroid: list[float] = []
        self._personal_centroid: list[float] = []

    async def initialize(self) -> None:
        self._knowledge_centroid = await self._compute_centroid(self._knowledge_seeds)
        self._personal_centroid = await self._compute_centroid(self._personal_seeds)
        logger.info(
            "Intent classifier initialized: %d knowledge seeds, %d personal seeds",
            len(self._knowledge_seeds),
            len(self._personal_seeds),
        )

    async def classify(self, query_embedding: list[float]) -> IntentClass:
        if not self._knowledge_centroid or not self._personal_centroid:
            return IntentClass.MIXED

        knowledge_sim = self._cosine_similarity(query_embedding, self._knowledge_centroid)
        personal_sim = self._cosine_similarity(query_embedding, self._personal_centroid)

        is_knowledge = knowledge_sim >= self._threshold
        is_personal = personal_sim >= self._threshold

        if is_knowledge and is_personal:
            return IntentClass.MIXED
        if is_knowledge:
            return IntentClass.KNOWLEDGE
        if is_personal:
            return IntentClass.PERSONAL

        return IntentClass.MIXED

    async def _compute_centroid(self, seeds: list[str]) -> list[float]:
        if not seeds:
            return []
        embeddings = await self._embedding_port.embed_documents(seeds)
        if not embeddings:
            return []
        dim = len(embeddings[0])
        centroid = [0.0] * dim
        for emb in embeddings:
            for i in range(dim):
                centroid[i] += emb[i]
        return [v / len(embeddings) for v in centroid]

    @staticmethod
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
