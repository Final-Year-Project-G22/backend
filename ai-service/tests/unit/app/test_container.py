from __future__ import annotations

import pytest
from sqlalchemy.ext.asyncio import AsyncSession

from app.container import Container
from infrastructure.database.repositories import (
    SqlAlchemyConversationRepository,
    SqlAlchemyQuotaRepository,
)


@pytest.mark.asyncio
async def test_container_wires_database_repositories() -> None:
    container = Container()

    quota_repository = container.quota_repository()
    conversation_repository = container.conversation_repository()

    assert isinstance(quota_repository, SqlAlchemyQuotaRepository)
    assert isinstance(conversation_repository, SqlAlchemyConversationRepository)
    assert isinstance(quota_repository._session, AsyncSession)
    assert isinstance(conversation_repository._session, AsyncSession)

    await quota_repository._session.close()
    await conversation_repository._session.close()
