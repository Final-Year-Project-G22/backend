from __future__ import annotations

import asyncio
from typing import Any, TypeGuard

import httpx
from tenacity import retry, retry_if_exception, stop_after_attempt, wait_exponential

from core.domain.exceptions import EmbeddingError
from core.ports.embedding import EmbeddingPort

COHERE_BATCH_SIZE = 96
COHERE_BATCH_DELAY_S = 2.0
COHERE_MAX_RETRIES = 5
_429_TOO_MANY_REQUESTS = 429


def _is_rate_limit(exc: BaseException) -> bool:
    if isinstance(exc, EmbeddingError):
        cause = exc.__cause__
        if (
            isinstance(cause, httpx.HTTPStatusError)
            and cause.response.status_code == _429_TOO_MANY_REQUESTS
        ):
            return True
    return False


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
        results: list[list[float]] = []
        for i in range(0, len(texts), COHERE_BATCH_SIZE):
            batch = texts[i : i + COHERE_BATCH_SIZE]
            batch_results = await self._embed_batch(texts=batch, input_type=input_type)
            results.extend(batch_results)
            if i + COHERE_BATCH_SIZE < len(texts):
                await asyncio.sleep(COHERE_BATCH_DELAY_S)
        return results

    async def embed_query(self, query: str) -> list[float]:
        results = await self._embed_batch(texts=[query], input_type="search_query")
        return results[0]

    @retry(
        stop=stop_after_attempt(COHERE_MAX_RETRIES),
        wait=wait_exponential(multiplier=2, min=2, max=30),
        retry=retry_if_exception(_is_rate_limit),
    )
    async def _embed_batch(
        self,
        *,
        texts: list[str],
        input_type: str | None,
    ) -> list[list[float]]:
        payload: dict[str, Any] = {
            "model": self._model,
            "texts": texts,
        }
        # embed-multilingual-v3.0 requires input_type; default to the
        # document side so callers embedding into the index can omit it.
        payload["input_type"] = input_type or "search_document"

        embed_url = "https://api.cohere.com/v1/embed"

        try:
            response = await self._http.post(
                embed_url,
                headers={
                    "Authorization": f"Bearer {self._api_key}",
                    "Content-Type": "application/json",
                },
                json=payload,
                timeout=30,
            )
            response.raise_for_status()
        except httpx.HTTPStatusError as exc:
            raise EmbeddingError(
                "cohere embedding request failed",
                details={
                    "provider": self.provider,
                    "status": exc.response.status_code,
                    "response": exc.response.text,
                },
            ) from exc
        except httpx.HTTPError as exc:
            raise EmbeddingError(
                "cohere embedding request failed",
                details={
                    "provider": self.provider,
                    "reason": type(exc).__name__,
                    "detail": str(exc),
                },
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
