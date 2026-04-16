from __future__ import annotations

import csv
import logging
from io import StringIO
from typing import Any

from core.ports.parser import ParsedDocument, ParsedDocumentSection, ParserPort

logger = logging.getLogger(__name__)


class CsvParser(ParserPort):
    CONTENT_TYPES = frozenset(
        [
            "text/csv",
            "text/csv; charset=utf-8",
            "application/csv",
            "application/csv; charset=utf-8",
        ]
    )

    def supports(self, content_type: str) -> bool:
        return content_type.lower() in self.CONTENT_TYPES

    async def parse(self, content: bytes, metadata: dict[str, Any] | None = None) -> ParsedDocument:
        text = content.decode("utf-8", errors="replace")

        sections: list[ParsedDocumentSection] = []
        section_order = 0

        try:
            reader = csv.reader(StringIO(text))
            rows = list(reader)

            if not rows:
                return ParsedDocument(
                    document_text="",
                    sections=[ParsedDocumentSection(heading=None, content="", order=0)],
                    metadata=metadata or {},
                )

            if rows:
                header = rows[0]
                header_text = ", ".join(header) if header else "CSV Data"
                sections.append(
                    ParsedDocumentSection(
                        heading="Header",
                        content=header_text,
                        order=section_order,
                    )
                )
                section_order += 1

                data_rows = rows[1:] if len(rows) > 1 else []
                if data_rows:
                    row_count = len(data_rows)
                    sample_size = min(5, row_count)
                    sample_rows = data_rows[:sample_size]
                    sample_text = "\n".join(", ".join(row) for row in sample_rows)
                    suffix = (
                        f"\n... and {row_count - sample_size} more rows"
                        if row_count > sample_size
                        else ""
                    )

                    sections.append(
                        ParsedDocumentSection(
                            heading=f"Data ({row_count} rows)",
                            content=sample_text + suffix,
                            order=section_order,
                        )
                    )
                    section_order += 1

            full_text = "\n".join(", ".join(row) for row in rows)

            return ParsedDocument(
                document_text=full_text.strip(),
                sections=sections,
                metadata=metadata or {},
            )

        except csv.Error as exc:
            logger.warning("CSV parsing error: %s, falling back to plain text", exc)
            return ParsedDocument(
                document_text=text.strip(),
                sections=[
                    ParsedDocumentSection(
                        heading="Content",
                        content=text.strip(),
                        order=0,
                    )
                ],
                metadata=metadata or {},
            )


__all__ = ["CsvParser"]
