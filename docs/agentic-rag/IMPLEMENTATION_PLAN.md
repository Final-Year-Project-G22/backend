# Agentic RAG — Phase-by-Phase Implementation Plan

## Overview

This plan breaks the Agentic RAG implementation into 9 sequential phases. Each phase produces a working, testable increment. Phases 1-6 can be mostly parallelized; phases 7-9 depend on earlier phases.

**Total estimated scope**: ~30 new/modified Python files, ~15 new/modified Go files, 1 proto change, 1 DB migration, prompt templates.

---

## Phase 1: Proto Changes & Core Data Model

**Goal**: Add `strategy`, `debug_mode`, and `ThinkingChunk` to the proto; add `tool_calls` and `agent_strategy` to the DB and domain model. No behavior change yet.

### 1.1 Proto Changes

**File: `proto/ai/inference/v1/service.proto`**
- Add to `AskRequest`:
  - `string strategy = 9` — values: `"simple"` or `"agentic"`, default `"simple"`
  - `bool debug_mode = 10` — default `false`
- Add to `AskStreamChunk.oneof chunk`:
  - `ThinkingChunk thinking = 7`
- Add new message:
  ```protobuf
  message ThinkingChunk {
    string text = 1;
  }
  ```

**Action**: Run `buf generate` to regenerate Go and Python stubs in `core-backend/pb/` and `ai-service/grpc_stubs/`.

### 1.2 DB Migration

**File: new migration SQL**

```sql
ALTER TABLE ai_chat_messages
  ADD COLUMN tool_calls JSONB,
  ADD COLUMN agent_strategy VARCHAR(20) NOT NULL DEFAULT 'simple';
```

### 1.3 Domain Model

**File: `ai-service/core/domain/models.py`**
- Add `ToolCallRecord` dataclass:
  ```python
  class ToolCallRecord(BaseModel):
      tool_name: str
      arguments: dict[str, Any]
      result_summary: str
      success: bool
      error_message: str | None = None
      execution_ms: int = 0
      iteration: int = 1
  ```
- Add to `AIChatMessage`:
  - `tool_calls: list[ToolCallRecord] | None = None`
  - `agent_strategy: str = "simple"`

### 1.4 SQLAlchemy Model

**File: `ai-service/infrastructure/database/models_sqlalchemy.py`**
- Add to `AIChatMessage`:
  - `tool_calls: Mapped[list[dict[str, Any]] | None] = mapped_column(JSONB, nullable=True)`
  - `agent_strategy: Mapped[str] = mapped_column(String(20), nullable=False, default="simple")`

### 1.5 Domain Enum

**File: `ai-service/core/domain/enums.py`**
- Add `AskStrategy` enum: `SIMPLE = "simple"`, `AGENTIC = "agentic"`

### 1.6 Mapper Update

**File: `ai-service/infrastructure/database/repositories/mappers.py`**
- Map `tool_calls` between domain model and SQLAlchemy model

### 1.7 gRPC Server Update

**File: `ai-service/infrastructure/rpc/inference_service.py`**
- Read `strategy` and `debug_mode` from the `AskRequest` and pass through to the use case command

### 1.8 gRPC Client Update (Go)

**File: `core-backend/internal/modules/ai/infrastructure/client/grpc_client.go`**
- Pass `strategy` and `debug_mode` fields in the proto request

### 1.9 Domain Port Update

**File: `core-backend/internal/modules/ai/domain/port/inference.go`**
- Add `Strategy` and `DebugMode` fields to `AskRequest` struct

### 1.10 Application Service Update

**File: `core-backend/internal/modules/ai/application/service/ask_service.go`**
- Thread `strategy` and `debug_mode` from DTO through to the port request

### 1.11 DTO Update

**File: `core-backend/internal/modules/ai/delivery/dto/ask_dto.go`**
- Add `Strategy` and `DebugMode` fields to ask request DTOs

**Verification**: Build both services. Existing tests pass. Simple RAG path works unchanged with `strategy="simple"`.

---

## Phase 2: Prompt Template System

**Goal**: Load Jinja2 templates at startup; render system prompts with tools, locale, and pre-fetched context. Both simple and agentic templates work.

### 2.1 Add Jinja2 Dependency

**File: `ai-service/pyproject.toml`** (or `requirements.txt`)
- Add `jinja2`

### 2.2 Create Template Files

**File: `ai-service/prompts/_persona.j2`**
Role description and domain expertise for Adisu Serategna AI Advisor. Includes locale instruction: respond in the user's locale (`{{ locale }}`).

**File: `ai-service/prompts/_guardrails.j2`**
Citation requirements, honesty (no hallucination), formatting guidelines, uncertainty admission.

**File: `ai-service/prompts/simple_system.j2`**
```
{% include '_persona.j2' %}
{% include '_guardrails.j2' %}
{{ kb_context }}
```
Simple RAG: includes pre-fetched KB context directly. No tool mentions.

**File: `ai-service/prompts/agentic_system.j2`**
```
{% include '_persona.j2' %}
{% include '_tools.j2' %}
{% include '_reasoning.j2' %}
{% include '_guardrails.j2' %}
```
Agentic: tool instructions + reasoning rules.

**File: `ai-service/prompts/_tools.j2`**
Lists available tools with descriptions and when to use each. Dynamic: renders `{% for tool in tools %}`.

**File: `ai-service/prompts/_reasoning.j2`**
Instructions for the ReAct pattern: think before calling, call only when needed, chain calls logically, know when to stop.

**File: `ai-service/prompts/tool_history.j2`**
Template for formatting summarized tool calls from previous turns into the prompt.

### 2.3 Prompt Loader

**File: `ai-service/infrastructure/prompts/__init__.py`** (new directory)
- `PromptLoader` class:
  - Loads all `.j2` files from `prompts/` directory at startup
  - `render_agentic(tools, locale, pre_fetch_context, history) → str`
  - `render_simple(locale, kb_context, history) → str`

### 2.4 Config Update

**File: `ai-service/app/config.py`**
- Remove `AI_PERSONA_SYSTEM_PROMPT` and `AI_RESTRICTIONS` constants (replace with template paths)
- Add `AI_PROMPT_DIR: str = "prompts/"`
- Add `AI_AGENTIC_ENABLED: bool = False` feature flag
- Add `AI_AGENTIC_MAX_ITERATIONS: int = 5`

### 2.5 DI Container Update

**File: `ai-service/app/container.py`**
- Register `PromptLoader` as a singleton

**Verification**: Unit test `PromptLoader` renders all templates without errors. Service starts and loads templates.

---

## Phase 3: Tool Registry

**Goal**: A unified `ToolRegistry` in the AI service that merges local tool definitions with remote tools from core-backend, dispatches execution, and formats results.

### 3.1 Local Tool: search_knowledge_base

**File: `ai-service/infrastructure/tools/local/search_knowledge_base.py`** (new)

```python
class SearchKnowledgeBaseTool:
    name = "search_knowledge_base"
    description = "Search the knowledge base for Ethiopian business regulations, tax codes, licenses, and formalization guides."
    parameter_schema = {
        "type": "object",
        "properties": {
            "query": {"type": "string", "description": "Search query"},
            "top_k": {"type": "integer", "description": "Max results", "default": 5}
        },
        "required": ["query"]
    }

    async def execute(self, arguments: dict, account_id: str, user_id: str) -> ToolResult:
        # Use existing KnowledgeRepository.search_hybrid()
        ...
```

### 3.2 Local Tool: search_trusted_web

**File: `ai-service/infrastructure/tools/local/search_trusted_web.py`** (new)

- Hardcoded domain-to-URL mapping
- Uses httpx to fetch pages from whitelisted domains
- Extracts text with BeautifulSoup, chunks, and returns top chunks

### 3.3 Tool Registry Implementation

**File: `ai-service/infrastructure/tools/tool_registry.py`** (new)

```python
class ToolRegistry:
    def __init__(self, grpc_client: AIToolGrpcClient, settings: AppSettings):
        self._local_tools: dict[str, LocalTool] = {}
        self._remote_tools: dict[str, RemoteToolDef] = {}
        self._grpc_client = grpc_client
        self._expected_tools = settings.AI_EXPECTED_REMOTE_TOOLS
        self._refresh_ttl = settings.AI_TOOL_REFRESH_TTL_SEC

    async def initialize(self):
        self._register_local(SearchKnowledgeBaseTool())
        self._register_local(SearchTrustedWebTool())
        await self._refresh_remote()

    def get_tool_definitions(self) -> list[dict]:
        """Returns combined tool list for the LLM"""
        ...

    async def execute_tool(self, name, arguments, account_id, user_id) -> str:
        """Dispatches to local or remote executor, returns formatted result text"""
        ...

    async def _refresh_remote(self):
        """Calls ListTools(), filters to expected, caches definitions"""
        ...
```

### 3.4 Port Definition

**File: `ai-service/core/ports/tool_registry.py`** (new)

```python
class ToolRegistryPort(ABC):
    @abstractmethod
    async def get_tool_definitions(self) -> list[ToolDefinition]: ...
    @abstractmethod
    async def execute_tool(self, name, arguments, account_id, user_id) -> str: ...
```

### 3.5 Config Addition

**File: `ai-service/app/config.py`**
- `AI_EXPECTED_REMOTE_TOOLS: list[str]` = `["search_guides", "get_user_profile", "find_template", "get_guide_progress", "check_compliance_status"]`
- `AI_TOOL_REFRESH_TTL_SEC: int = 300`

### 3.6 DI Container Update

**File: `ai-service/app/container.py`**
- Register `ToolRegistry` as a singleton, initialize on startup

**Verification**: Unit tests for local tool execution, remote tool definition fetching (mock gRPC), result formatting.

---

## Phase 4: ReAct Loop & Agentic Strategy

**Goal**: The core of the feature — the `AgenticAskStrategy` with full ReAct loop.

### 4.1 Ask Strategy Interface

**File: `ai-service/core/ports/ask_strategy.py`** (new)

```python
class AskStrategyPort(ABC):
    @abstractmethod
    async def execute(self, command: AskCommand) -> AskAIResult: ...
    @abstractmethod
    async def execute_stream(self, command: AskCommand) -> AsyncIterator[AskStreamEvent]: ...

class AskCommand(BaseModel):
    request_id: str
    user_id: str
    account_id: str
    query: str
    language: str
    session_id: str | None
    top_k: int
    strategy: str
    debug_mode: bool
```

### 4.2 Simple Ask Strategy (extract from existing `AskAIUseCase`)

**File: `ai-service/core/usecases/strategies/simple_ask.py`** (new)

- Copy the existing `AskAIUseCase.execute()` logic into `SimpleAskStrategy.execute()`
- Copy `execute_stream()` logic (without the regex auto-search)
- Remove the `generate_stream` for agentic path from existing `AskAIUseCase`
- Keep the pure RAG pipeline

### 4.3 Agentic Ask Strategy

**File: `ai-service/core/usecases/strategies/agentic_ask.py`** (new)

Core `execute()` flow:
```
1. Resolve/create conversation
2. Persist user message
3. Check cache (conversation_id + query[:100])
4. Embed query
5. Intent classification → pre-fetch selection
6. Execute pre-fetches (concurrent)
7. Build initial prompt (system + tools + pre-fetch results + history + query)
8. ReAct loop (max_iterations):
   a. Call LLM (with tool definitions)
   b. If response is final_answer → break
   c. If response is tool_calls:
      - Execute tools (concurrent if multiple)
      - Format results
      - Append to prompt: tool call + tool result
      - Increment iteration
9. If max iterations reached → force finalize
10. Stream/collect final answer
11. Persist AI message with tool_calls + agent_strategy="agentic"
12. Cache response (60s TTL for final answer)
13. Generate title (concurrent)
14. Publish event
15. Return result
```

`execute_stream()` — same flow but yields events at each step.

### 4.4 Stream Event Types

**File: `ai-service/core/domain/stream_events.py`** (new or extend existing)

```python
class AskStreamEventType(StrEnum):
    TEXT = "text"
    THINKING = "thinking"
    TOOL_CALL = "tool_call"
    TOOL_RESULT = "tool_result"
    CITATIONS = "citations"
    DONE = "done"
    ERROR = "error"

class AskStreamEvent:
    type: AskStreamEventType
    data: dict
```

### 4.5 Refactor AskAIUseCase

**File: `ai-service/core/usecases/ask_ai.py`**

- Remove the inline RAG pipeline logic (moved to `SimpleAskStrategy`)
- `AskAIUseCase` becomes a thin orchestrator:
  - Reads `command.strategy`
  - Selects `SimpleAskStrategy` or `AgenticAskStrategy`
  - Falls back to simple if provider doesn't support tool calling (Ollama)
  - Falls back to simple if `AI_AGENTIC_ENABLED` is `False`
  - Delegates to the selected strategy

### 4.6 Tool Result Formatting in Prompt

**File: `ai-service/infrastructure/tools/result_formatter.py`** (new)

- Converts raw tool results to concise text for the LLM prompt
- Handles each tool separately: KB results → formatted chunks, compliance → human-readable lines

**Verification**: Unit tests for the ReAct loop with mocked LLM (fixed tool_call response → assert next LLM call includes tool results). Integration test with real Cohere/Gemini.

---

## Phase 5: Intent Classifier & Pre-Fetch

**Goal**: Classify queries into `knowledge`/`personal`/`mixed` using embedding centroids; pre-fetch appropriate tools.

### 5.1 Intent Classifier Interface

**File: `ai-service/core/ports/intent_classifier.py`** (new)

```python
class IntentClass(StrEnum):
    KNOWLEDGE = "knowledge"
    PERSONAL = "personal"
    MIXED = "mixed"

class IntentClassifierPort(ABC):
    @abstractmethod
    async def classify(self, query_embedding: list[float]) -> IntentClass: ...
```

### 5.2 Embedding-Based Classifier

**File: `ai-service/infrastructure/tools/intent_classifier.py`** (new)

- At startup: compute centroids from ~20-30 seed queries per intent
  - `knowledge` seeds: "business registration process", "tax rates", "license types", "regulation for restaurants"...
  - `personal` seeds: "my profile", "my compliance status", "my guide progress", "what step am I on"...
- At query time: compute cosine similarity of query embedding vs each centroid
- Threshold: score > 0.6 for knowledge, score > 0.6 for personal, both/below → mixed/fallback

### 5.3 Pre-Fetch Pipeline

**File: `ai-service/infrastructure/prefetch/pipeline.py`** (new)

- Accepts intent class → returns pre-fetched results dict
  - `knowledge` → `search_knowledge_base` only
  - `personal` → `get_user_profile` + `get_guide_progress` + `check_compliance_status` (concurrent via ToolRegistry)
  - `mixed` → all four (concurrent)

### 5.4 Integration into AgenticAskStrategy

In `agentic_ask.py`, before the first LLM call:
1. Embed query (already done for RAG anyway)
2. Classify intent
3. Run pre-fetch pipeline
4. Inject pre-fetch results into the initial prompt

### 5.5 Seed Query Configuration

**File: `ai-service/app/config.py`**
- `AI_INTENT_SEED_QUERIES_KNOWLEDGE: list[str]`
- `AI_INTENT_SEED_QUERIES_PERSONAL: list[str]`
- `AI_INTENT_SIMILARITY_THRESHOLD: float = 0.6`

**Verification**: Unit tests for centroid computation, classification accuracy on known queries. Integration test with pre-fetch results appearing in the LLM prompt.

---

## Phase 6: Go-Side Tools

**Goal**: Implement the 4 new remote tools in core-backend following existing `ToolHandler` patterns.

### 6.1 Tool: get_user_profile

**File: `core-backend/internal/modules/iam/ai_tools.go`** (extend existing)

```go
type GetUserProfileTool struct {
    bpRepo repository.BusinessProfileRepository
}

func (t *GetUserProfileTool) Name() string { return "get_user_profile" }
func (t *GetUserProfileTool) Description() string {
    return "Get the user's business profile: sector, region, tags, business stage, and locale."
}
func (t *GetUserProfileTool) ParameterSchema() string {
    return `{"type":"object","properties":{},"required":[]}`
}
func (t *GetUserProfileTool) Execute(ctx context.Context, argsJSON string, accountID, userID uuid.UUID) (string, error) {
    // Fetch business profile(s) for account
    // Return JSON with sector names, region, tags, stage
}
```

Register via `fx.Annotate` → `group:"ai_tool_handlers"`.

### 6.2 Tool: find_template

**File: `core-backend/internal/modules/library/ai_tools.go`** (new)

```go
type FindTemplateTool struct {
    docRepo repository.DocumentRepository  // or existing library usecase
}

func (t *FindTemplateTool) Name() string { return "find_template" }
func (t *FindTemplateTool) Description() string {
    return "Find document templates (e.g., business registration forms, tax filing forms, license applications)."
}
func (t *FindTemplateTool) ParameterSchema() string {
    return `{"type":"object","properties":{"query":{"type":"string","description":"Template search query"},"source":{"type":"string","description":"Optional document source filter"}},"required":["query"]}`
}
func (t *FindTemplateTool) Execute(ctx context.Context, argsJSON string, accountID, userID uuid.UUID) (string, error) {
    // Search library documents marked as templates
    // Return JSON with title, description, download URL
}
```

Register in the library module's FX.

### 6.3 Tool: get_guide_progress

**File: `core-backend/internal/modules/guide/ai_tools.go`** (extend existing)

```go
type GuideProgressTool struct {
    progressRepo repository.GuideProgressRepository  // or existing usecase
}

func (t *GuideProgressTool) Name() string { return "get_guide_progress" }
func (t *GuideProgressTool) Description() string {
    return "Get the user's progress through a business formalization guide: which guide they're on, which steps are completed, and where they might be stuck."
}
func (t *GuideProgressTool) ParameterSchema() string {
    return `{"type":"object","properties":{"guide_id":{"type":"string","description":"Optional specific guide ID"}},"required":[]}`
}
func (t *GuideProgressTool) Execute(ctx context.Context, argsJSON string, accountID, userID uuid.UUID) (string, error) {
    // Fetch guide progress for the account
    // Return JSON with guide name, current step, completed steps, last activity
}
```

Register in the guide module's FX.

### 6.4 Tool: check_compliance_status

**File: `core-backend/internal/modules/notification/ai_tools.go`** (new)

```go
type CheckComplianceStatusTool struct {
    complianceRepo notifrepo.ComplianceEntryRepository
}

func (t *CheckComplianceStatusTool) Name() string { return "check_compliance_status" }
func (t *CheckComplianceStatusTool) Description() string {
    return "Check the expiry status of a user's compliance entries (trade license, TIN, business registration) and their reminder windows."
}
func (t *CheckComplianceStatusTool) ParameterSchema() string {
    return `{"type":"object","properties":{},"required":[]}`
}
func (t *CheckComplianceStatusTool) Execute(ctx context.Context, argsJSON string, accountID, userID uuid.UUID) (string, error) {
    // List compliance entries for account
    // Return JSON with type, expiry date, days remaining, reminder window, status
}
```

Register in the notification module's FX.

### 6.5 Cleanup: Remove or Keep list_sectors/list_tags

These stay registered (harmless) but the AI service doesn't present them to the LLM. No action needed.

**Verification**: `AIToolService.ListTools()` returns all 8 tools (4 existing + 4 new). Unit tests for each tool's Execute. Register in the notification module's FX wiring.

---

## Phase 7: Streaming & API Routes

**Goal**: Wire up streaming events for the ReAct loop in both services; add the admin debug endpoint in core-backend.

### 7.1 gRPC Inference Service (Python)

**File: `ai-service/infrastructure/rpc/inference_service.py`**

- `AskStream`: read `strategy` and `debug_mode` from request
- Select strategy
- Call `strategy.execute_stream(command)`
- Map `AskStreamEvent` → proto `AskStreamChunk`:
  - `TEXT` → `TextChunk`
  - `THINKING` → `ThinkingChunk` (only if `debug_mode`)
  - `TOOL_CALL` → `ToolUseChunk` (name + args if debug, name only if not)
  - `TOOL_RESULT` → `ToolResultChunk` (summary if debug, status only if not)
  - `CITATIONS` → `CitationsChunk`
  - `DONE` → `DoneChunk`
  - `ERROR` → `ErrorChunk`

### 7.2 gRPC Client (Go)

**File: `core-backend/internal/modules/ai/infrastructure/client/grpc_client.go`**

- `AskStream()`: handle new `ThinkingChunk` from proto → map to `port.AskStreamChunk`
- Add `ThinkingChunk` case in the goroutine that reads from the stream

### 7.3 SSE Handler (Go)

**File: `core-backend/internal/modules/ai/delivery/handler/ask_handler.go`** (extend)

- Handle `ThinkingChunk` events (if debug mode, forward as SSE; if not, suppress)
- Handle `ToolUseChunk` and `ToolResultChunk` (already handled but verify)

### 7.4 Admin Debug Route

**File: `core-backend/internal/modules/ai/delivery/routes/ask_routes.go`** (extend)

```go
// POST /api/v1/ai/ask/stream/debug
// Middlewares: AuthMiddleware, AccountStatusMiddleware, PermissionMiddleware(iam.admin.read, [super_admin])
// Passes strategy="agentic", debug_mode=true to AskRequest
```

### 7.5 REST DTO Update (Go)

**File: `core-backend/internal/modules/ai/delivery/dto/ask_dto.go`**
- Add `Strategy` and `DebugMode` fields to request body DTOs

**Verification**: End-to-end stream test with mock LLM. Admin endpoint gated by permission. User endpoint shows status events only.

---

## Phase 8: LLM Adapter Tool Support

**Goal**: Wire tool definitions into each LLM adapter so the LLM actually receives and can call them.

### 8.1 Update LLMPort

**File: `ai-service/core/ports/llm.py`**
- Ensure `ToolDefinition` and `ToolCall` dataclasses exist and match the ToolRegistry output
- Update `generate()` and `generate_stream()` signatures to accept `tools: list[ToolDefinition] | None`

### 8.2 Cohere Adapter

**File: `ai-service/infrastructure/llm/cohere_adapter.py`**
- Map `ToolDefinition` → Cohere's `tools` format in the API request
- Parse `tool_calls` from Cohere's response → `ToolCall` dataclass
- Handle multi-turn tool use (tool results in subsequent messages)

### 8.3 Gemini Adapter

**File: `ai-service/infrastructure/llm/gemini_adapter.py`**
- Map `ToolDefinition` → Gemini's `function_declarations` format
- Parse `function_call` from Gemini's response → `ToolCall` dataclass
- Handle multi-turn function calling

### 8.4 Ollama Adapter

**File: `ai-service/infrastructure/llm/ollama_adapter.py`**
- If `tools` is provided but Ollama model doesn't support tool calling → raise or warn → forces fallback to simple strategy
- No tool calling implementation needed

**Verification**: Unit tests for each adapter: send tool definitions → parse tool call response. Integration test with real Cohere/Gemini: ask a query that triggers a KB search, verify tool is called.

---

## Phase 9: Conversation Context with Tool History

**Goal**: Load summarized tool history from previous turns when building the prompt for follow-up queries.

### 9.1 History Loader

**File: `ai-service/core/usecases/context/history_loader.py`** (new)

```python
class ConversationHistoryLoader:
    def load_with_tool_history(
        self,
        conversation: AIConversationSession,
        max_messages: int = 6,
    ) -> list[dict]:
        """
        Loads the last N messages. For AI_RESPONSE messages with tool_calls,
        prepends a brief tool history summary before the AI text.
        Format: "[Previously: searched KB for 'licenses' (3 results), checked compliance (1 entry)]\nAI: ..."
        """
```

### 9.2 Prompt Building

Integrate `ConversationHistoryLoader` into both `SimpleAskStrategy` and `AgenticAskStrategy` prompt-building steps. The agentic strategy formats tool history using `tool_history.j2`.

### 9.3 Title Generation

**File: `ai-service/core/usecases/strategies/agentic_ask.py`** (integrate)

- After the ReAct loop, as the final answer starts streaming, kick off a concurrent task:
  - Call LLM (small prompt: user query + tools called → "Summarize in 5 words")
  - Update conversation title if it was auto-generated (first 80 chars)

**Verification**: Unit test history loading with tool calls. Integration test: ask a follow-up question after a KB search, verify the LLM prompt includes tool history summary.

---

## Phase 10: Testing, Feature Flag & Rollout

**Goal**: Comprehensive test coverage, feature flag wiring, and deployment preparation.

### 10.1 Unit Tests

**Files to create/extend:**

| Test file | What it tests |
|---|---|
| `tests/unit/core/usecases/strategies/test_simple_ask.py` | Simple strategy correctness (migrate from test_ask_ai.py) |
| `tests/unit/core/usecases/strategies/test_agentic_ask.py` | ReAct loop with mocked LLM — iteration limit, parallel execution, forced finalize, error handling |
| `tests/unit/core/usecases/test_ask_ai.py` | Strategy selection logic, fallback behavior |
| `tests/unit/infrastructure/tools/test_tool_registry.py` | Local/remote merging, execution dispatch, result formatting |
| `tests/unit/infrastructure/tools/test_local_tools.py` | search_knowledge_base, search_trusted_web execution |
| `tests/unit/infrastructure/tools/test_intent_classifier.py` | Centroid computation, classification accuracy |
| `tests/unit/infrastructure/prefetch/test_pipeline.py` | Pre-fetch selection based on intent |
| `tests/unit/infrastructure/llm/test_cohere_tools.py` | Tool definition → Cohere format, response parsing |
| `tests/unit/infrastructure/llm/test_gemini_tools.py` | Tool definition → Gemini format, response parsing |
| `tests/unit/infrastructure/prompts/test_prompt_loader.py` | Template loading, rendering with variables |

### 10.2 Go-Side Unit Tests

| Test file | What it tests |
|---|---|
| `.../iam/ai_tools_test.go` | `get_user_profile` execution |
| `.../library/ai_tools_test.go` | `find_template` execution |
| `.../guide/ai_tools_test.go` | `get_guide_progress` execution |
| `.../notification/ai_tools_test.go` | `check_compliance_status` execution |
| `.../ai/grpc_client_test.go` | Proto field mapping for strategy, debug_mode, ThinkingChunk |
| `.../ai/ask_handler_test.go` | Admin route permission gating |

### 10.3 Integration Tests

| Test file | What it tests |
|---|---|
| `tests/integration/rpc/test_agentic_ask.py` | Full agentic flow with real Cohere/Gemini, real ToolRegistry, real DB |
| `tests/integration/rpc/test_agentic_stream.py` | Streaming agentic flow, event types verified |
| `tests/integration/rpc/test_strategy_fallback.py` | Ollama → simple fallback, disabled flag → simple fallback |

### 10.4 Feature Flag Wiring

**File: `ai-service/app/config.py`**
- `AI_AGENTIC_ENABLED` — gate the entire agentic path. Default `False`.

**File: `core-backend/internal/modules/ai/module.go`** (or config)
- `AI_AGENTIC_ENABLED` — if false, always pass `strategy="simple"` to AI service regardless of client request.

### 10.5 Rollout Plan

1. Deploy with `AI_AGENTIC_ENABLED=false` in both services → no user impact
2. Enable on staging, run integration tests, manual QA with admin debug endpoint
3. Enable on production for admin users first (via debug endpoint + manual toggling)
4. Enable globally with `AI_AGENTIC_ENABLED=true`, monitor latency and error rates
5. If issues: flip flag back to `false` — all users fall back to simple RAG instantly

---

## File Change Index

### AI Service (Python) — New Files

| File | Phase |
|---|---|
| `ai-service/prompts/_persona.j2` | 2 |
| `ai-service/prompts/_guardrails.j2` | 2 |
| `ai-service/prompts/_tools.j2` | 2 |
| `ai-service/prompts/_reasoning.j2` | 2 |
| `ai-service/prompts/agentic_system.j2` | 2 |
| `ai-service/prompts/simple_system.j2` | 2 |
| `ai-service/prompts/tool_history.j2` | 2 |
| `ai-service/infrastructure/prompts/__init__.py` | 2 |
| `ai-service/infrastructure/tools/__init__.py` | 3 |
| `ai-service/infrastructure/tools/tool_registry.py` | 3 |
| `ai-service/infrastructure/tools/local/__init__.py` | 3 |
| `ai-service/infrastructure/tools/local/search_knowledge_base.py` | 3 |
| `ai-service/infrastructure/tools/local/search_trusted_web.py` | 3 |
| `ai-service/infrastructure/tools/result_formatter.py` | 4 |
| `ai-service/infrastructure/tools/intent_classifier.py` | 5 |
| `ai-service/infrastructure/prefetch/__init__.py` | 5 |
| `ai-service/infrastructure/prefetch/pipeline.py` | 5 |
| `ai-service/core/ports/tool_registry.py` | 3 |
| `ai-service/core/ports/ask_strategy.py` | 4 |
| `ai-service/core/ports/intent_classifier.py` | 5 |
| `ai-service/core/usecases/strategies/__init__.py` | 4 |
| `ai-service/core/usecases/strategies/simple_ask.py` | 4 |
| `ai-service/core/usecases/strategies/agentic_ask.py` | 4 |
| `ai-service/core/usecases/context/__init__.py` | 9 |
| `ai-service/core/usecases/context/history_loader.py` | 9 |
| `ai-service/core/domain/stream_events.py` | 4 |
| `ai-service/core/domain/tools.py` | 1 |

### AI Service (Python) — Modified Files

| File | Phase | Changes |
|---|---|---|
| `ai-service/app/config.py` | 2,3,4,5,6,10 | Remove old prompt constants, add new config: prompt dir, agentic enabled, expected tools, intent seeds, max iterations, TTLs |
| `ai-service/app/container.py` | 2,3,4,5 | Register PromptLoader, ToolRegistry, IntentClassifier, strategies |
| `ai-service/core/domain/models.py` | 1 | Add `ToolCallRecord`, `tool_calls` and `agent_strategy` fields |
| `ai-service/core/domain/enums.py` | 1 | Add `AskStrategy` enum |
| `ai-service/core/ports/llm.py` | 8 | Ensure `ToolDefinition`, `ToolCall` types; add `tools` param |
| `ai-service/core/usecases/ask_ai.py` | 4 | Refactor to strategy selector |
| `ai-service/infrastructure/database/models_sqlalchemy.py` | 1 | Add `tool_calls` JSONB, `agent_strategy` column |
| `ai-service/infrastructure/database/repositories/mappers.py` | 1 | Map new fields |
| `ai-service/infrastructure/rpc/inference_service.py` | 1,7 | Read strategy/debug_mode; map stream events to proto chunks |
| `ai-service/infrastructure/llm/cohere_adapter.py` | 8 | Tool definition mapping + tool call parsing |
| `ai-service/infrastructure/llm/gemini_adapter.py` | 8 | Function declaration mapping + function call parsing |
| `ai-service/infrastructure/llm/ollama_adapter.py` | 8 | Reject tool calls gracefully |

### Core Backend (Go) — New Files

| File | Phase |
|---|---|
| `core-backend/internal/modules/library/ai_tools.go` | 6 |
| `core-backend/internal/modules/library/ai_tools_test.go` | 10 |
| `core-backend/internal/modules/notification/ai_tools.go` | 6 |
| `core-backend/internal/modules/notification/ai_tools_test.go` | 10 |

### Core Backend (Go) — Modified Files

| File | Phase | Changes |
|---|---|---|
| `proto/ai/inference/v1/service.proto` | 1 | Add strategy, debug_mode, ThinkingChunk |
| `core-backend/internal/modules/ai/domain/port/inference.go` | 1 | Add Strategy, DebugMode to AskRequest struct |
| `core-backend/internal/modules/ai/infrastructure/client/grpc_client.go` | 1,7 | Pass new fields; handle ThinkingChunk |
| `core-backend/internal/modules/ai/delivery/dto/ask_dto.go` | 1,7 | Add Strategy, DebugMode to DTOs |
| `core-backend/internal/modules/ai/delivery/routes/ask_routes.go` | 7 | Add admin debug route |
| `core-backend/internal/modules/ai/delivery/handler/ask_handler.go` | 7 | Handle ThinkingChunk events |
| `core-backend/internal/modules/ai/application/service/ask_service.go` | 1 | Thread new fields through |
| `core-backend/internal/modules/iam/ai_tools.go` | 6 | Add get_user_profile tool |
| `core-backend/internal/modules/iam/module.go` | 6 | Register get_user_profile with FX |
| `core-backend/internal/modules/guide/ai_tools.go` | 6 | Add get_guide_progress tool |
| `core-backend/internal/modules/guide/module.go` | 6 | Register get_guide_progress with FX |
| `core-backend/internal/modules/library/module.go` | 6 | Register find_template with FX |
| `core-backend/internal/modules/notification/module.go` | 6 | Register check_compliance_status with FX |

### Database

| File | Phase |
|---|---|
| Migration SQL: add `tool_calls JSONB`, `agent_strategy VARCHAR(20)` to `ai_chat_messages` | 1 |

---

## Dependency Graph

```
Phase 1 (Proto + Model) ──► Phase 2 (Prompts) ──► Phase 4 (ReAct Loop) ──► Phase 7 (Streaming + API) ──► Phase 9 (History)
                          │                       │
                          └── Phase 3 (Tool Registry) ──┘
                                                      │
                          Phase 5 (Intent Classifier) ──┘
                          
Phase 6 (Go Tools) ──► Phase 7 (Streaming + API) ──► Phase 8 (LLM Adapters) ──► Phase 10 (Testing + Rollout)
```

- Phases 1 can start immediately
- Phases 2 and 3 can run in parallel after Phase 1
- Phase 4 depends on 2 and 3
- Phases 5 and 6 can run in parallel with 4
- Phase 7 depends on 4 and 6
- Phase 8 depends on 4
- Phase 9 depends on 4
- Phase 10 depends on all preceding phases
