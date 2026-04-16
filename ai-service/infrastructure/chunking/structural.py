from __future__ import annotations

import logging
from typing import Any

import tiktoken

from core.ports.chunking import Chunk, ChunkingPort, ChunkingStrategy, ChunkProvenance
from core.ports.parser import ParsedDocument, ParsedDocumentSection

logger = logging.getLogger(__name__)

_DEFAULT_ENCODING = "cl100k_base"
_TOKEN_OVERHEAD = 4


class StructuralChunkingAdapter(ChunkingPort):
    STRATEGY_TYPE = "structural"

    def __init__(self) -> None:
        try:
            self._encoding = tiktoken.get_encoding(_DEFAULT_ENCODING)
        except Exception as exc:
            logger.warning(
                "Failed to load tiktoken encoding: %s, falling back to word-based estimation", exc
            )
            self._encoding = None

    def supports(self, strategy_type: str) -> bool:
        return strategy_type.lower() == self.STRATEGY_TYPE

    async def chunk(
        self,
        document: ParsedDocument,
        strategy: ChunkingStrategy,
        metadata: dict[str, Any] | None = None,
    ) -> list[Chunk]:
        if not document.sections:
            return []

        chunks: list[Chunk] = []
        global_chunk_index = 0

        for section in document.sections:
            section_chunks = self._chunk_section(
                section=section,
                strategy=strategy,
                start_chunk_index=global_chunk_index,
                parser_version=metadata.get("parser_version") if metadata else None,
            )
            chunks.extend(section_chunks)
            global_chunk_index += len(section_chunks)

        return chunks

    def _chunk_section(
        self,
        section: ParsedDocumentSection,
        strategy: ChunkingStrategy,
        start_chunk_index: int,
        parser_version: str | None,
    ) -> list[Chunk]:
        if not section.content.strip():
            return []

        tokens = self._get_tokens(section.content)
        section_token_count = len(tokens)

        if section_token_count <= strategy.max_tokens:
            return [
                Chunk(
                    chunk_text=section.content.strip(),
                    token_count=section_token_count,
                    provenance=ChunkProvenance(
                        section_heading=section.heading,
                        section_order=section.order,
                        chunk_index=start_chunk_index,
                        parser_version=parser_version,
                    ),
                )
            ]

        return self._split_large_section(
            content=section.content,
            tokens=tokens,
            heading=section.heading,
            section_order=section.order,
            strategy=strategy,
            start_chunk_index=start_chunk_index,
            parser_version=parser_version,
        )

    def _split_large_section(
        self,
        content: str,
        tokens: list[int],
        heading: str | None,
        section_order: int,
        strategy: ChunkingStrategy,
        start_chunk_index: int,
        parser_version: str | None,
    ) -> list[Chunk]:
        chunks: list[Chunk] = []
        max_tokens = strategy.max_tokens
        overlap_tokens = min(strategy.overlap_tokens, max_tokens // 4)
        min_chunk_tokens = strategy.min_chunk_tokens

        start = 0
        chunk_idx = start_chunk_index

        while start < len(tokens):
            end = min(start + max_tokens, len(tokens))
            chunk_tokens = tokens[start:end]

            chunk_text = (
                self._encoding.decode(chunk_tokens)
                if self._encoding
                else self._fallback_decode(content, start, end)
            )

            if chunk_text.strip():
                chunks.append(
                    Chunk(
                        chunk_text=chunk_text.strip(),
                        token_count=len(chunk_tokens),
                        provenance=ChunkProvenance(
                            section_heading=heading,
                            section_order=section_order,
                            chunk_index=chunk_idx,
                            parser_version=parser_version,
                        ),
                    )
                )
                chunk_idx += 1

            if end >= len(tokens):
                break

            start = end - overlap_tokens if overlap_tokens > 0 else end

            remaining = len(tokens) - start
            if remaining < min_chunk_tokens:
                break

        return chunks

    def _get_tokens(self, text: str) -> list[int]:
        if self._encoding:
            try:
                return self._encoding.encode(text)
            except Exception as exc:
                logger.warning("Failed to encode text: %s", exc)
        return self._fallback_tokenize(text)

    def _fallback_tokenize(self, text: str) -> list[int]:
        words = text.split()
        tokens_per_word = 1.25
        estimated_tokens = int(len(words) * tokens_per_word) + _TOKEN_OVERHEAD
        return list(range(estimated_tokens))

    def _fallback_decode(self, text: str, start: int, end: int) -> str:
        words = text.split()
        tokens_per_word = 1.25
        start_word = int(start / tokens_per_word)
        end_word = int(end / tokens_per_word)
        return " ".join(words[start_word:end_word])


__all__ = ["StructuralChunkingAdapter"]
