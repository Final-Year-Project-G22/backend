# Document Ingestion Pipeline PRDs

This document contains a set of product requirements documents derived from ADR-001 (Document Ingestion Pipeline: Core Upload + AI Async Processing). It is intentionally split into multiple PRDs so delivery can be phased while preserving one end-to-end target architecture.

---

# PRD 1: Reliable Ingestion Intake and Secure Eventing

## Problem Statement

Users can upload documents today, but the platform does not have a production-grade ingestion intake flow that reliably converts uploads into asynchronous ingestion jobs. The current system lacks a dedicated finalize-upload workflow, transactional outbox publication, strict signed event envelopes, and safe idempotent delivery semantics. As a result, uploads are vulnerable to event loss, duplication side effects, and weak traceability.

## Solution

Build a core-owned ingestion intake workflow where clients upload directly to storage, call a finalize endpoint, and receive a durable ingestion job registration. Finalization validates object integrity, writes document metadata and outbox records atomically, and asynchronously publishes versioned signed events to the queue. The event contract becomes the source of truth for producer-consumer compatibility and replay-safe processing.

## User Stories

1. As a platform user, I want to upload files directly to object storage, so that upload latency is low and API servers are not a bottleneck.
2. As a platform user, I want a clear finalize step after upload, so that I know when ingestion has officially started.
3. As a platform user, I want failed finalization to explain what is wrong with my object metadata, so that I can fix the issue and retry.
4. As a platform user, I want duplicate finalize calls to be safe, so that transient network retries do not create duplicate jobs.
5. As a platform user, I want ingestion requests to survive temporary queue outages, so that my uploads are not lost.
6. As a platform administrator, I want every ingestion request to carry a unique event ID and idempotency key, so that I can audit and deduplicate processing.
7. As a platform administrator, I want ingestion events to be signed and key-versioned, so that consumers can verify authenticity during key rotation.
8. As a platform security engineer, I want events to include file pointers but never file bytes, so that sensitive raw content is not exposed in transit.
9. As a backend engineer, I want contract-versioned event names and schema versions, so that producers and consumers can evolve safely.
10. As a backend engineer, I want account and user identity in the event envelope, so that downstream authorization and observability remain tenant-aware.
11. As a backend engineer, I want replay counts and batch IDs in envelope metadata, so that retries and bulk ingestion can be tracked consistently.
12. As an SRE, I want at-least-once delivery with explicit idempotency handling, so that reliability is prioritized without exactly-once complexity.
13. As an SRE, I want failed publishes to be retried from outbox storage, so that temporary infrastructure failures self-heal.
14. As a product owner, I want ingestion toggles that can reject new jobs without dropping in-flight work, so that incident response is controlled.
15. As a compliance stakeholder, I want a complete audit trail from finalize request to published ingestion event, so that post-incident reviews are possible.
16. As an integration developer, I want a shared contract package consumed by both services, so that type drift is detected at build time.
17. As a release manager, I want backward-compatible event evolution rules, so that staged rollouts across services are safe.
18. As a support engineer, I want deterministic ingestion request identifiers exposed to operators, so that user-reported issues are easy to trace.

## Implementation Decisions

- Adopt a strict, shared ingestion event envelope with required metadata fields (event identity, schema versioning, producer identity, signature metadata, idempotency metadata, tenancy metadata, replay metadata, payload).
- Introduce a core ingestion finalization application service as a deep module that encapsulates: object verification, checksum/size/type validation, document metadata creation, and outbox write in one transaction.
- Introduce a transactional outbox deep module with a small interface (`enqueue`, `dispatch`, `mark_published`, `mark_failed`) to isolate delivery guarantees from request handlers.
- Define event naming and schema governance policy: explicit version suffixes, additive-safe evolution, and compatibility validation before rollout.
- Use HMAC signatures with key IDs and active/previous key support; key verification fails closed.
- Keep payload references pointer-only (storage key, checksum, content type, object metadata) and prohibit raw document content in transport.
- Enforce idempotency key semantics at finalize boundary and at consumer boundary.
- Preserve core ownership for ingestion request acceptance and queue publication; AI service remains downstream executor.
- Add global ingestion control toggle behavior: reject new finalize requests when disabled, preserve in-flight processing continuity.

## Testing Decisions

- A good test validates externally observable behavior (accepted/rejected finalize requests, outbox durability, publish semantics, envelope validity), not private helper internals.
- Modules to test: ingestion finalization service, outbox dispatcher, signature signer/verifier utilities, contract serializer/deserializer, idempotency guard logic, ingestion toggle behavior.
- Contract tests must prove schema compatibility and signature validation across producer and consumer boundaries.
- Integration tests must cover upload finalize -> persisted outbox -> published ingestion request event.
- Prior art in the repository includes contract-shape tests for ports, gRPC contract integration tests, and repository behavior tests; ingestion tests should mirror that style with protocol-level assertions and persistence assertions.
- Add negative tests for malformed envelope, stale key ID, invalid signature, missing required metadata fields, and duplicate idempotency key submissions.

## Out of Scope

- OCR and image-to-text extraction.
- End-user ingestion progress streaming UX.
- AI parsing/chunking/embedding execution internals.
- Admin UI for retry management (API-level controls only are covered in later PRDs).

## Further Notes

- This PRD establishes the trust and durability boundary that all subsequent ingestion execution work depends on.
- The design assumes at-least-once delivery and requires strict idempotency to avoid duplicate side effects.

---

# PRD 2: AI Ingestion Orchestration, Parsing, Quality Gates, and Indexing

## Problem Statement

Even with reliable ingestion request delivery, the AI service does not currently have a production worker pipeline that can parse uploaded files, chunk them predictably, generate embeddings, and index results with strong quality guarantees. The system also lacks explicit state-machine orchestration, retry classification, DLQ handling, and profile-aware embedding compatibility for future model changes.

## Solution

Build an AI-owned asynchronous worker pipeline that consumes ingestion request events, verifies signatures, enforces idempotency, and executes a deterministic ingestion state machine from validation through indexing. Add parser adapters for initial formats, a canonical parse model, structural-first chunking, language detection, quality gating, profile-aware embedding persistence, and retrieval compatibility signals including BM25 fallback metadata.

## User Stories

1. As a platform user, I want uploaded documents to be parsed and indexed automatically, so that I can query their content soon after upload.
2. As a platform user, I want ingestion failures to be classified and retried when transient, so that temporary outages do not require manual intervention.
3. As a platform user, I want unrecoverable ingestion failures to stop cleanly with status visibility, so that I do not wait indefinitely.
4. As a platform operator, I want a deterministic ingestion state machine, so that job progress and failure points are debuggable.
5. As a platform operator, I want replay-safe idempotent processing, so that repeated message deliveries do not create duplicate chunks.
6. As a platform operator, I want newer ingestion events to supersede stale lineage versions, so that obsolete work does not overwrite fresher data.
7. As an AI engineer, I want parser adapters behind a common parser port, so that adding new formats does not destabilize orchestration.
8. As an AI engineer, I want canonical parse output regardless of source format, so that chunking and embedding stay format-agnostic.
9. As an AI engineer, I want structural-first chunking with token caps and overlap, so that retrieval quality is high and context windows are respected.
10. As an AI engineer, I want chunk provenance metadata persisted, so that every indexed fragment is traceable to original structure.
11. As an AI engineer, I want language auto-detection with mismatch flagging against declared language, so that cross-language quality issues are surfaced.
12. As an AI engineer, I want quality gates that reject poor chunk sets, so that empty or low-quality indexes do not pollute retrieval.
13. As a retrieval engineer, I want embeddings tied to explicit embedding profiles, so that model upgrades can coexist safely.
14. As a retrieval engineer, I want active-profile-only vector retrieval, so that query vectors and stored vectors remain compatible.
15. As a retrieval engineer, I want BM25 fallback when vector coverage is low, so that relevance degrades gracefully.
16. As a client app developer, I want retrieval responses to include retrieval mode and vector coverage ratio, so that I can expose transparent relevance diagnostics.
17. As a security engineer, I want AI workers to fetch content only via short-lived signed URLs, so that long-lived storage exposure is avoided.
18. As a security engineer, I want signed URL refresh to require trusted internal service identity and document/object mapping validation, so that cross-document access is blocked.
19. As an SRE, I want exhausted retries routed to DLQ with persistent failure context, so that re-drive is auditable.
20. As an SRE, I want ingestion concurrency and retry windows configurable, so that throughput can be tuned without redeploying code.
21. As a product owner, I want ingestion success to be atomic for chunks, embeddings, and completion status, so that users never see partially committed success states.
22. As a compliance stakeholder, I want no raw content snippets in events or logs, so that privacy controls are maintained.

## Implementation Decisions

- Introduce an ingestion orchestration deep module with a simple command interface (`start_ingestion(event)`) that encapsulates state transitions, retry policy, and terminal resolution.
- Implement envelope verification and key rotation handling as a standalone trust module reused by both request-consumption and status-publication flows.
- Implement an idempotent event ledger module for processed event IDs/idempotency keys and supersede checks by logical document lineage.
- Implement parser architecture as a deep module with adapter registry plus shared normalization output (`document_text`, `sections`, normalized metadata).
- Implement chunking and quality gate engine as a deep module with configurable thresholds and deterministic acceptance/rejection outcomes.
- Introduce profile-aware embedding persistence model where chunk vectors are linked to embedding profile metadata and compatibility tags.
- Extend retrieval execution contract so the caller receives retrieval mode and vector coverage diagnostics.
- Implement signed URL acquisition flow via internal service call and refresh semantics with strict authorization checks.
- Implement retry classifier with transient/permanent categories, exponential backoff with jitter, and DLQ sink for exhausted attempts.
- Preserve atomic completion semantics by committing chunks, embeddings, and final completed state in one transactional unit.

## Testing Decisions

- A good test verifies behavior at module boundaries: state transition outputs, persisted records, published status events, and retry outcomes; it does not assert internal helper implementation details.
- Modules to test: orchestration state machine, signature verification module, idempotency/supersede ledger, parser adapters, chunking and quality gate engine, embedding profile selector, retrieval fallback policy, retry classifier, DLQ routing behavior.
- Add deterministic scenario tests for state transitions (queued -> validating -> fetching -> chunking -> embedding -> indexing -> completed, plus failed/cancelled terminals).
- Add contract tests for signed event envelope consumption and status event emission with schema/version compatibility.
- Add repository tests for profile-aware embedding reads/writes and active-profile filtering.
- Add failure-injection integration tests for transient fetch errors, embedding provider errors, and invalid parse outputs.
- Prior art in the repository includes use-case unit tests, repository mapping tests, and RPC contract integration tests; ingestion tests should follow this pyramid by isolating core logic and adding minimal but real integration hops.

## Out of Scope

- OCR support and scanned-image ingestion.
- Advanced semantic reranking improvements beyond current retrieval architecture.
- Automatic model-selection optimization by document type.
- End-user-facing visualization of per-chunk quality diagnostics.

## Further Notes

- This PRD assumes PRD 1 contract and outbox foundations are complete.
- The architecture is intentionally profile-aware to avoid lock-in to a single embedding model.

---

# PRD 3: Ingestion Status Projection, SSE Progress Streaming, and Operational Controls

## Problem Statement

Users and operators currently lack real-time, reliable visibility into ingestion progress. AI processing may happen asynchronously, but core APIs do not expose a robust projection of worker status, and there is no horizontally scalable SSE replay mechanism for reconnecting clients. Operational controls for re-drive and rollout safety are also under-defined without a mirrored status plane.

## Solution

Make the AI service the source of truth for ingestion status events and build a core-side status projection optimized for API reads and SSE delivery. Add Redis-backed multi-instance fan-out, resumable event replay via Last-Event-ID, periodic snapshots, and operational controls for DLQ re-drive auditing and ingestion toggle enforcement.

## User Stories

1. As a platform user, I want to see ingestion progress in near real time, so that I know whether my document is still processing.
2. As a platform user, I want progress updates to continue after reconnecting, so that I do not lose track when my network drops.
3. As a platform user, I want final success/failure status with timestamps, so that I can trust the document state shown in the UI.
4. As a platform user, I want stable stage names across clients, so that progress labels are predictable.
5. As a frontend engineer, I want SSE events that can be resumed with Last-Event-ID, so that reconnect behavior is simple and robust.
6. As a frontend engineer, I want periodic snapshot events in addition to stage transitions, so that progress bars remain smooth.
7. As a backend engineer, I want core APIs backed by a status projection store, so that query performance does not depend on worker internals.
8. As a backend engineer, I want stale status events ignored when superseded, so that clients never regress to older job states.
9. As a backend engineer, I want replay buffering with bounded retention, so that reconnect support does not create unbounded memory growth.
10. As an SRE, I want Redis-backed fan-out across core instances, so that horizontal scaling does not break SSE delivery.
11. As an SRE, I want measurable status event cadence policies, so that event volume remains predictable under load.
12. As an SRE, I want explicit ingestion toggle behavior, so that operational freezes are safe and deterministic.
13. As an SRE, I want DLQ re-drive actions to be core-controlled and audited, so that recovery operations are traceable.
14. As a support engineer, I want quick lookup endpoints for current ingestion status, so that I can answer user tickets immediately.
15. As a support engineer, I want correlation identifiers in status events, so that cross-service tracing is straightforward.
16. As a product manager, I want progress semantics to distinguish active, terminal failed, and cancelled outcomes, so that user messaging is accurate.
17. As a security engineer, I want status payloads to exclude raw document snippets, so that telemetry never leaks content.
18. As a security engineer, I want SSE streams authorized per account/user context, so that tenants cannot observe each other.
19. As a QA engineer, I want deterministic replay and resume behavior for disconnected clients, so that end-to-end validation is repeatable.
20. As an operations lead, I want burst smoke-test confidence on status streaming and projection writes, so that launch risk is reduced.

## Implementation Decisions

- Introduce a status projection deep module in core that consumes status events, applies supersede rules, and stores current state plus progression history metadata for APIs.
- Implement a streaming gateway deep module that translates projection updates into SSE frames, handles replay windows, and applies resumable cursor semantics.
- Use Redis as cross-instance event relay and replay buffer backing to support horizontal core scaling.
- Standardize status event emission policy from AI: stage transitions, terminal states, plus periodic snapshots driven by time and progress delta thresholds.
- Define status event ordering and monotonic update rules so out-of-order delivery cannot regress visible state.
- Add operational control module for ingestion toggle and audited re-drive initiation requests.
- Add projection read endpoints and stream endpoints with consistent filtering by account, user, document, and job identifiers.
- Enforce payload hygiene: status events and logs contain metadata-only progress details.

## Testing Decisions

- A good test confirms user-observable semantics (ordered status progression, reconnect replay correctness, authorization boundaries, toggle behavior), not transport implementation internals.
- Modules to test: status projection consumer, supersede resolver, SSE replay/cursor manager, Redis fan-out adapter, ingestion toggle gate, re-drive audit module.
- Add integration tests for AI status event -> core projection update -> SSE client delivery.
- Add reconnection tests that verify Last-Event-ID replay behavior and bounded replay retention.
- Add ordering tests for out-of-order and duplicate status events to confirm monotonic visibility.
- Add security tests for unauthorized stream attempts and cross-tenant access prevention.
- Add performance smoke tests for burst event throughput and SSE fan-out latency under agreed SLOs.
- Prior art in the repository includes RPC integration contract tests and async use-case tests; status/SSE tests should extend this with event-driven integration fixtures and transport-facing assertions.

## Out of Scope

- Rich UI visualization design for progress pages.
- Historical analytics dashboards for ingestion trends.
- Push-notification channels outside SSE.
- Full incident management workflow tooling beyond audited re-drive controls.

## Further Notes

- This PRD assumes PRD 1 event contracts and PRD 2 status publishing are available.
- Replay retention defaults should be configurable and validated against memory/cost budgets.
