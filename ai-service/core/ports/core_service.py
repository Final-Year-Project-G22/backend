from __future__ import annotations

import uuid
from abc import ABC, abstractmethod
from datetime import datetime

from pydantic import BaseModel, ConfigDict

from core.domain.enums import Language, Tier


class CoreUserProfile(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True)

    user_id: uuid.UUID
    tier: Tier
    preferred_language: Language | None = None


class SignedUrlResult(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True)

    signed_url: str
    expires_at: datetime
    content_type: str | None = None
    content_length: int | None = None


class CoreServicePort(ABC):
    @abstractmethod
    async def get_user_tier(self, user_id: uuid.UUID) -> Tier | None: ...

    @abstractmethod
    async def get_user_profile(self, user_id: uuid.UUID) -> CoreUserProfile | None: ...

    @abstractmethod
    async def get_signed_url(
        self,
        document_id: uuid.UUID,
        account_id: uuid.UUID,
        *,
        expires_in_seconds: int = 3600,
    ) -> SignedUrlResult: ...


__all__ = ["CoreServicePort", "CoreUserProfile", "SignedUrlResult"]
