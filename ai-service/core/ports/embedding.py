from __future__ import annotations

from abc import ABC, abstractmethod


class EmbeddingPort(ABC):
    @property
    @abstractmethod
    def dimensions(self) -> int: ...

    @property
    @abstractmethod
    def provider(self) -> str: ...

    @abstractmethod
    async def embed_documents(
        self,
        texts: list[str],
        *,
        input_type: str | None = None,
    ) -> list[list[float]]: ...

    @abstractmethod
    async def embed_query(self, query: str) -> list[float]: ...


__all__ = ["EmbeddingPort"]
