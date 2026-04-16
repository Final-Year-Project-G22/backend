from __future__ import annotations

import json
import logging
from typing import Any, cast

from core.ports.parser import ParsedDocument, ParsedDocumentSection, ParserPort

logger = logging.getLogger(__name__)

_MAX_DICT_PREVIEW_KEYS = 10
_MAX_LIST_PREVIEW_ITEMS = 5


class JsonParser(ParserPort):
    CONTENT_TYPES = frozenset(
        [
            "application/json",
            "application/json; charset=utf-8",
        ]
    )

    def supports(self, content_type: str) -> bool:
        return content_type.lower() in self.CONTENT_TYPES

    async def parse(self, content: bytes, metadata: dict[str, Any] | None = None) -> ParsedDocument:
        text = content.decode("utf-8", errors="replace")

        sections: list[ParsedDocumentSection] = []
        section_order = 0

        try:
            data = json.loads(text)
        except json.JSONDecodeError as exc:
            logger.warning("JSON parsing error: %s", exc)
            return ParsedDocument(
                document_text=text.strip(),
                sections=[
                    ParsedDocumentSection(
                        heading="Raw Content",
                        content=text.strip(),
                        order=0,
                    )
                ],
                metadata=metadata or {},
            )

        if isinstance(data, dict):
            dict_data = cast(dict[str, Any], data)
            key_count = len(dict_data)
            preview_items = list(dict_data.items())[:_MAX_DICT_PREVIEW_KEYS]
            preview_text = "\n".join(f"{k}: {v}" for k, v in preview_items)

            if key_count > _MAX_DICT_PREVIEW_KEYS:
                preview_text += f"\n... and {key_count - _MAX_DICT_PREVIEW_KEYS} more keys"

            sections.append(
                ParsedDocumentSection(
                    heading=f"Object ({key_count} keys)",
                    content=preview_text,
                    order=section_order,
                )
            )
        elif isinstance(data, list):
            list_data = cast(list[Any], data)
            item_count = len(list_data)
            preview_items = list_data[:_MAX_LIST_PREVIEW_ITEMS]
            preview_text = "\n".join(str(item) for item in preview_items)

            if item_count > _MAX_LIST_PREVIEW_ITEMS:
                preview_text += f"\n... and {item_count - _MAX_LIST_PREVIEW_ITEMS} more items"

            sections.append(
                ParsedDocumentSection(
                    heading=f"Array ({item_count} items)",
                    content=preview_text,
                    order=section_order,
                )
            )

        if not sections:
            sections.append(
                ParsedDocumentSection(
                    heading="Content",
                    content=text.strip(),
                    order=0,
                )
            )

        formatted_json = json.dumps(data, ensure_ascii=False, indent=2)

        return ParsedDocument(
            document_text=formatted_json.strip(),
            sections=sections,
            metadata=metadata or {},
        )


__all__ = ["JsonParser"]
