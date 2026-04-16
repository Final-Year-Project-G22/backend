from __future__ import annotations

from typing import Any


class AIServiceError(Exception):
    def __init__(
        self,
        message: str,
        *,
        code: str = "AI_SERVICE_ERROR",
        details: dict[str, Any] | None = None,
    ) -> None:
        super().__init__(message)
        self.message = message
        self.code = code
        self.details = details or {}


class ValidationError(AIServiceError):
    def __init__(self, message: str, *, details: dict[str, Any] | None = None) -> None:
        super().__init__(message, code="VALIDATION_ERROR", details=details)


class EmbeddingError(AIServiceError):
    def __init__(self, message: str, *, details: dict[str, Any] | None = None) -> None:
        super().__init__(message, code="EMBEDDING_ERROR", details=details)


class LLMError(AIServiceError):
    def __init__(self, message: str, *, details: dict[str, Any] | None = None) -> None:
        super().__init__(message, code="LLM_ERROR", details=details)


class RepositoryError(AIServiceError):
    def __init__(self, message: str, *, details: dict[str, Any] | None = None) -> None:
        super().__init__(message, code="REPOSITORY_ERROR", details=details)


class CacheError(AIServiceError):
    def __init__(self, message: str, *, details: dict[str, Any] | None = None) -> None:
        super().__init__(message, code="CACHE_ERROR", details=details)


class QuotaExceededError(AIServiceError):
    def __init__(self, message: str, *, details: dict[str, Any] | None = None) -> None:
        super().__init__(message, code="QUOTA_EXCEEDED", details=details)


class RetrievalError(AIServiceError):
    def __init__(self, message: str, *, details: dict[str, Any] | None = None) -> None:
        super().__init__(message, code="RETRIEVAL_ERROR", details=details)


class ConfigurationError(AIServiceError):
    def __init__(self, message: str, *, details: dict[str, Any] | None = None) -> None:
        super().__init__(message, code="CONFIGURATION_ERROR", details=details)


class InvalidStateTransitionError(AIServiceError):
    def __init__(self, message: str, *, details: dict[str, Any] | None = None) -> None:
        super().__init__(message, code="INVALID_STATE_TRANSITION", details=details)


__all__ = [
    "AIServiceError",
    "CacheError",
    "ConfigurationError",
    "EmbeddingError",
    "InvalidStateTransitionError",
    "LLMError",
    "QuotaExceededError",
    "RepositoryError",
    "RetrievalError",
    "ValidationError",
]
