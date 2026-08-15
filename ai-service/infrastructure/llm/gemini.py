from __future__ import annotations

import json
import logging
from collections.abc import AsyncIterator
from typing import Any, NotRequired, Protocol, TypedDict, TypeGuard, cast

import httpx

from core.domain.exceptions import LLMError
from core.ports.llm import LLMChunk, LLMPort, LLMResult, ToolCall, ToolDefinition
from infrastructure.llm.json_guard import is_json_object

logger = logging.getLogger(__name__)


class _GeminiFunctionCall(TypedDict):
    """Well-formed shape of a Gemini ``functionCall`` part."""

    name: NotRequired[str]
    args: NotRequired[object]


class _GeminiPart(TypedDict):
    """Well-formed shape of a single part of a Gemini candidate's content."""

    text: NotRequired[str]
    functionCall: NotRequired[_GeminiFunctionCall]


class _GeminiContent(TypedDict):
    parts: NotRequired[list[_GeminiPart]]


class _GeminiCandidate(TypedDict):
    content: NotRequired[_GeminiContent]


class _GeminiResponse(TypedDict):
    """Schema of a Gemini generateContent response (and its SSE chunks)."""

    candidates: NotRequired[list[_GeminiCandidate]]


class _RefreshableCredentials(Protocol):
    """The subset of google-auth credentials used for Vertex AI auth."""

    token: str | None

    def refresh(self, request: Any) -> None: ...


class GeminiLLMAdapter(LLMPort):
    def __init__(
        self,
        *,
        api_key: str,
        model: str,
        http_client: httpx.AsyncClient,
        use_vertex: bool = False,
        vertex_project: str = "",
        vertex_location: str = "us-central1",
    ) -> None:
        self._api_key = api_key
        self._model = model
        self._http = http_client
        self._use_vertex = use_vertex
        self._vertex_project = vertex_project
        self._vertex_location = vertex_location
        self._auth_token: str | None = None

    @property
    def provider(self) -> str:
        return "gemini"

    @property
    def model(self) -> str:
        return self._model

    def _ensure_auth(self) -> None:
        if self._use_vertex and not self._auth_token:
            try:
                import google.auth
                import google.auth.transport.requests

                auth_module: Any = google.auth
                raw_credentials, _ = auth_module.default(
                    scopes=["https://www.googleapis.com/auth/cloud-platform"],
                )
                credentials = cast(_RefreshableCredentials | None, raw_credentials)
                if credentials is not None:
                    if hasattr(credentials, "refresh"):
                        request = google.auth.transport.requests.Request()
                        credentials.refresh(request)
                    self._auth_token = credentials.token
            except Exception as e:
                logger.warning("Vertex AI auth failed: %s", e)
                self._auth_token = ""  # nosec B105 -- empty-string sentinel, not a credential

    def _build_url(self, action: str, *, alt_sse: bool = False) -> str:
        if self._use_vertex:
            url = (
                f"https://{self._vertex_location}-aiplatform.googleapis.com/v1/"
                f"projects/{self._vertex_project}/locations/{self._vertex_location}/"
                f"publishers/google/models/{self._model}:{action}"
            )
            if alt_sse:
                url += "?alt=sse"
            return url
        return f"https://generativelanguage.googleapis.com/v1beta/models/{self._model}:{action}"

    async def generate(
        self,
        prompt: str,
        *,
        system_prompt: str | None = None,
        tools: list[ToolDefinition] | None = None,
        max_tokens: int = 1024,
        temperature: float = 0.2,
    ) -> LLMResult:
        self._ensure_auth()
        payload = _build_payload(
            prompt=prompt,
            system_prompt=system_prompt,
            tools=tools,
            max_tokens=max_tokens,
            temperature=temperature,
        )
        url = self._build_url("generateContent")
        params, headers = self._build_request_params()
        try:
            response = await self._http.post(
                url,
                params=params or None,
                headers=headers or None,
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
        return _parse_response(data, provider=self.provider)

    def generate_stream(
        self,
        prompt: str,
        *,
        system_prompt: str | None = None,
        tools: list[ToolDefinition] | None = None,
        max_tokens: int = 1024,
        temperature: float = 0.2,
    ) -> AsyncIterator[LLMChunk]:
        self._ensure_auth()

        async def _stream() -> AsyncIterator[LLMChunk]:
            payload = _build_payload(
                prompt=prompt,
                system_prompt=system_prompt,
                tools=tools,
                max_tokens=max_tokens,
                temperature=temperature,
            )
            url = self._build_url("streamGenerateContent", alt_sse=True)
            params, headers = self._build_request_params()
            try:
                async with self._http.stream(
                    "POST",
                    url,
                    params=params or None,
                    headers=headers or None,
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

    def _build_request_params(self) -> tuple[dict[str, str], dict[str, str]]:
        if self._use_vertex:
            headers: dict[str, str] = {}
            if self._auth_token:
                headers["Authorization"] = f"Bearer {self._auth_token}"
            return {}, headers
        return {"key": self._api_key}, {}


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


def _is_gemini_response(raw: object) -> TypeGuard[_GeminiResponse]:  # noqa: PLR0911
    """Narrow raw JSON to the Gemini generateContent response schema.

    This is the audited parse boundary for this provider: the backbone of
    the payload (``candidates`` -> ``content`` -> ``parts``) is validated
    here, while individual parts are still inspected leniently by the
    extractors so the adapter keeps tolerating malformed provider payloads.
    """
    if not is_json_object(raw):
        return False
    candidates = raw.get("candidates")
    if candidates is None:
        return True
    if not isinstance(candidates, list):
        return False
    candidate_list = cast(list[object], candidates)
    if not candidate_list:
        return False
    first = candidate_list[0]
    if not is_json_object(first):
        return False
    content = first.get("content")
    if content is None:
        return True
    if not is_json_object(content):
        return False
    parts = content.get("parts")
    if parts is None:
        return True
    if not isinstance(parts, list):
        return False
    return bool(cast(list[object], parts))


def _parse_response(raw: object, *, provider: str) -> LLMResult:
    if not _is_gemini_response(raw):
        raise LLMError(
            "gemini response missing text",
            details={"provider": provider},
        )
    text = _extract_text_optional(raw)
    tool_calls = _extract_tool_calls(raw)
    if not text and not tool_calls:
        raise LLMError(
            "gemini response missing text",
            details={"provider": provider},
        )
    return LLMResult(text=text or "", tool_calls=tool_calls)


def _parts_of(data: _GeminiResponse) -> list[_GeminiPart] | None:
    """Return the parts of the first candidate, or None if there are none."""
    candidates = data.get("candidates")
    if not candidates:
        return None
    first = candidates[0]
    content = first.get("content")
    if content is None:
        return None
    return content.get("parts")


def _extract_text_optional(data: _GeminiResponse) -> str | None:
    parts = _parts_of(data)
    if not parts:
        return None
    for part in parts:
        if not is_json_object(part):
            continue
        part_payload: dict[str, object] = part
        text = part_payload.get("text")
        if isinstance(text, str):
            return text
    return None


def _extract_tool_calls(data: _GeminiResponse) -> list[ToolCall] | None:
    parts = _parts_of(data)
    if not parts:
        return None
    result: list[ToolCall] = []
    for part in parts:
        if not is_json_object(part):
            continue
        part_payload: dict[str, object] = part
        function_call = part_payload.get("functionCall")
        if not is_json_object(function_call):
            continue
        function_payload: dict[str, object] = function_call
        name = function_payload.get("name")
        if not isinstance(name, str) or not name:
            continue
        raw_args = function_payload.get("args")
        arguments = raw_args if is_json_object(raw_args) else {}
        result.append(ToolCall(name=name, arguments=arguments))
    return result or None


def _parse_stream_chunk(raw: object) -> LLMChunk | None:
    if not _is_gemini_response(raw):
        return None
    text = _extract_text_optional(raw)
    if text:
        return LLMChunk(text=text)
    return None


__all__ = ["GeminiLLMAdapter"]
