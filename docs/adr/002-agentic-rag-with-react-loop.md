# ADR 002: Agentic RAG with ReAct Loop

## Status
Proposed

## Date
2026-05-26

## Context

The current AI feature uses simple RAG: the user query is embedded, hybrid search (vector + BM25) retrieves chunks from the knowledge base, and the LLM generates an answer in a single pass. The LLM has no ability to decide which information sources to consult or to gather additional context beyond what the initial search returns.

Meanwhile, the infrastructure for tool calling already exists but is unwired:
- `ToolHandler` interface and `ToolRegistry` in core-backend (Go) with 4 registered tools
- `AIToolGrpcClient` in the AI service (Python) that can list and execute remote tools
- `ToolUseChunk`/`ToolResultChunk` stream event types in the proto
- The system prompt instructs the LLM to use `search_knowledge_base`, `find_template`, and `check_guide_progress` tools — but these are not registered tools, and the LLM receives `tools=None` at call time

The current "auto-search" pattern (regex-driven `search_guides` call before the LLM run) is fragile and not genuinely agentic — the LLM doesn't decide to search; the backend decides for it.

Users frequently ask questions requiring multi-source synthesis: "What licenses do I need for a restaurant, and can you find me the application template?" — which needs KB search + template lookup + guide progress check.

## Decision

We will implement **Agentic RAG** using a server-driven ReAct (Reason-Act-Observe) loop in the AI service. The following specific decisions were reached through a 19-question grill session:

### Strategy Pattern (Q1-Q3)

Two `AskStrategy` implementations coexist:
- `SimpleAskStrategy` — existing behavior: hybrid search → single LLM call → answer. Used for providers without tool support (Ollama) and as a fallback.
- `AgenticAskStrategy` — ReAct loop with tool calling. Used for Cohere and Gemini.

The `AskAIUseCase` selects the strategy based on the `strategy` field in the gRPC request (`"simple"` | `"agentic"`) and the provider's tool-calling capability. A feature flag (`AI_AGENTIC_ENABLED`) gates global availability.

### ReAct Loop Mechanics (Q4, Q5)

- Server-driven loop: the AI service owns execution. Client receives stream events.
- Max 5 iterations, with forced finalization if the cap is reached.
- Early exit when the LLM produces a final answer without calling more tools.
- Parallel execution when the LLM returns multiple independent tool calls in one turn.
- 1 retry for transient failures (timeout, connection); no retry for domain-logic failures (empty result, auth error). Agent always produces some answer — graceful degradation.

### Tool Architecture (Q6-Q10)

Seven tools total, split by ownership:

**Local (AI service / Python):**
1. `search_knowledge_base` — hybrid vector+BM25 search
2. `search_trusted_web` — fetch from hardcoded domain whitelist, no external search API

**Remote (core-backend / Go, via gRPC):**
3. `search_guides` — search business formalization guides
4. `get_user_profile` — business profile (sector, region, stage, tags, locale)
5. `find_template` — document template retrieval
6. `get_guide_progress` — guide step completion status
7. `check_compliance_status` — compliance entry expiry dates

All tools available to all user tiers. The regex-based auto-search for guides is removed.

### Tool Registry (Q7, Q8)

A unified `ToolRegistry` in the AI service:
- At startup, fetches remote tool definitions from `AIToolGrpcClient.ListTools()`, filtered to an explicit config list
- Registers local tools with their schemas and handler functions
- Caches tool definitions with a configurable TTL refresh
- Dispatches execution: local tools run directly, remote tools call `AIToolGrpcClient.ExecuteTool()`
- Formats all results as concise text summaries for the LLM

### Pre-Fetch and Intent Classification (Q11, Q12)

An embedding-based `IntentClassifier` classifies queries into `knowledge`, `personal`, or `mixed`:
- Pre-computed intent centroids from seed queries, compared via cosine similarity
- `knowledge` → pre-fetch `search_knowledge_base` only
- `personal` → pre-fetch `get_user_profile` + `get_guide_progress` + `check_compliance_status`
- `mixed` → pre-fetch all
- Pre-fetch runs concurrently with prompt building; results are injected into the initial LLM prompt

### Streaming Architecture (Q13, Q14)

The `AskStreamChunk` proto gains a `ThinkingChunk` type. Two streaming modes:
- **User endpoint** (`/api/v1/ai/ask/stream`): status-level events only — thinking indicator, tool call started, tool call completed, text chunks, citations, done. No raw reasoning text.
- **Admin debug endpoint** (`/api/v1/ai/ask/stream/debug`, gated by `iam.admin.read`): full ReAct loop internals — raw reasoning text, complete tool arguments and results, per-iteration latency. Uses the same gRPC RPC with `debug_mode: true`.

### Conversation Persistence (Q15)

Tool calls are stored as `ToolCallRecord` entries nested in a `tool_calls` JSONB field on the `AIChatMessage` AI_RESPONSE. When building context for follow-up turns, summarized tool history (50-100 tokens per turn) is injected into the prompt — enough for the LLM to know what was previously looked up without re-consuming full chunk text.

### Prompt Management (Q16, Q17)

File-based Jinja2 templates loaded at application startup:
```
prompts/
├── agentic_system.j2      # Full agentic system prompt
├── simple_system.j2       # Full simple RAG system prompt
├── _persona.j2            # Shared: role + domain expertise + locale response rule
├── _guardrails.j2         # Shared: citations, honesty, formatting
├── _tools.j2              # Agentic-only: tool descriptions and usage rules
├── _reasoning.j2          # Agentic-only: think-before-acting instructions
└── tool_history.j2        # Template for formatting summarized tool calls in history
```

Unified prompt content across providers; each LLM adapter handles provider-specific formatting (Cohere's tool-use API, Gemini's function declarations).

### Proto Changes (Q18)

`AskRequest` gains two fields:
- `strategy` (string): `"simple"` or `"agentic"`
- `debug_mode` (bool): enables verbose stream events

`AskStreamChunk` gains:
- `ThinkingChunk` (message): `string text`

### Performance Optimizations (Q19)

- Pre-fetch eliminates one ReAct iteration for common queries
- Parallel tool execution when tools are independent
- Prompt caching (Cohere, Gemini native) for the system prompt
- Pre-fetch results cached (3600s TTL), final answers cached (60s TTL) for duplicate-click protection
- Quota: 1 query = 1 unit regardless of internal tool call count

### Title Generation

After the ReAct loop, a concurrent lightweight LLM call generates a conversation title (5 words) from the user query + tools called. Runs in parallel with answer streaming; zero added latency.

## Considered Options

### Single-pass tool calling
The LLM receives tool definitions and can make one tool call, but cannot chain calls or react to results. Rejected because multi-source queries (KB + guides + compliance) require sequential tool use.

### Client-driven agent loop
The frontend orchestrates the ReAct loop by receiving tool calls, executing them, and sending results back. Rejected because it increases round trips (3+ per turn), exposes tool execution to the client, and duplicates tool registry logic.

### Plan-and-execute
The LLM first writes a full plan, then executes each step. Rejected as overengineered for the domain's typical query complexity (2-3 tool calls max) and adds latency for writing/parsing a plan.

### No pre-fetch, purely on-demand tool calls
Every tool call is LLM-driven. Rejected because ~70% of queries start with KB search, and pre-fetching it saves one full iteration (~500ms-1s) for common queries.

## Consequences

### Positive
- **LLM-driven tool selection** — the model decides what it needs, not heuristic regex
- **Multi-source synthesis** — single query can combine KB, guides, templates, compliance, and web results
- **Coexistence with simple RAG** — strategy pattern allows both paths to live side-by-side; fallback is automatic
- **Transparent debugging** — admin debug endpoint reveals the full decision chain
- **Incremental tool addition** — new tools require only implementing the `ToolHandler` interface in Go

### Negative
- **Increased latency per query** — multi-iteration ReAct means 2-4 LLM calls instead of 1 (mitigated by pre-fetch, parallel execution, prompt caching, and iteration cap)
- **Higher token cost per query** — system prompt + tool definitions + multiple call/response pairs increase total token consumption (mitigated by prompt caching and result summarization)
- **Added complexity** — strategy pattern, intent classifier, tool registry merging, result formatting add ~10-15 new files to the AI service
- **Proto breaking change** — `AskRequest` adds fields, requiring coordinated deployment of AI service + core-backend

### Dependencies
- Requires Cohere or Gemini as the LLM provider (Ollama falls back to simple RAG)
- Requires existing `AIToolGrpcClient` and `ToolHandler` infrastructure (already built)
- Requires new Go tools: `get_user_profile` (iam), `find_template` (library), `get_guide_progress` (guide), `check_compliance_status` (notification)
- Requires Jinja2 dependency in the Python AI service
- Requires DB migration: `tool_calls JSONB` and `agent_strategy VARCHAR(20)` columns on `ai_chat_messages`

## References
- 19-question grill session: agentic-loop-mechanics.md (this directory)
- Protobuf: `proto/ai/inference/v1/service.proto`, `proto/core/ai_tool/v1/tool_service.proto`
- Existing tool infrastructure: `core-backend/internal/modules/ai_tool/`, `core-backend/internal/modules/guide/ai_tools.go`
- Existing AI service: `ai-service/core/usecases/ask_ai.py`, `ai-service/infrastructure/rpc/ai_tool_grpc_client.py`
- CONTEXT.md AI and agentic RAG terms section (added 2026-05-26)
