from __future__ import annotations

import os
import uuid
from collections.abc import AsyncIterator

import asyncpg
import pytest
from sqlalchemy.engine import make_url
from sqlalchemy.schema import CreateTable

from infrastructure.database import models_sqlalchemy  # noqa: F401
from infrastructure.database.connection import Base, Database


@pytest.fixture
async def database() -> AsyncIterator[Database]:
    raw_url = os.environ.get("AI_TEST_DATABASE_URL")
    if raw_url is None:
        pytest.skip("AI_TEST_DATABASE_URL requires PostgreSQL with pgvector")
    url = make_url(raw_url)
    schema = f"ai247_{uuid.uuid4().hex}"
    admin = await asyncpg.connect(
        url.set(drivername="postgresql").render_as_string(hide_password=False)
    )
    try:
        await admin.execute("CREATE EXTENSION IF NOT EXISTS vector")
        await admin.execute(f'CREATE SCHEMA "{schema}"')
        db = Database(
            raw_url, connect_args={"server_settings": {"search_path": f"{schema},public"}}
        )
        try:
            async with db.engine.begin() as connection:
                for table in Base.metadata.sorted_tables:
                    await connection.execute(CreateTable(table))
            yield db
        finally:
            await db.engine.dispose()
            await admin.execute(f'DROP SCHEMA "{schema}" CASCADE')
    finally:
        await admin.close()
