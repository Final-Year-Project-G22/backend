# PRD-4: Ask/Chat REST API — Admin Test Panel and User Chat Interface

## Problem Statement

The document ingestion pipeline (PRD-1 through PRD-3) is complete. Users can upload documents which are parsed, chunked, embedded, and indexed. However, there is no way for:

- **Admins** to monitor ingestion health in real time, view failed events, re-drive them, or toggle ingestion on/off
- **Users** to query the knowledge base using natural language, manage their conversation history, or receive streamed AI responses

AI service has the full RAG pipeline (Ask, hybrid search, conversation management, LLM adapters with streaming support). Core backend has the gRPC client for unary Ask. What is missing is the REST API bridge and the streaming/conversation gRPC methods.

## Solution

### Admin Side

Wire existing ingestion infrastructure (implemented in PRD-3 but not exposed) to HTTP endpoints:

- SSE streaming for real-time ingestion progress by document or account
- DLQ event listing and re-drive (single and batch)
- Ingestion toggle (enable/disable globally)

### User Side

Build the Ask/Chat REST API:

- Non-streaming Ask: POST a question, receive full RAG-powered answer with citations in one response
- Streaming Ask: POST a question, receive tokens via SSE as the LLM generates them
- Conversation management: list past conversations, get a conversation with paginated messages, archive a conversation

AI service owns conversation data and exposes it via gRPC. Core backend bridges to REST.

## User Stories

### Admin — Ingestion Monitoring

1. As an admin, I want to see real-time ingestion progress for uploaded documents via SSE, so I can monitor batch uploads without polling or refreshing.
2. As an admin, I want to reconnect to an SSE stream and resume from where I left off using Last-Event-ID, so I don't miss updates during brief disconnections.
3. As an admin, I want to view the list of dead-letter queue (DLQ) events with pagination, so I can see what failed and why.
4. As an admin, I want to re-drive a single DLQ event, so I can recover from transient failures without re-uploading the document.
5. As an admin, I want to batch re-drive multiple DLQ events at once, so I can recover from widespread transient failures efficiently.
6. As an admin, I want to see the full event payload of a DLQ event, so I can diagnose the root cause of the failure.
7. As an admin, I want to enable or disable ingestion globally via a toggle, so I can halt processing during maintenance or incidents.
8. As an admin, I want to see which documents are queued, processing, completed, or failed, so I can track the overall health of the ingestion pipeline.

### User — Ask

9. As a user, I want to send a question and receive an answer from the AI, so I can search and query my uploaded documents using natural language.
10. As a user, I want to see which documents and chunks were used to generate the answer, so I can verify the sources and explore them further.
11. As a user, I want to receive my answer token-by-token in real time via streaming, so the experience feels responsive and faster than waiting for the full response.
12. As a user, I want to specify a language for my query, so I can search documents in a language I know they are written in.
13. As a user, I want to control how many context chunks are retrieved (topK), so I can balance answer quality against response speed.

### User — Conversations

14. As a user, I want to continue a conversation across multiple messages, so I can ask follow-up questions that maintain context from earlier exchanges.
15. As a user, I want to see a list of all my past conversations with their titles and timestamps, so I can find and resume a previous topic.
16. As a user, I want to open a conversation and see the message history, so I can review what was discussed previously.
17. As a user, I want to scroll back through older messages in a long conversation with pagination, so I can revisit earlier context without loading everything at once.
18. As a user, I want to archive a conversation, so I can declutter my conversation list while preserving the data.
19. As a user, I want to start a new conversation with an optional custom title, so I can organize topics meaningfully.
20. As a user, I want to see a cached response if I ask the same question again within the cache window, so I get faster answers and use fewer tokens.

### User — Errors and Limits

21. As a user, I want to receive a clear error message when my daily query limit is exceeded, so I understand why my request was rejected and when I can try again.
22. As a user, I want to receive a clear error message when the AI provider fails, so I understand what went wrong.
23. As a user, I want to receive a streamed error event if something goes wrong mid-stream, so I get feedback rather than a silent connection drop.

## Implementation Decisions

### 1. Communication Architecture

- Web/Mobile → Core Backend: REST over HTTP
- Core Backend → AI Service: gRPC (unary for Ask and conversation management, server-side streaming for AskStream)
- AI Service → Core Backend → Client: SSE over HTTP chunked transfer for streaming Ask
- Ingestion events: RabbitMQ (existing, unchanged)

### 2. Proto Separation

A new `AIConversationService` proto is introduced, separate from the existing `AIInferenceService`. `AIInferenceService` is extended with an `AskStream` RPC. Separating conversation CRUD from the inference operation allows independent versioning and keeps each service focused.

### 3. Streaming Transport

Ask streaming uses gRPC server-side streaming. AI service implements `AskStream` RPC, streaming chunks as the LLM generates tokens. Core backend reads the gRPC stream and concurrently writes SSE events to the HTTP response. Single connection, minimal latency. RabbitMQ is not used for streaming.

### 4. SSE Event Format

The streaming endpoint emits three event types:

- `chunk` — partial text tokens as they arrive: `{"text": "partial..."}`
- `citations` — full citation list sent at the end: `{"citations": [...]}`
- `done` — final metadata on clean completion: `{"latencyMs": 2340, "model": "..."}`

On error: a single `error` event: `{"code": "RESOURCE_EXHAUSTED", "message": "Daily limit exceeded"}`

An abrupt connection close is treated as an implicit error by the client.

### 5. Conversation Persistence in Streaming

The session and user message are created eagerly — before any streaming begins. The conversation appears in the conversation list immediately. The AI message with the full answer is persisted when the stream finishes.

### 6. Message Pagination

`GetConversation` accepts `limit` and `offset` for messages. The response includes `total_messages` so clients know when to stop fetching. Frontend loads most recent messages first (offset=0) and fetches older messages on scroll-up.

### 7. Conversation Archival

Archive is a soft delete. `deleted_at` is set on the session. Archived sessions are excluded from the conversation list but preserved for audit and recovery.

### 8. Language Parameter

The `language` parameter is optional on Ask. If provided, it filters documents by language. If omitted, English is used as the default.

### 9. Document Scoping

Ask queries are not scoped to user accounts or business profiles. All indexed documents are queryable by any user. Account/business-profile scoping will be added in a future PRD.

### 10. Citation Format

Citations are lean: `chunk_id`, `document_id`, `title`, `score`, and `source_type`. Chunk text is not included in responses. This keeps payloads small. Future PRD will define richer citation formats for the mobile app.

### 11. Custom Conversation Titles

New conversations may include an optional `title` field. If omitted, the title is auto-generated from the first 80 characters of the query.

### 12. Quota Enforcement

AI service is the authoritative quota check. `RESOURCE_EXHAUSTED` gRPC errors are mapped to HTTP 429 in core backend. No separate pre-check RPC is added.

### 13. Admin Endpoint Wiring

The SSE gateway, DLQ controller, and ingest toggle were implemented in PRD-3 but not wired to HTTP routes. These are wired as part of this PRD since the admin web panel depends on them.

### 14. REST API Endpoints

#### Admin

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/ai/ingestion/status/stream` | SSE stream by document or account ID |
| GET | `/api/v1/ai/ingestion/dlq/events` | List DLQ events with pagination |
| POST | `/api/v1/ai/ingestion/dlq/events/:id/redrive` | Re-drive single DLQ event |
| POST | `/api/v1/ai/ingestion/dlq/events/batch/redrive` | Batch re-drive DLQ events |
| GET | `/api/v1/ai/ingestion/toggle` | Get ingestion toggle state |
| PATCH | `/api/v1/ai/ingestion/toggle` | Enable/disable ingestion |

#### User

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/ai/ask` | Non-streaming Ask |
| POST | `/api/v1/ai/ask/stream` | Streaming Ask via SSE |
| GET | `/api/v1/ai/conversations` | List user's conversations |
| GET | `/api/v1/ai/conversations/:id` | Get conversation with paginated messages |
| DELETE | `/api/v1/ai/conversations/:id` | Archive a conversation |

### 15. Proto Contracts

#### New: AIConversationService

Unary RPCs: `ListConversations`, `GetConversation`, `ArchiveConversation`. `GetConversation` accepts message `limit` and `offset`. Responses include `total_messages`.

#### New: AskStream in AIInferenceService

Server-side streaming RPC. Streams `TextChunk`, `CitationsChunk`, `DoneChunk`, or `ErrorChunk` messages.

### 16. Modules to Build or Modify

#### Core Backend

- `AskHandler` — HTTP handlers for Ask, AskStream, conversation CRUD
- `AskService` — application service bridging REST to gRPC
- `AskDTOs` — request/response DTOs for all Ask endpoints
- `SSEWriter` — utility to write SSE events to HTTP response with framing and flushing
- `SSEHandler`, `DLQHandler`, `ToggleHandler` — handlers for existing admin infrastructure
- `AdminRoutes` — new route file for SSE, DLQ, toggle endpoints
- `AskRoutes` — new route file for Ask and conversation endpoints
- `AIInferencePort` — extend with ListConversations, GetConversation, ArchiveConversation, AskStream methods
- `gRPCClient` — implement new gRPC methods

#### AI Service

- `AIConversationService` (gRPC) — wraps `ConversationUseCase` for list, get, archive
- `AskStream` (gRPC) — streams `LLMPort.generate_stream` tokens via gRPC server-side streaming
- `container` — wire `AIConversationService` with dependencies

#### Proto

- `proto/ai/conversation/v1/service.proto` — new: `AIConversationService`
- `proto/ai/inference/v1/service.proto` — add: `AskStream` RPC and chunk message types

### 17. Error Handling

- `RESOURCE_EXHAUSTED` from gRPC → HTTP 429
- `INVALID_ARGUMENT` from gRPC → HTTP 400
- `INTERNAL` from gRPC → HTTP 500
- Mid-stream errors → SSE `error` event, then close

## Testing Decisions

### What Makes a Good Test

- Test external behavior only — HTTP request/response for handlers, gRPC request/response for services
- Do not test internal implementation details (e.g., do not test that chunks are streamed in a specific order internally, test that the SSE events arrive in the expected sequence)
- Mock at dependency boundaries (gRPC client, LLM port, repository)
- Use the same patterns as existing tests in the codebase

### Modules to Test

1. **`AskHandler`** (core-backend) — mock gRPC client, verify HTTP request/response mapping, error propagation
2. **`AskStream` handler** (ai-service) — mock LLM `generate_stream`, verify chunk sequencing, citation emission, done/error events
3. **`AIConversationService`** (ai-service) — mock `ConversationUseCase`, verify gRPC request/response mapping
4. **`SSEWriter`** (core-backend) — verify SSE framing, flush behavior, error handling
5. **DLQ handler** — mock `DLQController`, verify re-drive request/response mapping

### Prior Art

- `test_inference_service.py` in ai-service: tests gRPC handler with mocked use case
- `test_ask_handler.py` pattern in core-backend: HTTP handler with mocked gRPC client
- Existing SSE tests: verify event format and sequence
- Existing integration tests for gRPC contract end-to-end

## Out of Scope

- User profile / business profile injection into the LLM prompt context
- Account / business-profile scoped document filtering in search
- Streaming response citations rendered progressively during streaming
- Conversation title editing by user
- Answer regeneration or editing past AI responses
- User feedback or rating on answers
- Push notifications beyond SSE
- Analytics dashboards
- Historical ingestion analytics
- Rich citation format for mobile (article number, page reference, section heading)
- Per-user rate limiting at core-backend level
- Conversation history search

## Further Notes

- Feature flags (`AI_ASK_ENABLED`) should gate the Ask endpoints in both core-backend and ai-service for safe rollout
- Long streaming connections hold HTTP connection pools; connection timeouts in the server must be tuned appropriately
- Streaming SSE for conversation list updates (real-time sidebar) is deferred — polling is acceptable for the conversation list
- The conversation data lives in ai-service's database. Core backend has no local copy. This is intentional — conversation data belongs to the domain that manages it
- All endpoints require Bearer token authentication (reuse existing auth middleware)
