"""Shared mappers from domain objects to ai.inference/ai.conversation protos."""

from __future__ import annotations

from ai.inference.v1 import service_pb2

from core.domain.value_objects import SearchHit, TokenUsage

_EXCERPT_MAX_LENGTH = 300


def map_usage(token_usage: TokenUsage) -> service_pb2.Usage:
    return service_pb2.Usage(
        prompt_tokens=token_usage.prompt_tokens,
        completion_tokens=token_usage.completion_tokens,
        total_tokens=token_usage.total_tokens,
    )


def map_citation_from_hit(hit: SearchHit) -> service_pb2.Citation:
    return service_pb2.Citation(
        document_id=str(hit.document_id),
        chunk_id=str(hit.chunk_id),
        source_type="chunk",
        score=hit.score,
        title=hit.document_title or f"Chunk {hit.chunk_index}",
        excerpt=hit.chunk_text[:_EXCERPT_MAX_LENGTH],
    )
