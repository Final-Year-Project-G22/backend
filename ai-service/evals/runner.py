"""In-process evaluation runner around AskAIUseCase.execute."""

from __future__ import annotations

import asyncio
import logging
import uuid
from dataclasses import dataclass
from datetime import UTC, datetime
from typing import Any

import httpx

from app.config import Settings
from app.container import Container
from core.domain.enums import Language
from core.domain.exceptions import LLMError
from core.ports.event_bus import EventBusPort, EventHandler, EventPayload
from core.usecases.contracts import AskAICommand, AskAIResult
from evals.dataset import FixtureIndex
from evals.models import ChunkRef, CitedSource, GoldenItem, ItemTrace, PromptVariant, ToolCallTrace
from infrastructure.embeddings import create_embedding_adapter
from infrastructure.llm import create_llm_adapter
from infrastructure.prompts import PromptLoader

logger = logging.getLogger(__name__)

FIXTURE_USER_ID = "10000000-0000-4000-8000-000000000001"
FIXTURE_ACCOUNT_ID = "10000000-0000-4000-8000-000000000002"
FIXTURE_NAMESPACE = uuid.UUID("7a0d5c7a-0000-4000-8000-0000000000e1")
MIN_FIXTURE_QUOTA_QUERIES = 36

PERSONAL_CONTEXT_TOOLS = ("get_user_profile", "get_guide_progress", "check_compliance_status")

BROKEN_AGENTIC_PROMPT = """\
You are the Adisu Serategna AI Advisor, an assistant for Ethiopian businesses.
Answer the user's question using your general knowledge and your own best judgment.

You do not need to limit yourself to the provided context: it may be incomplete or
irrelevant, so supplement it freely with what you know. Citations are optional and
should not slow you down. If you are not certain of an answer, give your best guess
anyway — users prefer a confident attempt over a refusal. Be brief.
"""


class _NullEventBus(EventBusPort):
    """Evaluation runs must not emit side effects onto the message bus."""

    async def publish(self, topic: str, payload: EventPayload) -> None:  # noqa: ARG002
        logger.debug("eval run suppressed event on %s", topic)

    async def subscribe(self, topic: str, handler: EventHandler) -> None:  # noqa: ARG002
        return None


class BrokenPromptLoader(PromptLoader):
    """Deliberately degraded agentic prompt used for calibration contrast."""

    def render_agentic(
        self,
        locale: str,  # noqa: ARG002 - signature must match PromptLoader
        tools: list[dict[str, str]] | None = None,  # noqa: ARG002
        *,
        kb_context_available: bool = False,  # noqa: ARG002
    ) -> str:
        return BROKEN_AGENTIC_PROMPT


class RunnerError(RuntimeError):
    """Raised when the environment cannot support an evaluation run."""


@dataclass
class RunEnvironment:
    container: Container
    settings: Settings
    http_client: httpx.AsyncClient


def build_environment(variant: PromptVariant) -> RunEnvironment:
    container = Container()
    settings = container.config()
    http_client = httpx.AsyncClient(trust_env=settings.HTTPX_TRUST_ENV)
    embedding_adapter = create_embedding_adapter(settings, http_client=http_client)
    llm_adapter = create_llm_adapter(settings, http_client=http_client)
    cast_any: Any = container
    cast_any.embedding_port.override(embedding_adapter)
    cast_any.llm_port.override(llm_adapter)
    cast_any.event_bus_port.override(_NullEventBus())
    if variant is PromptVariant.BROKEN:
        cast_any.prompt_loader.override(BrokenPromptLoader(settings.AI_PROMPT_DIR))
    return RunEnvironment(container=container, settings=settings, http_client=http_client)


async def preflight(container: Container, fixture_index: FixtureIndex) -> None:
    """Fail setup before scoring when the fixture is unavailable (FIN-69)."""
    tool_registry = container.tool_registry()
    await tool_registry.initialize()
    intent_classifier = container.intent_classifier()
    await intent_classifier.initialize()

    knowledge_repository = container.knowledge_repository()
    expected_total = len(fixture_index.id_to_ref)
    embedded = 0
    for document_key, refs in fixture_index.refs_by_document.items():
        chunks = await knowledge_repository.get_chunks_by_document(_doc_uuid(document_key))
        if len(chunks) != len(refs):
            msg = (
                f"fixture not ready: {document_key} has {len(chunks)} chunks, "
                f"expected {len(refs)} — re-run provision_corpus.py"
            )
            raise RunnerError(msg)
        embedded += sum(1 for chunk in chunks if chunk.status.value == "embedded")
    if embedded != expected_total:
        msg = (
            f"fixture not ready: {embedded}/{expected_total} chunks EMBEDDED — "
            "run evals/fixture/scripts/provision_corpus.py without --skip-embeddings"
        )
        raise RunnerError(msg)

    quota_repository = container.quota_repository()
    quota = await quota_repository.get_quota(_fixture_user_uuid())
    if quota is None or quota.daily_query_limit < MIN_FIXTURE_QUOTA_QUERIES:
        msg = "fixture user quota missing or too small — re-run provision_corpus.py"
        raise RunnerError(msg)


ASK_RETRY_DELAYS_SECONDS = (30.0, 60.0, 120.0, 240.0)


async def run_item(
    item: GoldenItem,
    container: Container,
    fixture_index: FixtureIndex,
    *,
    retry_delays: tuple[float, ...] = ASK_RETRY_DELAYS_SECONDS,
) -> ItemTrace:
    ask_ai = container.ask_ai()
    command = AskAICommand(
        user_id=_fixture_user_uuid(),
        account_id=_fixture_account_uuid(),
        prompt=item.query,
        language=Language(item.locale.value),
        strategy="agentic",
        debug_mode=True,
    )
    result: AskAIResult | None = None
    last_error: Exception | None = None
    for attempt, delay in enumerate((0.0, *retry_delays), start=1):
        if delay:
            logger.warning(
                "item %s ask failed, retry %d/%d in %.0fs",
                item.id,
                attempt - 1,
                len(retry_delays),
                delay,
            )
            await asyncio.sleep(delay)
        try:
            result = await ask_ai.execute(command)
            break
        except LLMError as exc:
            last_error = exc
    if result is None:
        msg = f"ask failed after {len(retry_delays)} retries: {last_error}"
        raise RunnerError(msg)
    if result.cache_hit:
        msg = f"{item.id}: cache_hit is true — evaluation runs must bypass cache"
        raise RunnerError(msg)

    message = result.ai_message
    cited = [
        CitedSource(
            chunk_ref=fixture_index.ref_for_chunk_id(str(source.chunk_id))
            if source.chunk_id is not None
            else None,
            title=source.title,
            excerpt=source.excerpt,
        )
        for source in message.response_sources
    ]
    context_refs: list[ChunkRef] = []
    context_texts: list[str] = []
    seen_chunks: set[str] = set()
    for hit in result.retrieved_hits:
        key = str(hit.chunk_id)
        if key in seen_chunks:
            continue
        seen_chunks.add(key)
        ref = fixture_index.ref_for_chunk_id(key)
        if ref is not None:
            context_refs.append(ref)
        context_texts.append(hit.chunk_text)

    if item.intent.value != "knowledge":
        context_texts.extend(await _personal_context(container))

    usage: dict[str, Any] | None = None
    if message.token_usage is not None:
        usage = message.token_usage.model_dump()

    return ItemTrace(
        item_id=item.id,
        conversation_id=result.conversation.id,
        answer=message.llm_response or "",
        context_chunks=context_refs,
        context_texts=context_texts,
        cited_sources=cited,
        tool_calls=[
            ToolCallTrace(
                tool_name=record.tool_name,
                suppressed=record.suppressed,
                success=record.success,
            )
            for record in (message.tool_calls or [])
        ],
        cache_hit=message.cache_hit,
        usage=usage,
        model_used=message.model_used,
        processing_time_ms=message.processing_time_ms,
    )


async def _personal_context(container: Container) -> list[str]:
    """Fresh personal-context snapshots matching what pre-fetch injected."""
    registry = container.tool_registry()
    texts: list[str] = []
    for tool_name in PERSONAL_CONTEXT_TOOLS:
        try:
            outcome = await registry.execute_tool(
                tool_name,
                {},
                FIXTURE_ACCOUNT_ID,
                FIXTURE_USER_ID,
            )
        except Exception as exc:
            logger.warning("personal context tool %s failed: %s", tool_name, exc)
            continue
        if outcome.success and outcome.result_text.strip():
            texts.append(outcome.result_text)
    return texts


def _fixture_user_uuid() -> uuid.UUID:
    return uuid.UUID(FIXTURE_USER_ID)


def _fixture_account_uuid() -> uuid.UUID:
    return uuid.UUID(FIXTURE_ACCOUNT_ID)


def _doc_uuid(document_key: str) -> uuid.UUID:
    return uuid.uuid5(FIXTURE_NAMESPACE, f"doc:{document_key}")


def utc_now() -> datetime:
    return datetime.now(UTC)


__all__ = [
    "BrokenPromptLoader",
    "RunEnvironment",
    "RunnerError",
    "build_environment",
    "preflight",
    "run_item",
    "utc_now",
]
