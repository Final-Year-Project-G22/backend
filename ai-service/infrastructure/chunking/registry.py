from __future__ import annotations

import logging
from typing import Any

from core.ports.chunking import Chunk, ChunkingPort, ChunkingStrategy
from core.ports.parser import ParsedDocument
from infrastructure.chunking.structural import StructuralChunkingAdapter

logger = logging.getLogger(__name__)


class ChunkingRegistry:
    def __init__(self) -> None:
        self._chunkers: list[ChunkingPort] = [
            StructuralChunkingAdapter(),
        ]

    def get_chunker(self, strategy_type: str) -> ChunkingPort | None:
        for chunker in self._chunkers:
            if chunker.supports(strategy_type):
                logger.debug("Found chunker for strategy_type: %s", strategy_type)
                return chunker
        logger.warning("No chunker found for strategy_type: %s", strategy_type)
        return None

    async def chunk(
        self,
        strategy_type: str,
        document: ParsedDocument,
        strategy: ChunkingStrategy,
        metadata: dict[str, Any] | None = None,
    ) -> list[Chunk]:
        chunker = self.get_chunker(strategy_type)
        if chunker is None:
            msg = f"No chunker found for strategy type: {strategy_type}"
            raise ValueError(msg)
        return await chunker.chunk(document, strategy, metadata)


__all__ = ["ChunkingRegistry"]
