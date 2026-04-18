# ADR-002: Ask/Chat REST API — Admin Test Panel and User Chat Interface

- Status: Proposed
- Date: 2026-04-17
- Decision Makers: Core Backend + AI Service

## Context

Core backend handles authentication, storage, and event publishing. AI service owns retrieval, chunking, embedding, and inference. The document ingestion pipeline (ADR-001) is complete, giving users a searchable knowledge base.

Two distinct user personas need access to the AI layer:

**Admin (Web):** Uploads documents and needs to monitor ingestion health — real-time progress, failed events, and operational controls.

**User (Mobile):** Queries the knowledge base via natural language, manages conversation history, and receives streamed AI responses.

The AI service already has the Ask use case, conversation management, and LLM adapters (including streaming). Core backend already has the gRPC client for the unary Ask RPC. What is missing is the REST API bridge in core backend and the streaming/conversation gRPC methods in AI service.

## Decision

### 1) Communication Pattern

- Web → Core Backend: REST/HTTP
- Core Backend → AI Service: gRPC (unary for Ask and conversation management, server-side streaming for AskStream)
- AI Service → Core Backend → Client: SSE over HTTP chunked transfer (for streaming Ask)
- Ingestion events: RabbitMQ (existing, unchanged)

AI service owns conversation data (sessions, messages) and exposes it via gRPC. Core backend bridges to REST for web and mobile clients.

### 2) Proto Separation

A new `AIConversationService` proto is introduced, separate from the existing `AIInferenceService`.

Reason: `AIInferenceService` is focused on the Ask/RAG operation. Conversation CRUD (list, get, archive) is a different concern. Separating them allows independent versioning and keeps each service focused.

The existing `AIInferenceService` is extended with an `AskStream` RPC.

### 3) Streaming Transport

Ask streaming uses gRPC server-side streaming. AI service implements `AskStream` RPC, streaming chunks as the LLM generates tokens. Core backend reads the gRPC stream and concurrently writes SSE events to the HTTP response. A single connection handles the full flow with minimal latency.

RabbitMQ is not used for streaming — it adds unnecessary complexity and latency for a synchronous operation.

### 4) SSE Event Format

The streaming endpoint emits three event types:

- `chunk` — partial text tokens as they arrive from the LLM
- `citations` — full citation list sent at the end, after all chunks
- `done` — final metadata (latency, model, token usage) signaling clean stream completion

Error cases emit a single `error` event before closing the stream. Client treats an abrupt connection close as an implicit error.

### 5) Conversation Persistence in Streaming

The session and user message are created eagerly — before any streaming begins — so the conversation appears in the sidebar immediately. The AI message with the full answer is persisted when the stream finishes.

This gives the user immediate feedback that their message was received while the response streams in.

### 6) Conversation Archival

Archive is a soft delete. The `deleted_at` timestamp is set on the session. Archived sessions are excluded from the conversation list but preserved for audit and recovery.

### 7) Message Pagination

`GetConversation` accepts `limit` and `offset` for messages. The response includes `total_messages` so clients know when to stop fetching. Frontend loads the most recent messages first (offset=0) and fetches older messages on scroll-up.

### 8) Language Parameter

The `language` parameter is optional. If provided, it filters documents by language. If omitted, English is used as the default. This gives users control while defaulting to a sensible value.

### 9) Document Scoping

Ask queries are not scoped to user accounts or business profiles for this ADR. All indexed documents are queryable by any user. Account/business-profile scoping will be added in a future ADR.

### 10) Citation Format

Citations are lean: `chunk_id`, `document_id`, `title`, `score`, and `source_type`. The chunk text is not included in the response. This keeps payloads small. A future ADR will define richer citation formats (article number, page reference, section heading) for the mobile app.

### 11) Custom Conversation Titles

New conversations may include an optional `title` field in the request. If omitted, the title is auto-generated from the first 80 characters of the query. This gives flexibility while defaulting to a useful title.

### 12) Quota Enforcement

AI service is the authoritative quota enforcement point. Core backend handles the `RESOURCE_EXHAUSTED` gRPC error and maps it to HTTP 429. No separate pre-check RPC is added — the error handling on the main call is sufficient and avoids race conditions.

### 13) Admin Endpoint Wiring

The ingestion status SSE, DLQ re-drive, and ingest toggle infrastructure was implemented in PRD3 but not wired to HTTP routes. These are wired in this ADR as they are needed by the admin web panel.

## REST API Summary

### Admin (Existing Infrastructure, Wiring Only)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/ai/ingestion/status/stream` | SSE stream by document or account ID |
| GET | `/api/v1/ai/ingestion/dlq/events` | List dead-letter queue events |
| POST | `/api/v1/ai/ingestion/dlq/events/:id/redrive` | Re-drive single DLQ event |
| POST | `/api/v1/ai/ingestion/dlq/events/batch/redrive` | Batch re-drive DLQ events |
| GET | `/api/v1/ai/ingestion/toggle` | Get ingestion toggle state |
| PATCH | `/api/v1/ai/ingestion/toggle` | Enable/disable ingestion |

### User (New)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/ai/ask` | Non-streaming Ask |
| POST | `/api/v1/ai/ask/stream` | Streaming Ask via SSE |
| GET | `/api/v1/ai/conversations` | List user's conversations |
| GET | `/api/v1/ai/conversations/:id` | Get conversation with paginated messages |
| DELETE | `/api/v1/ai/conversations/:id` | Archive a conversation |

## Consequences

### Positive

- Web admin panel gets real-time ingestion monitoring and operational controls
- Users get a full chat interface with conversation history and streaming responses
- Proto separation keeps inference and conversation concerns independent
- Eager session persistence gives immediate UI feedback during streaming
- Soft-delete archive preserves data for audit

### Negative

- Two proto files to maintain instead of one
- SSE streaming requires careful connection management and error handling
- Conversation data now lives in two databases (core-backend and ai-service), increasing operational complexity

### Risks

- LLM provider rate limits or outages affect streaming quality; error event handling must be robust
- Long streaming connections may hold HTTP connection pools; connection timeouts must be tuned
- Future account/profile scoping will require changes to the search path in AI service
