from __future__ import annotations

from enum import StrEnum
from typing import Any

from core.domain.models import AIChatMessage, AIConversationSession
from core.domain.value_objects import SearchHit


class AskStreamEventType(StrEnum):
    TEXT = "text"
    THINKING = "thinking"
    TOOL_CALL = "tool_call"
    TOOL_RESULT = "tool_result"
    TOOL_SUPPRESSED = "tool_suppressed"
    CITATIONS = "citations"
    DONE = "done"
    ERROR = "error"


class AskStreamEvent:
    def __init__(
        self,
        type_: AskStreamEventType = AskStreamEventType.TEXT,
        text: str | None = None,
        tool_name: str | None = None,
        tool_arguments: dict[str, Any] | None = None,
        tool_result_summary: str | None = None,
        suppression_reason: str | None = None,
        matched_query: str | None = None,
        citations: list[dict[str, Any]] | None = None,
        done: AIConversationSession | None = None,
        ai_message: AIChatMessage | None = None,
        merged_hits: list[SearchHit] | None = None,
        error_code: str | None = None,
        error_message: str | None = None,
    ) -> None:
        self.type = type_
        self.text = text
        self.tool_name = tool_name
        self.tool_arguments = tool_arguments
        self.tool_result_summary = tool_result_summary
        self.suppression_reason = suppression_reason
        self.matched_query = matched_query
        self.citations = citations
        self.done = done
        self.ai_message = ai_message
        self.merged_hits = merged_hits
        self.error_code = error_code
        self.error_message = error_message

    @property
    def is_text(self) -> bool:
        return self.type is AskStreamEventType.TEXT

    @property
    def is_thinking(self) -> bool:
        return self.type is AskStreamEventType.THINKING

    @property
    def is_tool_call(self) -> bool:
        return self.type is AskStreamEventType.TOOL_CALL

    @property
    def is_tool_result(self) -> bool:
        return self.type is AskStreamEventType.TOOL_RESULT

    @property
    def is_tool_suppressed(self) -> bool:
        return self.type is AskStreamEventType.TOOL_SUPPRESSED

    @property
    def is_done(self) -> bool:
        return self.type is AskStreamEventType.DONE

    @property
    def is_error(self) -> bool:
        return self.type is AskStreamEventType.ERROR
