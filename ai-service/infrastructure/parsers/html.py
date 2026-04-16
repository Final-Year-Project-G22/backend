from __future__ import annotations

import logging
from typing import Any

from bs4 import BeautifulSoup

from core.ports.parser import ParsedDocument, ParsedDocumentSection, ParserPort

logger = logging.getLogger(__name__)


class HtmlParser(ParserPort):
    CONTENT_TYPES = frozenset(
        [
            "text/html",
            "application/xhtml+xml",
            "text/html; charset=utf-8",
            "text/html;charset=utf-8",
        ]
    )

    HEADING_TAGS = frozenset({"h1", "h2", "h3", "h4", "h5", "h6"})
    BLOCK_TAGS = frozenset({"p", "div", "section", "article", "blockquote", "li", "tr"})

    def supports(self, content_type: str) -> bool:
        return content_type.lower() in self.CONTENT_TYPES

    async def parse(self, content: bytes, metadata: dict[str, Any] | None = None) -> ParsedDocument:
        html_content = content.decode("utf-8", errors="replace")

        soup = BeautifulSoup(html_content, "lxml")

        for script in soup(["script", "style"]):
            script.decompose()

        text = soup.get_text(separator="\n")

        sections: list[ParsedDocumentSection] = []
        current_heading: str | None = None
        current_content: list[str] = []
        section_order = 0

        for element in soup.find_all(
            ["h1", "h2", "h3", "h4", "h5", "h6", "p", "div", "section", "article", "li"]
        ):
            tag = element.name

            if tag in self.HEADING_TAGS:
                if current_content:
                    text_content = element.get_text(separator=" ", strip=True)
                    if text_content:
                        sections.append(
                            ParsedDocumentSection(
                                heading=current_heading,
                                content="\n".join(current_content).strip(),
                                order=section_order,
                            )
                        )
                        section_order += 1
                        current_content = []

                current_heading = element.get_text(strip=True)
                current_content = []
            elif current_heading is not None or current_content:
                text_content = element.get_text(separator=" ", strip=True)
                if text_content:
                    current_content.append(text_content)

        if current_content:
            sections.append(
                ParsedDocumentSection(
                    heading=current_heading,
                    content="\n".join(current_content).strip(),
                    order=section_order,
                )
            )

        if not sections:
            cleaned_text = "\n".join(line.strip() for line in text.split("\n") if line.strip())
            sections.append(
                ParsedDocumentSection(
                    heading=None,
                    content=cleaned_text,
                    order=0,
                )
            )

        return ParsedDocument(
            document_text=text.strip(),
            sections=sections,
            metadata=metadata or {},
        )


__all__ = ["HtmlParser"]
