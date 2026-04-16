from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Any

from pydantic import BaseModel, Field


class ParsedDocumentSection(BaseModel):
    heading: str | None = None
    content: str
    order: int


class ParsedDocument(BaseModel):
    document_text: str
    sections: list[ParsedDocumentSection]
    metadata: dict[str, Any] = Field(default_factory=dict)


class ParserPort(ABC):
    @abstractmethod
    def supports(self, content_type: str) -> bool:
        raise NotImplementedError

    @abstractmethod
    async def parse(self, content: bytes, metadata: dict[str, Any] | None = None) -> ParsedDocument:
        raise NotImplementedError


__all__ = ["ParsedDocument", "ParsedDocumentSection", "ParserPort"]
