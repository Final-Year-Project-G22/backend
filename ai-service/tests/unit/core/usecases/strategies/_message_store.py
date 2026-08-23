from __future__ import annotations

import uuid

from core.domain.models import AIChatMessage


class _MessageStore:
    """Stateful stand-in for the conversation repository's message seam."""

    def __init__(self, *existing: AIChatMessage) -> None:
        self.messages: list[AIChatMessage] = list(existing)

    async def list(
        self,
        _conversation_id: uuid.UUID,
        *,
        limit: int = 100,
        offset: int = 0,
        descending: bool = False,
    ) -> list[AIChatMessage]:
        ordered = sorted(
            self.messages,
            key=lambda m: (m.message_order, m.created_at, m.id),
            reverse=descending,
        )
        if limit <= 0:
            return ordered[offset:]
        return ordered[offset : offset + limit]

    async def add(self, message: AIChatMessage) -> AIChatMessage:
        self.messages.append(message)
        return message

    async def update(self, message: AIChatMessage) -> AIChatMessage:
        return message
