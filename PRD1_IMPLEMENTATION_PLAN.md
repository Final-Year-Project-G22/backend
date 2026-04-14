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

Implemented:
- Extended storage abstraction with upload intent contract.
- Implemented SeaweedFS direct upload intent generation.
- Added MinIO non-implementation stub to keep interface compatibility.

Status:
- Completed and committed.

---

## Commit 5 - Finalize API + transactional finalize service

Goal:
- Implement finalize endpoint with idempotent transactional write.

Implemented:
- Added ingestion DTOs for upload intent and finalize endpoints.
- Added authenticated ingestion routes and handlers.
- Implemented transactional finalize flow with idempotency handling.
- Persisted ingestion document + outbox record atomically.

Status:
- Completed and committed.

---

## Commit 6 - Envelope signer + ingestion configs

Goal:
- Enforce signed envelopes with key rotation support.

Implemented:
- Added envelope signer service with HMAC signing.
- Added ingestion signing configuration.
- Added outbox dispatcher lifecycle integration.

Status:
- Completed and committed.

---

## Commit 7 - Outbox dispatcher + retry semantics

Goal:
- Publish asynchronously with durable retry behavior.

Implemented:
- Added configurable dispatcher batch and interval controls.
- Added exponential retry backoff with max delay.
- Added dead-letter transition after max attempts.
- Updated outbox repository contract for retry/dead-letter state transitions.

Status:
- Completed and committed.

---

## Commit 8 - AI-side envelope verification (minimal)

Goal:
- Consumer trust checks for signature/schema/event-shape.

Implemented:
- Added AI-side envelope verifier utility.
- Added schema/key/signature verification checks.
- Added unit tests for valid, invalid, unknown-key, and schema-mismatch cases.

Status:
- Completed and committed.

---

## Commit 9 - Integration + contract hardening

Goal:
- Lock non-negotiable test gates for PRD1.

Implemented (ongoing):
- Added outbox dispatcher unit coverage for success, retry, and dead-letter paths.
- Added ingestion service unit coverage for validation, idempotency, persistence, and failure paths.
- Remaining: broader end-to-end contract/integration wiring across service boundary.

Status:
- In progress.

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
- Commit 3: committed.
- Commit 4: committed.
- Commit 5: committed.
- Commit 6: committed.
- Commit 7: committed.
- Commit 8: committed.
- Commit 9: in progress (unit-level hardening completed).
