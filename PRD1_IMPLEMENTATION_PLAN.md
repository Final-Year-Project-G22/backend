# PRD 1 Implementation Plan (Detailed, Implementation-Ready)

This plan covers **PRD 1: Reliable Ingestion Intake and Secure Eventing** from `DOCUMENT_INGESTION_PIPELINE_PRDS.md`.

It is based on:
- `ADR_DOCUMENT_INGESTION_PIPELINE.md`
- Current repository structure and implementation state in `core-backend`, `ai-service`, and `proto`

---

## 1) Current State Analysis (Codebase Reality)

### What already exists (useful foundations)
- Core has RabbitMQ abstraction with `Publish`/`Subscribe` and topic exchange wiring in `core-backend/pkg/rabbitmq/client.go`.
- Core has transaction support across repositories via `Transactor.WithinTransaction` and context-bound tx propagation in `core-backend/internal/core/database.go` and `core-backend/internal/shared/repository/base_repository.go`.
- Core has storage abstractions (`Exists`, `GetInfo`, `GetPresignedURL`) in `core-backend/pkg/storage/storage.go` with SeaweedFS/MinIO implementations.
- Core has module wiring via FX and startup lifecycle hooks in `core-backend/internal/core/module.go`.
- Proto generation pipeline already exists via `proto/buf.gen.yaml` and `make proto-gen`.

### What is missing for PRD1
- No ingestion event contract package in `proto` yet (only inference/core user contracts are present).
- Core AI module is currently only an inference gRPC client; no HTTP ingestion routes/services/repositories (`core-backend/internal/modules/ai/module.go`).
- No transactional outbox entity/repository/dispatcher in core.
- No strict signed envelope utility (key-id, signature, rotation support) in core or AI.
- No finalize-upload API in core.
- No ingestion toggle and signing key config model in core config.
- AI service has event bus port shape but no concrete messagebus implementation for ingestion yet (`ai-service/core/ports/event_bus.py`, `ai-service/infrastructure/messagebus/`).

### Important constraints identified during analysis
- Current storage interface only exposes `GetPresignedURL` (download semantics), but ADR requires direct-to-storage upload URL flow. PRD1 therefore needs either:
  1) storage interface extension for upload signing, or
  2) provider-specific upload ticket abstraction.
- Current core publish pattern in IAM uses direct async publish (goroutine) without outbox durability (`core-backend/internal/modules/iam/application/service/auth_service.go`). PRD1 must not follow that pattern.
- AI module currently has no entity provider registered to schema manager; PRD1 DB entities under AI module require adding this wiring.

---

## 2) PRD1 Scope Lock

### In scope (must ship)
- Shared ingestion event contract v1 (strict envelope + requested payload schema).
- Core finalize-upload API with metadata verification.
- Core transactional persistence of ingestion request + outbox event.
- Core outbox dispatcher with retry/backoff.
- Signed envelope generation (HMAC + key id + rotation support).
- Idempotency handling at finalize boundary.
- Minimal AI-side envelope verification module + tests (contract safety).

### Out of scope for PRD1
- AI parsing/chunking/embedding/indexing orchestration.
- SSE/status mirror/projection.
- Internal core gRPC signed-URL refresh endpoint.
- DLQ re-drive admin APIs/UI.

---

## 3) Target PRD1 Architecture (Implementation View)

1. Client requests upload intent from core (direct upload flow support).
2. Client uploads object to storage.
3. Client calls finalize endpoint with object pointer + declared metadata + idempotency key.
4. Core verifies object metadata from storage and validates ownership/shape.
5. Core writes ingestion document metadata + outbox row in one DB transaction.
6. Background dispatcher signs envelope and publishes `document.ingestion.requested.v1`.
7. Outbox row is marked published or scheduled for retry.

Reliability model: **at-least-once delivery + idempotent consumer contract**.

---

## 4) Workstreams and Module Design

## A. Shared Contracts (Proto)

### Deliverables
- New package under `proto/ai/ingestion/v1` (recommended):
  - `events.proto` for envelope + payload messages
  - optional `contracts.proto` split if preferred
- Event type constants represented in payload/envelope fields:
  - `document.ingestion.requested.v1`
  - reserve status/lifecycle types for forward compatibility.

### Decisions
- Keep envelope mandatory fields strict per ADR.
- Keep payload pointer-only (no file bytes).
- Add schema evolution notes (additive only, no tag reuse).

### Verification
- `make proto-gen`
- compile check both services.

---

## B. Core Domain + Persistence (Ingestion + Outbox)

### Deep module 1: `IngestionFinalizeService`
Encapsulates:
- request validation
- object metadata verification
- idempotency resolution
- transactional write of ingestion doc + outbox

### Deep module 2: `OutboxDispatcher`
Encapsulates:
- polling pending rows
- envelope signing
- publish/retry scheduling
- durable result transitions

### Data model (core DB)
Recommended new entities/tables:
- `ingestion_documents`
  - `id`, `account_id`, `user_id`, `storage_key`, `content_type`, `size_bytes`, `checksum`, `status`, `schema_version`, `source_filename`, `batch_id`, `created_at`, `updated_at`
- `ingestion_outbox`
  - `id`, `event_id`, `event_type`, `schema_version`, `idempotency_key`, `aggregate_id`, `payload_json`, `signature_key_id`, `status`, `attempt_count`, `next_attempt_at`, `published_at`, `last_error`, `created_at`, `updated_at`

Indexes/constraints:
- unique: `event_id`
- unique: `(event_type, idempotency_key)`
- index: `(status, next_attempt_at)`
- index: `aggregate_id`

### File areas
- `core-backend/internal/modules/ai/domain/entity/`
- `core-backend/internal/modules/ai/domain/repository/`
- `core-backend/internal/modules/ai/infrastructure/repository/`
- `core-backend/internal/modules/ai/application/service/`
- `core-backend/internal/modules/ai/entities.go` (new, for schema manager registration)

---

## C. Core API Surface (Finalize Upload)

### Endpoints
- `POST /api/v1/ai/ingestion/uploads/finalize` (mandatory)
- `POST /api/v1/ai/ingestion/uploads/initiate` (recommended if direct-upload URL generation is included in PRD1 slice)

### Finalize request fields (minimum)
- `storage_key`
- `content_type`
- `size_bytes`
- `checksum`
- `idempotency_key`
- optional: `source_filename`, `batch_id`, `declared_language`

### Response fields (minimum)
- `ingestion_id`
- `document_id`
- `state=queued`
- `event_id`

### File areas
- `core-backend/internal/modules/ai/delivery/dto/`
- `core-backend/internal/modules/ai/delivery/handler/`
- `core-backend/internal/modules/ai/delivery/routes/`

---

## D. Storage Verification + Upload URL Gap Closure

### Gap
ADR requires direct-to-storage upload URL issuance, but storage interface currently has only download presign.

### PRD1 implementation decision
- Extend `Storage` interface with upload intent support, for example:
  - `GetPresignedUploadURL(...)` OR
  - `CreateUploadTicket(...)`

### Provider implementation strategy
- SeaweedFS: use assign flow + constrained upload metadata.
- MinIO: use presigned PUT.

### Files
- `core-backend/pkg/storage/storage.go`
- `core-backend/pkg/storage/seaweedfs.go`
- `core-backend/pkg/storage/minio.go`

---

## E. Envelope Signing and Key Rotation

### Deep module 3: `EnvelopeSigner`
Responsibilities:
- canonical serialization
- HMAC signing with active key
- attach `key_id`

### Deep module 4: `EnvelopeVerifier` (AI-side for contract safety)
Responsibilities:
- lookup key by `key_id`
- verify signature
- reject malformed/unknown schema versions

### Config additions
- Core: active key id/secret + previous keys list + ingestion toggle + retry knobs.
- AI: verification key ring + accepted schema versions.

### Files
- Core config model: `core-backend/internal/core/types.go`, `core-backend/internal/configs/config.yml`
- AI config model: `ai-service/app/config.py`
- New security helpers under module/infrastructure utilities.

---

## F. Outbox Dispatcher Lifecycle Integration

### Behavior
- Start worker loop via FX lifecycle on app start.
- Stop gracefully on shutdown.
- Batch size and poll interval are config-driven.
- Retry policy: exponential backoff + jitter.

### Files
- `core-backend/internal/core/module.go` (lifecycle hook wiring)
- `core-backend/internal/modules/ai/application/service/` (dispatcher/service)

---

## G. Test Strategy (PRD1 Gates)

### Required test layers
1. Contract tests:
   - envelope required fields
   - schema version compatibility
   - signature verify success/failure
2. Core unit tests:
   - finalize validation rules
   - idempotency behavior
   - signer behavior
3. Core repository tests:
   - outbox state transitions
   - retry scheduling semantics
4. Core integration tests:
   - finalize -> tx write -> outbox persisted -> dispatcher publish
5. AI unit tests:
   - envelope verifier with key rotation and malformed inputs

### Commands
- `make proto-gen`
- `make test-go`
- `make test-ai-unit`

---

## 5) Commit-by-Commit Execution Plan

## Commit 1 - Ingestion v1 contracts
**Goal:** establish shared envelope/payload contract.

Changes:
- add `proto/ai/ingestion/v1/*.proto`
- regenerate stubs via `make proto-gen`

Checks:
- generated files appear in `core-backend/pb` and `ai-service/grpc_stubs`
- both codebases compile.

---

## Commit 2 - Core AI module scaffolding + entity provider
**Goal:** make AI module capable of owning DB entities and routes.

Changes:
- add AI entity provider and register with schema manager
- expand `core-backend/internal/modules/ai/module.go` wiring for repositories/services/handlers/routes (initial scaffolding)

Checks:
- app boots
- no schema registration conflicts.

---

## Commit 3 - Core ingestion entities + repositories + migration
**Goal:** persist ingestion docs and outbox records durably.

Changes:
- add entities + repo interfaces + implementations
- generate migration (core migration workflow)
- apply migration locally

Checks:
- migration applies cleanly
- repository tests pass.

---

## Commit 4 - Storage upload intent support
**Goal:** close direct-upload URL gap.

Changes:
- extend storage interface for upload URL/ticket support
- implement SeaweedFS and MinIO upload URL generation
- add unit tests for both adapters

Checks:
- compile passes
- adapter tests pass
- generated URL/ticket includes expiry and key scope.

---

## Commit 5 - Finalize API + transactional finalize service
**Goal:** implement finalize endpoint with idempotent transactional writes.

Changes:
- add DTOs/handlers/routes for finalize
- implement `IngestionFinalizeService`
- perform metadata verification against storage
- write ingestion + outbox in one tx

Checks:
- handler/unit tests pass
- duplicate idempotency calls are stable and non-duplicating.

---

## Commit 6 - Envelope signer + core config extensions
**Goal:** enforce strict signed envelope generation.

Changes:
- add signer module and canonical serializer
- add config for active/previous keys, ingestion toggle, dispatcher knobs
- wire signer into outbox publish path

Checks:
- positive/negative signature tests pass
- malformed config fails fast on startup.

---

## Commit 7 - Outbox dispatcher lifecycle + retries
**Goal:** async publish with durable retries.

Changes:
- add dispatcher worker
- add FX lifecycle start/stop wiring
- implement retry/backoff/jitter and status transitions
- structured logging + metrics counters

Checks:
- integration test finalize->publish path passes
- simulated publish failure schedules retry correctly.

---

## Commit 8 - AI envelope verifier (minimal consumer-readiness)
**Goal:** ensure consumer-side trust model is in place for PRD1.

Changes:
- add AI verifier utility + config key ring support
- add tests: valid, invalid, unknown key, bad schema

Checks:
- AI unit suite green for verifier tests.

---

## Commit 9 - End-to-end contract/integration hardening
**Goal:** lock non-negotiable PRD1 validation gates.

Changes:
- add full contract tests and e2e finalize/outbox/publish test path
- add idempotency replay tests

Checks:
- CI green for proto + go + ai unit tests.

---

## Commit 10 - Operational docs and rollout controls
**Goal:** make rollout safe and repeatable.

Changes:
- add runbook for key rotation and ingestion toggle behavior
- add release checklist and rollback notes

Checks:
- ops docs reviewed and linked from root docs.

---

## 6) Ready-Made Execution Checklist

## A. Planning and Contracts
- [ ] Confirm PRD1 scope lock (no SSE/worker orchestration in this phase).
- [ ] Create ingestion v1 proto envelope + payload schema.
- [ ] Regenerate stubs and verify no breaking change warnings.
- [ ] Document schema evolution rules in proto package notes.

## B. Core Data and Module Wiring
- [ ] Add AI entity provider and schema registration.
- [ ] Create `ingestion_documents` entity.
- [ ] Create `ingestion_outbox` entity.
- [ ] Add repository interfaces and implementations.
- [ ] Add migration and apply locally.

## C. Storage and Finalize API
- [ ] Extend storage contract for upload URL/ticket support.
- [ ] Implement provider support (SeaweedFS + MinIO).
- [ ] Add finalize DTOs/handler/routes.
- [ ] Implement metadata verification logic.
- [ ] Implement transactional finalize service.
- [ ] Enforce idempotency key behavior.

## D. Security and Eventing
- [ ] Add core signing config (active + previous keys).
- [ ] Implement canonical envelope serializer/signing.
- [ ] Add ingestion toggle config and enforcement.
- [ ] Implement outbox dispatcher publish loop.
- [ ] Implement retry/backoff scheduling with persisted error reason.
- [ ] Ensure logs/events contain no raw document content.

## E. AI Consumer Safety
- [ ] Add AI verifier config/key ring.
- [ ] Add envelope verification utility.
- [ ] Add signature/schema/key-id verification tests.

## F. Testing and Validation
- [ ] Contract tests for envelope required fields.
- [ ] Contract tests for signature and schema version compatibility.
- [ ] Unit tests for finalize service and idempotency.
- [ ] Unit tests for outbox dispatcher transitions.
- [ ] Integration test finalize -> outbox -> publish.
- [ ] Failure-path integration test for retry scheduling.

## G. Rollout and Ops
- [ ] Add metrics (publish success/failure, pending outbox depth, retry count).
- [ ] Add alerts for repeated publish failures.
- [ ] Document key rotation runbook.
- [ ] Document ingestion toggle runbook.
- [ ] Execute staged rollout (toggle off -> deploy -> enable).

## H. Definition of Done
- [ ] All required tests pass in CI.
- [ ] Finalize path is transactionally durable.
- [ ] Outbox handles broker outage without data loss.
- [ ] Published events are always signed and schema-valid.
- [ ] Duplicate finalize requests do not create duplicate jobs.

---

## 7) Suggested Verification Sequence During Implementation

1. `make proto-gen`
2. `make test-go`
3. `make test-ai-unit`
4. run local finalize->outbox->publish integration scenario
5. simulate RabbitMQ outage and verify outbox retry persistence
6. rotate active signing key (active -> previous) and verify compatibility window
