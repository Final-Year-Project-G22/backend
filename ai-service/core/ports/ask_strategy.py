from __future__ import annotations

from abc import ABC, abstractmethod
from collections.abc import AsyncIterator

from core.domain.stream_events import AskStreamEvent
from core.usecases.contracts import AskAICommand, AskAIResult


class AskStrategyPort(ABC):
    @abstractmethod
    async def execute(self, command: AskAICommand) -> AskAIResult: ...

    @abstractmethod
    async def execute_stream(self, command: AskAICommand) -> AsyncIterator[AskStreamEvent]: ...
