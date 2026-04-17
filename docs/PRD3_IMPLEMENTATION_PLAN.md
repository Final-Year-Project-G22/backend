# PRD3 Implementation Plan

This document tracks the implementation plan for **PRD 3: Ingestion Status Projection, SSE Progress Streaming, and Operational Controls** and maps work to commit-sized slices.

## Scope (PRD3)

**User Stories covered:**
1. Near real-time progress visibility
2. Progress updates after reconnecting
3. Final success/failure status with timestamps
4. Stable stage names
5. SSE with Last-Event-ID resume
6. Periodic snapshot events
7. Core APIs backed by status projection store
8. Stale status events ignored when superseded
9. Replay buffering with bounded retention
10. Redis-backed fan-out across instances
11. Status event cadence policies
12. Ingestion toggle behavior
13. DLQ re-drive core-controlled and audited
14. Quick lookup endpoints
15. Correlation identifiers in status events
16. Progress semantics for active/failed/cancelled
17. Status payloads exclude raw content
18. SSE streams authorized per account/user
19. Deterministic replay and resume
20. Burst smoke-test confidence

**Implementation decisions:**
- Status projection module in core consuming status events
- Streaming gateway for SSE with replay support
- Redis-backed cross-instance fan-out
- Status event ordering and monotonic update rules
- Ingestion toggle operational controls
- DLQ re-drive auditing
- Periodic snapshots driven by time/progress delta
- Payload hygiene enforcement

**Out of scope:**
- Rich UI visualization design
- Historical analytics dashboards
- Push-notification channels outside SSE
- Full incident management workflow

---

## Commit Plan

## Commit 1 - Status projection model, entities, and consumer

Goal:
- Build status projection store that consumes AI status events and applies supersede rules.
- Store current state plus progression history metadata.

User stories: 7, 8, 16, 19

Implementation decisions:
- Status projection deep module in core
- Supersede rules for stale status events
- Monotonic update rules

Primary paths:
- `core-backend/internal/modules/ai/domain/entity/`
- `core-backend/internal/modules/ai/domain/repository/`
- `core-backend/internal/modules/ai/infrastructure/repository/`
- `core-backend/internal/modules/ai/application/service/status_projection_service.go`

Status:
- Completed (Commit: 0bd3f9c).

---

## Commit 2 - Status persistence tables and repository

Goal:
- Add status tables and repository for state + progression history.

User stories: 14, 15

Implementation decisions:
- Status tables with history metadata
- Repository pattern for projection reads

Primary paths:
- `core-backend/internal/modules/ai/infrastructure/repository/`
- `core-backend/internal/modules/ai/domain/entity/`

Status:
- Completed (Commit 3 merged with Commit 2).

---

## Commit 3 - Status API endpoints and ordering rules

Goal:
- Add read endpoints for current status by account/document/user.
- Define status event ordering and monotonic update rules.

User stories: 1, 3, 4, 14

Implementation decisions:
- Projection read endpoints with filtering
- Monotonic visibility enforcement

Primary paths:
- `core-backend/internal/modules/ai/delivery/`

Status:
- Completed (Commit ff66ad8).

---

## Commit 4 - SSE streaming gateway with replay

Goal:
- Implement streaming gateway that translates projection updates into SSE frames.
- Handle replay windows and resumable cursor semantics.

User stories: 2, 5, 19

Implementation decisions:
- Streaming gateway deep module
- Last-Event-ID replay support

Primary paths:
- `core-backend/internal/modules/ai/application/service/sse_gateway.go`
- `core-backend/internal/modules/ai/delivery/`

Status:
- Completed (Commit 979a4bb).

---

## Commit 5 - Redis fan-out and bounded replay buffer

Goal:
- Add Redis-backed event relay and replay buffer for horizontal scaling.

User stories: 9, 10

Implementation decisions:
- Redis as cross-instance event relay
- Bounded retention for replay buffer

Primary paths:
- `core-backend/internal/modules/ai/infrastructure/`

Status:
- Completed (Commit 2e37b27).

---

## Commit 6 - Ingestion toggle operational control

Goal:
- Add ingestion toggle behavior for operational freezes.

User stories: 12

Implementation decisions:
- Operational control module for toggle
- Safe and deterministic toggle behavior

Primary paths:
- `core-backend/internal/modules/ai/application/service/`
- `core-backend/internal/modules/ai/domain/`

Status:
- Completed (Commit b901ba7).

---

## Commit 7 - DLQ re-drive controller with auditing

Goal:
- Add audited re-drive initiation and auditing.

User stories: 13

Implementation decisions:
- DLQ re-drive controller
- Audited recovery operations

Primary paths:
- `core-backend/internal/modules/ai/application/service/`
- `core-backend/internal/modules/ai/infrastructure/`

Status:
- Planned.

---

## Commit 8 - Periodic snapshot emission and cadence policy

Goal:
- Add time/progress delta driven snapshot events.
- Define status event cadence policies.

User stories: 6, 11

Implementation decisions:
- Periodic snapshots driven by time and progress delta thresholds
- Measurable status event cadence

Primary paths:
- `ai-service/core/usecases/`
- `ai-service/core/domain/`

Status:
- Completed (Commit b9a4214 merged with Commit 7).

---

## Commit 9 - Security tests for tenant isolation

Goal:
- Add security tests for unauthorized stream attempts and cross-tenant access prevention.

User stories: 18

Implementation decisions:
- SSE streams authorized per account/user context
- Tenant isolation enforcement

Primary paths:
- `core-backend/internal/modules/ai/`
- `core-backend/tests/`

Status:
- Completed (Commit 0de0703 merged with Commit 9).

---

## Commit 10 - Integration tests and smoke tests

Goal:
- Add integration tests for AI status event -> core projection update -> SSE client delivery.
- Add reconnection tests and performance smoke tests.

User stories: 20

Implementation decisions:
- Integration tests for full event flow
- Reconnection tests for Last-Event-ID replay
- Performance smoke tests for burst throughput

Primary paths:
- `core-backend/tests/integration/`

Status:
- Planned.

---

## Testing Checklist (Per Commit)

- [ ] Ordered status progression tests
- [ ] Reconnect replay correctness tests  
- [ ] Authorization boundaries tests
- [ ] Toggle behavior tests
- [ ] Cross-tenant access prevention tests
- [ ] Event throughput and SSE fan-out latency under SLOs
- [ ] Monotonic visibility for out-of-order events

## Security Checklist (Per Commit)

- [ ] Status payloads exclude raw document snippets
- [ ] SSE streams authorized per account/user
- [ ] No content leakage in telemetry

## Current Execution State

- Commit 1-10: completed.