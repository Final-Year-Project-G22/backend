# AI Evaluation Fixture (FIN-76)

Deterministic, frozen fixture for the 36-item Agentic Ask evaluation (FIN-69).
**No live or demo data**: every document and account row is authored fixture
content, versioned by `manifest.json`.

## What it is

| Component | Location | Purpose |
|---|---|---|
| Frozen knowledge corpus | `corpus/en/*.md`, `corpus/am/*.md` | 12 authored documents (6 topics × 2 locales) covering VAT, TIN, business registration, trade license, compliance deadlines, and formalization steps |
| Manifest | `manifest.json` | Fixture version `1.0.0`, document keys (`locale:source:stem`), sources, external IDs, titles, effective dates, and the fixture account spec |
| Chunk refs | `chunk_refs.json` | Deterministic per-document chunk counts and token counts, computed with the real production chunker (structural, 512 tokens, 50 overlap) |
| Corpus provisioner | `scripts/provision_corpus.py` | Idempotent upsert of documents + chunks (uuid5 IDs from document keys), embeds via the configured `EmbeddingPort`, seeds fixture-user quota |
| Account seed | `scripts/seed_account.sql` | Core-backend fixture account: user, profile, compliance entries, guide, steps, translations, progress |
| Verifier | `scripts/verify_fixture.py` | Asserts manifest/chunk-refs consistency and account state; prints verification facts |

## Document keys

`document_key = {locale}:{source}:{stem}` — e.g. `en:tax_code:vat-registration`.
Golden items reference `expected_chunk_refs` as `document_key` + `chunk_index`;
chunk indices are stable because the corpus text and chunking strategy are
frozen (`chunk_refs.json` records the expected counts — re-run
`scripts/compute_chunk_refs.py` after any corpus edit and commit the diff).

## Provisioning

1. Core-backend DB (profile/compliance/guide state):

   ```sh
   psql "$CORE_DATABASE_URL" -f evals/fixture/scripts/seed_account.sql
   ```

2. ai-service DB (corpus + chunks + embeddings):

   ```sh
   cd ai-service
   uv run python evals/fixture/scripts/provision_corpus.py
   ```

   Requires a configured embedding provider (default Cohere) in the ai-service
   environment. With `--skip-embeddings` the script upserts the corpus
   structure with chunk status `PENDING` (for local checks without provider
   keys); the eval environment must re-run without the flag so chunks become
   `EMBEDDED` and retrievable.

3. Verify:

   ```sh
   cd ai-service
   DATABASE_URL=... CORE_DATABASE_URL=... \
     uv run python evals/fixture/scripts/verify_fixture.py
   ```

## Fixture account (eval-msme-01)

- Login: `eval-msme-01@fixture.local` / `EvalFixture2026!`
- Profile: Selam Coffee Export PLC — sector `crop-farming`, region `OROMIA`,
  stage `OPERATIONAL`, tags `[plc, op-exporter, has-employees, tax-vat]`
- Compliance: `trade_license` (expires 2027-01-10, 30-day reminder),
  `tax_registration` (2030-06-15, 45-day), `business_registration`
  (2031-06-15, 60-day)
- Guide `fixture-business-formalization`: 4 steps; 2 COMPLETED, 1
  IN_PROGRESS (`renew-trade-license`), 1 LOCKED
- Fixed UUIDs: user `10000000-…-0001`, account `10000000-…-0002`, profile
  `10000000-…-0003`, compliance `10000000-…-0004..6`, guide `20000000-…-0001`,
  steps `21000000-…-0001..4`, progress `30000000-…-0001..4`

## Verification facts (recorded 2026-08-18, fixture v1.0.0)

- Documents provisioned: **12** (6 EN, 6 AM; sources: 2×tax_code, 1×legal,
  1×government, 2×guide per locale)
- Chunks provisioned: **62** (5 per doc, 6 for `formalization-overview` both
  locales); chunk indices 0-based per document, deterministic
- Account profile: as above; compliance: 3 active entries; guide progress:
  2 COMPLETED / 1 IN_PROGRESS / 1 LOCKED
- Fixture-user quota: pro, 1000 queries/day, 200 conversations/day
- Embeddings: **PENDING** in this environment (no provider key); the eval
  environment must run `provision_corpus.py` without `--skip-embeddings` and
  re-verify (`EMBEDDED` chunks)
