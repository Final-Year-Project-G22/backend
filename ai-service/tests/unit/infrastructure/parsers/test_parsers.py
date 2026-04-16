from __future__ import annotations

import pytest

from infrastructure.parsers import (
    CsvParser,
    DocxParser,
    HtmlParser,
    JsonParser,
    MarkdownParser,
    ParserRegistry,
    PdfParser,
    TextPlainParser,
    XmlParser,
)

_EXPECTED_HEADING_COUNT = 2


class TestTextPlainParser:
    def test_supports_returns_true_for_plain_text(self) -> None:
        parser = TextPlainParser()
        assert parser.supports("text/plain") is True
        assert parser.supports("text/plain; charset=utf-8") is True
        assert parser.supports("text/plain;charset=utf-8") is True

    def test_supports_returns_false_for_non_plain_text(self) -> None:
        parser = TextPlainParser()
        assert parser.supports("application/json") is False
        assert parser.supports("text/html") is False

    @pytest.mark.asyncio
    async def test_parse_simple_text(self) -> None:
        parser = TextPlainParser()
        content = b"Hello world\nThis is a test"
        result = await parser.parse(content)

        assert result.document_text == "Hello world\nThis is a test"
        assert len(result.sections) > 0
        assert result.sections[0].content == "Hello world\nThis is a test"

    @pytest.mark.asyncio
    async def test_parse_empty_content(self) -> None:
        parser = TextPlainParser()
        content = b""
        result = await parser.parse(content)

        assert result.document_text == ""
        assert len(result.sections) == 1
        assert result.sections[0].content == ""

    @pytest.mark.asyncio
    async def test_parse_with_metadata(self) -> None:
        parser = TextPlainParser()
        content = b"Test content"
        metadata = {"source": "test.pdf", "filename": "test.txt"}
        result = await parser.parse(content, metadata)

        assert result.metadata == metadata


class TestMarkdownParser:
    def test_supports_returns_true_for_markdown(self) -> None:
        parser = MarkdownParser()
        assert parser.supports("text/markdown") is True
        assert parser.supports("text/x-markdown") is True

    def test_supports_returns_false_for_non_markdown(self) -> None:
        parser = MarkdownParser()
        assert parser.supports("text/plain") is False

    @pytest.mark.asyncio
    async def test_parse_markdown_with_headings(self) -> None:
        parser = MarkdownParser()
        content = b"# Heading 1\n\nSome content\n\n## Heading 2\n\nMore content"
        result = await parser.parse(content)

        assert result.document_text.strip().startswith("# Heading 1")
        assert len(result.sections) == _EXPECTED_HEADING_COUNT
        assert result.sections[0].heading == "Heading 1"
        assert result.sections[1].heading == "Heading 2"

    @pytest.mark.asyncio
    async def test_parse_markdown_without_headings(self) -> None:
        parser = MarkdownParser()
        content = b"Just plain markdown content without headings"
        result = await parser.parse(content)

        assert len(result.sections) == 1
        assert result.sections[0].heading is None


class TestHtmlParser:
    def test_supports_returns_true_for_html(self) -> None:
        parser = HtmlParser()
        assert parser.supports("text/html") is True
        assert parser.supports("application/xhtml+xml") is True

    def test_supports_returns_false_for_non_html(self) -> None:
        parser = HtmlParser()
        assert parser.supports("text/plain") is False

    @pytest.mark.asyncio
    async def test_parse_html_with_headings(self) -> None:
        parser = HtmlParser()
        content = b"<html><h1>Title</h1><p>Paragraph content</p></html>"
        result = await parser.parse(content)

        assert "Title" in result.document_text
        assert "Paragraph content" in result.document_text
        assert len(result.sections) >= 1


class TestCsvParser:
    def test_supports_returns_true_for_csv(self) -> None:
        parser = CsvParser()
        assert parser.supports("text/csv") is True
        assert parser.supports("application/csv") is True

    def test_supports_returns_false_for_non_csv(self) -> None:
        parser = CsvParser()
        assert parser.supports("text/plain") is False

    @pytest.mark.asyncio
    async def test_parse_csv_with_header(self) -> None:
        parser = CsvParser()
        content = b"name,age,city\nJohn,30,NYC\nJane,25,LA"
        result = await parser.parse(content)

        assert "name" in result.document_text
        assert "age" in result.document_text
        assert len(result.sections) >= 1
        assert result.sections[0].heading == "Header"

    @pytest.mark.asyncio
    async def test_parse_empty_csv(self) -> None:
        parser = CsvParser()
        content = b""
        result = await parser.parse(content)

        assert result.document_text == ""


class TestJsonParser:
    def test_supports_returns_true_for_json(self) -> None:
        parser = JsonParser()
        assert parser.supports("application/json") is True

    def test_supports_returns_false_for_non_json(self) -> None:
        parser = JsonParser()
        assert parser.supports("text/plain") is False

    @pytest.mark.asyncio
    async def test_parse_json_object(self) -> None:
        parser = JsonParser()
        content = b'{"name": "test", "value": 123}'
        result = await parser.parse(content)

        assert "name" in result.document_text
        assert "test" in result.document_text
        assert len(result.sections) >= 1

    @pytest.mark.asyncio
    async def test_parse_json_array(self) -> None:
        parser = JsonParser()
        content = b'[{"id": 1}, {"id": 2}, {"id": 3}]'
        result = await parser.parse(content)

        assert result.sections[0].heading is not None
        assert "Array" in result.sections[0].heading
        assert len(result.sections) >= 1


class TestXmlParser:
    def test_supports_returns_true_for_xml(self) -> None:
        parser = XmlParser()
        assert parser.supports("application/xml") is True
        assert parser.supports("text/xml") is True

    def test_supports_returns_false_for_non_xml(self) -> None:
        parser = XmlParser()
        assert parser.supports("text/plain") is False

    @pytest.mark.asyncio
    async def test_parse_simple_xml(self) -> None:
        parser = XmlParser()
        content = b"<root><child>text</child></root>"
        result = await parser.parse(content)

        assert "root" in result.document_text
        assert "child" in result.document_text
        assert len(result.sections) >= 1


class TestParserRegistry:
    def test_get_parser_returns_correct_parser_for_pdf(self) -> None:
        registry = ParserRegistry()
        parser = registry.get_parser("application/pdf")
        assert isinstance(parser, PdfParser)

    def test_get_parser_returns_correct_parser_for_docx(self) -> None:
        registry = ParserRegistry()
        parser = registry.get_parser(
            "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
        )
        assert isinstance(parser, DocxParser)

    def test_get_parser_returns_correct_parser_for_markdown(self) -> None:
        registry = ParserRegistry()
        parser = registry.get_parser("text/markdown")
        assert isinstance(parser, MarkdownParser)

    def test_get_parser_returns_correct_parser_for_html(self) -> None:
        registry = ParserRegistry()
        parser = registry.get_parser("text/html")
        assert isinstance(parser, HtmlParser)

    def test_get_parser_returns_correct_parser_for_csv(self) -> None:
        registry = ParserRegistry()
        parser = registry.get_parser("text/csv")
        assert isinstance(parser, CsvParser)

    def test_get_parser_returns_correct_parser_for_json(self) -> None:
        registry = ParserRegistry()
        parser = registry.get_parser("application/json")
        assert isinstance(parser, JsonParser)

    def test_get_parser_returns_correct_parser_for_xml(self) -> None:
        registry = ParserRegistry()
        parser = registry.get_parser("application/xml")
        assert isinstance(parser, XmlParser)

    def test_get_parser_returns_correct_parser_for_plain_text(self) -> None:
        registry = ParserRegistry()
        parser = registry.get_parser("text/plain")
        assert isinstance(parser, TextPlainParser)

    def test_get_parser_returns_none_for_unknown_type(self) -> None:
        registry = ParserRegistry()
        parser = registry.get_parser("application/octet-stream")
        assert parser is None

    def test_supports_returns_true_for_supported_types(self) -> None:
        registry = ParserRegistry()
        assert registry.supports("application/pdf") is True
        assert registry.supports("text/plain") is True
        assert registry.supports("application/json") is True

    def test_supports_returns_false_for_unsupported_types(self) -> None:
        registry = ParserRegistry()
        assert registry.supports("application/octet-stream") is False
        assert registry.supports("image/png") is False
