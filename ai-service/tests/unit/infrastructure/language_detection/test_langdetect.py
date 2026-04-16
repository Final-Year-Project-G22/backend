from __future__ import annotations

from core.domain.enums import Language
from core.ports.language_detection import DetectionResult
from infrastructure.language_detection import LangDetectAdapter

_MIN_CONFIDENCE_50 = 0.5
_EXPECTED_CONFIDENCE_95 = 0.95


class TestLangDetectAdapter:
    def test_detect_english_text(self) -> None:
        adapter = LangDetectAdapter()
        result = adapter.detect("This is a sample English text for testing.")

        assert result.detected_language == Language.ENGLISH
        assert result.confidence > _MIN_CONFIDENCE_50

    def test_detect_amharic_text(self) -> None:
        adapter = LangDetectAdapter()
        result = adapter.detect("ይህ ኢትዮጵያ ነው")

        assert result.detected_language == Language.AMHARIC
        assert result.confidence > _MIN_CONFIDENCE_50

    def test_detect_empty_text_returns_english(self) -> None:
        adapter = LangDetectAdapter()
        result = adapter.detect("")

        assert result.detected_language == Language.ENGLISH
        assert result.confidence == 0.0

    def test_detect_whitespace_only_returns_english(self) -> None:
        adapter = LangDetectAdapter()
        result = adapter.detect("   ")

        assert result.detected_language == Language.ENGLISH
        assert result.confidence == 0.0


class TestDetectionResult:
    def test_creation(self) -> None:
        result = DetectionResult(
            detected_language=Language.ENGLISH,
            confidence=_EXPECTED_CONFIDENCE_95,
        )
        assert result.detected_language == Language.ENGLISH
        assert result.confidence == _EXPECTED_CONFIDENCE_95
