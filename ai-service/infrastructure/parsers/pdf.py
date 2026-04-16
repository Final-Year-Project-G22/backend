from __future__ import annotations

import logging
from typing import Any

import pymupdf  # type: ignore[import]

from core.ports.parser import ParsedDocument, ParsedDocumentSection, ParserPort

logger = logging.getLogger(__name__)


class _PdfPage:
    def __init__(self, page: Any) -> None:
        self._page = page

    def get_text(self) -> str:
        result = self._page.get_text()
        return result if isinstance(result, str) else ""


class _PdfDocument:
    def __init__(self, doc: Any) -> None:
        self._doc = doc

    @property
    def page_count(self) -> int:
        return self._doc.page_count

    def load_page(self, page_num: int) -> _PdfPage:
        return _PdfPage(self._doc.load_page(page_num))

    def close(self) -> None:
        self._doc.close()


def _open_pdf(stream: bytes, filetype: str) -> _PdfDocument:
    doc = pymupdf.open(stream=stream, filetype=filetype)  # type: ignore[no-untyped-call]
    return _PdfDocument(doc)


class PdfParser(ParserPort):
    CONTENT_TYPES = frozenset(
        [
            "application/pdf",
            "application/pdf; charset=utf-8",
        ]
    )

    def supports(self, content_type: str) -> bool:
        return content_type.lower() in self.CONTENT_TYPES

    async def parse(self, content: bytes, metadata: dict[str, Any] | None = None) -> ParsedDocument:
        try:
            doc = _open_pdf(content, "pdf")
        except Exception:
            logger.exception("Failed to open PDF")
            raise

        sections: list[ParsedDocumentSection] = []
        text_parts: list[str] = []
        section_order = 0

        page_count = doc.page_count
        for page_num in range(page_count):
            page = doc.load_page(page_num)
            text = page.get_text()
            if text.strip():
                text_parts.append(text)
                sections.append(
                    ParsedDocumentSection(
                        heading=f"Page {page_num + 1}",
                        content=text.strip(),
                        order=section_order,
                    )
                )
                section_order += 1

        doc.close()

        full_text = "\n".join(text_parts)

        if not sections:
            sections.append(
                ParsedDocumentSection(
                    heading=None,
                    content=full_text.strip() if full_text else "",
                    order=0,
                )
            )

        return ParsedDocument(
            document_text=full_text.strip(),
            sections=sections,
            metadata=metadata or {},
        )


__all__ = ["PdfParser"]
