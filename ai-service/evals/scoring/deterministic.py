"""Deterministic gates: citations, grounding, tools, recall, unknown handling."""

from __future__ import annotations

from evals.models import NOT_APPLICABLE, ChunkRef, GoldenItem, ItemTrace, MetricValue

UNKNOWN_MARKERS_EN = (
    "do not have the verified",
    "don't have the verified",
    "cannot find",
    "can't find",
    "could not find",
    "couldn't find",
    "no verified ethiopian regulatory documents",
    "do not have enough verified information",
    "don't have enough verified information",
    "unable to verify this from the available documents",
    "not mentioned in the provided documents",
    "not specified in the provided documents",
    "the provided documents do not mention",
    "the documents do not specify",
    "i do not have information",
    "i don't have information",
)
UNKNOWN_MARKERS_AM = (
    "የተረጋገጠ መረጃ የላቝ",
    "የተረጋገጡ ሰነዶች የሉኝም",
    "ያረጋገጥኩት መረጃ የለም",
    "ማረጋገጫ ሰነዶች አልተገኙም",
    "በዚህ ጥያቄ ላይ ትክክለኛ መረጃ የለም",
    "የተረጋገጠ የኢትዮጵያ ሕግ መረጃ የለም",
    "መልስ መስጠት አልችልም",
    "ማወቅ አልቻልኩም",
    "አልተገኘም",
    "አያውቅም",
)


def _ref_set(refs: list[ChunkRef]) -> set[str]:
    return {ref.ref_id() for ref in refs}


def _cited_refs(trace: ItemTrace) -> list[ChunkRef]:
    return [source.chunk_ref for source in trace.cited_sources if source.chunk_ref is not None]


def citation_coverage(item: GoldenItem, trace: ItemTrace) -> MetricValue:
    if not item.required_citations:
        return NOT_APPLICABLE
    cited = _ref_set(_cited_refs(trace))
    covered = sum(1 for ref in item.required_citations if ref.ref_id() in cited)
    value = covered / len(item.required_citations)
    missing = [ref.ref_id() for ref in item.required_citations if ref.ref_id() not in cited]
    return MetricValue(value=value, detail={"missing": missing})


def grounding_precision(
    trace: ItemTrace,
    known_refs: set[str],
) -> MetricValue:
    """Fraction of cited sources that exist in the fixture and were actually retrieved.

    ``response_sources`` are machine-derived from retrieval today, so this gate
    is structural in v1; it guards against future regressions where citations
    diverge from the captured context.
    """
    cited = _cited_refs(trace)
    if not cited:
        return NOT_APPLICABLE
    retrieved = _ref_set(trace.context_chunks)
    grounded = sum(1 for ref in cited if ref.ref_id() in known_refs and ref.ref_id() in retrieved)
    ungrounded = [
        ref.ref_id()
        for ref in cited
        if ref.ref_id() not in known_refs or ref.ref_id() not in retrieved
    ]
    return MetricValue(
        value=grounded / len(cited),
        detail={"ungrounded": ungrounded},
    )


def tool_adherence(item: GoldenItem, trace: ItemTrace) -> tuple[MetricValue, bool]:
    """Returns the metric plus a hard-fail flag for wrong-family tool use."""
    executed = {call.tool_name for call in trace.tool_calls if not call.suppressed and call.success}
    wrong_family = sorted(executed - set(item.allowed_tools))
    if wrong_family:
        return (
            MetricValue(value=0.0, detail={"wrong_family": wrong_family}),
            True,
        )
    if not item.expected_tools:
        return MetricValue(value=1.0, detail={"executed": sorted(executed)}), False
    covered = len(set(item.expected_tools) & executed)
    return (
        MetricValue(
            value=covered / len(item.expected_tools),
            detail={"expected": item.expected_tools, "executed": sorted(executed)},
        ),
        False,
    )


def context_recall(item: GoldenItem, trace: ItemTrace) -> MetricValue:
    if not item.expected_chunk_refs:
        return NOT_APPLICABLE
    retrieved = _ref_set(trace.context_chunks)
    hits = sum(1 for ref in item.expected_chunk_refs if ref.ref_id() in retrieved)
    missing = [ref.ref_id() for ref in item.expected_chunk_refs if ref.ref_id() not in retrieved]
    return MetricValue(value=hits / len(item.expected_chunk_refs), detail={"missing": missing})


def unknown_handling(item: GoldenItem, answer: str) -> MetricValue:
    if not item.unknown_expected:
        return NOT_APPLICABLE
    lowered = answer.lower().strip()
    refused = any(marker in lowered for marker in UNKNOWN_MARKERS_EN) or any(
        marker in answer for marker in UNKNOWN_MARKERS_AM
    )
    return MetricValue(value=1.0 if refused else 0.0, detail={"refused": refused})


__all__ = [
    "citation_coverage",
    "context_recall",
    "grounding_precision",
    "tool_adherence",
    "unknown_handling",
]
