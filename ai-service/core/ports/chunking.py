from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Any

from pydantic import BaseModel, Field

from core.ports.parser import ParsedDocument


class ChunkProvenance(BaseModel):
    section_heading: str | None = None
    section_order: int
    chunk_index: int
    parser_version: str | None = None


class Chunk(BaseModel):
    chunk_text: str
    token_count: int
    provenance: ChunkProvenance
    metadata: dict[str, Any] = Field(default_factory=dict)


class ChunkingStrategy(BaseModel):
    max_tokens: int = Field(default=512, ge=50, le=4096)
    overlap_tokens: int = Field(default=50, ge=0, le=512)
    overlap_by_section: bool = True
    min_chunk_tokens: int = Field(default=20, ge=1)


class ChunkingPort(ABC):
    @abstractmethod
    def supports(self, strategy_type: str) -> bool:
        raise NotImplementedError

    @abstractmethod
    async def chunk(
        self,
        document: ParsedDocument,
        strategy: ChunkingStrategy,
        metadata: dict[str, Any] | None = None,
    ) -> list[Chunk]:
        raise NotImplementedError


__all__ = ["Chunk", "ChunkProvenance", "ChunkingPort", "ChunkingStrategy"]
