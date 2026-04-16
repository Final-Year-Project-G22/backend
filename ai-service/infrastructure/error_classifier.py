from __future__ import annotations

from core.domain.exceptions import AIServiceError, PermanentError, TransientError
from core.domain.retry_policy import RetryOutcome
from core.ports.error_classifier import ErrorClassifierPort

TRANSIENT_ERROR_CODES = frozenset(
    {
        "REPOSITORY_ERROR",
        "EMBEDDING_ERROR",
        "LLM_ERROR",
        "CACHE_ERROR",
        "TRANSIENT_ERROR",
    }
)

PERMANENT_ERROR_CODES = frozenset(
    {
        "VALIDATION_ERROR",
        "QUOTA_EXCEEDED",
        "RETRIEVAL_ERROR",
        "CONFIGURATION_ERROR",
        "INVALID_STATE_TRANSITION",
        "PERMANENT_ERROR",
    }
)


class DefaultErrorClassifier(ErrorClassifierPort):
    def classify(self, error: Exception) -> RetryOutcome:
        if isinstance(error, TransientError) or error.__class__ is TransientError:
            return RetryOutcome.TRANSIENT
        if isinstance(error, PermanentError) or error.__class__ is PermanentError:
            return RetryOutcome.PERMANENT

        if isinstance(error, AIServiceError):
            code = error.code
            if code in TRANSIENT_ERROR_CODES:
                return RetryOutcome.TRANSIENT
            if code in PERMANENT_ERROR_CODES:
                return RetryOutcome.PERMANENT

        return RetryOutcome.TRANSIENT

    def is_transient_exception(self, error: Exception) -> bool:
        return self.classify(error) is RetryOutcome.TRANSIENT

    def is_permanent_exception(self, error: Exception) -> bool:
        return self.classify(error) is RetryOutcome.PERMANENT


__all__ = ["DefaultErrorClassifier"]
