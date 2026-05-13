from __future__ import annotations

import json
from collections.abc import AsyncIterator
from typing import Any

import httpx

from core.domain.exceptions import LLMError
from core.ports.llm import LLMChunk, LLMPort, LLMResult, ToolCall, ToolDefinition


class GeminiLLMAdapter(LLMPort):
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
        return "gemini"

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
            prompt=prompt,
            system_prompt=system_prompt,
            tools=tools,
            max_tokens=max_tokens,
            temperature=temperature,
        )
        try:
            response = await self._http.post(
                f"https://generativelanguage.googleapis.com/v1beta/models/{self._model}:generateContent",
                params={"key": self._api_key},
                json=payload,
                timeout=60,
            )
            response.raise_for_status()
        except httpx.HTTPError as exc:
            raise LLMError(
                "gemini generation failed",
                details={"provider": self.provider},
            ) from exc

        data = response.json()
        return _parse_response(data)

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
                prompt=prompt,
                system_prompt=system_prompt,
                tools=tools,
                max_tokens=max_tokens,
                temperature=temperature,
            )
            try:
                async with self._http.stream(
                    "POST",
                    f"https://generativelanguage.googleapis.com/v1beta/models/{self._model}:streamGenerateContent",
                    params={"key": self._api_key},
                    json=payload,
                    timeout=60,
                ) as response:
                    response.raise_for_status()
                    async for line in response.aiter_lines():
                        if not line:
                            continue
                        payload_line = line
                        if payload_line.startswith("data:"):
                            payload_line = payload_line.removeprefix("data:").strip()
                        try:
                            data = json.loads(payload_line)
                        except json.JSONDecodeError:
                            continue
                        chunk = _parse_stream_chunk(data)
                        if chunk is not None:
                            yield chunk
            except httpx.HTTPError as exc:
                raise LLMError(
                    "gemini streaming generation failed",
                    details={"provider": self.provider},
                ) from exc

        return _stream()


def _build_payload(
    *,
    prompt: str,
    system_prompt: str | None,
    tools: list[ToolDefinition] | None,
    max_tokens: int,
    temperature: float,
) -> dict[str, Any]:
    payload: dict[str, Any] = {
        "contents": [{"role": "user", "parts": [{"text": prompt}]}],
        "generationConfig": {
            "maxOutputTokens": max_tokens,
            "temperature": temperature,
        },
    }
    if system_prompt:
        payload["systemInstruction"] = {"parts": [{"text": system_prompt}]}
    if tools:
        payload["tools"] = [
            {
                "functionDeclarations": [
                    {
                        "name": t.name,
                        "description": t.description,
                        "parameters": json.loads(t.parameter_schema_json)
                        if t.parameter_schema_json
                        else {"type": "object", "properties": {}},
                    }
                    for t in tools
                ]
            }
        ]
    return payload


def _parse_response(data: dict[str, Any]) -> LLMResult:
    text = _extract_text_optional(data)
    tool_calls = _extract_tool_calls(data)
    return LLMResult(text=text or "", tool_calls=tool_calls)


def _extract_text_optional(data: dict[str, Any]) -> str | None:
    candidates = data.get("candidates")
    if not _is_object_list(candidates) or not candidates:
        return None
    first = candidates[0]
    if not _is_object_dict(first):
        return None
    content = first.get("content")
    if not _is_object_dict(content):
        return None
    parts = content.get("parts")
    if not _is_object_list(parts) or not parts:
        return None
    for part in parts:
        if not _is_object_dict(part):
            continue
        text = part.get("text")
        if isinstance(text, str):
            return text
    return None


def _extract_tool_calls(data: dict[str, Any]) -> list[ToolCall] | None:
    candidates = data.get("candidates")
    if not _is_object_list(candidates) or not candidates:
        return None
    first = candidates[0]
    if not _is_object_dict(first):
        return None
    content = first.get("content")
    if not _is_object_dict(content):
        return None
    parts = content.get("parts")
    if not _is_object_list(parts) or not parts:
        return None
    result: list[ToolCall] = []
    for part in parts:
        if not _is_object_dict(part):
            continue
        fc = part.get("functionCall")
        if not _is_object_dict(fc):
            continue
        name = fc.get("name")
        if not isinstance(name, str) or not name:
            continue
        args = fc.get("args")
        result.append(ToolCall(name=name, arguments=args if isinstance(args, dict) else {}))
    return result or None


def _parse_stream_chunk(data: dict[str, Any]) -> LLMChunk | None:
    text = _extract_text_optional(data)
    if text:
        return LLMChunk(text=text)
    return None


def _is_object_dict(raw: object) -> bool:
    return isinstance(raw, dict)


def _is_object_list(raw: object) -> bool:
    return isinstance(raw, list)


__all__ = ["GeminiLLMAdapter"]
