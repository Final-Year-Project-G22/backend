from __future__ import annotations

import logging
from typing import Any

from core.ports.parser import ParsedDocument, ParsedDocumentSection, ParserPort

logger = logging.getLogger(__name__)


class TextPlainParser(ParserPort):
    CONTENT_TYPES = frozenset(
        ["text/plain", "text/plain; charset=utf-8", "text/plain;charset=utf-8"]
    )
    UPPERCASE_HEADING_MAX_LENGTH = 80
    COLON_HEADING_MAX_LENGTH = 60

    def supports(self, content_type: str) -> bool:
        return content_type.lower() in self.CONTENT_TYPES

    async def parse(self, content: bytes, metadata: dict[str, Any] | None = None) -> ParsedDocument:
        text = content.decode("utf-8", errors="replace")

        lines = text.split("\n")
        sections: list[ParsedDocumentSection] = []
        current_section_content: list[str] = []
        section_order = 0

        for line in lines:
            stripped = line.strip()
            if not stripped:
                if current_section_content:
                    current_section_content.append(line)
                continue

            if self._is_heading_line(stripped):
                if current_section_content:
                    sections.append(
                        ParsedDocumentSection(
                            heading=None,
                            content="\n".join(current_section_content).strip(),
                            order=section_order,
                        )
                    )
                    section_order += 1
                    current_section_content = []

                sections.append(
                    ParsedDocumentSection(
                        heading=stripped,
                        content="",
                        order=section_order,
                    )
                )
                section_order += 1
            else:
                current_section_content.append(line)

        if current_section_content:
            sections.append(
                ParsedDocumentSection(
                    heading=None,
                    content="\n".join(current_section_content).strip(),
                    order=section_order,
                )
            )

        if not sections:
            sections.append(
                ParsedDocumentSection(
                    heading=None,
                    content=text.strip(),
                    order=0,
                )
            )

        return ParsedDocument(
            document_text=text.strip(),
            sections=sections,
            metadata=metadata or {},
        )

    def _is_heading_line(self, line: str) -> bool:
        if line.startswith("#"):
            return True
        if line.isupper() and len(line) < self.UPPERCASE_HEADING_MAX_LENGTH:
            return True
        return bool(line.endswith(":") and len(line) < self.COLON_HEADING_MAX_LENGTH)


__all__ = ["TextPlainParser"]
