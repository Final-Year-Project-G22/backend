from __future__ import annotations

from typing import Any, cast

import httpx

from core.domain.exceptions import EmbeddingError
from core.ports.embedding import EmbeddingPort


class OllamaEmbeddingAdapter(EmbeddingPort):
    def __init__(
        self,
        *,
        base_url: str,
        model: str,
        dimensions: int,
        http_client: httpx.AsyncClient,
    ) -> None:
        self._base_url = base_url.rstrip("/")
        self._model = model
        self._dimensions = dimensions
        self._http = http_client

    @property
    def dimensions(self) -> int:
        return self._dimensions

    @property
    def provider(self) -> str:
        return "ollama"

    async def embed_documents(
        self,
        texts: list[str],
        *,
        input_type: str | None = None,
    ) -> list[list[float]]:
        _ = input_type
        return [await self._embed_single(text) for text in texts]

    async def embed_query(self, query: str) -> list[float]:
        return await self._embed_single(query)

    async def _embed_single(self, text: str) -> list[float]:
        payload: dict[str, Any] = {
            "model": self._model,
            "prompt": text,
        }
        try:
            response = await self._http.post(
                f"{self._base_url}/api/embeddings",
                json=payload,
                timeout=30,
            )
            response.raise_for_status()
        except httpx.HTTPError as exc:
            raise EmbeddingError(
                "ollama embedding request failed",
                details={"provider": self.provider},
            ) from exc

        data = response.json()
        embedding = data.get("embedding")
        if not isinstance(embedding, list):
            raise EmbeddingError(
                "ollama embedding response missing embedding",
                details={"provider": self.provider},
            )
        return _ensure_float_list(cast(list[object], embedding))


def _ensure_float_list(raw: list[object]) -> list[float]:
    return [_ensure_float(value) for value in raw]


def _ensure_float(raw: object) -> float:
    if not isinstance(raw, (int, float)):
        raise EmbeddingError("invalid embedding value")
    return float(raw)


__all__ = ["OllamaEmbeddingAdapter"]
