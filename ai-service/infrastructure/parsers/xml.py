from __future__ import annotations

import importlib
import logging
from typing import Any, Protocol, cast

from core.ports.parser import ParsedDocument, ParsedDocumentSection, ParserPort

logger = logging.getLogger(__name__)


class _EtreeProtocol(Protocol):
    def fromstring(self, text: bytes) -> Any: ...
    def tostring(self, element: Any, *, pretty_print: bool, encoding: str) -> str: ...  # noqa: V107


def _load_etree() -> _EtreeProtocol:
    module = importlib.import_module("lxml.etree")
    return cast(_EtreeProtocol, module)


class _XmlElement:
    def __init__(self, elem: Any) -> None:
        self._elem = elem

    @property
    def tag(self) -> str:
        return str(self._elem.tag)

    @property
    def attrib(self) -> dict[str, str]:
        return {str(k): str(v) for k, v in dict(self._elem.attrib).items()}

    @property
    def text(self) -> str | None:
        value = self._elem.text
        return None if value is None else str(value)

    def iter_children(self) -> list[_XmlElement]:
        return [_XmlElement(child) for child in self._elem]


class _XmlDocument:
    def __init__(self, root: Any) -> None:
        self._root = root

    @property
    def root(self) -> _XmlElement:
        return _XmlElement(self._root)

    def get_raw_root(self) -> Any:
        return self._root


def _parse_xml(content: bytes) -> _XmlDocument:
    etree = _load_etree()
    root = etree.fromstring(content)
    return _XmlDocument(root)


def _serialize_xml(doc: _XmlDocument) -> str:
    etree = _load_etree()
    raw = doc.get_raw_root()
    return etree.tostring(raw, pretty_print=True, encoding="unicode")


class XmlParser(ParserPort):
    CONTENT_TYPES = frozenset(
        [
            "application/xml",
            "text/xml",
            "application/xml; charset=utf-8",
            "text/xml; charset=utf-8",
        ]
    )

    def supports(self, content_type: str) -> bool:
        return content_type.lower() in self.CONTENT_TYPES

    async def parse(self, content: bytes, metadata: dict[str, Any] | None = None) -> ParsedDocument:
        text = content.decode("utf-8", errors="replace")

        sections: list[ParsedDocumentSection] = []
        section_order = 0

        try:
            doc = _parse_xml(content)
            root_elem = doc.root
            root_tag = root_elem.tag
            pretty_xml = _serialize_xml(doc)
        except Exception as exc:
            logger.warning("XML parsing error: %s, falling back to plain text", exc)
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

        def describe_element(elem: _XmlElement, depth: int = 0) -> list[str]:
            result: list[str] = []
            indent = "  " * depth

            tag = elem.tag
            attributes = " ".join(f'@{k}="{v}"' for k, v in elem.attrib.items())
            if attributes:
                result.append(f"{indent}<{tag} {attributes}>")
            else:
                result.append(f"{indent}<{tag}>")

            if elem.text and elem.text.strip():
                result.append(f"{indent}  {elem.text.strip()}")

            for child in elem.iter_children():
                result.extend(describe_element(child, depth + 1))

            result.append(f"{indent}</{tag}>")
            return result

        element_description = describe_element(root_elem)

        max_lines = 100
        if len(element_description) > max_lines:
            content_text = "\n".join(element_description[:max_lines])
            content_text += f"\n... and {len(element_description) - max_lines} more lines"
        else:
            content_text = "\n".join(element_description)

        sections.append(
            ParsedDocumentSection(
                heading=f"Root: <{root_tag}>",
                content=content_text,
                order=section_order,
            )
        )

        return ParsedDocument(
            document_text=pretty_xml.strip(),
            sections=sections,
            metadata=metadata or {},
        )


__all__ = ["XmlParser"]
