from __future__ import annotations

import json
from collections.abc import AsyncIterator
from typing import Any

import httpx

from core.domain.exceptions import LLMError
from core.ports.llm import LLMChunk, LLMPort, LLMResult, ToolCall, ToolDefinition

_CHAT_URL = "https://api.cohere.com/v2/chat"


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


def _parse_response(data: dict[str, Any], *, provider: str) -> LLMResult:
    message: object = data.get("message")
    if not isinstance(message, dict):
        raise LLMError(
            "cohere response missing message",
            details={"provider": provider},
        )

    content: object = message.get("content")
    text = ""
    if isinstance(content, list) and content:
        first = content[0]
        if isinstance(first, dict):
            t = first.get("text")
            if isinstance(t, str):
                text = t

    tool_calls = _parse_tool_calls(message.get("tool_calls"))
    return LLMResult(text=text, tool_calls=tool_calls)


def _parse_tool_calls(raw: object) -> list[ToolCall] | None:
    if not isinstance(raw, list) or not raw:
        return None
    result: list[ToolCall] = []
    for tc in raw:
        if not isinstance(tc, dict):
            continue
        func: object = tc.get("function")
        if not isinstance(func, dict):
            continue
        name = func.get("name", "")
        if not isinstance(name, str) or not name:
            continue
        args_raw = func.get("arguments")
        if isinstance(args_raw, str):
            try:
                args = json.loads(args_raw)
            except json.JSONDecodeError:
                args = {}
        elif isinstance(args_raw, dict):
            args = args_raw
        else:
            args = {}
        result.append(ToolCall(name=name, arguments=args))
    return result or None


def _parse_stream_chunk(data: dict[str, Any]) -> LLMChunk | None:  # noqa: PLR0912
    event_type = data.get("type")
    if event_type == "content-delta":
        delta: object = data.get("delta")
        if isinstance(delta, dict):
            msg: object = delta.get("message")
            if isinstance(msg, dict):
                content: object = msg.get("content")
                if isinstance(content, dict):
                    text = content.get("text")
                    if isinstance(text, str):
                        return LLMChunk(text=text)
    if event_type == "tool-calls-chunk":
        tool_calls_raw = data.get("tool_call_delta", {}).get("tool_calls")
        if isinstance(tool_calls_raw, list) and tool_calls_raw:
            for tc in tool_calls_raw:
                if not isinstance(tc, dict):
                    continue
                func: object = tc.get("function")
                if not isinstance(func, dict):
                    continue
                name = func.get("name", "")
                if isinstance(name, str) and name:
                    args_str = func.get("arguments", "{}")
                    if isinstance(args_str, str):
                        try:
                            args = json.loads(args_str)
                        except json.JSONDecodeError:
                            args = {}
                    elif isinstance(args_str, dict):
                        args = args_str
                    else:
                        args = {}
                    return LLMChunk(tool_call=ToolCall(name=name, arguments=args))
    return None


__all__ = ["CohereLLMAdapter"]
