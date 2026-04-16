# PRD2 Implementation Plan

This document tracks the implementation plan for **PRD 2: AI Ingestion Orchestration, Parsing, Quality Gates, and Indexing** and maps work to commit-sized slices.

## Scope (PRD2)

- AI-worker owned ingestion orchestration from event consumption to terminal state.
- Strict envelope verification + idempotent/supersede-safe processing.
- Deterministic parsing and chunking pipeline with quality gates.
- Profile-aware embedding persistence and indexing.
- Retry classifier, backoff policy, and DLQ routing.
- Status event publication for downstream projection (PRD3 dependency).

Out of scope for PRD2:

- Core status projection and SSE replay endpoints (PRD3).
- OCR/scanned-image ingestion.
- Advanced semantic reranking beyond current retrieval approach.
- End-user UX for chunk-level quality diagnostics.

---

## Delivery Assumptions

- PRD1 contracts and finalize/outbox flow are already live.
- Transport remains RabbitMQ with at-least-once delivery.
- AI service keeps content-privacy guarantees (no raw snippets in events/logs).
- Migration approach remains reset/reseed-friendly in non-prod environments.

---

## Commit Plan

## Commit 1 - Ingestion worker bootstrap + queue consumer

Goal:
- Introduce a dedicated ingestion worker runtime that consumes `document.ingestion.requested.v1` events.

Planned implementation:
- Add worker entrypoint and consumer loop.
- Add message decode/ack/nack plumbing with bounded prefetch.
- Add DI wiring for worker-specific dependencies and configuration.

Primary paths:
- `ai-service/workers/`
- `ai-service/infrastructure/messagebus/`
- `ai-service/app/container.py`
- `ai-service/app/config.py`

Status:
- Completed.

---

## Commit 2 - Envelope trust gate + ingestion event decoding

Goal:
- Fail closed on invalid signatures/schema before orchestration starts.

Planned implementation:
- Reuse/extend envelope verifier for ingestion worker consumption path.
- Validate accepted event types and payload schema versions.
- Add canonical error mapping for reject/retry handling.

Primary paths:
- `ai-service/core/security/envelope_verifier.py`
- `ai-service/core/domain/ingestion_events.py`
- `ai-service/workers/`

Status:
- Completed.

---

## Commit 3 - Idempotency ledger + supersede checks

Goal:
- Ensure replay-safe processing and stale-lineage suppression.

Planned implementation:
- Add processed-event ledger (event_id/idempotency_key uniqueness).
- Add lineage version comparison and supersede guard.
- Define no-op behavior for duplicate and stale events.

Primary paths:
- `ai-service/core/ports/`
- `ai-service/infrastructure/database/models_sqlalchemy.py`
- `ai-service/infrastructure/database/repositories/`
- `ai-service/alembic/versions/`

Status:
- Completed.

---

## Commit 4 - Orchestration state machine (deterministic core)

Goal:
- Implement a single orchestration use case that drives all state transitions.

Planned implementation:
- Add `start_ingestion(event)`-style application service.
- Encode state machine transitions: queued -> validating -> fetching -> chunking -> embedding -> indexing -> completed, with failed/cancelled terminals.
- Emit structured transition context for observability and retries.

Primary paths:
- `ai-service/core/usecases/`
- `ai-service/core/domain/`

Status:
- Completed.

---

## Commit 5 - Parser port + PR1 format adapters

Goal:
- Make parsing extensible while producing one canonical parse output.

Planned implementation:
- Add parser port contract and adapter registry.
- Implement initial adapters for PR1 formats: pdf, docx, txt, md, html, csv, json, xml.
- Normalize outputs to canonical shape (`document_text`, `sections`, normalized metadata).

Primary paths:
- `ai-service/core/ports/`
- `ai-service/core/domain/`
- `ai-service/infrastructure/`

Status:
- Planned.

---

## Commit 6 - Chunking engine + provenance persistence

Goal:
- Generate deterministic, traceable chunks for embedding/indexing.

Planned implementation:
- Add structural-first chunking with token cap + overlap controls.
- Persist chunk provenance (`chunk_index`, source section/location, parser version).
- Add chunk schema/repository layer required by orchestration.

Primary paths:
- `ai-service/core/usecases/`
- `ai-service/core/domain/`
- `ai-service/infrastructure/database/models_sqlalchemy.py`
- `ai-service/infrastructure/database/repositories/`
- `ai-service/alembic/versions/`

Status:
- Planned.

---

## Commit 7 - Language detection + quality gates

Goal:
- Reject low-quality chunk sets early and deterministically.

Planned implementation:
- Add language auto-detection and mismatch flag against declared language.
- Add quality gate policy (`accepted_chunks > 0`, rejected-ratio threshold, configurable minimums).
- Encode permanent-failure outcomes for quality failures.

Primary paths:
- `ai-service/core/usecases/`
- `ai-service/core/domain/`
- `ai-service/app/config.py`

Status:
- Planned.

---

## Commit 8 - Embedding profile model + active-profile compatibility

Goal:
- Support profile-aware vector persistence and safe model evolution.

Planned implementation:
- Add embedding profile tables and chunk-embedding linkage.
- Store profile metadata/tags with vectors.
- Enforce active-profile filtering in retrieval-facing repository paths.

Primary paths:
- `ai-service/infrastructure/database/models_sqlalchemy.py`
- `ai-service/infrastructure/database/repositories/knowledge_repository.py`
- `ai-service/alembic/versions/`

Status:
- Planned.

---

## Commit 9 - Retry classifier + backoff + DLQ sink

Goal:
- Make transient failures self-healing and terminal failures auditable.

Planned implementation:
- Add transient/permanent error classifier.
- Implement exponential backoff with jitter and max-attempt policy.
- Route exhausted attempts to DLQ with persistent failure context.

Primary paths:
- `ai-service/workers/`
- `ai-service/core/usecases/`
- `ai-service/infrastructure/messagebus/`
- `ai-service/app/config.py`

Status:
- Planned.

---

## Commit 10 - Signed URL fetch/refresh flow (internal core trust)

Goal:
- Fetch document bytes via short-lived signed URLs only, with strict internal authorization.

Planned implementation:
- Add internal core-service client call for signed URL acquisition/refresh.
- Validate document/object mapping and service identity assumptions in request flow.
- Handle refresh mid-pipeline without restarting job state.

Primary paths:
- `ai-service/infrastructure/rpc/core_service.py`
- `ai-service/core/ports/core_service.py`
- `ai-service/core/usecases/`

Status:
- Planned.

---

## Commit 11 - Status publisher contract + emission policy

Goal:
- Publish ingestion status updates for downstream projection and UX progress.

Planned implementation:
- Emit `document.ingestion.status.updated.v1` events at stage transitions and terminal states.
- Add periodic snapshot emission policy (time/progress delta driven).
- Include correlation and lineage-safe metadata needed by core projection.

Primary paths:
- `ai-service/core/usecases/`
- `ai-service/core/ports/event_bus.py`
- `ai-service/infrastructure/messagebus/`
- `proto/ai/ingestion/v1/events.proto` (only if additive payload fields are required)

Status:
- Planned.

---

## Commit 12 - Atomic completion transaction + hardening tests

Goal:
- Guarantee no partial-success visibility and lock PRD2 non-negotiable gates.

Planned implementation:
- Commit chunks + embeddings + completed state atomically.
- Add deterministic transition tests, contract tests, repository tests, and failure-injection integration tests.
- Add smoke scenarios for concurrency and retry tuning defaults.

Primary paths:
- `ai-service/core/usecases/`
- `ai-service/infrastructure/database/repositories/`
- `ai-service/tests/unit/`
- `ai-service/tests/integration/`

Status:
- Planned.

---

## Review Checklist (Per Commit)

- Worker behavior remains idempotent under duplicate delivery.
- No raw content snippets in logs/events/errors.
- Retry vs permanent-failure classification is explicit and tested.
- State transitions are monotonic and terminal states are final.
- Schema changes are additive-only and contract tests pass.
- Build/lint/tests pass for touched layers.

## Current Execution State

- Commit 1: completed.
- Commit 2: completed.
- Commit 3: completed.
- Commit 4: completed.
- Commit 5-12: planned.
