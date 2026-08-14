from __future__ import annotations

import logging
from collections.abc import AsyncIterator

from core.domain.stream_events import AskStreamEvent
from core.ports.ask_strategy import AskStrategyPort
from core.ports.llm import LLMPort
from core.usecases.contracts import AskAICommand, AskAIResult
from core.usecases.strategies.agentic_ask import AgenticAskStrategy
from core.usecases.strategies.simple_ask import SimpleAskStrategy

logger = logging.getLogger(__name__)


class AskAIUseCase:
    def __init__(
        self,
        simple_strategy: SimpleAskStrategy,
        agentic_strategy: AgenticAskStrategy | None = None,
        *,
        agentic_enabled: bool = True,
    ) -> None:
        self._simple_strategy = simple_strategy
        self._agentic_strategy = agentic_strategy
        self._agentic_enabled = agentic_enabled

    @property
    def _llm_port(self) -> LLMPort | None:
        return self._simple_strategy.llm_port

    def _select_strategy(self, command: AskAICommand) -> AskStrategyPort:
        if (
            self._agentic_enabled
            and command.strategy == "agentic"
            and self._agentic_strategy is not None
        ):
            return self._agentic_strategy
        return self._simple_strategy

    async def execute(self, command: AskAICommand) -> AskAIResult:
        strategy = self._select_strategy(command)
        return await strategy.execute(command)

    async def execute_stream_with_tools(
        self,
        command: AskAICommand,
    ) -> AsyncIterator[AskStreamEvent]:
        strategy = self._select_strategy(command)
        async for event in strategy.execute_stream(command):
            yield event


__all__ = ["AskAIUseCase"]
