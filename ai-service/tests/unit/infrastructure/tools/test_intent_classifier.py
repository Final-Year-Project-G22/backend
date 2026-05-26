from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from core.ports.intent_classifier import IntentClass
from infrastructure.tools.intent_classifier import EmbeddingIntentClassifier


@pytest.fixture
def mock_embedding_port() -> MagicMock:
    port = MagicMock()
    port.embed_documents = AsyncMock(
        side_effect=[
            # Knowledge seeds → knowledge centroid
            [[0.9, 0.1, 0.0], [0.8, 0.2, 0.0]],
            # Personal seeds → personal centroid
            [[0.1, 0.9, 0.0], [0.2, 0.8, 0.0]],
        ]
    )
    return port


@pytest.fixture
async def classifier(mock_embedding_port: MagicMock) -> EmbeddingIntentClassifier:
    c = EmbeddingIntentClassifier(
        embedding_port=mock_embedding_port,
        knowledge_seeds=["business registration", "tax rates"],
        personal_seeds=["my profile", "my status"],
        threshold=0.6,
    )
    await c.initialize()
    return c


def test_knowledge_class(classifier: EmbeddingIntentClassifier) -> None:
    assert IntentClass.KNOWLEDGE.value == "knowledge"
    assert IntentClass.PERSONAL.value == "personal"
    assert IntentClass.MIXED.value == "mixed"


@pytest.mark.asyncio
async def test_initializes_centroids(mock_embedding_port: MagicMock) -> None:
    c = EmbeddingIntentClassifier(
        embedding_port=mock_embedding_port,
        knowledge_seeds=["seed1"],
        personal_seeds=["seed2"],
        threshold=0.6,
    )
    await c.initialize()
    mock_embedding_port.embed_documents.assert_awaited()
    assert len(c._knowledge_centroid) == 3
    assert len(c._personal_centroid) == 3


@pytest.mark.asyncio
async def test_classify_mixed_without_centroids() -> None:
    c = EmbeddingIntentClassifier(
        embedding_port=MagicMock(),
        knowledge_seeds=[],
        personal_seeds=[],
    )
    result = await c.classify([0.1, 0.2, 0.3])
    assert result == IntentClass.MIXED


@pytest.mark.asyncio
async def test_classify_knowledge_when_above_threshold(
    classifier: EmbeddingIntentClassifier,
) -> None:
    similar_to_knowledge = [0.88, 0.12, 0.0]
    result = await classifier.classify(similar_to_knowledge)
    assert result == IntentClass.KNOWLEDGE


@pytest.mark.asyncio
async def test_classify_personal_when_above_threshold(
    classifier: EmbeddingIntentClassifier,
) -> None:
    similar_to_personal = [0.1, 0.85, 0.05]
    result = await classifier.classify(similar_to_personal)
    assert result == IntentClass.PERSONAL


def test_cosine_similarity_identical() -> None:
    v = [1.0, 2.0, 3.0]
    sim = EmbeddingIntentClassifier._cosine_similarity(v, v)
    assert abs(sim - 1.0) < 1e-9


def test_cosine_similarity_orthogonal() -> None:
    a = [1.0, 0.0]
    b = [0.0, 1.0]
    sim = EmbeddingIntentClassifier._cosine_similarity(a, b)
    assert abs(sim) < 1e-9


def test_cosine_similarity_zero_denominator() -> None:
    sim = EmbeddingIntentClassifier._cosine_similarity([0.0, 0.0], [0.0, 0.0])
    assert sim == 0.0


def test_cosine_similarity_mismatched_length() -> None:
    sim = EmbeddingIntentClassifier._cosine_similarity([1.0], [1.0, 2.0])
    assert sim == 0.0
