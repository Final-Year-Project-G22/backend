# PRD1 Implementation Plan

This document tracks the implementation plan for **PRD 1: Reliable Ingestion Intake and Secure Eventing** and maps work to commit-sized slices.

## Scope (PRD1)

- Shared, versioned ingestion event contracts.
- Core finalize-upload intake path (API + validation + transactional write).
- Transactional outbox for durable async publishing.
- Signed envelope generation (HMAC + key id + rotation support).
- Idempotency handling at finalize boundary.
- Retry-capable outbox dispatcher.

Out of scope for PRD1:

- AI parsing/chunking/embedding execution.
- SSE status projection and replay APIs.
- OCR.

---

## Commit Plan

## Commit 1 - Ingestion v1 contracts

Goal:
- Establish shared envelope/payload contract for RabbitMQ transport.

Implemented:
- `proto/ai/ingestion/v1/events.proto`
- `proto/ai/ingestion/v1/doc.md`
- Generated Go stubs:
  - `core-backend/pb/ai/ingestion/v1/events.pb.go`
  - `core-backend/pb/ai/ingestion/v1/events.pb.validate.go`
- Core AI event contract helpers:
  - `core-backend/internal/modules/ai/domain/event/ingestion_events.go`
  - `core-backend/internal/modules/ai/domain/event/ingestion_envelope.go`

Notes:
- Contract is message-definition based (no ingestion gRPC service in this commit).
- Transport remains RabbitMQ topic exchange.
- Envelope aligns with ADR strict fields.

Status:
- Completed and committed.

---

## Commit 2 - AI module scaffolding + schema provider wiring

Goal:
- Make AI module schema-aware and ready for ingestion feature expansion.

Implemented:
- AI entity provider scaffold:
  - `core-backend/internal/modules/ai/entities.go`
- AI module wiring expansion:
  - `core-backend/internal/modules/ai/module.go`
- Ingestion service scaffold:
  - `core-backend/internal/modules/ai/application/service/ingestion_service.go`
- Handler scaffold:
  - `core-backend/internal/modules/ai/delivery/handler/ingestion_handler.go`
- Route scaffolding:
  - `core-backend/internal/modules/ai/delivery/routes/routes.go`
  - `core-backend/internal/modules/ai/delivery/routes/ingestion_routes.go`

Status:
- Completed and committed.

---

## Commit 3 - Ingestion entities + repositories (without migration)

Goal:
- Add domain/persistence model and repositories needed for finalize + outbox flow.

Implemented:

Entities:
- `core-backend/internal/modules/ai/domain/entity/ingestion_document.go`
- `core-backend/internal/modules/ai/domain/entity/ingestion_outbox.go`

Repository interfaces:
- `core-backend/internal/modules/ai/domain/repository/ingestion_document_repository.go`
- `core-backend/internal/modules/ai/domain/repository/ingestion_outbox_repository.go`

Repository implementations:
- `core-backend/internal/modules/ai/infrastructure/repository/ingestion_document_repository.go`
- `core-backend/internal/modules/ai/infrastructure/repository/ingestion_outbox_repository.go`

Module/entity wiring updates:
- `core-backend/internal/modules/ai/entities.go`
- `core-backend/internal/modules/ai/module.go`

Not implemented here:
- SQL migration file generation (intentionally skipped; user will generate).

Status:
- Code complete, pending your review + migration file.

---

## Commit 4 - Storage upload intent support

Goal:
- Add direct-to-storage upload signing capability for PRD flow.

Planned:
- Extend storage abstraction for upload presign/ticket.
- Implement provider-specific support (SeaweedFS + MinIO).
- Add unit coverage for both adapters.

Status:
- Not started.

---

## Commit 5 - Finalize API + transactional finalize service

Goal:
- Implement finalize endpoint with idempotent transactional write.

Planned:
- Add finalize DTO + handler + routes.
- Validate object metadata (`exists`, `size`, `checksum`, `content-type`).
- Write `ingestion_documents` + `ingestion_outbox` in one DB tx.

Status:
- Not started.

---

## Commit 6 - Envelope signer + ingestion configs

Goal:
- Enforce signed envelopes with key rotation support.

Planned:
- Add signing utility and canonical payload serialization.
- Add config for active/previous keys.
- Add ingestion toggle and dispatcher knobs.

Status:
- Not started.

---

## Commit 7 - Outbox dispatcher + retry semantics

Goal:
- Publish asynchronously with durable retry behavior.

Planned:
- Background dispatcher lifecycle wiring.
- Retry/backoff/jitter.
- Persist publish/fail transitions.

Status:
- Not started.

---

## Commit 8 - AI-side envelope verification (minimal)

Goal:
- Consumer trust checks for signature/schema/event-shape.

Planned:
- Add envelope verification utility in AI service.
- Add validation tests (valid/invalid sig, unknown key, schema mismatch).

Status:
- Not started.

---

## Commit 9 - Integration + contract hardening

Goal:
- Lock non-negotiable test gates for PRD1.

Planned:
- Contract tests (schema/version/signature).
- Finalize -> outbox -> publish integration tests.
- Idempotency and retry-path tests.

Status:
- Not started.

---

## Commit 10 - Operational readiness docs

Goal:
- Rollout and key-rotation readiness.

Planned:
- Add runbook for key rotation.
- Add ingest toggle behavior and rollback notes.
- Add alerting/metrics checklist.

Status:
- Not started.

---

## Review Checklist (Per Commit)

- Contract compatibility (proto fields + event naming).
- DB model correctness (constraints, indexes, nullable/default semantics).
- Transaction boundaries respected.
- Idempotency guarantees explicit.
- No raw file bytes or snippets in events/logs.
- Build/lint/tests passing for touched areas.

## Current Execution State

- Commit 1: committed.
- Commit 2: committed.
- Commit 3: implemented in code; awaiting migration generation and your review before commit.
