# PRD-4 PR Implementation Plan

## 1) PR Decomposition Summary

The PRD-4 Ask/Chat API is decomposed into 7 PRs following the vertical slice pattern:

- **Infrastructure first**: Proto contracts and gRPC infrastructure (PRs 1-2)
- **Core service development**: AI service gRPC handlers and streaming (PRs 3-4)
- **Backend bridge**: REST handlers and wiring (PRs 5-6)
- **Integration and polish**: Admin endpoints and feature flag wiring (PR 7)

This separation ensures contract changes are merged before consumer wiring, and each PR can be reviewed and tested independently.

---

## 2) Clarifications Resolved

| Question | Decision |
|----------|----------|
| Conversation data ownership | Core-backend will have a read-through cache; AI service remains the source of truth |
| Feature flags | AI_ASK_ENABLED will gate Ask endpoints in both services |
| Auth middleware | Existing Bearer token auth is sufficient for all endpoints |
| AI service database | SQLAlchemy migrations already in place; conversations table exists |
| Admin role checks | Existing auth middleware; no additional role checks in this plan |

---

## 3) PR Plan

### PR 1 — Proto Contracts: Conversation Service and AskStream

- **Goal**: Define gRPC contracts for conversation CRUD and streaming Ask
- **Scope**: New proto definitions only; no implementation
- **Out of scope**: Handler implementation, REST endpoints, caching
- **Dependencies**: None
- **Files/modules likely touched**:
  - `proto/ai/conversation/v1/service.proto` (new)
  - `proto/ai/conversation/v1/models.proto` (new)
  - `proto/ai/inference/v1/service.proto` (add AskStream)
- **Commit checkpoints**:
  1. Add `AIConversationService` proto with ListConversations, GetConversation, ArchiveConversation RPCs
  2. Add chunk message types to `ai/inference/v1`: TextChunk, CitationsChunk, DoneChunk, ErrorChunk
  3. Add AskStream RPC to AIInferenceService
  4. Generate Go and Python gRPC code
- **Tests**:
  - Contract: Verify proto syntax and validate generated code compiles
- **Acceptance criteria**:
  - All proto files compile with `buf lint`
  - Generated Go code in `core-backend/pb/`
  - Generated Python stubs in `ai-service/grpc_stubs/`
- **Risk level**: Low (contract-only, no runtime changes)
- **Notes**: This PR establishes the contract. All subsequent PRs depend on it.

---

### PR 2 — AI Service: Conversation gRPC Handlers

- **Goal**: Implement AIConversationService gRPC handlers in AI service
- **Scope**: Wrap existing ConversationUseCase with gRPC handlers
- **Out of scope**: Streaming Ask, REST API, caching in core-backend
- **Dependencies**: PR 1 (proto contracts)
- **Files/modules likely touched**:
  - `ai-service/grpc/services/conversation_service.py` (new)
  - `ai-service/app.py` (wire new service)
  - `ai-service/core/usecases/conversation.py` (existing, no changes)
- **Commit checkpoints**:
  1. Scaffold AIConversationService with ListConversations, GetConversation, ArchiveConversation
  2. Map gRPC request/response to/from domain models
  3. Wire service in app container
  4. Add basic handler tests with mocked use case
- **Tests**:
  - Unit: Handler tests with mocked ConversationUseCase
  - Contract: gRPC contract tests verifying request/response mapping
- **Acceptance criteria**:
  - All three RPCs return correct response shapes
  - gRPC server starts and registers the new service
  - Existing ConversationUseCase tests still pass
- **Risk level**: Low (additive, no breaking changes)
- **Notes**: Uses existing ConversationUseCase; only adds gRPC adapter layer.

---

### PR 3 — AI Service: AskStream Implementation

- **Goal**: Implement streaming Ask RPC in AI service
- **Scope**: Server-side streaming RPC that streams tokens as LLM generates them
- **Out of scope**: REST endpoints, core-backend SSE wiring
- **Dependencies**: PR 1 (proto contracts for chunk types)
- **Files/modules likely touched**:
  - `ai-service/grpc/services/inference_service.py` (add AskStream handler)
  - `ai-service/core/ports/llm_port.py` (add generate_stream method if not exists)
  - `ai-service/core/usecases/ask.py` (add streaming use case)
- **Commit checkpoints**:
  1. Add TextChunk, CitationsChunk, DoneChunk, ErrorChunk to inference service
  2. Implement AskStream handler that streams LLM tokens
  3. Handle session creation (eager, before streaming)
  4. Handle citations emission at end of stream
  5. Handle error events mid-stream
  6. Add streaming tests
- **Tests**:
  - Unit: Mock LLM generate_stream, verify chunk sequencing
  - Integration: Test full streaming flow with mocked LLM
  - Error-path: Verify error events are sent on failure
- **Acceptance criteria**:
  - AskStream returns stream of chunks in correct order: text → citations → done
  - Errors during stream emit ErrorChunk and close stream
  - Session created before streaming begins
- **Risk level**: Medium (streaming adds complexity; test thoroughly)
- **Notes**: Ensure proper backpressure handling and connection cleanup on errors.

---

### PR 4 — Core Backend: gRPC Client Extensions

- **Goal**: Extend core-backend gRPC client with new methods
- **Scope**: Add client methods for conversation CRUD and AskStream
- **Out of scope**: REST handlers, route registration
- **Dependencies**: PR 1 (generated gRPC code), PR 2 (AI service handlers ready)
- **Files/modules likely touched**:
  - `core-backend/internal/modules/ai/infrastructure/client/grpc_client.go`
  - `core-backend/internal/modules/ai/domain/port/ai_inference_port.go`
- **Commit checkpoints**:
  1. Add ListConversations, GetConversation, ArchiveConversation to AIInferencePort
  2. Implement gRPC client methods for conversation CRUD
  3. Add AskStream client method with streaming support
  4. Add tests for new client methods
- **Tests**:
  - Unit: Mock gRPC server, verify client method calls
  - Integration: Test against running AI service (optional)
- **Acceptance criteria**:
  - All new client methods compile and have basic test coverage
  - AskStream client returns a stream interface
- **Risk level**: Low (additive client code)
- **Notes**: The AskStream client needs to handle gRPC streaming properly.

---

### PR 5 — Core Backend: Ask REST Handlers

- **Goal**: Implement REST handlers for Ask, AskStream, and conversation endpoints
- **Scope**: HTTP handlers, DTOs, SSE writer utility
- **Out of scope**: Admin endpoints, feature flag wiring
- **Dependencies**: PR 4 (gRPC client ready)
- **Files/modules likely touched**:
  - `core-backend/internal/modules/ai/delivery/handler/ask_handler.go` (new)
  - `core-backend/internal/modules/ai/delivery/dto/ask.go` (new)
  - `core-backend/internal/modules/ai/application/service/sse_writer.go` (new)
  - `core-backend/internal/modules/ai/application/service/ask_service.go` (new)
- **Commit checkpoints**:
  1. Create DTOs for AskRequest, AskResponse, conversation list, messages
  2. Create AskService application service
  3. Implement AskHandler with non-streaming Ask
  4. Implement AskHandler with streaming Ask (SSE)
  5. Implement conversation CRUD handlers (list, get, archive)
  6. Add SSEWriter utility for proper event framing and flushing
  7. Add handler tests with mocked gRPC client
- **Tests**:
  - Unit: Handler tests with mocked gRPC client
  - Integration: Test SSE event format and sequencing
  - Error-path: Test 429, 400, 500 error mapping from gRPC
- **Acceptance criteria**:
  - POST /api/v1/ai/ask returns answer with citations
  - POST /api/v1/ai/ask/stream streams tokens via SSE
  - GET /api/v1/ai/conversations returns paginated list
  - GET /api/v1/ai/conversations/:id returns messages with pagination
  - DELETE /api/v1/ai/conversations/:id archives conversation
- **Risk level**: Medium (core user-facing functionality)
- **Notes**: SSEWriter must handle flush on each event for real-time feel.

---

### PR 6 — Core Backend: Admin Ingestion Endpoints

- **Goal**: Wire existing admin infrastructure to HTTP endpoints
- **Scope**: SSE for ingestion status, DLQ handlers, ingest toggle
- **Out of scope**: Feature flags, new infrastructure
- **Dependencies**: PR 4 (client ready), PR 5 (patterns established)
- **Files/modules likely touched**:
  - `core-backend/internal/modules/ai/delivery/handler/sse_handler.go` (new)
  - `core-backend/internal/modules/ai/delivery/handler/dlq_handler.go` (new)
  - `core-backend/internal/modules/ai/delivery/handler/toggle_handler.go` (new)
  - `core-backend/internal/modules/ai/delivery/routes/admin_routes.go` (new)
  - Existing: `SSEGateway`, `DLQController`, `IngestToggle` (from PRD-3)
- **Commit checkpoints**:
  1. Wire SSEGateway to SSEHandler for document/account streaming
  2. Wire DLQController to DLQHandler for list/redrive endpoints
  3. Wire IngestToggle to ToggleHandler
  4. Add handler tests
- **Tests**:
  - Unit: Handler tests with mocked infrastructure
  - Integration: Test SSE reconnection with Last-Event-ID
- **Acceptance criteria**:
  - GET /api/v1/ai/ingestion/status/stream streams status
  - GET /api/v1/ai/ingestion/dlq/events lists DLQ with pagination
  - POST /api/v1/ai/ingestion/dlq/events/:id/redrive re-drives single
  - POST /api/v1/ai/ingestion/dlq/events/batch/redrive batch re-drives
  - GET /api/v1/ai/ingestion/toggle returns current state
  - PATCH /api/v1/ai/ingestion/toggle sets state
- **Risk level**: Low (wiring existing infrastructure)
- **Notes**: These use existing PRD-3 infrastructure; only exposing via HTTP.

---

### PR 7 — Feature Flags and Integration

- **Goal**: Add feature flags, wire everything together, final integration
- **Scope**: Feature flag wiring, route registration, end-to-end tests
- **Out of scope**: None
- **Dependencies**: PRs 1-6 (all previous PRs)
- **Files/modules likely touched**:
  - `core-backend/internal/modules/ai/module.go` (register routes, feature flag)
  - `ai-service/main.py` or `app.py` (feature flag wiring)
  - Config files for feature flags
- **Commit checkpoints**:
  1. Add AI_ASK_ENABLED feature flag to config
  2. Gate Ask endpoints in handlers with feature flag
  3. Gate gRPC methods in AI service with feature flag
  4. Register all new routes in module
  5. Add integration tests covering full flow
  6. Update README or docs
- **Tests**:
  - Integration: End-to-end test: REST request → gRPC → LLM → response
  - Regression: Existing tests still pass
  - Feature-flag: Test behavior when flag is off
- **Acceptance criteria**:
  - Feature flag disabled → 404 or 503 on Ask endpoints
  - Feature flag enabled → full functionality works
  - All routes registered correctly
  - No regressions in existing functionality
- **Risk level**: Medium (integration across services)
- **Notes**: This is the merging checkpoint. Only merge after all PRs tested.

---

## 4) Recommended Order

| Order | PR | Dependencies |
|-------|-----|--------------|
| 1 | PR 1: Proto Contracts | None |
| 2 | PR 2: AI Conversation gRPC | PR 1 |
| 3 | PR 3: AskStream Implementation | PR 1 |
| 4 | PR 4: Core gRPC Client | PR 1, PR 2 |
| 5 | PR 5: Ask REST Handlers | PR 4 |
| 6 | PR 6: Admin Endpoints | PR 4, PR 5 |
| 7 | PR 7: Feature Flags & Integration | PRs 1-6 |

**Merge strategy**: 
- PRs 1-4 can be developed in parallel (different services)
- PRs 5-6 depend on PR 4 being merged
- PR 7 requires all previous PRs merged

---

## 5) Test Strategy Across the Full Epic

| Test Type | PR(s) | Description |
|-----------|-------|-------------|
| Proto contract validation | PR 1 | `buf lint`, code generation |
| gRPC handler unit | PR 2, PR 3 | Mocked use case/LLM |
| Streaming chunk sequencing | PR 3 | Verify text→citations→done order |
| gRPC client unit | PR 4 | Mocked server |
| REST handler HTTP mapping | PR 5 | Mocked gRPC client |
| SSE event format | PR 5 | Verify chunk, citations, done, error events |
| Error propagation | PR 5 | gRPC error → HTTP status code |
| Admin infrastructure wiring | PR 6 | Mocked DLQ/SSE controller |
| SSE reconnection | PR 6 | Last-Event-ID handling |
| Feature flag off | PR 7 | 404/503 response |
| End-to-end integration | PR 7 | Full REST→gRPC→LLM flow |

---

## 6) Open Questions Remaining

1. **Caching invalidation**: When should core-backend invalidate its read-through cache? (e.g., on archive, new message)
2. **Rate limiting**: PRD mentions quota enforcement by AI service. Should core-backend add rate limiting at HTTP level?
3. **Streaming timeouts**: Have you configured appropriate HTTP connection timeouts for long-running streams?
4. **DLQ payload size**: Should DLQ event payload be truncated in list view, full on detail?
5. **Conversation title auto-generation**: What algorithm for the 80-char default title?

These are low-priority for implementation start but should be clarified before PR 5-7.