from __future__ import annotations

import json
from collections.abc import AsyncIterator
from typing import Any

import httpx

from core.domain.exceptions import LLMError
from core.ports.llm import LLMPort


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
        max_tokens: int = 1024,
        temperature: float = 0.2,
    ) -> str:
        payload = _build_payload(prompt, max_tokens=max_tokens, temperature=temperature)
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
        return _extract_text(data, provider=self.provider)

    def generate_stream(
        self,
        prompt: str,
        *,
        max_tokens: int = 1024,
        temperature: float = 0.2,
    ) -> AsyncIterator[str]:
        async def _stream() -> AsyncIterator[str]:
            payload = _build_payload(prompt, max_tokens=max_tokens, temperature=temperature)
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
                        text = _extract_text(data, provider=self.provider, allow_empty=True)
                        if text:
                            yield text
            except httpx.HTTPError as exc:
                raise LLMError(
                    "gemini streaming generation failed",
                    details={"provider": self.provider},
                ) from exc

        return _stream()


def _build_payload(prompt: str, *, max_tokens: int, temperature: float) -> dict[str, Any]:
    return {
        "contents": [{"role": "user", "parts": [{"text": prompt}]}],
        "generationConfig": {
            "maxOutputTokens": max_tokens,
            "temperature": temperature,
        },
    }


def _extract_text(
    data: dict[str, Any],
    *,
    provider: str,
    allow_empty: bool = False,
) -> str:
    text = _extract_text_optional(data)
    if isinstance(text, str):
        return text
    if allow_empty:
        return ""
    raise LLMError(
        "gemini response missing text",
        details={"provider": provider},
    )


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

    first_part = parts[0]
    if not _is_object_dict(first_part):
        return None

    text = first_part.get("text")
    return text if isinstance(text, str) else None


def _is_object_dict(raw: object) -> bool:
    return isinstance(raw, dict)


def _is_object_list(raw: object) -> bool:
    return isinstance(raw, list)


__all__ = ["GeminiLLMAdapter"]
