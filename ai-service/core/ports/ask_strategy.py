from __future__ import annotations

from abc import ABC, abstractmethod
from collections.abc import AsyncIterator
from typing import TYPE_CHECKING

from core.domain.stream_events import AskStreamEvent

if TYPE_CHECKING:
    from core.ports.llm import LLMPort
    from core.usecases.contracts import AskAICommand, AskAIResult


class AskStrategyPort(ABC):
    @property
    @abstractmethod
    def llm_port(self) -> LLMPort: ...

    @abstractmethod
    async def execute(self, command: AskAICommand) -> AskAIResult: ...

    @abstractmethod
    def execute_stream(self, command: AskAICommand) -> AsyncIterator[AskStreamEvent]: ...
