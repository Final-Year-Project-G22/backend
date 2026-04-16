from __future__ import annotations

from core.domain.retry_policy import BackoffType, RetryContext, RetryPolicy


def calculate_delay_ms(policy: RetryPolicy, context: RetryContext) -> int:
    attempt = context.retry_count + 1
    base = policy.base_delay_ms
    max_delay = policy.max_delay_ms

    if policy.backoff_type is BackoffType.EXPONENTIAL:
        delay = base * (2 ** (attempt - 1))
    else:
        delay = base * attempt

    return min(delay, max_delay)


__all__ = ["calculate_delay_ms"]
