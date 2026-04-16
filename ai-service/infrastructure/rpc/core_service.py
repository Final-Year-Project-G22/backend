from __future__ import annotations

import importlib
import uuid
from dataclasses import dataclass
from datetime import UTC, datetime
from functools import lru_cache
from types import ModuleType, SimpleNamespace
from typing import Any, Protocol

import grpc
from core.domain.enums import Language, Tier
from core.domain.exceptions import ConfigurationError, RepositoryError
from core.ports.core_service import CoreServicePort, CoreUserProfile, SignedUrlResult

GRPC_NOT_FOUND_CODE = 5


@dataclass(frozen=True)
class CoreUserResponse:
    user_id: uuid.UUID
    account_id: uuid.UUID
    tier: str
    preferred_language: str | None = None


@dataclass(frozen=True)
class SignedUrlResponse:
    signed_url: str
    expires_at: int
    content_type: str | None = None
    content_length: int | None = None


class CoreServiceClient(Protocol):
    async def get_user(self, user_id: uuid.UUID) -> CoreUserResponse | None: ...


class DocumentFetchClient(Protocol):
    async def get_signed_url(
        self,
        document_id: uuid.UUID,
        account_id: uuid.UUID,
        expires_in_seconds: int,
    ) -> SignedUrlResponse: ...


class CoreServiceGrpcAdapter(CoreServicePort):
    def __init__(
        self,
        *,
        endpoint: str,
        client: CoreServiceClient,
        document_fetch_channel: Any | None = None,
    ) -> None:
        if not endpoint.strip():
            raise ConfigurationError("core grpc endpoint is required")
        self._endpoint = endpoint
        self._client = client
        object.__setattr__(self, "_document_fetch_channel", document_fetch_channel)

    async def get_user_tier(self, user_id: uuid.UUID) -> Tier | None:
        response = await self._fetch_user(user_id)
        if response is None:
            return None
        return _parse_tier(response.tier, user_id=user_id)

    async def get_user_profile(self, user_id: uuid.UUID) -> CoreUserProfile | None:
        response = await self._fetch_user(user_id)
        if response is None:
            return None

        tier = _parse_tier(response.tier, user_id=user_id)
        preferred_language = _parse_language(response.preferred_language, user_id=user_id)

        return CoreUserProfile(
            user_id=response.user_id,
            tier=tier,
            preferred_language=preferred_language,
        )

    async def get_signed_url(
        self,
        document_id: uuid.UUID,
        account_id: uuid.UUID,
        *,
        expires_in_seconds: int = 3600,
    ) -> SignedUrlResult:
        channel = getattr(self, "_document_fetch_channel", None)
        if channel is None:
            raise RepositoryError(
                "document fetch channel not configured",
                details={"endpoint": self._endpoint},
            )

        document_fetch_client = _build_document_fetch_client(channel)
        if document_fetch_client is None:
            raise RepositoryError(
                "document fetch stubs are not available",
                details={"endpoint": self._endpoint},
            )

        try:
            request = _build_get_signed_url_request(
                document_id,
                account_id,
                expires_in_seconds,
            )
            response = await document_fetch_client.get_signed_url(request)
        except Exception as exc:
            raise RepositoryError(
                "failed to fetch signed url from core service",
                details={
                    "document_id": str(document_id),
                    "account_id": str(account_id),
                    "endpoint": self._endpoint,
                },
            ) from exc

        expires_at = datetime.fromtimestamp(
            response.expires_at,
            tz=UTC,
        )

        return SignedUrlResult(
            signed_url=response.signed_url,
            expires_at=expires_at,
            content_type=response.content_type,
            content_length=response.content_length,
        )

    async def _fetch_user(self, user_id: uuid.UUID) -> CoreUserResponse | None:
        try:
            response = await self._client.get_user(user_id)
        except Exception as exc:
            if _is_not_found_error(exc):
                return None
            raise RepositoryError(
                "failed to fetch user from core service",
                details={"user_id": str(user_id), "endpoint": self._endpoint},
            ) from exc

        if response is None:
            return None
        return response


class CoreUserGrpcClient(CoreServiceClient):
    def __init__(self, *, endpoint: str) -> None:
        if not endpoint.strip():
            raise ConfigurationError("core grpc endpoint is required")

        self._endpoint = endpoint
        self._channel = _create_async_channel(endpoint)
        self._stub: Any | None = _build_core_user_stub(self._channel)

    async def get_user(self, user_id: uuid.UUID) -> CoreUserResponse | None:
        stub = self._stub
        if stub is None:
            stub = _build_core_user_stub(self._channel)
            if stub is None:
                raise RepositoryError(
                    "core grpc stubs are not available",
                    details={"endpoint": self._endpoint},
                )
            self._stub = stub

        request: Any = _build_get_user_request(str(user_id))
        response = await stub.GetUserProfile(request)

        parsed_user_id = _parse_uuid(str(getattr(response, "user_id", "")), field="user_id")
        parsed_account_id = _parse_uuid(
            str(getattr(response, "account_id", "")), field="account_id"
        )

        preferred_language_raw = getattr(response, "preferred_language", "")
        preferred_language = str(preferred_language_raw) or None

        return CoreUserResponse(
            user_id=parsed_user_id,
            account_id=parsed_account_id,
            tier=str(getattr(response, "tier", "")),
            preferred_language=preferred_language,
        )


def _parse_tier(raw: str, *, user_id: uuid.UUID) -> Tier:
    try:
        return Tier(raw)
    except ValueError as exc:
        raise RepositoryError(
            "core service returned unknown tier",
            details={"tier": raw, "user_id": str(user_id)},
        ) from exc


def _parse_language(raw: str | None, *, user_id: uuid.UUID) -> Language | None:
    if raw is None:
        return None

    try:
        return Language(raw)
    except ValueError as exc:
        raise RepositoryError(
            "core service returned unknown language",
            details={"language": raw, "user_id": str(user_id)},
        ) from exc


def _parse_uuid(raw: str, *, field: str) -> uuid.UUID:
    try:
        return uuid.UUID(raw)
    except ValueError as exc:
        raise RepositoryError(
            "core service returned invalid identifier",
            details={"field": field, "value": raw},
        ) from exc


def _is_not_found_error(exc: Exception) -> bool:
    code_attr = getattr(exc, "code", None)
    if not callable(code_attr):
        return False

    status = code_attr()
    if status is None:
        return False

    name = getattr(status, "name", None)
    if isinstance(name, str) and name == "NOT_FOUND":
        return True

    if status in {"NOT_FOUND", GRPC_NOT_FOUND_CODE}:
        return True

    return str(status).endswith("NOT_FOUND")


def _create_async_channel(endpoint: str) -> Any:
    grpc_aio: Any = getattr(grpc, "aio")  # noqa: B009
    insecure_channel: Any = getattr(grpc_aio, "insecure_channel")  # noqa: B009
    return insecure_channel(endpoint)


@lru_cache(maxsize=1)
def _load_user_proto_modules() -> tuple[ModuleType | None, ModuleType | None]:
    service_pb2_mod = _import_optional_module("core.user.v1.service_pb2")
    service_pb2_grpc_mod = _import_optional_module("core.user.v1.service_pb2_grpc")
    return service_pb2_mod, service_pb2_grpc_mod


def _import_optional_module(module_path: str) -> ModuleType | None:
    try:
        return importlib.import_module(module_path)
    except Exception:
        return None


def _build_core_user_stub(channel: Any) -> Any | None:
    _, service_pb2_grpc = _load_user_proto_modules()
    if service_pb2_grpc is None:
        return None

    stub_ctor: Any = getattr(service_pb2_grpc, "CoreUserServiceStub", None)
    if not callable(stub_ctor):
        return None

    return stub_ctor(channel)


def _build_get_user_request(user_id: str) -> Any:
    service_pb2, _ = _load_user_proto_modules()
    if service_pb2 is not None:
        request_ctor: Any = getattr(service_pb2, "GetUserProfileRequest", None)
        if callable(request_ctor):
            return request_ctor(user_id=user_id)

    return SimpleNamespace(user_id=user_id)


@lru_cache(maxsize=1)
def _load_document_proto_modules() -> tuple[ModuleType | None, ModuleType | None]:
    service_pb2_mod = _import_optional_module("core.document.v1.document_pb2")
    service_pb2_grpc_mod = _import_optional_module("core.document.v1.document_pb2_grpc")
    return service_pb2_mod, service_pb2_grpc_mod


def _build_document_fetch_client(channel: Any) -> Any | None:
    _, service_pb2_grpc = _load_document_proto_modules()
    if service_pb2_grpc is None:
        return None

    stub_ctor: Any = getattr(service_pb2_grpc, "DocumentFetchServiceStub", None)
    if not callable(stub_ctor):
        return None

    return stub_ctor(channel)


def _build_get_signed_url_request(
    document_id: uuid.UUID,
    account_id: uuid.UUID,
    expires_in_seconds: int,
) -> Any:
    service_pb2, _ = _load_document_proto_modules()
    if service_pb2 is not None:
        request_ctor: Any = getattr(service_pb2, "GetSignedUrlRequest", None)
        if callable(request_ctor):
            return request_ctor(
                document_id=str(document_id),
                account_id=str(account_id),
                expires_in_seconds=expires_in_seconds,
            )

    return SimpleNamespace(
        document_id=str(document_id),
        account_id=str(account_id),
        expires_in_seconds=expires_in_seconds,
    )


__all__ = [
    "CoreServiceClient",
    "CoreServiceGrpcAdapter",
    "CoreUserGrpcClient",
    "CoreUserResponse",
    "DocumentFetchClient",
    "SignedUrlResponse",
]
