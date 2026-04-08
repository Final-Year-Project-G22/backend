from __future__ import annotations

import json
from collections.abc import AsyncIterator
from typing import Any

import httpx

from core.domain.exceptions import LLMError
from core.ports.llm import LLMPort


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
        max_tokens: int = 1024,
        temperature: float = 0.2,
    ) -> str:
        payload = _build_payload(
            model=self._model,
            prompt=prompt,
            max_tokens=max_tokens,
            temperature=temperature,
            stream=False,
        )
        try:
            response = await self._http.post(
                f"{self._base_url}/api/generate",
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
        text = data.get("response")
        if not isinstance(text, str):
            raise LLMError(
                "ollama response missing text",
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
            payload = _build_payload(
                model=self._model,
                prompt=prompt,
                max_tokens=max_tokens,
                temperature=temperature,
                stream=True,
            )
            try:
                async with self._http.stream(
                    "POST",
                    f"{self._base_url}/api/generate",
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
                        chunk = data.get("response")
                        if isinstance(chunk, str) and chunk:
                            yield chunk
                        if data.get("done") is True:
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
    max_tokens: int,
    temperature: float,
    stream: bool,
) -> dict[str, Any]:
    return {
        "model": model,
        "prompt": prompt,
        "stream": stream,
        "options": {
            "temperature": temperature,
            "num_predict": max_tokens,
        },
    }


__all__ = ["OllamaLLMAdapter"]
