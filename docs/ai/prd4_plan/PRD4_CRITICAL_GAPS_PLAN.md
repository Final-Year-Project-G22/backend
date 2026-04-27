# PRD-4 Critical Gaps and Mismatches Plan

This document outlines the critical gaps and mismatches found during an implementation audit of PRD-4 (Ask/Chat REST API) and provides a plan to address them in three focused pull requests.

## Critical Gaps and Mismatches Identified

1. **Route Path Parameters and Missing Document-Level SSE**
   - Ask routes use incorrect Huma path syntax (`:sessionId` instead of `{sessionId}`).
   - Ingestion SSE stream by document ID is implemented (`StreamStatusByDocument`) but not exposed via HTTP route.

2. **Streaming Persistence, Custom Title, and Timestamp Issues**
   - AI message is persisted before streaming completes (causing potential divergence).
   - Custom conversation title from `AskRequest.title` is not carried through to conversation creation.
   - Timestamps in conversation responses are set to `time.Now()` instead of using gRPC-provided values.

3. **Missing Read-Through Conversation Cache**
   - Core backend makes a gRPC call for every conversation list/get request, despite having a read-through cache design clarified in discussions.

## Proposed PRs to Address Gaps

### PR A — Route fixes & document-level ingestion SSE
- **Goal**: Correct Huma path syntax and add missing document‑based SSE endpoint.
- **Scope**:
  - Change `:sessionId` and `:documentId` to `{sessionId}` and `{documentId}` in ask routes.
  - Register `StreamStatusByDocument` under `/api/v1/ai/ingestion/status/stream/document/{documentId}`.
- **Out of scope**: Business logic changes, caching.
- **Dependencies**: None.
- **Files**:
  - `core-backend/internal/modules/ai/delivery/routes/ask_routes.go`
  - `core-backend/internal/modules/ai/delivery/routes/status_routes.go`
  - `core-backend/internal/modules/ai/delivery/routes/routes.go`
- **Checkpoints**:
  1. Update path parameters in `ask_routes.go` to use `{...}` syntax.
  2. Add handler func for document‑stream in `status_routes.go` (call `SSEHandler.StreamStatusByDocument`).
  3. Register the new route in `routes.go`.
  4. Add unit tests for path parsing.
- **Acceptance Criteria**:
  - All ask/conversation routes resolve without 404.
  - GET `/api/v1/ai/ingestion/status/stream/document/{id}` returns SSE stream with `Content-Type: text/event-stream`.

### PR B — Streaming persistence, custom title, timestamp fix
- **Goal**: Ensure AI message is persisted only after stream completes, allow custom title from AskRequest, and return gRPC‑provided timestamps.
- **Scope**:
  - Modify `AIInferenceService.AskStream` to defer persistence until after the stream finishes.
  - Pass through optional `title` from AskRequest to conversation creation.
  - Update `AskHandler` to use timestamps from gRPC responses (Ask, AskStream, List/Get/Archive).
- **Out of scope**: DLQ/toggle, ingestion streams.
- **Dependencies**: PR A (for route stability) but can be developed in parallel.
- **Files**:
  - `ai-service/infrastructure/rpc/services/inference_service.py`
  - `ai-service/core/usecases/ask_ai.py`
  - `core-backend/internal/modules/ai/delivery/dto/ask_dto.go` (add Title field)
  - `core-backend/internal/modules/ai/delivery/handler/ask_handler.go`
  - `core-backend/internal/modules/ai/application/service/ask_service.go` (if needed)
- **Checkpoints**:
  1. Add `Title` field to AskRequest DTO and map it through to gRPC AskRequest.
  2. Adjust `AskAIUseCase._resolve_conversation` to use provided title or fallback to prompt[:80].
  3. Refactor `AIInferenceService.AskStream` to:
     - Create session and persist user message before streaming.
     - Stream LLM chunks via `generate_stream`.
     - After stream ends, persist AI message with full answer and citations.
  4. Update `AskHandler.HandleAsk` and `HandleAskStream` to map `SessionID`, `CreatedAt`, `UpdatedAt` from gRPC responses.
  5. Adjust List/Get/Archive handlers to use timestamps from gRPC.
- **Tests**:
  - Unit: Mock LLM `generate_stream` to verify chunk ordering and that persistence occurs after final chunk.
  - Unit: Verify custom title is stored and returned.
  - Unit: Verify timestamps in Ask/AskStream/List/Get/Archive responses equal those from gRPC.
  - Integration: End‑to‑end stream request → verify SSE events → verify persisted conversation has correct timestamps and title.
- **Acceptance Criteria**:
  - Non‑stream Ask returns answer with correct timestamps and optional title.
  - Stream Ask yields SSE `chunk`, `citations`, `done` events; after stream closes, conversation is stored with AI message containing full answer and correct timestamps.
  - Custom title provided in request appears in conversation list/get.

### PR C — Read‑through conversation cache
- **Goal**: Add a read‑through cache layer in core‑backend Ask flow so repeated conversation reads are served from cache while still going to AI service for writes.
- **Scope**:
  - Introduce a `ConversationCache` interface (get/set) backed by existing `core.Cache`.
  - Wrap `AIInferencePort` methods `ListConversations`, `GetConversation` (and optionally `ArchiveConversation` invalidation) with cache look‑through.
  - On write (create/archive), update or invalidate cache entries.
- **Out of scope**: Changing AI service, modifying proto.
- **Dependencies**: PR B (so caching layer sits atop stable AskService).
- **Files**:
  - `core-backend/internal/modules/ai/domain/port/conversation_cache.go` (new)
  - `core-backend/internal/modules/ai/application/service/ask_service.go` (inject cache, wrap calls)
  - `core-backend/internal/modules/ai/infrastructure/client/grpc_client.go` (no change; used via port)
  - `core-backend/internal/modules/ai/module.go` (provide cache instance)
- **Checkpoints**:
  1. Define `ConversationCachePort` with `Get`, `Set`, `Invalidate`.
  2. Implement `cachingConversationService` struct that delegates to underlying `AIInferencePort` and caches read results.
  3. Update `AskService` to accept the caching wrapper (or create it internally) and use it for List/Get.
  4. On `ArchiveConversation`, invalidate the cached session.
  5. Wire `core.Cache` via fx in `module.go` and provide the caching wrapper as `AIInferencePort`.
- **Tests**:
  - Unit: Verify cache hit/miss for List/Get.
  - Unit: Verify cache invalidation on archive.
  - Integration: Make two identical ListConversations requests; second should be faster (mock timing) and not hit gRPC.
- **Acceptance Criteria**:
  - Repeated conversation list/get calls hit cache after first population.
  - Writes (create/archive) bypass cache or update it appropriately.
  - No regression in Ask/AskStream functionality.

## Recommended Order
1. PR A (route fixes) – safe, unlocks correct endpoint usage.
2. PR B (streaming & title) – depends on stable routes.
3. PR C (caching) – depends on PR B’s AskService interface.

## Test Strategy Across the Epic
- **Contract tests**: Already exist for protobuf; ensure any new fields (title) are validated.
- **gRPC service tests**: Updated in PR B to verify title handling and deferred persistence.
- **Handler HTTP tests**: Added in PR A (routing) and PR B (request/response mapping, SSE).
- **Cache tests**: Unit and integration in PR C.
- **End‑to‑end**: Full Ask/AskStream flow validated in PR B (SSE events + persisted data).

## Open Questions Remaining
- Should the conversation cache be LRU or TTL‑based? (Decision: use existing `core.Cache` with configurable TTL.)
- On stream failure part‑way, should we persist a partial AI message? (PRD implies only on clean completion; we will not persist on error.)

--- 
*This plan is ready to be passed to implementation.*