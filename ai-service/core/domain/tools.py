from __future__ import annotations

from typing import Any

from pydantic import BaseModel, Field

from core.domain.value_objects import SearchHit


class ToolResult(BaseModel):
    tool_name: str
    arguments: dict[str, Any] = Field(default_factory=dict)
    result_text: str = ""
    success: bool = True
    error_message: str | None = None
    execution_ms: int = 0
    hits: list[SearchHit] = Field(default_factory=list)


__all__ = ["ToolResult"]
