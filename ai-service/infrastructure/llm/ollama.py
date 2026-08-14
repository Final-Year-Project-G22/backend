from __future__ import annotations

import json
from collections.abc import AsyncIterator
from typing import Any, NotRequired, TypedDict, TypeGuard

import httpx

from core.domain.exceptions import LLMError
from core.ports.llm import LLMChunk, LLMPort, LLMResult, ToolCall, ToolDefinition
from infrastructure.llm.json_guard import is_json_object


class _OllamaFunction(TypedDict):
    """Well-formed shape of the function descriptor of an Ollama tool call."""

    name: NotRequired[str]
    arguments: NotRequired[object]


class _OllamaToolCall(TypedDict):
    function: NotRequired[_OllamaFunction]


class _OllamaMessage(TypedDict):
    content: NotRequired[str]
    tool_calls: NotRequired[list[_OllamaToolCall]]


class _OllamaResponse(TypedDict):
    """Schema of an Ollama chat response (and its stream chunks)."""

    message: _OllamaMessage
    done: NotRequired[bool]


class OllamaLLMAdapter(LLMPort):
    def __init__(
        self,
        *,
        base_url: str,
        model: str,
        http_client: httpx.AsyncClient,
    ) -> None:
        self._base_url = base_url.rstrip("/")
        self._model = model
        self._http = http_client

    @property
    def provider(self) -> str:
        return "ollama"

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
                f"{self._base_url}/api/chat",
                json=payload,
                timeout=60,
            )
            response.raise_for_status()
        except httpx.HTTPError as exc:
            raise LLMError(
                "ollama generation failed",
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
                    f"{self._base_url}/api/chat",
                    json=payload,
                    timeout=60,
                ) as response:
                    response.raise_for_status()
                    async for line in response.aiter_lines():
                        if not line:
                            continue
                        try:
                            data = json.loads(line)
                        except json.JSONDecodeError:
                            continue
                        chunk = _parse_stream_chunk(data)
                        if chunk is not None:
                            yield chunk
                        if is_json_object(data) and data.get("done") is True:
                            break
            except httpx.HTTPError as exc:
                raise LLMError(
                    "ollama streaming generation failed",
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
    messages: list[dict[str, str]] = []
    if system_prompt:
        messages.append({"role": "system", "content": system_prompt})
    messages.append({"role": "user", "content": prompt})

    payload: dict[str, Any] = {
        "model": model,
        "messages": messages,
        "options": {
            "temperature": temperature,
            "num_predict": max_tokens,
        },
        "stream": stream,
    }
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


def _is_ollama_response(raw: object) -> TypeGuard[_OllamaResponse]:
    """Narrow raw JSON to the Ollama chat response schema.

    This is the audited parse boundary for this provider: only the
    ``message`` backbone is validated here, while message fields are still
    inspected leniently by the extractors.
    """
    if not is_json_object(raw):
        return False
    return isinstance(raw.get("message"), dict)


def _parse_response(raw: object, *, provider: str) -> LLMResult:
    if not _is_ollama_response(raw):
        raise LLMError(
            "ollama response missing message",
            details={"provider": provider},
        )
    message = raw["message"]
    content = message.get("content")
    text = content if isinstance(content, str) else ""
    tool_calls = _parse_tool_calls(message.get("tool_calls"))
    if not text and not tool_calls:
        raise LLMError(
            "ollama response missing text",
            details={"provider": provider},
        )
    return LLMResult(text=text, tool_calls=tool_calls)


def _parse_tool_calls(raw: list[_OllamaToolCall] | None) -> list[ToolCall] | None:
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
        raw_args = function_payload.get("arguments")
        arguments = raw_args if is_json_object(raw_args) else {}
        result.append(ToolCall(name=name, arguments=arguments))
    return result or None


def _parse_stream_chunk(raw: object) -> LLMChunk | None:
    if not _is_ollama_response(raw):
        return None
    message = raw["message"]
    content = message.get("content")
    if isinstance(content, str) and content:
        return LLMChunk(text=content)
    tool_calls = _parse_tool_calls(message.get("tool_calls"))
    if tool_calls:
        return LLMChunk(tool_call=tool_calls[0])
    return None


__all__ = ["OllamaLLMAdapter"]
