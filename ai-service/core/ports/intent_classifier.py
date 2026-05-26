from __future__ import annotations

from abc import ABC, abstractmethod
from enum import StrEnum


class IntentClass(StrEnum):
    KNOWLEDGE = "knowledge"
    PERSONAL = "personal"
    MIXED = "mixed"


class IntentClassifierPort(ABC):
    @abstractmethod
    async def classify(self, query_embedding: list[float]) -> IntentClass: ...

    @abstractmethod
    async def initialize(self) -> None: ...
