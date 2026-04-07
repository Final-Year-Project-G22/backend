from __future__ import annotations

from abc import ABC, abstractmethod
from collections.abc import Awaitable, Callable
from typing import Any

EventPayload = dict[str, Any]
EventHandler = Callable[[EventPayload], Awaitable[None]]


class EventBusPort(ABC):
    @abstractmethod
    async def publish(self, topic: str, payload: EventPayload) -> None: ...

    @abstractmethod
    async def subscribe(self, topic: str, handler: EventHandler) -> None: ...


__all__ = ["EventBusPort", "EventHandler", "EventPayload"]
