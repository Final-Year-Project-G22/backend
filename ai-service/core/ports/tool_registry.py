from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Any

from core.domain.tools import ToolResult
from core.ports.llm import ToolDefinition


class ToolRegistryPort(ABC):
    @abstractmethod
    async def get_tool_definitions(self) -> list[ToolDefinition]: ...

    @abstractmethod
    async def execute_tool(
        self,
        name: str,
        arguments: dict[str, Any],
        account_id: str,
        user_id: str,
    ) -> ToolResult: ...
