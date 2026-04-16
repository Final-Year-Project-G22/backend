from __future__ import annotations

from dataclasses import dataclass
from enum import Enum


class RetryOutcome(Enum):
    TRANSIENT = "transient"
    PERMANENT = "permanent"


class BackoffType(Enum):
    EXPONENTIAL = "exponential"
    LINEAR = "linear"


@dataclass(frozen=True)
class RetryPolicy:
    max_retries: int
    base_delay_ms: int
    max_delay_ms: int
    jitter_factor: float
    backoff_type: BackoffType = BackoffType.EXPONENTIAL


@dataclass(frozen=True)
class RetryContext:
    retry_count: int
    last_error: str
    retry_outcome: RetryOutcome


__all__ = [
    "BackoffType",
    "RetryContext",
    "RetryOutcome",
    "RetryPolicy",
]
