from __future__ import annotations

from abc import ABC, abstractmethod

from core.domain.retry_policy import RetryOutcome


class ErrorClassifierPort(ABC):
    @abstractmethod
    def classify(self, error: Exception) -> RetryOutcome: ...

    @abstractmethod
    def is_transient_exception(self, error: Exception) -> bool: ...

    @abstractmethod
    def is_permanent_exception(self, error: Exception) -> bool: ...


__all__ = ["ErrorClassifierPort"]
