from __future__ import annotations

from enum import StrEnum


class Language(StrEnum):
    ENGLISH = "en"
    AMHARIC = "am"


class DocumentSource(StrEnum):
    GOVERNMENT = "government"
    LEGAL = "legal"
    TAX_CODE = "tax_code"
    GUIDE = "guide"
    STEP = "step"
    FAQ = "faq"


class DocumentStatus(StrEnum):
    PROCESSING = "processing"
    ACTIVE = "active"
    ARCHIVED = "archived"
    FAILED = "failed"


class ChunkStatus(StrEnum):
    PENDING = "pending"
    EMBEDDED = "embedded"
    FAILED = "failed"


class Tier(StrEnum):
    BASIC = "basic"
    PRO = "pro"
    PREMIUM = "premium"


class SessionStatus(StrEnum):
    ACTIVE = "active"
    ARCHIVED = "archived"


class MessageType(StrEnum):
    USER_QUERY = "user_query"
    AI_RESPONSE = "ai_response"


__all__ = [
    "ChunkStatus",
    "DocumentSource",
    "DocumentStatus",
    "Language",
    "MessageType",
    "SessionStatus",
    "Tier",
]
