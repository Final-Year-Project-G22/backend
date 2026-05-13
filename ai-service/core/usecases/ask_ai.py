from __future__ import annotations

import contextlib
import uuid
from datetime import UTC, datetime
from typing import Any

from core.domain.enums import DocumentSource, Language, MessageType
from core.domain.exceptions import AIServiceError
from core.domain.models import AIChatMessage, AIConversationSession
from core.domain.value_objects import ResponseSource, SearchFilters, SearchHit
from core.ports.cache import CachePort
from core.ports.conversation_repository import ConversationRepositoryPort
from core.ports.embedding import EmbeddingPort
from core.ports.event_bus import EventBusPort
from core.ports.knowledge_repository import KnowledgeRepositoryPort
from core.ports.llm import LLMPort
from core.usecases.contracts import AskAICommand, AskAIResult, CreateSessionCommand
from core.usecases.conversation import ConversationUseCase
from core.usecases.defaults import (
    DEFAULT_LLM_TEMPERATURE,
    DEFAULT_MAX_CONTEXT_HITS,
    DEFAULT_MAX_OUTPUT_TOKENS,
)
from core.usecases.quota_guard import QuotaGuardUseCase


def _utc_now() -> datetime:
    return datetime.now(UTC)


class AskAIUseCase:
    def __init__(
        self,
        conversation: ConversationUseCase,
        quota_guard: QuotaGuardUseCase,
        knowledge_repository: KnowledgeRepositoryPort,
        embedding_port: EmbeddingPort,
        llm_port: LLMPort,
        *,
        cache: CachePort | None = None,
        event_bus: EventBusPort | None = None,
    ) -> None:
        self._conversation = conversation
        self._quota_guard = quota_guard
        self._knowledge_repository = knowledge_repository
        self._embedding_port = embedding_port
        self._llm_port = llm_port
        self._cache = cache
        self._event_bus = event_bus

    async def execute(self, command: AskAICommand) -> AskAIResult:
        now = _utc_now()
        conversation = await self._resolve_conversation(command, now)
        user_message = await self._persist_user_message(command, conversation, now)

        cache_key = self._build_cache_key(command, conversation.id)
        cached_response = await self._try_cache(cache_key)
        if cached_response is not None:
            ai_message = await self._persist_ai_message(
                command,
                conversation,
                cached_response,
                now,
                cache_hit=True,
            )
            return self._build_result(conversation, user_message, ai_message, cache_hit=True)

        query_embedding = await self._embedding_port.embed_query(command.prompt)
        user_message = await self._update_user_message_embedding(user_message, query_embedding, now)

        vector_hits, bm25_hits = await self._retrieve_context(command, query_embedding)
        merged_hits = self._merge_and_dedupe_hits(vector_hits, bm25_hits)

        prompt = self._build_prompt(command.prompt, merged_hits, command.language)
        llm_response = await self._llm_port.generate(
            prompt,
            max_tokens=DEFAULT_MAX_OUTPUT_TOKENS,
            temperature=DEFAULT_LLM_TEMPERATURE,
        )

        ai_message = await self._persist_ai_message(
            command,
            conversation,
            llm_response,
            now,
            retrieved_hits=merged_hits,
        )

        await self._cache_response(cache_key, llm_response)
        await self._publish_query_event(command, conversation, ai_message)

        return self._build_result(conversation, user_message, ai_message, merged_hits)

    async def _resolve_conversation(
        self,
        command: AskAICommand,
        now: datetime,
    ) -> AIConversationSession:
        if command.conversation_id is not None:
            existing = await self._conversation.get_session(command.conversation_id)
            if existing is not None:
                return existing

        create_command = CreateSessionCommand(
            user_id=command.user_id,
            title=(
                command.title.strip()
                if command.title and command.title.strip()
                else command.prompt[:80]
            ),
            language=command.language,
        )
        return await self._conversation.create_session(create_command, at=now)

    async def _persist_user_message(
        self,
        command: AskAICommand,
        conversation: AIConversationSession,
        now: datetime,
    ) -> AIChatMessage:
        message_order = await self._next_message_order(conversation.id)
        user_message = AIChatMessage(
            user_id=command.user_id,
            conversation_id=conversation.id,
            message_type=MessageType.USER_QUERY,
            user_query=command.prompt,
            query_language=command.language,
            message_order=message_order,
            created_at=now,
            updated_at=now,
        )
        return await self._conversation_repository.add_message(user_message)

    @property
    def _conversation_repository(self) -> ConversationRepositoryPort:
        return self._conversation.conversation_repository

    async def _update_user_message_embedding(
        self,
        message: AIChatMessage,
        embedding: list[float],
        now: datetime,
    ) -> AIChatMessage:
        updated = message.model_copy(update={"query_embedding": embedding, "updated_at": now})
        return await self._conversation_repository.update_message(updated)

    async def _retrieve_context(
        self,
        command: AskAICommand,
        query_embedding: list[float],
    ) -> tuple[list[SearchHit], list[SearchHit]]:
        return await self._gather_hits(command, query_embedding)

    async def _gather_hits(
        self,
        command: AskAICommand,
        query_embedding: list[float],
    ) -> tuple[list[SearchHit], list[SearchHit]]:
        filters = SearchFilters(
            language=command.language,
            only_active=True,
        )

        vector_hits = await self._knowledge_repository.search_vector(
            query_embedding,
            top_k=command.vector_top_k,
            filters=filters,
        )
        bm25_hits = await self._knowledge_repository.search_bm25(
            command.prompt,
            top_k=command.bm25_top_k,
            filters=filters,
        )

        return vector_hits, bm25_hits

    def _merge_and_dedupe_hits(
        self,
        vector_hits: list[SearchHit],
        bm25_hits: list[SearchHit],
    ) -> list[SearchHit]:
        seen: set[uuid.UUID] = set()
        merged: list[SearchHit] = []

        for hit in vector_hits:
            if hit.chunk_id not in seen:
                seen.add(hit.chunk_id)
                merged.append(hit)

        for hit in bm25_hits:
            if hit.chunk_id not in seen:
                seen.add(hit.chunk_id)
                merged.append(hit)

        return merged[:DEFAULT_MAX_CONTEXT_HITS]

    def _build_prompt(
        self, user_query: str, context_hits: list[SearchHit], language: Language = Language.ENGLISH
    ) -> str:
        context_parts: list[str] = []
        for i, hit in enumerate(context_hits, 1):
            context_parts.append(f"[{i}] {hit.chunk_text}")
        context_block = "\n\n".join(context_parts) if context_parts else "No context available."

        lang_instruction = (
            f"\n\nAnswer in {language.value.upper()}." if language != Language.ENGLISH else ""
        )
        return (
            f"Context:\n{context_block}\n\n"
            f"Question: {user_query}\n\n"
            f"Answer based on the context above. If the context doesn't contain "
            f"relevant information, say so clearly.{lang_instruction}"
        )

    async def _persist_ai_message(
        self,
        command: AskAICommand,
        conversation: AIConversationSession,
        llm_response: str,
        now: datetime,
        *,
        retrieved_hits: list[SearchHit] | None = None,
        cache_hit: bool = False,
    ) -> AIChatMessage:
        message_order = await self._next_message_order(conversation.id)
        hits = retrieved_hits or []
        ai_message = AIChatMessage(
            user_id=command.user_id,
            conversation_id=conversation.id,
            message_type=MessageType.AI_RESPONSE,
            llm_response=llm_response,
            query_language=command.language,
            retrieved_chunk_ids=[h.chunk_id for h in hits],
            context_chunks=hits,
            response_sources=[
                ResponseSource(
                    source=DocumentSource(h.source.value),
                    document_id=h.document_id,
                    chunk_id=h.chunk_id,
                    title=getattr(h, "document_title", None) or f"Chunk {h.chunk_index}",
                    excerpt=h.chunk_text[:300],
                    score=h.score,
                )
                for h in hits
            ],
            message_order=message_order,
            cache_hit=cache_hit,
            created_at=now,
            updated_at=now,
        )
        return await self._conversation_repository.add_message(ai_message)

    async def _next_message_order(self, conversation_id: uuid.UUID) -> int:
        messages = await self._conversation.list_messages(conversation_id, limit=1, offset=0)
        if not messages:
            return 1
        return messages[0].message_order + 1

    def _build_cache_key(self, command: AskAICommand, conversation_id: uuid.UUID) -> str:
        return f"ai:cache:{conversation_id}:{command.prompt[:100]}"

    async def _try_cache(self, cache_key: str) -> str | None:
        if self._cache is None:
            return None
        try:
            return await self._cache.get(cache_key)
        except AIServiceError:
            return None

    async def _cache_response(self, cache_key: str, response: str) -> None:
        if self._cache is None:
            return
        with contextlib.suppress(AIServiceError):
            await self._cache.set(cache_key, response, ttl_seconds=3600)

    async def _publish_query_event(
        self,
        command: AskAICommand,
        conversation: AIConversationSession,
        ai_message: AIChatMessage,
    ) -> None:
        if self._event_bus is None:
            return
        try:
            payload: dict[str, Any] = {
                "user_id": str(command.user_id),
                "conversation_id": str(conversation.id),
                "message_id": str(ai_message.id),
                "language": command.language.value,
                "cache_hit": ai_message.cache_hit,
            }
            await self._event_bus.publish("ai.query.completed", payload)
        except AIServiceError:
            pass

    def _build_result(
        self,
        conversation: AIConversationSession,
        user_message: AIChatMessage,
        ai_message: AIChatMessage,
        retrieved_hits: list[SearchHit] | None = None,
        *,
        cache_hit: bool = False,
    ) -> AskAIResult:
        return AskAIResult(
            conversation=conversation,
            user_message=user_message,
            ai_message=ai_message,
            retrieved_hits=retrieved_hits or [],
            cache_hit=cache_hit,
        )


__all__ = ["AskAIUseCase"]
