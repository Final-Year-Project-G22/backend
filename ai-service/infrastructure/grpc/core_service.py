from __future__ import annotations

import uuid
from dataclasses import dataclass
from typing import Protocol

from core.domain.enums import Language, Tier
from core.domain.exceptions import ConfigurationError, RepositoryError
from core.ports.core_service import CoreServicePort, CoreUserProfile

GRPC_NOT_FOUND_CODE = 5


@dataclass(frozen=True)
class CoreUserResponse:
    user_id: uuid.UUID
    tier: str
    preferred_language: str | None = None


class CoreServiceClient(Protocol):
    async def get_user(self, user_id: uuid.UUID) -> CoreUserResponse | None: ...


class CoreServiceGrpcAdapter(CoreServicePort):
    def __init__(
        self,
        *,
        endpoint: str,
        client: CoreServiceClient,
    ) -> None:
        if not endpoint.strip():
            raise ConfigurationError("core grpc endpoint is required")
        self._endpoint = endpoint
        self._client = client

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


__all__ = ["CoreServiceClient", "CoreServiceGrpcAdapter", "CoreUserResponse"]
