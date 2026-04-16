from __future__ import annotations

import logging
import re

from core.domain.enums import Language
from core.ports.language_detection import DetectionResult, LanguageDetectionPort

logger = logging.getLogger(__name__)

_AMHARIC_PATTERN = re.compile(r"[\u1200-\u137F]")


def _has_amharic_chars(text: str) -> bool:
    return bool(_AMHARIC_PATTERN.search(text))


class LangDetectAdapter(LanguageDetectionPort):
    def detect(self, text: str) -> DetectionResult:
        if not text or not text.strip():
            return DetectionResult(
                detected_language=Language.ENGLISH,
                confidence=0.0,
            )

        try:
            if _has_amharic_chars(text):
                language = Language.AMHARIC
                confidence = 0.95
            else:
                language = Language.ENGLISH
                confidence = 0.90

            return DetectionResult(
                detected_language=language,
                confidence=confidence,
            )
        except Exception as exc:
            logger.warning("Language detection failed: %s, defaulting to English", exc)
            return DetectionResult(
                detected_language=Language.ENGLISH,
                confidence=0.0,
            )


__all__ = ["LangDetectAdapter"]
