from __future__ import annotations

import inspect

from core.ports.cache import CachePort
from core.ports.conversation_repository import ConversationRepositoryPort
from core.ports.core_service import CoreServicePort
from core.ports.embedding import EmbeddingPort
from core.ports.event_bus import EventBusPort
from core.ports.knowledge_repository import KnowledgeRepositoryPort
from core.ports.llm import LLMPort
from core.ports.quota_repository import QuotaRepositoryPort


def test_embedding_port_contract_shape() -> None:
    assert inspect.isabstract(EmbeddingPort)
    expected = {"dimensions", "provider", "embed_documents", "embed_query"}
    assert expected.issubset(EmbeddingPort.__abstractmethods__)


def test_llm_port_contract_shape() -> None:
    assert inspect.isabstract(LLMPort)
    expected = {"provider", "model", "generate", "generate_stream"}
    assert expected.issubset(LLMPort.__abstractmethods__)


def test_llm_port_generate_stream_is_not_coroutine_function() -> None:
    assert inspect.iscoroutinefunction(LLMPort.generate) is True
    assert inspect.iscoroutinefunction(LLMPort.generate_stream) is False

    signature = inspect.signature(LLMPort.generate_stream)
    assert str(signature.return_annotation) == "AsyncIterator[LLMChunk]"


def test_knowledge_repository_port_contract_shape() -> None:
    assert inspect.isabstract(KnowledgeRepositoryPort)
    expected = {
        "delete_chunks_by_document",
        "get_chunks_by_document",
        "get_document",
        "get_document_by_external_id",
        "list_documents",
        "search_bm25",
        "search_vector",
        "soft_delete_document",
        "upsert_chunks",
        "upsert_document",
    }
    assert expected.issubset(KnowledgeRepositoryPort.__abstractmethods__)


def test_conversation_repository_port_contract_shape() -> None:
    assert inspect.isabstract(ConversationRepositoryPort)
    expected = {
        "add_message",
        "create_session",
        "get_message",
        "get_session",
        "list_messages",
        "list_sessions_by_user",
        "soft_delete_session",
        "update_session",
    }
    assert expected.issubset(ConversationRepositoryPort.__abstractmethods__)


def test_quota_repository_port_contract_shape() -> None:
    assert inspect.isabstract(QuotaRepositoryPort)
    expected = {"get_quota", "increment_usage", "reset_daily_usage", "upsert_quota"}
    assert expected.issubset(QuotaRepositoryPort.__abstractmethods__)


def test_cache_port_contract_shape() -> None:
    assert inspect.isabstract(CachePort)
    expected = {"delete", "delete_pattern", "get", "set"}
    assert expected.issubset(CachePort.__abstractmethods__)


def test_event_bus_port_contract_shape() -> None:
    assert inspect.isabstract(EventBusPort)
    expected = {"publish", "subscribe"}
    assert expected.issubset(EventBusPort.__abstractmethods__)


def test_core_service_port_contract_shape() -> None:
    assert inspect.isabstract(CoreServicePort)
    expected = {"get_user_profile", "get_user_tier"}
    assert expected.issubset(CoreServicePort.__abstractmethods__)
