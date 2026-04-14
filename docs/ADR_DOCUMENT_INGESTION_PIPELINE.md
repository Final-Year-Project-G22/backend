# ADR-001: Document Ingestion Pipeline (Core Upload + AI Async Processing)

- Status: Accepted
- Date: 2026-04-09
- Decision Makers: Core Backend + AI Service

## Context

Core backend already owns upload/auth/storage entrypoints. AI service already owns retrieval and inference primitives, but there is no production-grade async ingestion pipeline for parsing, chunking, embedding, and indexing uploaded documents at scale.

The design must support:
- high burst upload traffic,
- reliability (no event loss, retry-safe),
- security (signed events, least privilege file access),
- observability and user-facing progress streaming,
- future embedding profile evolution.

## Decision

### 1) Ownership and Flow
- Core remains the ingestion request entrypoint.
- Upload path uses direct-to-storage presigned upload URLs.
- Client must call FinalizeUpload in core after upload.
- Core verifies object metadata (existence, size, checksum, content-type), writes document metadata + outbox in one transaction, then publishes ingestion request asynchronously.
- AI service worker process owns orchestration and execution.

### 2) Eventing and Delivery
- Transport: RabbitMQ.
- Delivery semantics: at-least-once.
- Event granularity: one event per document; optional batch_id.
- Event contracts are shared and versioned (v1).
- Versioned event names:
  - document.ingestion.requested.v1
  - document.ingestion.status.updated.v1
  - document.lifecycle.archived.v1 / document.lifecycle.removed.v1

### 3) Event Envelope (Strict)
All ingestion events must include:
- event_id
- event_type
- schema_version
- occurred_at
- producer
- key_id
- signature
- idempotency_key
- account_id
- user_id
- batch_id
- replay_count
- payload

### 4) Security
- Events are HMAC signed with key_id and key rotation support (active + previous keys).
- File payload in events is pointer metadata only (never file bytes).
- AI fetches content through short-lived signed URL (15 minutes) with refresh via internal core gRPC method.
- Core validates caller/service identity and document/object mapping before issuing refresh URL.
- No raw content/snippets in logs/events.

### 5) Orchestration and Status
- AI worker state machine:
  - queued -> validating -> fetching -> chunking -> embedding -> indexing -> completed
  - terminal: failed, cancelled
- Newer events supersede older ones for same logical document version lineage.
- Success is atomic: chunks + embeddings + completed status are committed together.
- Retry policy: exponential backoff + jitter + transient/permanent classification.
- Exhausted retries go to DLQ; failures persisted; admin re-drive is core-controlled and audited.

### 6) Parsing, Chunking, and Quality
- PR1 formats (no OCR): pdf, docx, txt, md, html, csv, json, xml.
- Parser architecture: unified parser port with format adapters + shared normalization.
- Canonical parser output: document_text + sections[] + metadata.
- Chunking: structural-first + token cap + overlap.
- Chunk provenance persisted: document/chunk identifiers, chunk_index, section heading, source location, parser_version, embedding_profile_id.
- Language: auto-detected at ingest; declared language retained; mismatch flagged.
- Chunk quality gates applied; document fails if accepted_chunks == 0 or rejected_ratio > 70%.
- Thresholds are ai-service config driven.

### 7) Embeddings and Retrieval Compatibility
- Dynamic model switching is supported.
- Storage is profile-aware (chunk embeddings linked to embedding profile metadata).
- Retrieval queries active compatible embedding profile only; BM25 fallback when vector coverage is low.
- Response metadata includes retrieval_mode and vector_coverage_ratio.

### 8) Progress Delivery to Clients
- AI service is source of truth for ingestion status.
- AI emits status update events; core maintains projection/mirror for APIs.
- Client progress channel: SSE from core.
- Multi-instance core SSE fan-out: Redis-backed.
- SSE resume via Last-Event-ID with replay buffer (30 minutes or 500 events/job).
- Status event emission policy: stage transitions + terminal + periodic snapshots (5s OR 5% progress delta).

### 9) Rollout and Control
- Rollout uses a global ingestion toggle.
- Toggle off behavior: reject new ingestion requests; allow in-flight jobs to finish; keep status/SSE available.

## Consequences

### Positive
- Reliable ingestion under burst load with strong idempotency.
- Clear ownership boundaries and operational controls.
- Secure internal trust model and auditable re-drive.
- Scalable UX visibility via SSE.

### Tradeoffs
- More components (outbox, workers, status projection, SSE replay).
- Requires tighter observability and alerting discipline.

## Alternatives Considered

- Synchronous core->ai ingestion call: rejected due to latency and burst fragility.
- Single queue without tiering: rejected due to fairness concerns.
- In-memory SSE fan-out only: rejected for horizontal scaling.
- Exactly-once delivery: rejected for complexity; idempotent at-least-once chosen.

## Non-Negotiable Validation Gates

- Contract tests (envelope/signature/version compatibility).
- End-to-end integration (upload -> outbox -> worker -> status -> completed).
- Idempotency/supersede/replay tests.
- Retry + DLQ classification tests.
- Burst performance smoke tests against agreed SLOs.

---

## Execution Checklist (Mapped to Files/Modules)

### A) Shared Contracts (proto)
- [ ] Add ingestion event contracts and envelope schema
  - `proto/` (new files under `ingestion/v1`)
- [ ] Generate language stubs
  - `proto/buf.gen.yaml`
  - generated outputs in core and ai service protobuf directories

### B) Core: Direct Upload Finalization + Outbox Publish
- [ ] Add finalize upload DTO/handler/route
  - `core-backend/internal/modules/ai/delivery/dto/`
  - `core-backend/internal/modules/ai/delivery/handler/`
  - `core-backend/internal/modules/ai/delivery/routes/`
- [ ] Implement storage object verification service logic
  - `core-backend/pkg/storage/`
  - module-level application service in `core-backend/internal/modules/ai/application/`
- [ ] Add outbox entity/repository and migration
  - `core-backend/internal/modules/.../infrastructure/repository/`
  - `core-backend/internal/core/` DB wiring + migration registration
- [ ] Add outbox dispatcher lifecycle hook
  - `core-backend/internal/core/module.go`

### C) Core: Ingestion Status Mirror + SSE
- [ ] Add status projection model/repo (consumed from ai status events)
  - `core-backend/internal/modules/ai/infrastructure/repository/`
- [ ] Subscribe to `document.ingestion.status.updated.v1`
  - `core-backend/internal/handlers/handlers.go` or dedicated module handler
- [ ] Add SSE endpoint and polling status endpoint
  - `core-backend/internal/modules/ai/delivery/routes/`
  - `core-backend/internal/modules/ai/delivery/handler/`
- [ ] Add Redis pub/sub fan-out + replay buffer support
  - `core-backend/internal/core/cache.go` (or dedicated SSE bus helper)

### D) Core: Internal Signed URL Refresh gRPC
- [ ] Add core internal gRPC method for ingestion URL refresh
  - core gRPC module: `core-backend/internal/modules/coregrpc/`
  - proto/service stubs in shared `proto/` and generated pb
- [ ] Validate service token + document/object mapping before issuing URL
  - auth/guard logic in core internal service layer

### E) AI Service: Worker + Orchestration
- [ ] Add worker runtime and queue consumer
  - `ai-service/workers/` (new worker module(s))
- [ ] Implement envelope verification, key rotation, idempotency handling
  - `ai-service/core/usecases/` + `ai-service/infrastructure/messagebus/`
- [ ] Implement orchestration state machine and status publisher
  - `ai-service/core/usecases/`
  - `ai-service/core/ports/event_bus.py`

### F) AI Service: Parsing + Chunking + Embeddings
- [ ] Add parser port and adapters for PR1 formats
  - `ai-service/core/ports/` (new parser port)
  - `ai-service/infrastructure/` (format adapters)
- [ ] Add canonical parse output model and chunking pipeline
  - `ai-service/core/domain/`
  - `ai-service/core/usecases/`
- [ ] Add quality gates and language auto-detection integration
  - `ai-service/core/usecases/`

### G) AI Service: Profile-Aware Embedding Storage
- [ ] Add schema/models for embedding profiles and chunk embeddings
  - `ai-service/infrastructure/database/models_sqlalchemy.py`
  - `ai-service/alembic/versions/`
- [ ] Reset/reseed-friendly migration path (non-prod assumption)
  - `ai-service/alembic/versions/`
- [ ] Update repository logic to active-profile query + BM25 fallback metadata
  - `ai-service/infrastructure/database/repositories/knowledge_repository.py`

### H) AI Service: Config and Concurrency Controls
- [ ] Add ingestion config knobs
  - `ai-service/app/config.py`
  - e.g., queue concurrency, retry windows, rejection threshold, snapshot cadence
- [ ] Wire dependencies in DI
  - `ai-service/app/container.py`

### I) Security and Keys
- [ ] Add signing key config (active + previous + key_id)
  - core config: `core-backend/internal/core/types.go`, `core-backend/internal/configs/config.yml`
  - ai config: `ai-service/app/config.py`
- [ ] Ensure no payload content logging
- logging call sites in core handlers and ai workers

### J) Test Matrix (Must Pass)
- [ ] Contract tests (event envelope/signature/schema)
  - core + ai unit/integration test dirs
- [ ] E2E ingestion integration
  - upload finalize -> outbox -> consume -> completed
- [ ] Replay/idempotency/supersede tests
- [ ] Retry/DLQ classification tests
- [ ] Burst smoke/perf tests with SLO assertions
