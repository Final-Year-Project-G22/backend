from __future__ import annotations

import json
from collections.abc import AsyncIterator
from typing import Any, NotRequired, TypedDict, TypeGuard

import httpx

from core.domain.exceptions import LLMError
from core.ports.llm import LLMChunk, LLMPort, LLMResult, ToolCall, ToolDefinition
from infrastructure.llm.json_guard import is_json_object

_CHAT_URL = "https://api.cohere.com/v2/chat"


class _CohereContentBlock(TypedDict):
    """Well-formed shape of a content block in a Cohere chat message."""

    text: NotRequired[str]


class _CohereFunction(TypedDict):
    """Well-formed shape of the function descriptor of a Cohere tool call."""

    name: NotRequired[str]
    arguments: NotRequired[object]


class _CohereToolCall(TypedDict):
    function: NotRequired[_CohereFunction]


class _CohereMessage(TypedDict):
    content: NotRequired[list[_CohereContentBlock]]
    tool_calls: NotRequired[list[_CohereToolCall]]


class _CohereResponse(TypedDict):
    """Schema of a Cohere v2 chat response."""

    message: _CohereMessage


class _CohereStreamDeltaContent(TypedDict):
    text: NotRequired[str]


class _CohereStreamDeltaMessage(TypedDict):
    content: NotRequired[_CohereStreamDeltaContent]


class _CohereStreamDelta(TypedDict):
    message: NotRequired[_CohereStreamDeltaMessage]


class _CohereContentDeltaEvent(TypedDict):
    """Schema of a Cohere ``content-delta`` stream event."""

    type: str
    delta: NotRequired[_CohereStreamDelta]


class _CohereToolCallDelta(TypedDict):
    tool_calls: NotRequired[list[_CohereToolCall]]


class _CohereToolCallsChunkEvent(TypedDict):
    """Schema of a Cohere ``tool-calls-chunk`` stream event."""

    type: str
    tool_call_delta: NotRequired[_CohereToolCallDelta]


class CohereLLMAdapter(LLMPort):
    def __init__(
        self,
        *,
        api_key: str,
        model: str,
        http_client: httpx.AsyncClient,
    ) -> None:
        self._api_key = api_key
        self._model = model
        self._http = http_client

    @property
    def provider(self) -> str:
        return "cohere"

    @property
    def model(self) -> str:
        return self._model

    async def generate(
        self,
        prompt: str,
        *,
        system_prompt: str | None = None,
        tools: list[ToolDefinition] | None = None,
        max_tokens: int = 1024,
        temperature: float = 0.2,
    ) -> LLMResult:
        payload = _build_payload(
            model=self._model,
            prompt=prompt,
            system_prompt=system_prompt,
            tools=tools,
            max_tokens=max_tokens,
            temperature=temperature,
            stream=False,
        )
        try:
            response = await self._http.post(
                _CHAT_URL,
                headers={
                    "Authorization": f"Bearer {self._api_key}",
                    "Content-Type": "application/json",
                },
                json=payload,
                timeout=60,
            )
            response.raise_for_status()
        except httpx.HTTPError as exc:
            raise LLMError(
                "cohere generation failed",
                details={"provider": self.provider},
            ) from exc

        data = response.json()
        return _parse_response(data, provider=self.provider)

    def generate_stream(
        self,
        prompt: str,
        *,
        system_prompt: str | None = None,
        tools: list[ToolDefinition] | None = None,
        max_tokens: int = 1024,
        temperature: float = 0.2,
    ) -> AsyncIterator[LLMChunk]:
        async def _stream() -> AsyncIterator[LLMChunk]:
            payload = _build_payload(
                model=self._model,
                prompt=prompt,
                system_prompt=system_prompt,
                tools=tools,
                max_tokens=max_tokens,
                temperature=temperature,
                stream=True,
            )
            try:
                async with self._http.stream(
                    "POST",
                    _CHAT_URL,
                    headers={
                        "Authorization": f"Bearer {self._api_key}",
                        "Content-Type": "application/json",
                        "Accept": "text/event-stream",
                    },
                    json=payload,
                    timeout=60,
                ) as response:
                    response.raise_for_status()
                    async for line in response.aiter_lines():
                        if not line:
                            continue
                        if not line.startswith("data:"):
                            continue
                        chunk = line.removeprefix("data:").strip()
                        if chunk == "[DONE]":
                            break
                        try:
                            data = json.loads(chunk)
                        except json.JSONDecodeError:
                            continue
                        parsed = _parse_stream_chunk(data)
                        if parsed:
                            yield parsed
            except httpx.HTTPError as exc:
                raise LLMError(
                    "cohere streaming generation failed",
                    details={"provider": self.provider},
                ) from exc

        return _stream()


def _build_payload(
    *,
    model: str,
    prompt: str,
    system_prompt: str | None,
    tools: list[ToolDefinition] | None,
    max_tokens: int,
    temperature: float,
    stream: bool,
) -> dict[str, Any]:
    messages: list[dict[str, Any]] = [{"role": "user", "content": prompt}]
    if system_prompt:
        messages.insert(0, {"role": "system", "content": system_prompt})

    payload: dict[str, Any] = {
        "model": model,
        "messages": messages,
        "max_tokens": max_tokens,
        "temperature": temperature,
    }
    if stream:
        payload["stream"] = True
    if tools:
        payload["tools"] = [
            {
                "type": "function",
                "function": {
                    "name": t.name,
                    "description": t.description,
                    "parameters": json.loads(t.parameter_schema_json)
                    if t.parameter_schema_json
                    else {"type": "object", "properties": {}},
                },
            }
            for t in tools
        ]
    return payload


def _is_cohere_response(raw: object) -> TypeGuard[_CohereResponse]:
    """Narrow raw JSON to the Cohere v2 chat response schema.

    This is the audited parse boundary for this provider's responses. Only
    the ``message`` backbone is validated here; content blocks and tool
    calls are still inspected leniently by the extractors.
    """
    if not is_json_object(raw):
        return False
    return isinstance(raw.get("message"), dict)


def _parse_response(raw: object, *, provider: str) -> LLMResult:
    if not _is_cohere_response(raw):
        raise LLMError(
            "cohere response missing message",
            details={"provider": provider},
        )
    message = raw["message"]
    text = _extract_text(message)
    tool_calls = _parse_tool_calls(message.get("tool_calls"))
    if not text and not tool_calls:
        raise LLMError(
            "cohere response missing content",
            details={"provider": provider},
        )
    return LLMResult(text=text, tool_calls=tool_calls)


def _extract_text(message: _CohereMessage) -> str:
    content = message.get("content")
    if not isinstance(content, list) or not content:
        return ""
    first = content[0]
    if not is_json_object(first):
        return ""
    first_payload: dict[str, object] = first
    text = first_payload.get("text")
    return text if isinstance(text, str) else ""


def _parse_tool_calls(raw: list[_CohereToolCall] | None) -> list[ToolCall] | None:
    if not raw:
        return None
    result: list[ToolCall] = []
    for tool_call in raw:
        if not is_json_object(tool_call):
            continue
        tool_call_payload: dict[str, object] = tool_call
        function = tool_call_payload.get("function")
        if not is_json_object(function):
            continue
        function_payload: dict[str, object] = function
        name = function_payload.get("name")
        if not isinstance(name, str) or not name:
            continue
        result.append(
            ToolCall(
                name=name,
                arguments=_parse_arguments(function_payload.get("arguments")),
            )
        )
    return result or None


def _parse_arguments(raw: object) -> dict[str, object]:
    if isinstance(raw, str):
        try:
            parsed = json.loads(raw)
        except json.JSONDecodeError:
            return {}
        return parsed if is_json_object(parsed) else {}
    if is_json_object(raw):
        return raw
    return {}


def _is_content_delta_event(raw: object) -> TypeGuard[_CohereContentDeltaEvent]:  # noqa: PLR0911
    if not is_json_object(raw):
        return False
    if raw.get("type") != "content-delta":
        return False
    delta = raw.get("delta")
    if delta is None:
        return True
    if not is_json_object(delta):
        return False
    message = delta.get("message")
    if message is None:
        return True
    if not is_json_object(message):
        return False
    content = message.get("content")
    if content is None:
        return True
    if not is_json_object(content):
        return False
    text = content.get("text")
    return text is None or isinstance(text, str)


def _is_tool_calls_chunk_event(raw: object) -> TypeGuard[_CohereToolCallsChunkEvent]:
    if not is_json_object(raw):
        return False
    if raw.get("type") != "tool-calls-chunk":
        return False
    tool_call_delta = raw.get("tool_call_delta")
    if tool_call_delta is None:
        return True
    if not is_json_object(tool_call_delta):
        return False
    tool_calls = tool_call_delta.get("tool_calls")
    if tool_calls is None:
        return True
    return isinstance(tool_calls, list)


def _parse_stream_chunk(raw: object) -> LLMChunk | None:
    if not is_json_object(raw):
        return None
    event_type = raw.get("type")
    if event_type == "content-delta" and _is_content_delta_event(raw):
        return _parse_content_delta(raw)
    if event_type == "tool-calls-chunk" and _is_tool_calls_chunk_event(raw):
        return _parse_tool_calls_chunk(raw)
    return None


def _parse_content_delta(event: _CohereContentDeltaEvent) -> LLMChunk | None:
    delta = event.get("delta")
    if delta is None:
        return None
    message = delta.get("message")
    if message is None:
        return None
    content = message.get("content")
    if content is None:
        return None
    text = content.get("text")
    if text is None:
        return None
    return LLMChunk(text=text)


def _parse_tool_calls_chunk(event: _CohereToolCallsChunkEvent) -> LLMChunk | None:
    tool_call_delta = event.get("tool_call_delta")
    if tool_call_delta is None:
        return None
    tool_calls = tool_call_delta.get("tool_calls")
    if not tool_calls:
        return None
    for tool_call in tool_calls:
        if not is_json_object(tool_call):
            continue
        tool_call_payload: dict[str, object] = tool_call
        function = tool_call_payload.get("function")
        if not is_json_object(function):
            continue
        function_payload: dict[str, object] = function
        name = function_payload.get("name")
        if isinstance(name, str) and name:
            return LLMChunk(
                tool_call=ToolCall(
                    name=name,
                    arguments=_parse_arguments(function_payload.get("arguments")),
                )
            )
    return None


__all__ = ["CohereLLMAdapter"]
