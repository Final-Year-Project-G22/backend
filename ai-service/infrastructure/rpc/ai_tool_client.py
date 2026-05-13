from __future__ import annotations

import json
import logging
from typing import Any

import grpc
from core.ports.llm import ToolDefinition
from infrastructure.rpc.grpc_stub_loader import (
    build_execute_tool_request,
    build_list_tools_request,
    get_ai_tool_stub,
)

logger = logging.getLogger(__name__)


class AIToolGrpcClient:
    def __init__(self, endpoint: str) -> None:
        self._endpoint = endpoint
        self._channel: grpc.aio.Channel | None = None
        self._stub: Any = None

    async def _ensure_connected(self) -> Any:
        if self._stub is None:
            self._channel = grpc.aio.insecure_channel(self._endpoint)
            self._stub = get_ai_tool_stub(self._channel)
        return self._stub

    async def list_tools(self) -> list[ToolDefinition]:
        stub = await self._ensure_connected()
        if stub is None:
            logger.error("AIToolService stub not available")
            return []
        try:
            req = build_list_tools_request()
            resp = await stub.ListTools(req)
            return [
                ToolDefinition(
                    name=t.name,
                    description=t.description,
                    parameter_schema_json=t.parameter_schema_json,
                )
                for t in resp.tools
            ]
        except grpc.RpcError:
            logger.exception("Failed to list AI tools")
            return []

    async def execute_tool(
        self,
        name: str,
        arguments: dict[str, Any],
        account_id: str,
        user_id: str,
    ) -> dict[str, Any]:
        stub = await self._ensure_connected()
        if stub is None:
            return {"error": "AIToolService stub not available"}
        try:
            req = build_execute_tool_request(
                tool=name,
                arguments_json=json.dumps(arguments),
                account_id=account_id,
                user_id=user_id,
            )
            resp = await stub.ExecuteTool(req)
            if resp.success:
                return json.loads(resp.result_json)
            logger.exception("Tool %s failed", name)
            return {"error": resp.error_message}  # noqa: TRY300
        except grpc.RpcError as e:
            logger.exception("Failed to execute tool %s", name)
            return {"error": str(e)}

    async def close(self) -> None:
        if self._channel is not None:
            await self._channel.close()
            self._channel = None
            self._stub = None


__all__ = ["AIToolGrpcClient"]
