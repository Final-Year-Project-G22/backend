from __future__ import annotations

import json
from collections.abc import AsyncIterator
from typing import Any

import httpx

from core.domain.exceptions import LLMError
from core.ports.llm import LLMPort


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
            "message": prompt,
            "max_tokens": max_tokens,
            "temperature": temperature,
        }
        try:
            response = await self._http.post(
                "https://api.cohere.com/v2/chat",
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
        text = data.get("message", {}).get("content")
        if not isinstance(text, str):
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
                "message": prompt,
                "max_tokens": max_tokens,
                "temperature": temperature,
                "stream": True,
            }
            try:
                async with self._http.stream(
                    "POST",
                    "https://api.cohere.com/v2/chat",
                    headers={
                        "Authorization": f"Bearer {self._api_key}",
                        "Content-Type": "application/json",
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
                        text = data.get("message", {}).get("content")
                        if isinstance(text, str) and text:
                            yield text
            except httpx.HTTPError as exc:
                raise LLMError(
                    "cohere streaming generation failed",
                    details={"provider": self.provider},
                ) from exc

        return _stream()


__all__ = ["CohereLLMAdapter"]
