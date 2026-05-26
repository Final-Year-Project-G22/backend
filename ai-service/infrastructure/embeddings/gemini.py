from __future__ import annotations

import asyncio
from collections.abc import Sequence
from typing import Any, cast

import httpx
from tenacity import retry, retry_if_exception, stop_after_attempt, wait_exponential

from core.domain.exceptions import EmbeddingError
from core.ports.embedding import EmbeddingPort

BATCH_SIZE = 50
BATCH_DELAY_S = 2.0
MAX_RETRIES = 5
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


class GeminiEmbeddingAdapter(EmbeddingPort):
    def __init__(
        self,
        *,
        api_key: str,
        model: str,
        dimensions: int,
        http_client: httpx.AsyncClient,
        use_vertex: bool = False,
        vertex_project: str = "",
        vertex_location: str = "us-central1",
    ) -> None:
        self._api_key = api_key
        self._model = model
        self._dimensions = dimensions
        self._http = http_client
        self._use_vertex = use_vertex
        self._vertex_project = vertex_project
        self._vertex_location = vertex_location
        self._url: str | None = None
        self._params: dict[str, str] = {}
        self._headers: dict[str, str] = {}

    @property
    def dimensions(self) -> int:
        return self._dimensions

    @property
    def provider(self) -> str:
        return "gemini"

    async def embed_documents(
        self,
        texts: list[str],
        *,
        input_type: str | None = None,
    ) -> list[list[float]]:
        self._ensure_request_built()
        if self._use_vertex:
            return await self._embed_batches_vertex(texts, input_type=input_type)
        results: list[list[float]] = []
        for text in texts:
            result = await self._embed_single_gemini(text, input_type=input_type)
            results.append(result)
        return results

    async def embed_query(self, query: str) -> list[float]:
        self._ensure_request_built()
        if self._use_vertex:
            results = await self._embed_batches_vertex([query], input_type="search_query")
            return results[0]
        return await self._embed_single_gemini(query, input_type="search_query")

    def _ensure_request_built(self) -> None:
        if self._url is not None:
            return
        url, params, headers, auth_ok = self._build_request()
        if self._use_vertex and not auth_ok:
            raise EmbeddingError(
                "Vertex AI auth failed — set GOOGLE_APPLICATION_CREDENTIALS or configure a service account",
                details={"provider": self.provider},
            )
        self._url = url
        self._params = params
        self._headers = headers

    async def _embed_batches_vertex(
        self,
        texts: list[str],
        *,
        input_type: str | None,
    ) -> list[list[float]]:
        results: list[list[float]] = []
        for i in range(0, len(texts), BATCH_SIZE):
            batch = texts[i : i + BATCH_SIZE]
            batch_results = await self._embed_batch_vertex(batch, input_type=input_type)
            results.extend(batch_results)
            if i + BATCH_SIZE < len(texts):
                await asyncio.sleep(BATCH_DELAY_S)
        return results

    async def _embed_batch_vertex(
        self,
        texts: Sequence[str],
        *,
        input_type: str | None,
    ) -> list[list[float]]:
        return await self._embed_batch_vertex_with_retry(texts, input_type=input_type)

    @retry(
        stop=stop_after_attempt(MAX_RETRIES),
        wait=wait_exponential(multiplier=2, min=2, max=60),
        retry=retry_if_exception(_is_rate_limit),
    )
    async def _embed_batch_vertex_with_retry(
        self,
        texts: Sequence[str],
        *,
        input_type: str | None,
    ) -> list[list[float]]:
        task_type = _map_task_type(input_type)
        payload: dict[str, Any] = {
            "instances": [{"content": t} for t in texts],
        }
        if task_type is not None:
            payload["parameters"] = {"taskType": task_type}

        try:
            response = await self._http.post(
                self._url,
                params=self._params or None,
                headers=self._headers or None,
                json=payload,
                timeout=60,
            )
            response.raise_for_status()
        except httpx.HTTPError as exc:
            raise EmbeddingError(
                "gemini embedding request failed",
                details={"provider": self.provider},
            ) from exc

        data = response.json()
        predictions = data.get("predictions")
        if not isinstance(predictions, list) or len(predictions) != len(texts):
            raise EmbeddingError(
                "vertex ai embedding response missing or mismatched predictions",
                details={"provider": self.provider},
            )
        return [
            _ensure_float_list(cast(list[object], p["embeddings"]["values"])) for p in predictions
        ]

    @retry(
        stop=stop_after_attempt(MAX_RETRIES),
        wait=wait_exponential(multiplier=1, min=1, max=30),
        retry=retry_if_exception(_is_rate_limit),
    )
    async def _embed_single_gemini(
        self,
        text: str,
        *,
        input_type: str | None,
    ) -> list[float]:
        task_type = _map_task_type(input_type)
        payload: dict[str, Any] = {
            "content": {
                "parts": [{"text": text}],
            },
        }
        if task_type is not None:
            payload["taskType"] = task_type
        params = {**self._params, "key": self._api_key}

        try:
            response = await self._http.post(
                self._url,
                params=params or None,
                headers=self._headers or None,
                json=payload,
                timeout=30,
            )
            response.raise_for_status()
        except httpx.HTTPError as exc:
            raise EmbeddingError(
                "gemini embedding request failed",
                details={"provider": self.provider},
            ) from exc

        data = response.json()
        embedding = data.get("embedding", {}).get("values")
        if not isinstance(embedding, list):
            raise EmbeddingError(
                "gemini embedding response missing values",
                details={"provider": self.provider},
            )
        return _ensure_float_list(cast(list[object], embedding))

    def _build_request(self) -> tuple[str, dict[str, str], dict[str, str], bool]:
        if self._use_vertex:
            url = (
                f"https://{self._vertex_location}-aiplatform.googleapis.com/v1/"
                f"projects/{self._vertex_project}/locations/{self._vertex_location}/"
                f"publishers/google/models/{self._model}:predict"
            )
            headers: dict[str, str] = {}
            params: dict[str, str] = {}
            try:
                import google.auth
                import google.auth.transport.requests

                credentials, _ = google.auth.default(
                    scopes=["https://www.googleapis.com/auth/cloud-platform"],
                )
                creds = credentials
                if hasattr(creds, "refresh"):
                    try:
                        request = google.auth.transport.requests.Request()
                        creds.refresh(request)
                    except Exception:
                        pass
                token = creds.token
                if token:
                    headers["Authorization"] = f"Bearer {token}"
                    return url, params, headers, True
            except Exception:
                pass
            return url, params, headers, False
        url = f"https://generativelanguage.googleapis.com/v1beta/models/{self._model}:embedContent"
        params = {"key": self._api_key}
        headers = {}
        return url, params, headers, True


def _map_task_type(input_type: str | None) -> str | None:
    if input_type is None:
        return None
    if input_type == "search_query":
        return "RETRIEVAL_QUERY"
    if input_type == "search_document":
        return "RETRIEVAL_DOCUMENT"
    return None


def _ensure_float_list(raw: list[object]) -> list[float]:
    return [_ensure_float(value) for value in raw]


def _ensure_float(raw: object) -> float:
    if not isinstance(raw, (int, float)):
        raise EmbeddingError("invalid embedding value")
    return float(raw)


__all__ = ["GeminiEmbeddingAdapter"]
