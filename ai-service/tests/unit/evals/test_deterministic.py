"""Deterministic gate tests."""

from __future__ import annotations

from evals.models import ChunkRef, CitedSource, ToolCallTrace
from evals.scoring import (
    citation_coverage,
    context_recall,
    grounding_precision,
    tool_adherence,
    unknown_handling,
)
from tests.unit.evals.conftest import make_item, make_trace

VAT0 = ChunkRef(document_key="en:tax_code:vat-registration", chunk_index=0)
VAT1 = ChunkRef(document_key="en:tax_code:vat-registration", chunk_index=1)
TIN3 = ChunkRef(document_key="en:tax_code:tin-registration", chunk_index=3)


def test_citation_coverage_full_and_partial() -> None:
    item = make_item(required_citations=[VAT0, VAT1])
    hit = make_trace(
        cited_sources=[
            CitedSource(chunk_ref=VAT0, title="t"),
            CitedSource(chunk_ref=TIN3, title="u"),
        ]
    )
    score = citation_coverage(item, hit)
    assert score.applicable
    assert score.value == 0.5
    assert score.detail["missing"] == [VAT1.ref_id()]


def test_citation_coverage_not_applicable_without_requirements() -> None:
    score = citation_coverage(make_item(required_citations=[]), make_trace())
    assert not score.applicable


def test_grounding_precision_counts_unknown_refs_as_ungrounded() -> None:
    hallucinated = ChunkRef(document_key="en:tax_code:made-up-doc", chunk_index=9)
    hit = make_trace(
        cited_sources=[
            CitedSource(chunk_ref=VAT0, title="t"),
            CitedSource(chunk_ref=hallucinated, title="h"),
        ]
    )
    known_refs = {VAT0.ref_id()}
    score = grounding_precision(hit, known_refs)
    assert score.value == 0.5
    assert score.detail["ungrounded"] == [hallucinated.ref_id()]


def test_tool_adherence_wrong_family_is_hard_fail() -> None:
    item = make_item(allowed_tools=["search_knowledge_base"])
    trace = make_trace(tool_calls=[ToolCallTrace(tool_name="search_trusted_web")])
    metric, hard_fail = tool_adherence(item, trace)
    assert hard_fail
    assert metric.value == 0.0
    assert metric.detail["wrong_family"] == ["search_trusted_web"]


def test_tool_adherence_no_expected_tools_scores_full() -> None:
    item = make_item(expected_tools=[], allowed_tools=["search_knowledge_base"])
    trace = make_trace(tool_calls=[ToolCallTrace(tool_name="search_knowledge_base")])
    metric, hard_fail = tool_adherence(item, trace)
    assert not hard_fail
    assert metric.value == 1.0


def test_context_recall_measures_expected_evidence_retrieved() -> None:
    item = make_item(expected_chunk_refs=[VAT0, VAT1])
    trace = make_trace(context_chunks=[VAT1])
    score = context_recall(item, trace)
    assert score.value == 0.5
    assert score.detail["missing"] == [VAT0.ref_id()]


def test_unknown_handling_detects_english_refusal() -> None:
    item = make_item(id="kn-en-m2-u", unknown_expected=True)
    refused = unknown_handling(
        item,
        "I currently do not have the verified Ethiopian regulatory documents to answer this specific question.",
    )
    assert refused.value == 1.0
    fabricated = unknown_handling(item, "The VAT rate in Ethiopia is 15 percent.")
    assert fabricated.value == 0.0


def test_unknown_handling_detects_amharic_refusal() -> None:
    item = make_item(id="kn-am-m2-u", locale="am", unknown_expected=True)
    refused = unknown_handling(item, "የተረጋገጡ ሰነዶች የሉኝም፤ መጠኑን መጥቀስ አልችልም።")
    assert refused.value == 1.0
