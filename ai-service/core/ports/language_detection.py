from __future__ import annotations

from abc import ABC, abstractmethod

from pydantic import BaseModel

from core.domain.enums import Language


class DetectionResult(BaseModel):
    detected_language: Language
    confidence: float


class LanguageDetectionPort(ABC):
    @abstractmethod
    def detect(self, text: str) -> DetectionResult:
        raise NotImplementedError


__all__ = ["DetectionResult", "LanguageDetectionPort"]
