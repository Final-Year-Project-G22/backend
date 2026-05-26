from __future__ import annotations

from core.domain.enums import AskStrategy


def test_ask_strategy_values() -> None:
    assert AskStrategy.SIMPLE.value == "simple"
    assert AskStrategy.AGENTIC.value == "agentic"


def test_ask_strategy_is_str_enum() -> None:
    assert str(AskStrategy.SIMPLE) == "simple"
    assert str(AskStrategy.AGENTIC) == "agentic"
    assert AskStrategy("simple") is AskStrategy.SIMPLE
    assert AskStrategy("agentic") is AskStrategy.AGENTIC
