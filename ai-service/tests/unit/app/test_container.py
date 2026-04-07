from __future__ import annotations

import pytest
from sqlalchemy.ext.asyncio import AsyncSession

from app.container import Container
from infrastructure.database.repositories import (
    SqlAlchemyConversationRepository,
    SqlAlchemyKnowledgeRepository,
    SqlAlchemyQuotaRepository,
)


@pytest.mark.asyncio
async def test_container_wires_database_repositories() -> None:
    container = Container()

    quota_repository = container.quota_repository()
    conversation_repository = container.conversation_repository()
    knowledge_repository = container.knowledge_repository()

    assert isinstance(quota_repository, SqlAlchemyQuotaRepository)
    assert isinstance(conversation_repository, SqlAlchemyConversationRepository)
    assert isinstance(knowledge_repository, SqlAlchemyKnowledgeRepository)
    assert isinstance(quota_repository.session, AsyncSession)
    assert isinstance(conversation_repository.session, AsyncSession)
    assert isinstance(knowledge_repository.session, AsyncSession)

    await quota_repository.session.close()
    await conversation_repository.session.close()
    await knowledge_repository.session.close()
