from __future__ import annotations

import pytest

from core.ports.chunking import ChunkingStrategy, ChunkProvenance
from core.ports.parser import ParsedDocument, ParsedDocumentSection
from infrastructure.chunking import ChunkingRegistry, StructuralChunkingAdapter

_EXPECTED_TWO_SECTIONS = 2
_MAX_TOKENS_100 = 100
_DEFAULT_MAX_TOKENS = 512
_DEFAULT_OVERLAP_TOKENS = 50
_DEFAULT_MIN_CHUNK_TOKENS = 20
_CUSTOM_MAX_TOKENS = 256
_CUSTOM_OVERLAP_TOKENS = 25


class TestStructuralChunkingAdapter:
    def test_supports_returns_true_for_structural(self) -> None:
        chunker = StructuralChunkingAdapter()
        assert chunker.supports("structural") is True

    def test_supports_returns_false_for_other(self) -> None:
        chunker = StructuralChunkingAdapter()
        assert chunker.supports("recursive") is False
        assert chunker.supports("semantic") is False

    @pytest.mark.asyncio
    async def test_chunk_empty_document_returns_empty(self) -> None:
        chunker = StructuralChunkingAdapter()
        doc = ParsedDocument(document_text="", sections=[])
        strategy = ChunkingStrategy()
        result = await chunker.chunk(doc, strategy)
        assert result == []

    @pytest.mark.asyncio
    async def test_chunk_single_section_small(self) -> None:
        chunker = StructuralChunkingAdapter()
        doc = ParsedDocument(
            document_text="Hello world",
            sections=[
                ParsedDocumentSection(heading="Intro", content="Hello world", order=0),
            ],
        )
        strategy = ChunkingStrategy()
        result = await chunker.chunk(doc, strategy)

        assert len(result) == 1
        assert result[0].chunk_text == "Hello world"
        assert result[0].provenance.section_heading == "Intro"
        assert result[0].provenance.section_order == 0
        assert result[0].provenance.chunk_index == 0

    @pytest.mark.asyncio
    async def test_chunk_multiple_sections(self) -> None:
        chunker = StructuralChunkingAdapter()
        doc = ParsedDocument(
            document_text="Intro content\n\nSection 1 content",
            sections=[
                ParsedDocumentSection(heading="Intro", content="Intro content", order=0),
                ParsedDocumentSection(heading="Section 1", content="Section 1 content", order=1),
            ],
        )
        strategy = ChunkingStrategy()
        result = await chunker.chunk(doc, strategy)

        assert len(result) == _EXPECTED_TWO_SECTIONS
        assert result[0].provenance.section_order == 0
        assert result[1].provenance.section_order == 1
        assert result[1].provenance.chunk_index == 1

    @pytest.mark.asyncio
    async def test_chunk_respects_max_tokens(self) -> None:
        chunker = StructuralChunkingAdapter()
        long_text = "word " * 300
        doc = ParsedDocument(
            document_text=long_text,
            sections=[
                ParsedDocumentSection(heading=None, content=long_text, order=0),
            ],
        )
        strategy = ChunkingStrategy(max_tokens=100)
        result = await chunker.chunk(doc, strategy)

        assert len(result) > 1
        for chunk in result:
            assert chunk.token_count <= _MAX_TOKENS_100

    @pytest.mark.asyncio
    async def test_chunk_with_custom_metadata(self) -> None:
        chunker = StructuralChunkingAdapter()
        doc = ParsedDocument(
            document_text="Test",
            sections=[
                ParsedDocumentSection(heading="Title", content="Test", order=0),
            ],
        )
        strategy = ChunkingStrategy()
        result = await chunker.chunk(doc, strategy, {"parser_version": "1.0.0"})

        assert len(result) == 1
        assert result[0].provenance.parser_version == "1.0.0"


class TestChunkingRegistry:
    def test_get_chunker_returns_structural(self) -> None:
        registry = ChunkingRegistry()
        chunker = registry.get_chunker("structural")
        assert isinstance(chunker, StructuralChunkingAdapter)

    def test_get_chunker_returns_none_for_unknown(self) -> None:
        registry = ChunkingRegistry()
        chunker = registry.get_chunker("unknown")
        assert chunker is None

    @pytest.mark.asyncio
    async def test_chunk_delegates_to_chunker(self) -> None:
        registry = ChunkingRegistry()
        doc = ParsedDocument(
            document_text="Test content",
            sections=[
                ParsedDocumentSection(heading="Test", content="Test content", order=0),
            ],
        )
        strategy = ChunkingStrategy()
        result = await registry.chunk("structural", doc, strategy)

        assert len(result) == 1
        assert result[0].chunk_text == "Test content"

    @pytest.mark.asyncio
    async def test_chunk_raises_for_unknown_strategy(self) -> None:
        registry = ChunkingRegistry()
        doc = ParsedDocument(document_text="", sections=[])
        strategy = ChunkingStrategy()

        with pytest.raises(ValueError, match="No chunker found"):
            await registry.chunk("unknown", doc, strategy)


class TestChunkingStrategy:
    def test_default_values(self) -> None:
        strategy = ChunkingStrategy()
        assert strategy.max_tokens == _DEFAULT_MAX_TOKENS
        assert strategy.overlap_tokens == _DEFAULT_OVERLAP_TOKENS
        assert strategy.overlap_by_section is True
        assert strategy.min_chunk_tokens == _DEFAULT_MIN_CHUNK_TOKENS

    def test_custom_values(self) -> None:
        strategy = ChunkingStrategy(
            max_tokens=_CUSTOM_MAX_TOKENS, overlap_tokens=_CUSTOM_OVERLAP_TOKENS
        )
        assert strategy.max_tokens == _CUSTOM_MAX_TOKENS
        assert strategy.overlap_tokens == _CUSTOM_OVERLAP_TOKENS


class TestChunkProvenance:
    def test_creation(self) -> None:
        provenance = ChunkProvenance(
            section_heading="Test",
            section_order=1,
            chunk_index=0,
            parser_version="1.0",
        )
        assert provenance.section_heading == "Test"
        assert provenance.section_order == 1
        assert provenance.chunk_index == 0
        assert provenance.parser_version == "1.0"

    def test_optional_fields(self) -> None:
        provenance = ChunkProvenance(section_order=0, chunk_index=0)
        assert provenance.section_heading is None
        assert provenance.parser_version is None
