from __future__ import annotations

from abc import ABC, abstractmethod
from enum import StrEnum
from typing import Any

from pydantic import BaseModel, Field

from core.domain.enums import Language
from core.ports.language_detection import DetectionResult


class QualityGateResultStatus(StrEnum):
    PASSED = "passed"
    FAILED = "failed"
    WARNING = "warning"


class QualityGateResult(BaseModel):
    status: QualityGateResultStatus
    accepted_chunks: int
    rejected_chunks: int
    total_chunks: int
    language_mismatches: int
    failed_checks: list[str] = Field(default_factory=list)
    warnings: list[str] = Field(default_factory=list)


class QualityGatePolicy(BaseModel):
    min_accepted_chunks: int = Field(default=1, ge=0)
    max_rejection_ratio: float = Field(default=0.5, ge=0.0, le=1.0)
    min_token_count_per_chunk: int = Field(default=10, ge=1)
    check_language_match: bool = True
    warn_on_low_confidence_detection: bool = True
    min_detection_confidence: float = Field(default=0.7, ge=0.0, le=1.0)


class QualityGatePort(ABC):
    @abstractmethod
    async def evaluate(
        self,
        chunks: list[Any],
        policy: QualityGatePolicy,
        declared_language: Language | None = None,
        detection_result: DetectionResult | None = None,
    ) -> QualityGateResult:
        raise NotImplementedError


__all__ = [
    "QualityGatePolicy",
    "QualityGatePort",
    "QualityGateResult",
    "QualityGateResultStatus",
]
