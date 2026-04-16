from __future__ import annotations

import logging
from typing import Any

from core.domain.enums import Language
from core.ports.chunking import Chunk
from core.ports.language_detection import DetectionResult
from core.ports.quality_gates import (
    QualityGatePolicy,
    QualityGatePort,
    QualityGateResult,
    QualityGateResultStatus,
)

logger = logging.getLogger(__name__)


def _evaluate_chunks(chunks: list[Any], policy: QualityGatePolicy) -> tuple[int, int, list[str]]:
    accepted = 0
    rejected = 0
    failed: list[str] = []

    for chunk in chunks:
        if not isinstance(chunk, Chunk):
            rejected += 1
            continue
        if chunk.token_count < policy.min_token_count_per_chunk:
            rejected += 1
            continue
        accepted += 1

    total = accepted + rejected
    rejection_ratio = rejected / total if total > 0 else 0.0

    if accepted < policy.min_accepted_chunks:
        failed.append(f"accepted_chunks_below_minimum: {accepted} < {policy.min_accepted_chunks}")
    if rejection_ratio > policy.max_rejection_ratio:
        failed.append(
            f"rejection_ratio_too_high: {rejection_ratio:.2f} > {policy.max_rejection_ratio}"
        )

    return accepted, rejected, failed


def _check_language_mismatch(
    policy: QualityGatePolicy,
    declared: Language | None,
    detection: DetectionResult | None,
    accepted_chunks: int,
) -> tuple[int, bool, list[str]]:
    if not (policy.check_language_match and declared and detection):
        return 0, False, []

    if declared != detection.detected_language:
        return (
            accepted_chunks,
            True,
            [
                f"language_mismatch: declared={declared.value}, detected={detection.detected_language.value}"
            ],
        )
    return 0, False, []


def _check_low_confidence(
    policy: QualityGatePolicy, detection: DetectionResult | None
) -> list[str]:
    if (
        policy.warn_on_low_confidence_detection
        and detection
        and detection.confidence < policy.min_detection_confidence
    ):
        return [
            f"low_detection_confidence: {detection.confidence:.2f} < {policy.min_detection_confidence}"
        ]
    return []


def _determine_status(
    *, language_mismatch_detected: bool = False, failed_checks: list[str], accepted_count: int
) -> QualityGateResultStatus:
    if language_mismatch_detected:
        return QualityGateResultStatus.FAILED
    if not failed_checks:
        return QualityGateResultStatus.PASSED
    if accepted_count == 0:
        return QualityGateResultStatus.FAILED
    return QualityGateResultStatus.WARNING


class DefaultQualityGateAdapter(QualityGatePort):
    async def evaluate(
        self,
        chunks: list[Any],
        policy: QualityGatePolicy,
        declared_language: Language | None = None,
        detection_result: DetectionResult | None = None,
    ) -> QualityGateResult:
        accepted, rejected, failed = _evaluate_chunks(chunks, policy)
        total = accepted + rejected

        lang_mismatches, lang_failure, lang_checks = _check_language_mismatch(
            policy, declared_language, detection_result, accepted
        )
        failed.extend(lang_checks)

        warnings = _check_low_confidence(policy, detection_result)

        status = _determine_status(
            language_mismatch_detected=lang_failure,
            failed_checks=failed,
            accepted_count=accepted,
        )

        return QualityGateResult(
            status=status,
            accepted_chunks=accepted,
            rejected_chunks=rejected,
            total_chunks=total,
            language_mismatches=lang_mismatches,
            failed_checks=failed,
            warnings=warnings,
        )


__all__ = ["DefaultQualityGateAdapter"]
