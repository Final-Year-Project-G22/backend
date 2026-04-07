from __future__ import annotations

import uuid
from abc import ABC, abstractmethod

from pydantic import BaseModel, ConfigDict

from core.domain.enums import Language, Tier


class CoreUserProfile(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True)

    user_id: uuid.UUID
    tier: Tier
    preferred_language: Language | None = None


class CoreServicePort(ABC):
    @abstractmethod
    async def get_user_tier(self, user_id: uuid.UUID) -> Tier | None: ...

    @abstractmethod
    async def get_user_profile(self, user_id: uuid.UUID) -> CoreUserProfile | None: ...


__all__ = ["CoreServicePort", "CoreUserProfile"]
