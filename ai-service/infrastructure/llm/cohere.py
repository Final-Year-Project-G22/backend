from __future__ import annotations

import json
from collections.abc import AsyncIterator
from typing import Any

import httpx

from core.domain.exceptions import LLMError
from core.ports.llm import LLMPort

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
        max_tokens: int = 1024,
        temperature: float = 0.2,
    ) -> str:
        payload: dict[str, Any] = {
            "model": self._model,
            "messages": [{"role": "user", "content": prompt}],
            "max_tokens": max_tokens,
            "temperature": temperature,
        }
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
        text = _extract_text(data)
        if text is None:
            raise LLMError(
                "cohere generation response missing content",
                details={"provider": self.provider},
            )
        return text

    def generate_stream(
        self,
        prompt: str,
        *,
        max_tokens: int = 1024,
        temperature: float = 0.2,
    ) -> AsyncIterator[str]:
        async def _stream() -> AsyncIterator[str]:
            payload: dict[str, Any] = {
                "model": self._model,
                "messages": [{"role": "user", "content": prompt}],
                "max_tokens": max_tokens,
                "temperature": temperature,
                "stream": True,
            }
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
                        text = _extract_stream_text(data)
                        if text:
                            yield text
            except httpx.HTTPError as exc:
                raise LLMError(
                    "cohere streaming generation failed",
                    details={"provider": self.provider},
                ) from exc

        return _stream()


def _extract_text(data: dict[str, Any]) -> str | None:
    message: object = data.get("message")
    if not isinstance(message, dict):
        return None
    content: object = message.get("content")  # type: ignore[reportUnknownMemberType]
    if not isinstance(content, list) or not content:
        return None
    first: object = content[0]  # type: ignore[reportUnknownVariableType]
    if not isinstance(first, dict):
        return None
    result: object = first.get("text")  # type: ignore[reportUnknownMemberType]
    return result if isinstance(result, str) else None  # type: ignore[reportUnknownVariableType]


def _extract_stream_text(data: dict[str, Any]) -> str | None:
    if data.get("type") != "content-delta":
        return None
    delta: object = data.get("delta")
    if not isinstance(delta, dict):
        return None
    msg: object = delta.get("message")  # type: ignore[reportUnknownMemberType]
    if not isinstance(msg, dict):
        return None
    content: object = msg.get("content")  # type: ignore[reportUnknownMemberType]
    if not isinstance(content, dict):
        return None
    text: object = content.get("text")  # type: ignore[reportUnknownMemberType]
    return text if isinstance(text, str) else None  # type: ignore[reportUnknownVariableType]


__all__ = ["CohereLLMAdapter"]
