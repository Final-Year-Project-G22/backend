from __future__ import annotations

from core.domain.exceptions import (
    AIServiceError,
    CacheError,
    ConfigurationError,
    EmbeddingError,
    LLMError,
    QuotaExceededError,
    RepositoryError,
    RetrievalError,
    ValidationError,
)


def test_base_exception_defaults() -> None:
    error = AIServiceError("boom")

    assert str(error) == "boom"
    assert error.code == "AI_SERVICE_ERROR"
    assert error.details == {}


def test_base_exception_details_are_preserved() -> None:
    details = {"field": "language", "reason": "unsupported"}
    error = AIServiceError("invalid", code="CUSTOM", details=details)

    assert error.code == "CUSTOM"
    assert error.details == details


def test_validation_error_code() -> None:
    error = ValidationError("bad input")
    assert error.code == "VALIDATION_ERROR"


def test_embedding_error_code() -> None:
    error = EmbeddingError("embedding provider unavailable")
    assert error.code == "EMBEDDING_ERROR"


def test_llm_error_code() -> None:
    error = LLMError("generation timeout")
    assert error.code == "LLM_ERROR"


def test_repository_error_code() -> None:
    error = RepositoryError("db error")
    assert error.code == "REPOSITORY_ERROR"


def test_cache_error_code() -> None:
    error = CacheError("cache unavailable")
    assert error.code == "CACHE_ERROR"


def test_quota_exceeded_error_code() -> None:
    error = QuotaExceededError("daily limit reached")
    assert error.code == "QUOTA_EXCEEDED"


def test_retrieval_error_code() -> None:
    error = RetrievalError("retrieval pipeline failed")
    assert error.code == "RETRIEVAL_ERROR"


def test_configuration_error_code() -> None:
    error = ConfigurationError("missing api key")
    assert error.code == "CONFIGURATION_ERROR"


def test_all_domain_exceptions_inherit_base_error() -> None:
    exceptions = [
        ValidationError("x"),
        EmbeddingError("x"),
        LLMError("x"),
        RepositoryError("x"),
        CacheError("x"),
        QuotaExceededError("x"),
        RetrievalError("x"),
        ConfigurationError("x"),
    ]

    for error in exceptions:
        assert isinstance(error, AIServiceError)
