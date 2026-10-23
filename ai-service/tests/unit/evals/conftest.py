"""Shared fixtures for eval-harness unit tests."""

from __future__ import annotations

import uuid
from typing import Any

import pytest

from evals.dataset import FixtureIndex
from evals.models import ChunkRef, GoldenItem, ItemTrace


def make_item(**overrides: Any) -> GoldenItem:
    base: dict[str, Any] = {
        "id": "kn-en-e1-test",
        "intent": "knowledge",
        "locale": "en",
        "difficulty": "easy",
        "query": "What is VAT?",
        "expected_tools": [],
        "allowed_tools": ["search_knowledge_base"],
        "required_citations": [
            ChunkRef(document_key="en:tax_code:vat-registration", chunk_index=0)
        ],
        "expected_chunk_refs": [
            ChunkRef(document_key="en:tax_code:vat-registration", chunk_index=0)
        ],
        "golden_answer": "VAT is a consumption tax.",
        "golden_claims": ["VAT is a consumption tax."],
        "unknown_expected": False,
        "fixture_account": "eval-msme-01",
        "dataset_version": "1.0.0",
    }
    base.update(overrides)
    return GoldenItem.model_validate(base)


def make_trace(**overrides: Any) -> ItemTrace:
    base: dict[str, Any] = {
        "item_id": "kn-en-e1-test",
        "conversation_id": uuid.UUID(int=1),
        "answer": "VAT is a consumption tax applied in Ethiopia.",
        "context_chunks": [ChunkRef(document_key="en:tax_code:vat-registration", chunk_index=0)],
        "context_texts": ["Value Added Tax (VAT) is a consumption tax."],
    }
    base.update(overrides)
    return ItemTrace.model_validate(base)


@pytest.fixture
def fixture_index() -> FixtureIndex:
    namespace = uuid.UUID("7a0d5c7a-0000-4000-8000-0000000000e1")
    ref = ChunkRef(document_key="en:tax_code:vat-registration", chunk_index=0)
    chunk_id = uuid.uuid5(namespace, f"chunk:{ref.document_key}:{ref.chunk_index}")
    return FixtureIndex(
        fixture_version="test",
        id_to_ref={str(chunk_id): ref},
        refs_by_document={"en:tax_code:vat-registration": [ref]},
    )
