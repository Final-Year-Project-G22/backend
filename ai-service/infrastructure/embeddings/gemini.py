from __future__ import annotations

import logging
from typing import Any, cast

import httpx

from core.domain.exceptions import EmbeddingError
from core.ports.embedding import EmbeddingPort

logger = logging.getLogger(__name__)


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
        return [await self._embed_single(text, input_type=input_type) for text in texts]

    async def embed_query(self, query: str) -> list[float]:
        return await self._embed_single(query, input_type="search_query")

    async def _embed_single(
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

        url, params, headers = self._build_request()
        extra_params = {"key": self._api_key} if not self._use_vertex else {}
        params = {**extra_params, **params}

        try:
            response = await self._http.post(
                url,
                params=params or None,
                headers=headers or None,
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

    def _build_request(self) -> tuple[str, dict[str, str], dict[str, str]]:
        if self._use_vertex:
            url = (
                f"https://{self._vertex_location}-aiplatform.googleapis.com/v1/"
                f"projects/{self._vertex_project}/locations/{self._vertex_location}/"
                f"publishers/google/models/{self._model}:embedContent"
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
                    except Exception:  # nosec
                        pass
                token = creds.token
                if token:
                    headers["Authorization"] = f"Bearer {token}"
            except Exception:
                logger.warning("Vertex AI auth failed, falling back to API key")
                params["key"] = self._api_key
        else:
            url = f"https://generativelanguage.googleapis.com/v1beta/models/{self._model}:embedContent"
            params = {"key": self._api_key}
            headers = {}
        return url, params, headers


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
