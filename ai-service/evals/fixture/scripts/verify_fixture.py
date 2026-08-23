"""Verify the evaluation fixture is provisioned and print verification facts.

Checks (all must pass for the fixture to be valid):
  1. All 12 manifest documents exist in knowledge_documents with the
     documented source/language/external_id and are ACTIVE.
  2. Every document has exactly the chunk count recorded in chunk_refs.json
     (deterministic — the chunker must not have changed).
  3. Fixture account exists in the core-backend DB with the documented
     profile, compliance entries, and guide progress.

Usage (from ai-service/):
    uv run python evals/fixture/scripts/verify_fixture.py
    CORE_DATABASE_URL=postgresql://... uv run python evals/fixture/scripts/verify_fixture.py
"""

from __future__ import annotations

import asyncio
import json
import os
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[3]))

from sqlalchemy import text

from infrastructure.database.connection import async_session_factory

FIXTURE_DIR = Path(__file__).resolve().parents[1]
_failures: list[str] = []


def check(label: str, ok: bool, detail: str = "") -> None:
    if not ok:
        _failures.append(f"{label}: {detail}")
    print(f"{'PASS' if ok else 'FAIL'}  {label}" + (f" — {detail}" if detail else ""))


async def verify_ai_db() -> None:
    manifest = json.loads((FIXTURE_DIR / "manifest.json").read_text())
    chunk_refs = json.loads((FIXTURE_DIR / "chunk_refs.json").read_text())

    async with async_session_factory() as session:
        for doc in manifest["documents"]:
            row = (
                (
                    await session.execute(
                        text(
                            "SELECT status, title, language FROM knowledge_documents "
                            "WHERE external_id = :ext AND source = :src"
                        ),
                        {"ext": doc["external_id"], "src": doc["source"]},
                    )
                )
                .mappings()
                .first()
            )
            ok = row is not None and row["status"] == "active" and row["language"] == doc["locale"]
            check(
                f"document {doc['document_key']}",
                ok,
                f"status={row['status'] if row else 'MISSING'}",
            )

            expected_chunks = chunk_refs[doc["document_key"]]["chunk_count"]
            actual_chunks = (
                await session.execute(
                    text(
                        "SELECT COUNT(*) FROM document_chunks c "
                        "JOIN knowledge_documents d ON d.id = c.document_id "
                        "WHERE d.external_id = :ext AND d.source = :src"
                    ),
                    {"ext": doc["external_id"], "src": doc["source"]},
                )
            ).scalar()
            check(
                f"chunks {doc['document_key']}",
                actual_chunks == expected_chunks,
                f"expected {expected_chunks}, got {actual_chunks}",
            )

        total_docs = (
            await session.execute(
                text(
                    "SELECT COUNT(*) FROM knowledge_documents WHERE metadata ->> 'fixture' = 'true'"
                )
            )
        ).scalar()
        total_chunks = (
            await session.execute(
                text(
                    "SELECT COUNT(*) FROM document_chunks c "
                    "JOIN knowledge_documents d ON d.id = c.document_id "
                    "WHERE d.metadata ->> 'fixture' = 'true'"
                )
            )
        ).scalar()
        check("fixture documents total", total_docs == len(manifest["documents"]), str(total_docs))
        check(
            "fixture chunks total",
            total_chunks == sum(c["chunk_count"] for c in chunk_refs.values()),
            str(total_chunks),
        )

        quota = (
            (
                await session.execute(
                    text(
                        "SELECT tier, daily_query_limit, max_conversations_per_day FROM ai_user_quotas "
                        "WHERE user_id = '10000000-0000-4000-8000-000000000001'"
                    )
                )
            )
            .mappings()
            .first()
        )
        check(
            "fixture user quota",
            quota is not None and quota["daily_query_limit"] >= 500,
            str(dict(quota) if quota else None),
        )


async def verify_core_db() -> None:
    core_url = os.environ.get(
        "CORE_DATABASE_URL",
        "postgresql://user_adisu:o8%3DG2%295EU3X6@172.28.0.3:5432/adisu_db",
    )
    from sqlalchemy.ext.asyncio import create_async_engine

    engine = create_async_engine(core_url.replace("postgresql://", "postgresql+asyncpg://"))
    async with engine.connect() as conn:
        profile = (
            (
                await conn.execute(
                    text(
                        "SELECT bp.company_name, bp.region, bp.stage, s.slug AS sector "
                        "FROM business_profiles bp "
                        "LEFT JOIN sectors s ON s.id = bp.sector_id "
                        "WHERE bp.account_id = '10000000-0000-4000-8000-000000000002'"
                    )
                )
            )
            .mappings()
            .first()
        )
        check(
            "fixture profile",
            profile is not None and profile["company_name"] == "Selam Coffee Export PLC",
            str(dict(profile) if profile else None),
        )

        compliance = (
            await conn.execute(
                text(
                    "SELECT compliance_type, expiry_date, reminder_days_before FROM compliance_entries "
                    "WHERE account_id = '10000000-0000-4000-8000-000000000002' ORDER BY expiry_date"
                )
            )
        ).all()
        check("compliance entries", len(compliance) == 3, str([c[0] for c in compliance]))

        progress = (
            await conn.execute(
                text(
                    "SELECT status, COUNT(*) FROM user_guide_progresses "
                    "WHERE account_id = '10000000-0000-4000-8000-000000000002' GROUP BY status"
                )
            )
        ).all()
        statuses = {r[0]: r[1] for r in progress}
        check(
            "guide progress",
            statuses.get("COMPLETED") == 2 and statuses.get("IN_PROGRESS") == 1,
            str(statuses),
        )

        guides = (
            await conn.execute(
                text("SELECT slug FROM guides WHERE id = '20000000-0000-4000-8000-000000000001'")
            )
        ).scalar()
        check("fixture guide", guides == "fixture-business-formalization", str(guides))
    await engine.dispose()


async def main() -> None:
    await verify_ai_db()
    await verify_core_db()
    if _failures:
        print(f"\nFIXTURE VERIFICATION FAILED ({len(_failures)} failures)")
        sys.exit(1)
    print("\nFIXTURE VERIFICATION PASSED")


if __name__ == "__main__":
    asyncio.run(main())
