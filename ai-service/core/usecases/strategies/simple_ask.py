from __future__ import annotations

import contextlib
import logging
from collections.abc import AsyncIterator
from datetime import UTC, datetime
from typing import Any

from core.domain.enums import DocumentSource, Language, MessageType
from core.domain.exceptions import AIServiceError
from core.domain.models import AIChatMessage, AIConversationSession
from core.domain.stream_events import AskStreamEvent, AskStreamEventType
from core.domain.value_objects import ResponseSource, SearchFilters, SearchHit
from core.ports.ask_strategy import AskStrategyPort
from core.ports.cache import CachePort
from core.ports.conversation_repository import ConversationRepositoryPort
from core.ports.embedding import EmbeddingPort
from core.ports.event_bus import EventBusPort
from core.ports.knowledge_repository import KnowledgeRepositoryPort
from core.ports.llm import LLMPort
from core.ports.tool_registry import ToolRegistryPort
from core.usecases.contracts import AskAICommand, AskAIResult, CreateSessionCommand
from core.usecases.conversation import ConversationUseCase
from core.usecases.defaults import (
    DEFAULT_LLM_TEMPERATURE,
    DEFAULT_MAX_CONTEXT_HITS,
    DEFAULT_MAX_OUTPUT_TOKENS,
)
from infrastructure.prompts import PromptLoader

logger = logging.getLogger(__name__)

HISTORY_LIMIT = 6


def _utc_now() -> datetime:
    return datetime.now(UTC)


class SimpleAskStrategy(AskStrategyPort):
    def __init__(
        self,
        conversation: ConversationUseCase,
        knowledge_repository: KnowledgeRepositoryPort,
        embedding_port: EmbeddingPort,
        llm_port: LLMPort,
        prompt_loader: PromptLoader,
        tool_registry: ToolRegistryPort | None = None,
        *,
        cache: CachePort | None = None,
        event_bus: EventBusPort | None = None,
    ) -> None:
        self._conversation = conversation
        self._knowledge_repository = knowledge_repository
        self._embedding_port = embedding_port
        self._llm_port = llm_port
        self._prompt_loader = prompt_loader
        self._tool_registry = tool_registry
        self._cache = cache
        self._event_bus = event_bus

    @property
    def llm_port(self) -> LLMPort:
        return self._llm_port

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

        system_prompt = self._prompt_loader.render_simple(
            locale=command.language.value,
            kb_context=self._format_kb_context(merged_hits),
        )
        history = await self._load_history(conversation.id, conversation.message_count)
        prompt = self._build_prompt(command.prompt, history, language=command.language)

        result = await self._llm_port.generate(
            prompt,
            system_prompt=system_prompt,
            tools=None,
            max_tokens=DEFAULT_MAX_OUTPUT_TOKENS,
            temperature=DEFAULT_LLM_TEMPERATURE,
        )
        llm_response = result.text.strip() if result.text else "I'll help you with that."

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

    async def execute_stream(self, command: AskAICommand) -> AsyncIterator[AskStreamEvent]:
        now = _utc_now()
        conversation = await self._resolve_conversation(command, now)
        user_message = await self._persist_user_message(command, conversation, now)

        query_embedding = await self._embedding_port.embed_query(command.prompt)
        user_message = await self._update_user_message_embedding(user_message, query_embedding, now)

        vector_hits, bm25_hits = await self._retrieve_context(command, query_embedding)
        merged_hits = self._merge_and_dedupe_hits(vector_hits, bm25_hits)

        system_prompt = self._prompt_loader.render_simple(
            locale=command.language.value,
            kb_context=self._format_kb_context(merged_hits),
        )
        history = await self._load_history(conversation.id, conversation.message_count)
        prompt = self._build_prompt(command.prompt, history, language=command.language)

        full_response_parts: list[str] = []
        async for chunk in self._llm_port.generate_stream(
            prompt,
            system_prompt=system_prompt,
            tools=None,
            max_tokens=DEFAULT_MAX_OUTPUT_TOKENS,
            temperature=DEFAULT_LLM_TEMPERATURE,
        ):
            if chunk.text is not None:
                full_response_parts.append(chunk.text)
                yield AskStreamEvent(type_=AskStreamEventType.TEXT, text=chunk.text)

        full_response = "".join(full_response_parts).strip() or "I'll help you with that."
        ai_message = await self._persist_ai_message(
            command,
            conversation,
            full_response,
            now,
            retrieved_hits=merged_hits,
        )

        cache_key = self._build_cache_key(command, conversation.id)
        await self._cache_response(cache_key, full_response)
        await self._publish_query_event(command, conversation, ai_message)

        yield AskStreamEvent(
            type_=AskStreamEventType.DONE,
            done=conversation,
            ai_message=ai_message,
            merged_hits=merged_hits,
        )

    def _format_kb_context(self, hits: list[SearchHit]) -> str:
        if not hits:
            return ""
        lines: list[str] = []
        for i, hit in enumerate(hits, 1):
            lines.append(f"[{i}] {hit.chunk_text}")
        return "From uploaded documents:\n" + "\n\n".join(lines)

    def _build_prompt(
        self,
        user_query: str,
        history: list[AIChatMessage] | None = None,
        language: Language = Language.ENGLISH,
    ) -> str:
        parts: list[str] = []

        if history:
            history_parts: list[str] = []
            for msg in history:
                if msg.message_type == MessageType.USER_QUERY and msg.user_query:
                    history_parts.append(f"User: {msg.user_query}")
                elif msg.message_type == MessageType.AI_RESPONSE and msg.llm_response:
                    tool_summary = ""
                    if msg.tool_calls:
                        summaries = [
                            f"  [{tc.tool_name}] {tc.result_summary}" for tc in msg.tool_calls
                        ]
                        tool_summary = "\n" + "\n".join(summaries)
                    history_parts.append(f"Assistant: {msg.llm_response}{tool_summary}")
            if history_parts:
                parts.append("Conversation history:\n" + "\n".join(reversed(history_parts)) + "\n")

        lang_instruction = (
            f"\n\nAnswer in {language.value.upper()}." if language != Language.ENGLISH else ""
        )
        parts.append(f"\nQuestion: {user_query}\n\n{lang_instruction}")
        return "\n\n".join(parts)

    async def _load_history(
        self,
        conversation_id: Any,
        message_count: int = 0,
    ) -> list[AIChatMessage]:
        try:
            offset = max(0, message_count - HISTORY_LIMIT)
            return await self._conversation.list_messages(
                conversation_id, limit=HISTORY_LIMIT, offset=offset
            )
        except Exception:
            return []

    async def _resolve_conversation(
        self, command: AskAICommand, now: datetime
    ) -> AIConversationSession:
        if command.conversation_id is not None:
            existing = await self._conversation.get_session(command.conversation_id)
            if existing is not None:
                return existing
        create_cmd = CreateSessionCommand(
            user_id=command.user_id,
            title=(
                command.title.strip()
                if command.title and command.title.strip()
                else command.prompt[:80]
            ),
            language=command.language,
        )
        return await self._conversation.create_session(create_cmd, at=now)

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
        filters = SearchFilters(language=command.language, only_active=True)
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
        seen: set[Any] = set()
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

    async def _next_message_order(self, conversation_id: Any) -> int:
        messages = await self._conversation.list_messages(conversation_id, limit=1, offset=0)
        if not messages:
            return 1
        return messages[0].message_order + 1

    def _build_cache_key(self, command: AskAICommand, conversation_id: Any) -> str:
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
