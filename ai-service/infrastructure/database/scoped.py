from __future__ import annotations

from collections.abc import Awaitable, Callable
from typing import Any, Literal, cast

from core.domain.models import AIChatMessage, AIConversationSession
from core.ports.conversation_repository import ConversationRepositoryPort
from core.ports.core_service import CoreServicePort
from core.usecases.contracts import CreateSessionCommand
from core.usecases.conversation import ConversationUseCase
from core.usecases.quota_guard import QuotaGuardUseCase
from infrastructure.database.connection import Database

RepositoryName = Literal["conversation", "quota", "knowledge", "ledger"]


class ScopedRepository:
    """Open a short transaction per repository operation; never retain a session."""

    def __init__(self, database: Database, name: RepositoryName) -> None:
        self._database = database
        self._name = name

    def __getattr__(self, method: str) -> Callable[..., Awaitable[Any]]:
        async def invoke(*args: Any, **kwargs: Any) -> Any:
            async with self._database.command() as command:
                repository = getattr(command, self._name)
                operation = cast(Callable[..., Awaitable[Any]], getattr(repository, method))
                return await operation(*args, **kwargs)

        return invoke


class ScopedConversation(ConversationUseCase):
    """Compose quota and conversation creation in one owned database command."""

    def __init__(self, database: Database, core_service: CoreServicePort | None) -> None:
        super().__init__(
            cast(ConversationRepositoryPort, ScopedRepository(database, "conversation"))
        )
        self._database = database
        self._core_service = core_service

    async def append_message(self, message: AIChatMessage) -> AIChatMessage:
        async with self._database.command() as repositories:
            return await repositories.conversation.add_message_in_order(message)

    async def create_session(
        self, command: CreateSessionCommand, *, at: Any = None
    ) -> AIConversationSession:
        async with self._database.command() as repositories:
            usecase = ConversationUseCase(
                repositories.conversation,
                quota_guard=QuotaGuardUseCase(repositories.quota, core_service=self._core_service),
            )
            return await usecase.create_session(command, at=at)
