from __future__ import annotations

import json
import logging
import time
from typing import Any

from core.domain.tools import ToolResult
from core.ports.llm import ToolDefinition
from core.ports.tool_registry import ToolRegistryPort
from infrastructure.rpc.ai_tool_client import AIToolGrpcClient
from infrastructure.tools.local.search_knowledge_base import SearchKnowledgeBaseTool
from infrastructure.tools.local.search_trusted_web import SearchTrustedWebTool

logger = logging.getLogger(__name__)

_MAX_SUMMARY_LENGTH = 500


class ToolRegistry(ToolRegistryPort):
    def __init__(
        self,
        ai_tool_client: AIToolGrpcClient,
        expected_remote_tools: list[str],
        *,
        search_knowledge_base: SearchKnowledgeBaseTool,
        search_trusted_web: SearchTrustedWebTool,
    ) -> None:
        self._client = ai_tool_client
        self._expected_remote_tools = expected_remote_tools

        self._local_tools: dict[str, Any] = {}
        self._register_local(search_knowledge_base)
        self._register_local(search_trusted_web)

        self._remote_defs: list[ToolDefinition] = []
        self._remote_map: dict[str, ToolDefinition] = {}

    def _register_local(self, tool: Any) -> None:
        self._local_tools[tool.name] = tool

    async def initialize(self) -> None:
        await self._refresh_remote()

    async def _refresh_remote(self) -> None:
        try:
            remote_tools = await self._client.list_tools()
        except Exception:
            logger.exception("Failed to fetch remote tool definitions")
            return

        filtered = [t for t in remote_tools if t.name in self._expected_remote_tools]

        self._remote_defs = filtered
        self._remote_map = {t.name: t for t in filtered}

        found = {t.name for t in filtered}
        missing = set(self._expected_remote_tools) - found
        if missing:
            logger.warning("Expected remote tools not available: %s", sorted(missing))

    async def get_tool_definitions(self) -> list[ToolDefinition]:
        defs: list[ToolDefinition] = []
        for name, tool in self._local_tools.items():
            defs.append(
                ToolDefinition(
                    name=name,
                    description=tool.description,
                    parameter_schema_json=json.dumps(tool.parameter_schema),
                )
            )
        defs.extend(self._remote_defs)
        return defs

    async def execute_tool(
        self,
        name: str,
        arguments: dict[str, Any],
        account_id: str,
        user_id: str,
    ) -> ToolResult:
        if name in self._local_tools:
            return await self._execute_local(name, arguments, account_id, user_id)
        if name in self._remote_map:
            return await self._execute_remote(name, arguments, account_id, user_id)

        return ToolResult(
            tool_name=name,
            arguments=arguments,
            result_text=f"Unknown tool: {name}",
            success=False,
            error_message=f"Tool '{name}' is not registered",
        )

    async def _execute_local(
        self,
        name: str,
        arguments: dict[str, Any],
        account_id: str,
        user_id: str,
    ) -> ToolResult:
        tool = self._local_tools[name]
        start = time.perf_counter()
        try:
            result = await tool.execute(arguments, account_id, user_id)
        except Exception as e:
            elapsed = int((time.perf_counter() - start) * 1000)
            logger.exception("Local tool %s failed", name)
            return ToolResult(
                tool_name=name,
                arguments=arguments,
                result_text=f"Tool execution failed: {e}",
                success=False,
                error_message=str(e),
                execution_ms=elapsed,
            )
        else:
            elapsed = int((time.perf_counter() - start) * 1000)
            result.execution_ms = elapsed
            return result

    async def _execute_remote(
        self,
        name: str,
        arguments: dict[str, Any],
        account_id: str,
        user_id: str,
    ) -> ToolResult:
        start = time.perf_counter()
        try:
            raw = await self._client.execute_tool(name, arguments, account_id, user_id)
        except Exception as e:
            elapsed = int((time.perf_counter() - start) * 1000)
            logger.exception("Remote tool %s RPC failed", name)
            return ToolResult(
                tool_name=name,
                arguments=arguments,
                result_text=f"Tool execution failed: {e}",
                success=False,
                error_message=str(e),
                execution_ms=elapsed,
            )

        elapsed = int((time.perf_counter() - start) * 1000)

        if "error" in raw:
            return ToolResult(
                tool_name=name,
                arguments=arguments,
                result_text=raw.get("error", "Unknown error"),
                success=False,
                error_message=raw.get("error"),
                execution_ms=elapsed,
            )

        summary = json.dumps(raw, default=str, ensure_ascii=False)
        if len(summary) > _MAX_SUMMARY_LENGTH:
            summary = summary[: _MAX_SUMMARY_LENGTH - 3] + "..."

        return ToolResult(
            tool_name=name,
            arguments=arguments,
            result_text=summary,
            success=True,
            execution_ms=elapsed,
        )
