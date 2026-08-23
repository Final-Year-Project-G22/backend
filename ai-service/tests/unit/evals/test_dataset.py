"""Golden-set matrix validation tests."""

from __future__ import annotations

from pathlib import Path

import pytest

from evals.dataset import DatasetError, build_fixture_index, load_golden_set, validate_matrix
from evals.models import ChunkRef
from tests.unit.evals.conftest import make_item

GOLDEN_PATH = Path(__file__).resolve().parents[3] / "evals" / "golden.jsonl"


def test_shipped_golden_set_is_valid() -> None:
    items = load_golden_set(GOLDEN_PATH)
    assert len(items) == 36


def test_duplicate_item_id_rejected() -> None:
    item = make_item()
    with pytest.raises(DatasetError, match="duplicate item id"):
        validate_matrix([item, item])


def test_unknown_expected_restricted_to_knowledge() -> None:
    item = make_item(id="ps-en-e1-x", intent="personal", unknown_expected=True)
    with pytest.raises(DatasetError, match="unknown_expected"):
        validate_matrix([item])


def test_expected_tools_must_be_allowed() -> None:
    item = make_item(expected_tools=["search_trusted_web"])
    with pytest.raises(DatasetError, match="subset of allowed_tools"):
        validate_matrix([item])


def test_required_citations_within_expected() -> None:
    item = make_item(
        required_citations=[ChunkRef(document_key="en:tax_code:tin-registration", chunk_index=0)]
    )
    with pytest.raises(DatasetError, match="within expected_chunk_refs"):
        validate_matrix([item])


def test_incomplete_cell_rejected() -> None:
    items = [make_item(id=f"kn-en-{i}") for i in range(5)]
    with pytest.raises(DatasetError, match="expected 6 items"):
        validate_matrix(items)


def test_build_fixture_index_maps_all_chunks() -> None:
    index = build_fixture_index()
    assert index.fixture_version == "1.0.0"
    assert len(index.id_to_ref) == 62
    assert sum(len(refs) for refs in index.refs_by_document.values()) == 62
