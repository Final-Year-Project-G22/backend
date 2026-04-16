from __future__ import annotations

import pytest

from core.domain.enums import Language
from core.ports.chunking import Chunk, ChunkProvenance
from core.ports.language_detection import DetectionResult
from core.ports.quality_gates import QualityGatePolicy, QualityGateResultStatus
from infrastructure.quality_gates import DefaultQualityGateAdapter

_TWO_CHUNKS = 2
_DEFAULT_MAX_REJECTION = 0.5
_DEFAULT_MIN_TOKENS = 10
_DEFAULT_MIN_CONFIDENCE = 0.7
_DEFAULT_MIN_ACCEPTED = 1
_CUSTOM_MIN_ACCEPTED = 5
_CUSTOM_MAX_REJECTION = 0.3


class TestDefaultQualityGateAdapter:
    @pytest.mark.asyncio
    async def test_all_chunks_accepted_when_meeting_criteria(self) -> None:
        adapter = DefaultQualityGateAdapter()
        chunks = [
            Chunk(
                chunk_text="Test content 1",
                token_count=50,
                provenance=ChunkProvenance(section_order=0, chunk_index=0),
            ),
            Chunk(
                chunk_text="Test content 2",
                token_count=60,
                provenance=ChunkProvenance(section_order=0, chunk_index=1),
            ),
        ]
        policy = QualityGatePolicy()

        result = await adapter.evaluate(chunks, policy)

        assert result.status == QualityGateResultStatus.PASSED
        assert result.accepted_chunks == _TWO_CHUNKS
        assert result.rejected_chunks == 0
        assert result.total_chunks == _TWO_CHUNKS

    @pytest.mark.asyncio
    async def test_rejects_chunks_below_min_token_count(self) -> None:
        adapter = DefaultQualityGateAdapter()
        chunks = [
            Chunk(
                chunk_text="Short",
                token_count=5,
                provenance=ChunkProvenance(section_order=0, chunk_index=0),
            ),
            Chunk(
                chunk_text="Longer content here",
                token_count=50,
                provenance=ChunkProvenance(section_order=0, chunk_index=1),
            ),
        ]
        policy = QualityGatePolicy(min_token_count_per_chunk=10)

        result = await adapter.evaluate(chunks, policy)

        assert result.accepted_chunks == 1
        assert result.rejected_chunks == 1

    @pytest.mark.asyncio
    async def test_fails_when_accepted_chunks_below_minimum(self) -> None:
        adapter = DefaultQualityGateAdapter()
        chunks = [
            Chunk(
                chunk_text="Short",
                token_count=5,
                provenance=ChunkProvenance(section_order=0, chunk_index=0),
            ),
        ]
        policy = QualityGatePolicy(min_accepted_chunks=2)

        result = await adapter.evaluate(chunks, policy)

        assert result.status == QualityGateResultStatus.FAILED
        assert result.accepted_chunks == 0

    @pytest.mark.asyncio
    async def test_warning_when_rejection_ratio_too_high(self) -> None:
        adapter = DefaultQualityGateAdapter()
        chunks = [
            Chunk(
                chunk_text="Short",
                token_count=5,
                provenance=ChunkProvenance(section_order=0, chunk_index=0),
            ),
            Chunk(
                chunk_text="Short",
                token_count=5,
                provenance=ChunkProvenance(section_order=0, chunk_index=1),
            ),
            Chunk(
                chunk_text="Longer content",
                token_count=50,
                provenance=ChunkProvenance(section_order=0, chunk_index=2),
            ),
        ]
        policy = QualityGatePolicy(max_rejection_ratio=0.3, min_token_count_per_chunk=10)

        result = await adapter.evaluate(chunks, policy)

        assert result.status == QualityGateResultStatus.WARNING

    @pytest.mark.asyncio
    async def test_language_mismatch_fails_check(self) -> None:
        adapter = DefaultQualityGateAdapter()
        chunks = [
            Chunk(
                chunk_text="Test content",
                token_count=50,
                provenance=ChunkProvenance(section_order=0, chunk_index=0),
            ),
        ]
        policy = QualityGatePolicy(check_language_match=True)

        result = await adapter.evaluate(
            chunks,
            policy,
            declared_language=Language.ENGLISH,
            detection_result=DetectionResult(detected_language=Language.AMHARIC, confidence=0.9),
        )

        assert result.status == QualityGateResultStatus.FAILED
        assert result.language_mismatches == 1
        assert "language_mismatch" in result.failed_checks[0]

    @pytest.mark.asyncio
    async def test_warns_on_low_detection_confidence(self) -> None:
        adapter = DefaultQualityGateAdapter()
        chunks = [
            Chunk(
                chunk_text="Test content",
                token_count=50,
                provenance=ChunkProvenance(section_order=0, chunk_index=0),
            ),
        ]
        policy = QualityGatePolicy(
            warn_on_low_confidence_detection=True, min_detection_confidence=0.8
        )

        result = await adapter.evaluate(
            chunks,
            policy,
            declared_language=Language.ENGLISH,
            detection_result=DetectionResult(detected_language=Language.ENGLISH, confidence=0.5),
        )

        assert len(result.warnings) == 1
        assert "low_detection_confidence" in result.warnings[0]


class TestQualityGatePolicy:
    def test_default_values(self) -> None:
        policy = QualityGatePolicy()
        assert policy.min_accepted_chunks == _DEFAULT_MIN_ACCEPTED
        assert policy.max_rejection_ratio == _DEFAULT_MAX_REJECTION
        assert policy.min_token_count_per_chunk == _DEFAULT_MIN_TOKENS
        assert policy.check_language_match is True
        assert policy.warn_on_low_confidence_detection is True
        assert policy.min_detection_confidence == _DEFAULT_MIN_CONFIDENCE

    def test_custom_values(self) -> None:
        policy = QualityGatePolicy(
            min_accepted_chunks=_CUSTOM_MIN_ACCEPTED, max_rejection_ratio=_CUSTOM_MAX_REJECTION
        )
        assert policy.min_accepted_chunks == _CUSTOM_MIN_ACCEPTED
        assert policy.max_rejection_ratio == _CUSTOM_MAX_REJECTION
