from __future__ import annotations

from typing import Any, TypeGuard

import httpx

from core.domain.exceptions import EmbeddingError
from core.ports.embedding import EmbeddingPort


class CohereEmbeddingAdapter(EmbeddingPort):
    def __init__(
        self,
        *,
        api_key: str,
        model: str,
        dimensions: int,
        http_client: httpx.AsyncClient,
    ) -> None:
        self._api_key = api_key
        self._model = model
        self._dimensions = dimensions
        self._http = http_client

    @property
    def dimensions(self) -> int:
        return self._dimensions

    @property
    def provider(self) -> str:
        return "cohere"

    async def embed_documents(
        self,
        texts: list[str],
        *,
        input_type: str | None = None,
    ) -> list[list[float]]:
        return await self._embed(texts=texts, input_type=input_type)

    async def embed_query(self, query: str) -> list[float]:
        embeddings = await self._embed(texts=[query], input_type="search_query")
        return embeddings[0]

    async def _embed(
        self,
        *,
        texts: list[str],
        input_type: str | None,
    ) -> list[list[float]]:
        payload: dict[str, Any] = {
            "model": self._model,
            "texts": texts,
        }
        if input_type is not None:
            payload["input_type"] = input_type

        try:
            response = await self._http.post(
                "https://api.cohere.com/v1/embed",
                headers={
                    "Authorization": f"Bearer {self._api_key}",
                    "Content-Type": "application/json",
                },
                json=payload,
                timeout=30,
            )
            response.raise_for_status()
        except httpx.HTTPError as exc:
            raise EmbeddingError(
                "cohere embedding request failed",
                details={"provider": self.provider},
            ) from exc

        data = response.json()
        if not _is_object_dict(data):
            raise EmbeddingError(
                "cohere embedding response has invalid shape",
                details={"provider": self.provider},
            )

        embeddings = data.get("embeddings")
        if not _is_object_list(embeddings):
            raise EmbeddingError(
                "cohere embedding response missing embeddings",
                details={"provider": self.provider},
            )

        return _parse_embeddings(embeddings)


def _parse_embeddings(raw: list[object]) -> list[list[float]]:
    return [_ensure_float_list(item) for item in raw]


def _ensure_float_list(raw: object) -> list[float]:
    if not _is_object_list(raw):
        raise EmbeddingError("invalid embedding vector")
    return [_ensure_float(value) for value in raw]


def _ensure_float(raw: object) -> float:
    if not isinstance(raw, (int, float)):
        raise EmbeddingError("invalid embedding value")
    return float(raw)


def _is_object_dict(raw: Any) -> TypeGuard[dict[str, object]]:
    return isinstance(raw, dict)


def _is_object_list(raw: Any) -> TypeGuard[list[object]]:
    return isinstance(raw, list)


__all__ = ["CohereEmbeddingAdapter"]
