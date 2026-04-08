from __future__ import annotations

import uuid

import pytest

from core.domain.enums import Language, Tier
from core.domain.exceptions import ConfigurationError, RepositoryError
from infrastructure.grpc.core_service import CoreServiceGrpcAdapter, CoreUserResponse


class _Status:
    def __init__(self, name: str) -> None:
        self.name = name


class _RpcError(Exception):
    def __init__(self, status_name: str) -> None:
        self._status_name = status_name

    def code(self) -> _Status:
        return _Status(self._status_name)


class _ClientStub:
    def __init__(
        self,
        *,
        response: CoreUserResponse | None = None,
        error: Exception | None = None,
    ) -> None:
        self._response = response
        self._error = error

    async def get_user(self, user_id: uuid.UUID) -> CoreUserResponse | None:
        _ = user_id
        if self._error is not None:
            raise self._error
        return self._response


def test_adapter_requires_non_empty_endpoint() -> None:
    with pytest.raises(ConfigurationError, match="endpoint is required"):
        CoreServiceGrpcAdapter(endpoint="   ", client=_ClientStub())


@pytest.mark.asyncio
async def test_get_user_tier_returns_tier_from_core_response() -> None:
    user_id = uuid.uuid4()
    client = _ClientStub(
        response=CoreUserResponse(
            user_id=user_id,
            tier="pro",
            preferred_language="am",
        )
    )
    adapter = CoreServiceGrpcAdapter(endpoint="localhost:50052", client=client)

    result = await adapter.get_user_tier(user_id)

    assert result is Tier.PRO


@pytest.mark.asyncio
async def test_get_user_profile_returns_profile_with_language() -> None:
    user_id = uuid.uuid4()
    client = _ClientStub(
        response=CoreUserResponse(
            user_id=user_id,
            tier="basic",
            preferred_language="en",
        )
    )
    adapter = CoreServiceGrpcAdapter(endpoint="localhost:50052", client=client)

    profile = await adapter.get_user_profile(user_id)

    assert profile is not None
    assert profile.user_id == user_id
    assert profile.tier is Tier.BASIC
    assert profile.preferred_language is Language.ENGLISH


@pytest.mark.asyncio
async def test_get_user_profile_allows_missing_preferred_language() -> None:
    user_id = uuid.uuid4()
    client = _ClientStub(
        response=CoreUserResponse(
            user_id=user_id,
            tier="premium",
            preferred_language=None,
        )
    )
    adapter = CoreServiceGrpcAdapter(endpoint="localhost:50052", client=client)

    profile = await adapter.get_user_profile(user_id)

    assert profile is not None
    assert profile.preferred_language is None


@pytest.mark.asyncio
async def test_get_user_profile_returns_none_for_not_found() -> None:
    user_id = uuid.uuid4()
    client = _ClientStub(error=_RpcError("NOT_FOUND"))
    adapter = CoreServiceGrpcAdapter(endpoint="localhost:50052", client=client)

    result = await adapter.get_user_profile(user_id)

    assert result is None


@pytest.mark.asyncio
async def test_get_user_tier_returns_none_for_not_found() -> None:
    user_id = uuid.uuid4()
    client = _ClientStub(error=_RpcError("NOT_FOUND"))
    adapter = CoreServiceGrpcAdapter(endpoint="localhost:50052", client=client)

    result = await adapter.get_user_tier(user_id)

    assert result is None


@pytest.mark.asyncio
async def test_get_user_tier_raises_on_unknown_tier() -> None:
    user_id = uuid.uuid4()
    client = _ClientStub(response=CoreUserResponse(user_id=user_id, tier="starter"))
    adapter = CoreServiceGrpcAdapter(endpoint="localhost:50052", client=client)

    with pytest.raises(RepositoryError, match="unknown tier"):
        await adapter.get_user_tier(user_id)


@pytest.mark.asyncio
async def test_get_user_profile_raises_on_unknown_language() -> None:
    user_id = uuid.uuid4()
    client = _ClientStub(
        response=CoreUserResponse(
            user_id=user_id,
            tier="pro",
            preferred_language="fr",
        )
    )
    adapter = CoreServiceGrpcAdapter(endpoint="localhost:50052", client=client)

    with pytest.raises(RepositoryError, match="unknown language"):
        await adapter.get_user_profile(user_id)


@pytest.mark.asyncio
async def test_get_user_profile_wraps_non_not_found_grpc_errors() -> None:
    user_id = uuid.uuid4()
    client = _ClientStub(error=_RpcError("UNAVAILABLE"))
    adapter = CoreServiceGrpcAdapter(endpoint="localhost:50052", client=client)

    with pytest.raises(RepositoryError, match="failed to fetch user from core service"):
        await adapter.get_user_profile(user_id)
