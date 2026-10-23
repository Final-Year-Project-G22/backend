"""Golden-set loading and validation against the frozen fixture."""

from __future__ import annotations

import json
import uuid
from dataclasses import dataclass
from pathlib import Path

from evals.models import ChunkRef, Difficulty, GoldenItem, Intent, Locale

EVALS_DIR = Path(__file__).resolve().parent
GOLDEN_PATH = EVALS_DIR / "golden.jsonl"
FIXTURE_MANIFEST_PATH = EVALS_DIR / "fixture" / "manifest.json"
CHUNK_REFS_PATH = EVALS_DIR / "fixture" / "chunk_refs.json"

INTENTS_PER_CELL = {Intent.KNOWLEDGE, Intent.PERSONAL, Intent.MIXED}
LOCALES = {Locale.EN, Locale.AM}
DIFFICULTIES = {Difficulty.EASY, Difficulty.MEDIUM, Difficulty.HARD}
ITEMS_PER_CELL = 6
EASY_PER_CELL = 2
MEDIUM_PER_CELL = 2
HARD_PER_CELL = 2
UNKNOWN_ITEMS_PER_LOCALE = 2


@dataclass(frozen=True)
class FixtureIndex:
    """Reverse map from deterministic chunk id to stable chunk ref."""

    fixture_version: str
    id_to_ref: dict[str, ChunkRef]
    refs_by_document: dict[str, list[ChunkRef]]

    def ref_for_chunk_id(self, chunk_id: str) -> ChunkRef | None:
        return self.id_to_ref.get(chunk_id)


def load_golden_set(path: Path = GOLDEN_PATH) -> list[GoldenItem]:
    items: list[GoldenItem] = []
    with path.open(encoding="utf-8") as handle:
        for line_no, line in enumerate(handle, start=1):
            stripped = line.strip()
            if not stripped:
                continue
            try:
                items.append(GoldenItem.model_validate(json.loads(stripped)))
            except ValueError as exc:
                msg = f"golden.jsonl line {line_no}: invalid item"
                raise DatasetError(msg) from exc
    if not items:
        msg = "golden set is empty"
        raise DatasetError(msg)
    validate_matrix(items)
    return items


class DatasetError(RuntimeError):
    """Raised when the golden set or fixture index is inconsistent."""


def _validate_item(item: GoldenItem, seen_ids: set[str]) -> None:
    if item.id in seen_ids:
        msg = f"duplicate item id: {item.id}"
        raise DatasetError(msg)
    seen_ids.add(item.id)
    if item.unknown_expected and item.intent is not Intent.KNOWLEDGE:
        msg = f"{item.id}: unknown_expected is restricted to knowledge items"
        raise DatasetError(msg)
    disallowed = set(item.expected_tools) - set(item.allowed_tools)
    if disallowed:
        msg = f"{item.id}: expected_tools must be a subset of allowed_tools"
        raise DatasetError(msg)
    required_refs = {ref.ref_id() for ref in item.required_citations}
    expected_refs = {ref.ref_id() for ref in item.expected_chunk_refs}
    missing = required_refs - expected_refs
    if missing:
        msg = f"{item.id}: required_citations must be within expected_chunk_refs"
        raise DatasetError(msg)


def _validate_cell(intent: Intent, locale: Locale, cell_items: list[GoldenItem]) -> None:
    if len(cell_items) != ITEMS_PER_CELL:
        msg = f"cell {intent.value}/{locale.value}: expected {ITEMS_PER_CELL} items, got {len(cell_items)}"
        raise DatasetError(msg)
    counts: dict[Difficulty, int] = dict.fromkeys(DIFFICULTIES, 0)
    for cell_item in cell_items:
        counts[cell_item.difficulty] += 1
    if (
        counts[Difficulty.EASY] != EASY_PER_CELL
        or counts[Difficulty.MEDIUM] != MEDIUM_PER_CELL
        or counts[Difficulty.HARD] != HARD_PER_CELL
    ):
        msg = f"cell {intent.value}/{locale.value}: difficulty split must be 2/2/2, got {counts}"
        raise DatasetError(msg)


def _validate_unknowns(locale: Locale, items: list[GoldenItem]) -> None:
    unknowns = [item for item in items if item.locale is locale and item.unknown_expected]
    if len(unknowns) != UNKNOWN_ITEMS_PER_LOCALE:
        msg = f"locale {locale.value}: expected exactly {UNKNOWN_ITEMS_PER_LOCALE} unknown_expected items, got {len(unknowns)}"
        raise DatasetError(msg)
    if {item.difficulty for item in unknowns} != {Difficulty.MEDIUM, Difficulty.HARD}:
        msg = f"locale {locale.value}: unknown items must be one medium and one hard"
        raise DatasetError(msg)


def validate_matrix(items: list[GoldenItem]) -> None:
    seen_ids: set[str] = set()
    for item in items:
        _validate_item(item, seen_ids)

    cells: dict[tuple[Intent, Locale], list[GoldenItem]] = {}
    for item in items:
        cells.setdefault((item.intent, item.locale), []).append(item)

    for intent in INTENTS_PER_CELL:
        for locale in LOCALES:
            _validate_cell(intent, locale, cells.get((intent, locale), []))

    for locale in LOCALES:
        _validate_unknowns(locale, items)


def build_fixture_index() -> FixtureIndex:
    manifest = json.loads(FIXTURE_MANIFEST_PATH.read_text(encoding="utf-8"))
    chunk_refs = json.loads(CHUNK_REFS_PATH.read_text(encoding="utf-8"))
    namespace = uuid.UUID("7a0d5c7a-0000-4000-8000-0000000000e1")

    id_to_ref: dict[str, ChunkRef] = {}
    refs_by_document: dict[str, list[ChunkRef]] = {}
    for document in manifest["documents"]:
        document_key = document["document_key"]
        entry = chunk_refs.get(document_key)
        if entry is None:
            msg = f"chunk_refs.json missing entry for {document_key}"
            raise DatasetError(msg)
        refs: list[ChunkRef] = []
        for chunk_index in range(entry["chunk_count"]):
            ref = ChunkRef(document_key=document_key, chunk_index=chunk_index)
            chunk_id = uuid.uuid5(namespace, f"chunk:{document_key}:{chunk_index}")
            id_to_ref[str(chunk_id)] = ref
            refs.append(ref)
        refs_by_document[document_key] = refs

    return FixtureIndex(
        fixture_version=manifest["fixture_version"],
        id_to_ref=id_to_ref,
        refs_by_document=refs_by_document,
    )


__all__ = [
    "GOLDEN_PATH",
    "DatasetError",
    "FixtureIndex",
    "build_fixture_index",
    "load_golden_set",
    "validate_matrix",
]
