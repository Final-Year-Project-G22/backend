from __future__ import annotations

import time
from typing import Any, ClassVar

from core.domain.tools import ToolResult
from core.domain.value_objects import SearchFilters
from core.ports.embedding import EmbeddingPort
from core.ports.knowledge_repository import KnowledgeRepositoryPort

_MAX_EXCERPT_LENGTH = 500


class SearchKnowledgeBaseTool:
    name = "search_knowledge_base"

    description: str = (
        "Search the knowledge base for Ethiopian business regulations, "
        "tax codes, licenses, compliance requirements, and business formalization guides."
    )

    parameter_schema: ClassVar[dict[str, Any]] = {
        "type": "object",
        "properties": {
            "query": {
                "type": "string",
                "description": "The search query (English or Amharic).",
            },
            "top_k": {
                "type": "integer",
                "description": "Maximum number of results to return.",
                "default": 5,
            },
        },
        "required": ["query"],
    }

    def __init__(
        self,
        knowledge_repository: KnowledgeRepositoryPort,
        embedding_port: EmbeddingPort,
    ) -> None:
        self._knowledge_repository = knowledge_repository
        self._embedding_port = embedding_port

    async def execute(
        self,
        arguments: dict[str, Any],
        account_id: str = "",
        user_id: str = "",
    ) -> ToolResult:
        query = arguments.get("query", "")
        top_k = arguments.get("top_k", 5)

        if not query.strip():
            return ToolResult(
                tool_name=self.name,
                arguments=arguments,
                result_text="No search query provided.",
                success=False,
                error_message="Empty query",
            )

        start = time.perf_counter()
        try:
            query_embedding = await self._embedding_port.embed_query(query)
            vector_hits = await self._knowledge_repository.search_vector(
                query_embedding,
                top_k=top_k,
                filters=SearchFilters(),
            )
            bm25_hits = await self._knowledge_repository.search_bm25(
                query,
                top_k=top_k,
                filters=SearchFilters(),
            )
        except Exception as e:
            elapsed = int((time.perf_counter() - start) * 1000)
            return ToolResult(
                tool_name=self.name,
                arguments=arguments,
                result_text=f"Search failed: {e}",
                success=False,
                error_message=str(e),
                execution_ms=elapsed,
            )

        elapsed = int((time.perf_counter() - start) * 1000)

        seen: set[str] = set()
        merged: list[Any] = []
        for hit in list(vector_hits) + list(bm25_hits):
            key = str(hit.chunk_id)
            if key not in seen:
                seen.add(key)
                merged.append(hit)

        merged = merged[:top_k]

        if not merged:
            return ToolResult(
                tool_name=self.name,
                arguments=arguments,
                result_text=f"No results found for query: {query}",
                execution_ms=elapsed,
            )

        lines: list[str] = []
        for i, hit in enumerate(merged, 1):
            lines.append(f"[{i}] {hit.document_title} (score: {round(hit.score, 3)})")
            text = hit.chunk_text
            if len(text) > _MAX_EXCERPT_LENGTH:
                text = text[:_MAX_EXCERPT_LENGTH] + "..."
            lines.append(f"    {text}")
            lines.append("")

        result_text = "\n".join(lines)
        return ToolResult(
            tool_name=self.name,
            arguments=arguments,
            result_text=result_text,
            success=True,
            execution_ms=elapsed,
            hits=merged,
        )
