from __future__ import annotations

import contextlib
import logging
from collections.abc import AsyncIterator
from datetime import UTC, datetime
from typing import Any

from core.domain.enums import DocumentSource, MessageType
from core.domain.exceptions import AIServiceError
from core.domain.models import AIChatMessage, AIConversationSession, ToolCallRecord
from core.domain.stream_events import AskStreamEvent, AskStreamEventType
from core.domain.value_objects import ResponseSource, SearchFilters, SearchHit
from core.ports.ask_strategy import AskStrategyPort
from core.ports.cache import CachePort
from core.ports.conversation_repository import ConversationRepositoryPort
from core.ports.embedding import EmbeddingPort
from core.ports.event_bus import EventBusPort
from core.ports.intent_classifier import IntentClassifierPort
from core.ports.knowledge_repository import KnowledgeRepositoryPort
from core.ports.llm import LLMPort, ToolCall
from core.ports.tool_registry import ToolRegistryPort
from core.usecases.contracts import AskAICommand, AskAIResult, CreateSessionCommand
from core.usecases.conversation import ConversationUseCase
from core.usecases.defaults import (
    DEFAULT_LLM_TEMPERATURE,
    DEFAULT_MAX_CONTEXT_HITS,
    DEFAULT_MAX_OUTPUT_TOKENS,
)
from core.usecases.strategies.kb_query_guard import KBQueryGuard, SuppressionReason
from infrastructure.prefetch.pipeline import PreFetchPipeline
from infrastructure.prompts import PromptLoader

logger = logging.getLogger(__name__)

HISTORY_LIMIT = 6
MAX_AGENTIC_ITERATIONS = 5
SUPPRESSION_TRIPWIRE = 2
SEARCH_KNOWLEDGE_BASE_TOOL = "search_knowledge_base"
FINAL_ANSWER_FALLBACK = "I've retrieved what I could. Please refine your question."


def _utc_now() -> datetime:
    return datetime.now(UTC)


class AgenticAskStrategy(AskStrategyPort):
    def __init__(
        self,
        conversation: ConversationUseCase,
        knowledge_repository: KnowledgeRepositoryPort,
        embedding_port: EmbeddingPort,
        llm_port: LLMPort,
        prompt_loader: PromptLoader,
        tool_registry: ToolRegistryPort,
        intent_classifier: IntentClassifierPort,
        pre_fetch_pipeline: PreFetchPipeline,
        max_iterations: int = MAX_AGENTIC_ITERATIONS,
        *,
        kb_query_guard: KBQueryGuard | None = None,
        suppression_tripwire: int = SUPPRESSION_TRIPWIRE,
        cache: CachePort | None = None,
        event_bus: EventBusPort | None = None,
    ) -> None:
        self._conversation = conversation
        self._knowledge_repository = knowledge_repository
        self._embedding_port = embedding_port
        self._llm_port = llm_port
        self._prompt_loader = prompt_loader
        self._tool_registry = tool_registry
        self._intent_classifier = intent_classifier
        self._pre_fetch_pipeline = pre_fetch_pipeline
        self._max_iterations = max_iterations
        self._kb_query_guard = kb_query_guard or KBQueryGuard(embedding_port)
        self._suppression_tripwire = suppression_tripwire
        self._cache = cache
        self._event_bus = event_bus

    @property
    def llm_port(self) -> LLMPort:
        return self._llm_port

    async def execute(self, command: AskAICommand) -> AskAIResult:
        tool_calls_flat: list[ToolCallRecord] = []
        iteration = 0
        final_answer = ""
        merged_hits: list[SearchHit] = []
        done_event: AskStreamEvent | None = None

        async for event in self._execute_react_loop(command, tool_calls_flat, iteration):
            if event.is_text and event.text:
                final_answer += event.text
            if event.is_done:
                done_event = event
                merged_hits = event.merged_hits or []
                break

        if done_event is None or done_event.done is None:
            msg = "agentic strategy stream ended without a done event"
            raise AIServiceError(msg)

        conversation = done_event.done
        ai_message = AIChatMessage(
            user_id=command.user_id,
            conversation_id=conversation.id,
            message_type=MessageType.AI_RESPONSE,
            llm_response=final_answer.strip() or "I'll help you with that.",
            tool_calls=tool_calls_flat or None,
            agent_strategy="agentic",
            message_order=1,
        )

        return AskAIResult(
            conversation=conversation,
            user_message=None,
            ai_message=ai_message,
            retrieved_hits=merged_hits,
        )

    async def execute_stream(self, command: AskAICommand) -> AsyncIterator[AskStreamEvent]:
        tool_calls_flat: list[ToolCallRecord] = []
        iteration = 0

        async for event in self._execute_react_loop(command, tool_calls_flat, iteration):
            yield event

    async def _execute_react_loop(
        self,
        command: AskAICommand,
        tool_calls_flat: list[ToolCallRecord],
        iteration: int,
    ) -> AsyncIterator[AskStreamEvent]:
        now = _utc_now()
        conversation = await self._resolve_conversation(command, now)
        user_message = await self._persist_user_message(command, conversation, now)

        query_embedding = await self._embedding_port.embed_query(command.prompt)
        user_message = await self._update_user_message_embedding(user_message, query_embedding, now)

        vector_hits, bm25_hits = await self._retrieve_context(command, query_embedding)
        merged_hits = self._merge_and_dedupe_hits(vector_hits, bm25_hits)

        intent = await self._intent_classifier.classify(query_embedding)
        # The language-filtered hybrid search above is the single KB fetch for this turn.
        pre_fetch = await self._pre_fetch_pipeline.pre_fetch(
            intent,
            command.prompt,
            str(command.account_id),
            str(command.user_id),
            include_kb=False,
        )
        kb_context_available = bool(merged_hits or pre_fetch.get("kb"))
        self._kb_query_guard.start_turn(
            command.prompt,
            query_embedding,
            prompt_covered=kb_context_available,
        )

        tool_defs = await self._tool_registry.get_tool_definitions()
        history = await self._load_history(conversation.id, conversation.message_count)

        system_prompt = self._prompt_loader.render_agentic(
            locale=command.language.value,
            tools=[{"name": t.name, "description": t.description} for t in tool_defs],
            kb_context_available=kb_context_available,
        )

        messages = [{"role": "system", "content": system_prompt}]

        if history:
            history_text = self._format_history(history)
            messages.append({"role": "assistant", "content": history_text})

        user_context_parts: list[str] = []
        if merged_hits:
            user_context_parts.append(self._format_kb_context(merged_hits))
        if pre_fetch.get("kb"):
            user_context_parts.append(pre_fetch["kb"])
        if pre_fetch.get("profile"):
            user_context_parts.append(f"User profile: {pre_fetch['profile']}")
        if pre_fetch.get("progress"):
            user_context_parts.append(f"Guide progress: {pre_fetch['progress']}")
        if pre_fetch.get("compliance"):
            user_context_parts.append(f"Compliance status: {pre_fetch['compliance']}")

        user_content = command.prompt
        if user_context_parts:
            user_content = "Context:\n" + "\n\n".join(user_context_parts) + "\n\n" + user_content
        messages.append({"role": "user", "content": user_content})

        final_answer_parts: list[str] = []
        consecutive_suppressions = 0
        last_suppression_reason: SuppressionReason | None = None
        forced_finalization = False

        while iteration < self._max_iterations:
            iteration += 1

            current_prompt = messages[-1]["content"] if messages else ""

            result = await self._llm_port.generate(
                current_prompt,
                system_prompt=system_prompt,
                tools=tool_defs,
                max_tokens=DEFAULT_MAX_OUTPUT_TOKENS,
                temperature=DEFAULT_LLM_TEMPERATURE,
            )

            if result.tool_calls:
                tool_results: list[tuple[ToolCall, str]] = []
                for tc in result.tool_calls:
                    query = tc.arguments.get("query")
                    if (
                        tc.name == SEARCH_KNOWLEDGE_BASE_TOOL
                        and isinstance(query, str)
                        and query.strip()
                    ):
                        decision = await self._kb_query_guard.check(query)
                        if not decision.allowed:
                            consecutive_suppressions += 1
                            last_suppression_reason = decision.reason
                            suppression_reason = decision.reason.value if decision.reason else None
                            tool_calls_flat.append(
                                ToolCallRecord(
                                    tool_name=tc.name,
                                    arguments=tc.arguments,
                                    iteration=iteration,
                                    suppressed=True,
                                    suppression_reason=suppression_reason,
                                )
                            )
                            if command.debug_mode:
                                yield AskStreamEvent(
                                    type_=AskStreamEventType.TOOL_SUPPRESSED,
                                    tool_name=tc.name,
                                    suppression_reason=suppression_reason,
                                    matched_query=decision.matched_query,
                                )
                            if consecutive_suppressions >= self._suppression_tripwire:
                                forced_finalization = True
                                break
                            continue
                        self._kb_query_guard.register_executed(query, decision.embedding)

                    consecutive_suppressions = 0

                    yield AskStreamEvent(
                        type_=AskStreamEventType.TOOL_CALL,
                        tool_name=tc.name,
                        tool_arguments=tc.arguments,
                    )

                    if command.debug_mode:
                        yield AskStreamEvent(
                            type_=AskStreamEventType.THINKING,
                            text=f"Calling {tc.name} with {tc.arguments}",
                        )

                    tool_result = await self._tool_registry.execute_tool(
                        tc.name,
                        tc.arguments,
                        str(command.account_id),
                        str(command.user_id),
                    )

                    tool_calls_flat.append(
                        ToolCallRecord(
                            tool_name=tc.name,
                            arguments=tc.arguments,
                            result_summary=tool_result.result_text[:200],
                            success=tool_result.success,
                            error_message=tool_result.error_message,
                            execution_ms=tool_result.execution_ms,
                            iteration=iteration,
                        )
                    )

                    tool_results.append((tc, tool_result.result_text))

                    if (
                        tool_result.success
                        and tc.name == SEARCH_KNOWLEDGE_BASE_TOOL
                        and tool_result.hits
                    ):
                        merged_hits = self._merge_tool_hits(merged_hits, tool_result.hits)

                    yield AskStreamEvent(
                        type_=AskStreamEventType.TOOL_RESULT,
                        tool_name=tc.name,
                        tool_result_summary=tool_result.result_text[:200]
                        if tool_result.success
                        else f"Failed: {tool_result.error_message}",
                    )

                if forced_finalization:
                    break

                if not tool_results:
                    messages.append(
                        {
                            "role": "user",
                            "content": self._suppression_nudge(
                                command.prompt, last_suppression_reason
                            )
                            + "\n\n"
                            + user_content,
                        }
                    )
                    continue

                tool_call_text = "\n\n".join(
                    f"Tool: {tc.name}\n{tc.arguments.get('query', '') or tc.arguments.get('url', '')}"
                    for tc, _ in tool_results
                )
                tool_result_text = "\n\n".join(f"Result: {text[:500]}" for _, text in tool_results)
                messages.append(
                    {
                        "role": "assistant",
                        "content": f"I'll use the following tools:\n{tool_call_text}",
                    }
                )
                messages.append(
                    {
                        "role": "user",
                        "content": f"Here are the results:\n{tool_result_text}\n\nBased on this, provide a final answer to the user. You may call more tools if needed.",
                    }
                )
            else:
                if result.text:
                    final_answer_parts.append(result.text)

                full_response = "".join(final_answer_parts).strip() or "I'll help you with that."
                yield AskStreamEvent(
                    type_=AskStreamEventType.TEXT,
                    text=full_response,
                )

                ai_message = await self._persist_ai_message(
                    command,
                    conversation,
                    full_response,
                    now,
                    retrieved_hits=merged_hits,
                    tool_calls=tool_calls_flat,
                )

                cache_key = self._build_cache_key(command, conversation.id)
                await self._cache_response(cache_key, full_response)
                await self._publish_query_event(command, conversation, ai_message)
                await self._try_update_title(conversation, command, tool_calls_flat)

                yield AskStreamEvent(
                    type_=AskStreamEventType.DONE,
                    done=conversation,
                    ai_message=ai_message,
                    merged_hits=merged_hits,
                )
                return

        finalize_text = await self._finalize_from_gathered_context(
            messages, system_prompt, command.prompt
        )
        if finalize_text:
            final_answer_parts.append(finalize_text)

        final_answer = "".join(final_answer_parts).strip() or FINAL_ANSWER_FALLBACK
        yield AskStreamEvent(type_=AskStreamEventType.TEXT, text=final_answer)

        ai_message = await self._persist_ai_message(
            command,
            conversation,
            final_answer,
            now,
            retrieved_hits=merged_hits,
            tool_calls=tool_calls_flat,
        )

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

    def _merge_tool_hits(
        self,
        merged_hits: list[SearchHit],
        tool_hits: list[SearchHit],
    ) -> list[SearchHit]:
        seen = {str(h.chunk_id) for h in merged_hits}
        merged = list(merged_hits)
        for hit in tool_hits:
            if str(hit.chunk_id) not in seen:
                seen.add(str(hit.chunk_id))
                merged.append(hit)
        return merged

    def _suppression_nudge(self, prompt: str, reason: SuppressionReason | None) -> str:
        if reason is SuppressionReason.DRIFT:
            return (
                f"That search drifted away from the user's question: '{prompt}'. "
                "Stay on the original question and answer it from the context already "
                "gathered. If the user wants a different topic, invite them to ask it "
                "in a new message."
            )
        return (
            "The knowledge base was already searched with that query this turn. "
            "Do not repeat it. Use the context provided so far, or search the "
            "knowledge base only with a genuinely different query if more evidence "
            "is needed."
        )

    async def _finalize_from_gathered_context(
        self,
        messages: list[dict[str, str]],
        system_prompt: str,
        prompt: str,
    ) -> str:
        gathered_parts: list[str] = []
        for message in messages[1:]:
            content = message.get("content", "")
            if content.startswith(("Context:", "Here are the results:")):
                gathered_parts.append(content)
        finalize_note = (
            f"The user asked: '{prompt}'. You have searched enough for this turn. "
            "Provide the final answer now using only the information gathered below. "
            "Do not call any tools."
        )
        if gathered_parts:
            gathered = "\n\n".join(gathered_parts)
            finalize_note += f"\n\nInformation gathered:\n{gathered[:6000]}"
        messages.append({"role": "user", "content": finalize_note})
        try:
            result = await self._llm_port.generate(
                messages[-1]["content"],
                system_prompt=system_prompt,
                max_tokens=DEFAULT_MAX_OUTPUT_TOKENS,
                temperature=DEFAULT_LLM_TEMPERATURE,
            )
        except Exception:
            logger.warning("Forced finalization LLM call failed")
            return ""
        return (result.text or "").strip()

    def _format_history(self, history: list[AIChatMessage]) -> str:
        parts: list[str] = []
        for msg in history:
            if msg.message_type == MessageType.USER_QUERY and msg.user_query:
                parts.append(f"User: {msg.user_query}")
            elif msg.message_type == MessageType.AI_RESPONSE and msg.llm_response:
                tool_history = ""
                if msg.tool_calls:
                    summaries = [
                        f"  - {tc.tool_name}: {tc.result_summary}"
                        for tc in msg.tool_calls
                        if not tc.suppressed
                    ]
                    if summaries:
                        tool_history = "\n" + "\n".join(summaries)
                parts.append(f"Assistant: {msg.llm_response}{tool_history}")
        if parts:
            return "Conversation history:\n" + "\n".join(reversed(parts))
        return ""

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
        tool_calls: list[ToolCallRecord] | None = None,
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
            tool_calls=tool_calls,
            agent_strategy="agentic",
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
                "agent_strategy": "agentic",
                "cache_hit": ai_message.cache_hit,
            }
            await self._event_bus.publish("ai.query.completed", payload)
        except AIServiceError:
            pass

    async def _try_update_title(
        self,
        conversation: AIConversationSession,
        command: AskAICommand,
        tool_calls_flat: list[ToolCallRecord],
    ) -> None:
        if not tool_calls_flat or conversation.title != command.prompt[:80]:
            return

        tool_names = ", ".join(sorted({tc.tool_name for tc in tool_calls_flat}))
        title_prompt = (
            f"Summarize this Ethiopian business query in 1-5 words: "
            f"'{command.prompt}' (tools used: {tool_names})"
        )
        try:
            result = await self._llm_port.generate(
                title_prompt,
                max_tokens=20,
                temperature=0.3,
            )
            new_title = result.text.strip().strip("'\"").strip()
            max_title_len = 100
            if new_title and len(new_title) <= max_title_len:
                await self._conversation.update_session_title(conversation.id, new_title)
        except Exception:
            logger.warning("Failed to generate conversation title for %s", conversation.id)
