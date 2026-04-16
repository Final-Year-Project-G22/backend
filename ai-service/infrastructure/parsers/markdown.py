from __future__ import annotations

import logging
import re
from typing import Any

from core.ports.parser import ParsedDocument, ParsedDocumentSection, ParserPort

logger = logging.getLogger(__name__)


class MarkdownParser(ParserPort):
    CONTENT_TYPES = frozenset(
        [
            "text/markdown",
            "text/x-markdown",
            "text/markdown; charset=utf-8",
            "text/markdown;charset=utf-8",
        ]
    )

    HEADING_PATTERN = re.compile(r"^(#{1,6})\s+(.+)$", re.MULTILINE)

    def supports(self, content_type: str) -> bool:
        return content_type.lower() in self.CONTENT_TYPES

    async def parse(self, content: bytes, metadata: dict[str, Any] | None = None) -> ParsedDocument:
        text = content.decode("utf-8", errors="replace")

        matches = list(self.HEADING_PATTERN.finditer(text))

        sections: list[ParsedDocumentSection] = []

        if not matches:
            sections.append(
                ParsedDocumentSection(
                    heading=None,
                    content=text.strip(),
                    order=0,
                )
            )
        else:
            for idx, match in enumerate(matches):
                heading = match.group(2).strip()
                start = match.end()
                end = matches[idx + 1].start() if idx + 1 < len(matches) else len(text)
                section_content = text[start:end].strip()

                sections.append(
                    ParsedDocumentSection(
                        heading=heading,
                        content=section_content,
                        order=idx,
                    )
                )

        return ParsedDocument(
            document_text=text.strip(),
            sections=sections,
            metadata=metadata or {},
        )


__all__ = ["MarkdownParser"]
