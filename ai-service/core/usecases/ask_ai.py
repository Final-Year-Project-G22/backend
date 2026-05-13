from __future__ import annotations

import contextlib
import logging
import re
import uuid
from collections.abc import AsyncGenerator
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
from core.ports.llm import LLMPort, ToolCall
from core.usecases.contracts import AskAICommand, AskAIResult, CreateSessionCommand
from core.usecases.conversation import ConversationUseCase
from core.usecases.defaults import (
    DEFAULT_LLM_TEMPERATURE,
    DEFAULT_MAX_CONTEXT_HITS,
    DEFAULT_MAX_OUTPUT_TOKENS,
)
from core.usecases.quota_guard import QuotaGuardUseCase
from infrastructure.rpc.ai_tool_client import AIToolGrpcClient

logger = logging.getLogger(__name__)


def _utc_now() -> datetime:
    return datetime.now(UTC)


HISTORY_LIMIT = 6


class AskAIUseCase:
    def __init__(
        self,
        conversation: ConversationUseCase,
        quota_guard: QuotaGuardUseCase,
        knowledge_repository: KnowledgeRepositoryPort,
        embedding_port: EmbeddingPort,
        llm_port: LLMPort,
        ai_tool_client: AIToolGrpcClient | None = None,
        persona_prompt: str | None = None,
        restrictions: str | None = None,
        *,
        cache: CachePort | None = None,
        event_bus: EventBusPort | None = None,
    ) -> None:
        self._conversation = conversation
        self._quota_guard = quota_guard
        self._knowledge_repository = knowledge_repository
        self._embedding_port = embedding_port
        self._llm_port = llm_port
        self._ai_tool_client = ai_tool_client
        self._persona_prompt = persona_prompt
        self._restrictions = restrictions
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

        system_prompt = self._build_system_prompt()
        history = await self._load_history(conversation.id)
        prompt = self._build_prompt(command.prompt, merged_hits, command.language, history)
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

    async def execute_stream_with_tools(
        self,
        command: AskAICommand,
    ) -> AsyncGenerator[AskStreamEvent, None]:
        now = _utc_now()
        conversation = await self._resolve_conversation(command, now)
        user_message = await self._persist_user_message(command, conversation, now)

        query_embedding = await self._embedding_port.embed_query(command.prompt)
        user_message = await self._update_user_message_embedding(user_message, query_embedding, now)

        vector_hits, bm25_hits = await self._retrieve_context(command, query_embedding)
        merged_hits = self._merge_and_dedupe_hits(vector_hits, bm25_hits)

        system_prompt = self._build_system_prompt()
        history = await self._load_history(conversation.id)

        auto_tool_text, search_attempted = await self._auto_search_guides(command)
        prompt = self._build_prompt(
            command.prompt,
            merged_hits,
            command.language,
            history,
            tool_results_text=auto_tool_text,
            search_attempted=search_attempted,
        )
        llm_port = self._llm_port

        full_response_parts: list[str] = []
        async for chunk in llm_port.generate_stream(
            prompt,
            system_prompt=system_prompt,
            tools=None,
            max_tokens=DEFAULT_MAX_OUTPUT_TOKENS,
            temperature=DEFAULT_LLM_TEMPERATURE,
        ):
            if chunk.text is not None:
                full_response_parts.append(chunk.text)
                yield AskStreamEvent(text=chunk.text)

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

        yield AskStreamEvent(done=conversation, ai_message=ai_message, merged_hits=merged_hits)

    def _build_system_prompt(self) -> str:
        parts: list[str] = []
        if self._persona_prompt:
            parts.append(self._persona_prompt)
        if self._restrictions:
            parts.append(self._restrictions)
        parts.append(
            "You have no access to external functions. Do not mention or invent any tools, search capabilities, or plugins."
        )
        return "\n\n".join(parts) if parts else ""

    _SEARCH_KEYWORDS = re.compile(
        r"\b(guide|guides|find|search|list|show|what.*exist|what.*available|"
        r"is. there|are. there|how.*(do|to)|tell me about)\b",
        re.IGNORECASE,
    )

    def _should_auto_search(self, query: str) -> bool:
        return bool(self._SEARCH_KEYWORDS.search(query))

    _STOP_WORDS = re.compile(
        r"\b(what|which|how|where|when|who|why|is|are|was|were|do|does|did|"
        r"can|could|will|would|shall|should|may|might|have|has|had|hasn|"
        r"the|a|an|on|in|at|to|for|of|by|with|about|show|tell|list|find|"
        r"search|give|get|exist|available|guide|guides|me|i|my|you|your|"
        r"please|there|some|any|all|every|each|both|not|no|none|let)\b",
        re.IGNORECASE,
    )

    async def _auto_search_guides(self, command: AskAICommand) -> tuple[str | None, bool]:
        """Returns (tool_results_text, search_was_attempted)."""
        if not self._should_auto_search(command.prompt):
            return None, False
        if self._ai_tool_client is None:
            return None, False
        try:
            keyword = self._STOP_WORDS.sub("", command.prompt).strip().replace("-", " ")
            if not keyword:
                keyword = command.prompt[:64].replace("-", " ")

            seen_kw: set[str] = set()
            all_guides: list[dict[str, Any]] = []
            guide_ids: set[str] = set()

            async def _try_search(kw: str) -> None:
                if not kw or kw in seen_kw:
                    return
                seen_kw.add(kw)
                tc = self._ai_tool_client
                if tc is None:
                    return
                result = await tc.execute_tool(
                    name="search_guides",
                    arguments={"keyword": kw},
                    account_id=str(command.account_id),
                    user_id=str(command.user_id),
                )
                if not isinstance(result, list):
                    return
                for item in result:
                    if not isinstance(item, dict):
                        continue
                    gid = str(item.get("id") or item.get("slug") or "")  # type: ignore[reportUnknownArgumentType,reportUnknownMemberType]
                    gid = gid.strip()
                    if gid and gid not in guide_ids:
                        guide_ids.add(gid)
                        all_guides.append(item)

            await _try_search(keyword)

            _min_word_len = 3
            if not all_guides:
                words = [
                    w
                    for w in sorted(keyword.split(), key=len, reverse=True)
                    if len(w) > _min_word_len
                ]
                for w in words:
                    await _try_search(w)
                    if all_guides:
                        break

            lines: list[str] = []
            for g in all_guides:
                name = g.get("name", g.get("slug", "Unknown"))
                desc = g.get("description") or ""
                lines.append(f"- {name}" + (f": {desc}" if desc else ""))
            if not lines:
                return None, True
            return "Available guides:\n" + "\n".join(lines), True
        except Exception:
            return None, True

    async def _load_history(self, conversation_id: uuid.UUID) -> list[AIChatMessage]:
        try:
            return await self._conversation.list_messages(
                conversation_id, limit=HISTORY_LIMIT, offset=0
            )
        except Exception:
            return []

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
        self,
        user_query: str,
        context_hits: list[SearchHit],
        language: Language = Language.ENGLISH,
        history: list[AIChatMessage] | None = None,
        tool_results_text: str | None = None,
        search_attempted: bool = False,  # noqa: FBT001,FBT002
    ) -> str:
        parts: list[str] = []

        if history:
            history_parts: list[str] = []
            for msg in history:
                if msg.message_type == MessageType.USER_QUERY and msg.user_query:
                    history_parts.append(f"User: {msg.user_query}")
                elif msg.message_type == MessageType.AI_RESPONSE and msg.llm_response:
                    history_parts.append(f"Assistant: {msg.llm_response}")
            if history_parts:
                parts.append("Conversation history:\n" + "\n".join(reversed(history_parts)) + "\n")

        has_context = bool(context_hits)
        has_tool_results = bool(tool_results_text)

        if has_context:
            context_parts: list[str] = []
            for i, hit in enumerate(context_hits, 1):
                context_parts.append(f"[{i}] {hit.chunk_text}")
            parts.append("From uploaded documents:\n" + "\n\n".join(context_parts))

        if has_tool_results:
            parts.append(f"From guide search:\n{tool_results_text}")

        instruction = "Answer the question"
        if has_tool_results:
            instruction += " using the guide search results"
        if has_context:
            instruction += (
                ", supplemented by the uploaded documents"
                if has_tool_results
                else " using the uploaded documents"
            )
        instruction += ". Do not mention or invent any tools, functions, or capabilities. If you don't know, say so."
        if not has_context and not has_tool_results and search_attempted:
            instruction += " No relevant guides were found."
        elif not has_context and not has_tool_results:
            instruction += " Use your general knowledge."

        lang_instruction = (
            f"\n\nAnswer in {language.value.upper()}." if language != Language.ENGLISH else ""
        )
        parts.append(f"\nQuestion: {user_query}\n\n{instruction}{lang_instruction}")
        return "\n\n".join(parts)

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


class AskStreamEvent:
    def __init__(
        self,
        text: str | None = None,
        tool_uses: list[ToolCall] | None = None,
        tool_result: ToolCall | None = None,
        tool_result_data: dict[str, Any] | None = None,
        done: AIConversationSession | None = None,
        ai_message: AIChatMessage | None = None,
        merged_hits: list[SearchHit] | None = None,
    ) -> None:
        self.text = text
        self.tool_uses = tool_uses
        self.tool_result = tool_result
        self.tool_result_data = tool_result_data
        self.done = done
        self.ai_message = ai_message
        self.merged_hits = merged_hits

    @property
    def is_text(self) -> bool:
        return self.text is not None

    @property
    def is_tool_use(self) -> bool:
        return self.tool_uses is not None

    @property
    def is_tool_result(self) -> bool:
        return self.tool_result is not None

    @property
    def is_done(self) -> bool:
        return self.done is not None


__all__ = ["AskAIUseCase", "AskStreamEvent"]
