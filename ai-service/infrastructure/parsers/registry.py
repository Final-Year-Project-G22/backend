from __future__ import annotations

import logging

from core.ports.parser import ParserPort
from infrastructure.parsers.csv import CsvParser
from infrastructure.parsers.docx import DocxParser
from infrastructure.parsers.html import HtmlParser
from infrastructure.parsers.json import JsonParser
from infrastructure.parsers.markdown import MarkdownParser
from infrastructure.parsers.pdf import PdfParser
from infrastructure.parsers.plain import TextPlainParser
from infrastructure.parsers.xml import XmlParser

logger = logging.getLogger(__name__)


class ParserRegistry:
    def __init__(self) -> None:
        self._parsers: list[ParserPort] = [
            PdfParser(),
            DocxParser(),
            MarkdownParser(),
            HtmlParser(),
            CsvParser(),
            JsonParser(),
            XmlParser(),
            TextPlainParser(),
        ]

    def supports(self, content_type: str) -> bool:
        return any(parser.supports(content_type) for parser in self._parsers)

    def get_parser(self, content_type: str) -> ParserPort | None:
        for parser in self._parsers:
            if parser.supports(content_type):
                logger.debug("Found parser for content_type: %s", content_type)
                return parser
        logger.warning("No parser found for content_type: %s", content_type)
        return None


__all__ = ["ParserRegistry"]
