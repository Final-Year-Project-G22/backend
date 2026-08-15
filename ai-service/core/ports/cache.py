from __future__ import annotations

from abc import ABC, abstractmethod


class CachePort(ABC):
    @abstractmethod
    async def get(self, key: str) -> str | None: ...

    @abstractmethod
    async def set(self, key: str, value: str, *, ttl_seconds: int = 3600) -> None: ...  # noqa: V107

    @abstractmethod
    async def delete(self, key: str) -> None: ...

    @abstractmethod
    async def delete_pattern(self, pattern: str) -> int: ...  # noqa: V107


__all__ = ["CachePort"]
