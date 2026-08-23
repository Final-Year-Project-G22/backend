# Core Backend Notifications Context

This context defines the shared business language for user-facing notifications emitted by backend modules and delivered through the notification module. It exists to keep event contracts, delivery intent, and migration decisions consistent across modules.

## Language

### Core notification terms

**Notification Type**:
The business meaning of a notification intent, such as welcome, account verification, password reset, or account alert.
_Avoid_: Channel, medium

**Channel**:
The transport used to deliver a notification to a user: in-app, email, push, or SMS.
_Avoid_: Type

**Channel Policy**:
A rule describing whether a notification type targets one explicit channel (`single`) or all enabled channels for a user (`all_enabled`).
_Avoid_: Type policy

**All Enabled Channels**:
Delivery to every channel that is both configured for the notification type and enabled/eligible for the target user.
_Avoid_: Broadcast to everything

### Event and delivery terms

**Canonical Notification Event Envelope**:
A versioned event shape used by publishers that includes event identity, account identity, notification type, channel policy, variables, and metadata.
_Avoid_: Ad-hoc event payload

**Publisher-Owned Variables**:
Business variables required by templates are assembled by the source module that emits the event.
_Avoid_: Notification-enriched business context

**Idempotency Key**:
A deterministic key that uniquely identifies one notification intent so retries do not cause duplicate sends.
_Avoid_: Random dedupe token

**Skipped Delivery**:
A non-error channel outcome where delivery is intentionally not attempted due to policy, preference, eligibility, or expiry.
_Avoid_: Failure

**Failed Delivery**:
A channel outcome where delivery was attempted but did not succeed after retry policy.
_Avoid_: Skipped

### Security notification terms

**Critical Account Alert**:
An account security notification that is mandatory and not user-disableable.
_Avoid_: Optional alert

**Informational Account Alert**:
An account security-related notification that users can control with preferences.
_Avoid_: Mandatory alert

## Relationships

- A **Notification Type** can target one or more **Channels** according to its **Channel Policy**.
- **All Enabled Channels** includes **in-app** by default and filters by user preferences and channel eligibility.
- A **Canonical Notification Event Envelope** includes one **Notification Type**, one target account, and publisher-owned variables.
- One notification intent is protected by one **Idempotency Key**, and each channel can produce independent outcomes: delivered, skipped, or failed.
- **Critical Account Alert** uses always-on delivery policy and bypasses user opt-out preferences.
- **Informational Account Alert** uses standard preference-controlled delivery.

## Example dialogue

> **Dev:** "For password reset, should we send email and push if both are enabled?"
> **Domain expert:** "Yes, because the channel policy is all enabled channels, but only if each channel is eligible for that account and the reset notification is still within TTL."

## Flagged ambiguities

- "notification type" was initially used to mean delivery transports (email/push/SMS); resolved: transports are **Channels**, while **Notification Type** means business intent.
- "account alert" was initially treated as one class of message; resolved into **Critical Account Alert** and **Informational Account Alert** to separate mandatory vs preference-controlled behavior.

### Scheduled alert terms

**Scheduled Alert**:
A user-created notification that fires at a future time, with user-defined title, body, and delivery channel. Supports cancellation and rescheduling. Bound by a pro-tier limit of 3 pending items for non-pro users.
_Avoid_: Reminder, custom notification, personal alert

**Scheduled Alert Template**:
A seeded template that pre-fills the title and body of a Scheduled Alert. Users pick a template (e.g., tax filing, license renewal, custom) and may override the content.
_Avoid_: Preset, example

### Compliance terms

**Compliance Entry**:
A tracked deadline tied to a Business Profile, representing an official registration or license with an expiry date and a user-set reminder window. Examples: tax registration (TIN), trade license, business registration.
_Avoid_: Compliance record, license entry, deadline item

**Compliance Type**:
A seeded classification of compliance entries (e.g., `tax_registration`, `trade_license`, `business_registration`). Extensible by adding new seed rows.
_Avoid_: Category, kind

**Business Alert**:
A system-generated notification triggered when a Compliance Entry's expiry date falls within its configured reminder window. Delivered via the standard notification pipeline (queue → history + inbox).
_Avoid_: Compliance notification, auto-reminder

**Compliance Calendar**:
A read-only view showing upcoming Compliance Entry deadlines and active Scheduled Alerts on a timeline. Displayed as a widget on the Home dashboard and as a full view inside the Notifications tab.
_Avoid_: Deadline dashboard, compliance timeline

## Relationships (additions)

- A **Compliance Entry** belongs to exactly one **Business Profile**.
- A **Compliance Entry** has one **Compliance Type**.
- A **Business Alert** is triggered by a **Compliance Entry** reaching its reminder window.
- A **Scheduled Alert** is optionally based on a **Scheduled Alert Template**.
- A **Scheduled Alert** targets exactly one **Channel**.
- A **Compliance Calendar** aggregates **Compliance Entries** and **Scheduled Alerts** into a unified timeline.

### Localization terms

**User Locale**:
A user's language preference stored on their account profile. Controls the language of all API responses (error messages, success messages) and notification delivery (email, in-app, push). The system supports English (`en`) and Amharic (`am`).
_Avoid_: Language, region setting

**Request Locale**:
The locale detected from an incoming API request, derived from the `Accept-Language` header. Used to resolve error and success messages for that request when no user session is available.
_Avoid_: Request language, locale param

**Canonical Locale Resolution**:
A single middleware layer that extracts the request locale, stores it in the request context, and makes it available to all downstream handlers and error adapters. Eliminates ad-hoc locale extraction spread across packages.
_Avoid_: Multiple getLocale() paths

**Localized Message**:
A user-facing string (error or success) resolved at request time against the user's locale, using a dot-notated key (e.g., `notification.errors.scheduledAlertPastDue`) and the translation file for the detected locale. Falls back to English when a locale or key is missing.
_Avoid_: Hardcoded English string, raw message

**Notification Locale**:
The locale carried inside a **Canonical Notification Event Envelope**, set by the publisher from the target user's locale at the time of event creation. Ensures that asynchronously delivered notifications reach the user in their preferred language.
_Avoid_: Language field, locale metadata

## Relationships (localization)

- A **User** has one **User Locale**.
- A **Request Locale** is determined per API call; it may differ from the **User Locale**.
- All API responses use **Localized Messages** resolved against the **Request Locale**.
- A **Canonical Notification Event Envelope** carries a **Notification Locale** set from the target user's **User Locale**.
- The **Canonical Locale Resolution** middleware populates the **Request Locale** and supersedes all prior ad-hoc extraction points.

## Example dialogue

> **Dev:** "When we send a scheduled alert notification, what locale should the email use?"
> **Domain expert:** "The notification locale in the envelope — which was set from the user's stored locale at the time the event was published. If the user changed their preference afterward, it takes effect on the next notification."

## Flagged ambiguities

- "language" was initially used interchangeably with "locale"; resolved: **locale** is the canonical term, encompassing both language and regional formatting conventions.
- "notification language" was used to mean both the user's display preference and the template output language; resolved: **User Locale** governs account-level preference, **Notification Locale** is the resolved locale carried in the event envelope.

### Permission and authorization terms

**Permission Code**:
A dot-notated string that identifies one discrete action in the system, such as `library.read`, `iam.admin.list`, or `guide.write`. Permission codes follow the pattern `module.action` — where `module` is the owning module's namespace. Permissions are the atomic unit of authorization.
_Avoid_: Right, scope, privilege

**Module Permission Namespace**:
The prefix of a permission code that identifies the owning module (e.g., `library.*` for the library module, `iam.*` for the IAM module). Each module defines its own namespace, keeping authorization concerns co-located with the code they protect.
_Avoid_: Permission category, permission group

**Seed Permission**:
A permission registered at application startup by a module through the FX group injection mechanism (`group:"permission_seeds"`). Upserted into the permissions table by the Role Permission Seeder on every start. Exists for the lifetime of the application.
_Avoid_: Dynamic permission, runtime permission

**Global Role**:
A role entity stored in the IAM module that can hold **Permission Codes** from any module. Roles are not scoped to a single module. Assigning a **Global Role** to an account grants all linked permissions across all modules.
_Avoid_: Module role, scoped role

**System Role**:
A non-mutable **Global Role** seeded at startup, such as `super_admin` or `iam_admin`. Cannot be deleted or renamed through the API.
_Avoid_: Default role, built-in role

**Custom Role**:
A mutable **Global Role** created through the admin API. Can be assigned any subset of seeded permissions from any module.
_Avoid_: User role, custom permission set

**Role-Permission Assignment**:
The mapping between a **Global Role** and a **Permission Code** via the `role_permissions` join table. Determines what permissions are granted to any account that holds the role.
_Avoid_: Role permission link, grant record

**Role Bypass**:
A middleware optimization that skips the permission lookup for accounts holding an enumerated set of role codes (e.g., `super_admin`). Used when the role implicitly grants all applicable permissions.
_Avoid_: Super admin bypass, role skip

**Idempotent Permission Seed**:
A seed operation that upserts a **Seed Permission** only when its code does not already exist in the database. Prevents duplicate permission rows across restarts.
_Avoid_: Create-or-skip, first-run seed

**Super Admin Role**:
The **System Role** with code `super_admin`. Receives every **Seed Permission** from every module automatically via the seeder. Uses **Role Bypass** in all middleware instances to avoid per-request permission checks.
_Avoid_: Root, admin, global admin

**IAM Admin Role**:
The **System Role** with code `iam_admin`. Receives an explicit subset of `iam.*` permissions. Does not receive permissions from other modules (library, community, guide, notification).
_Avoid_: Identity admin, user admin

## Relationships (permissions)

- A **Permission Code** belongs to exactly one **Module Permission Namespace**.
- A **Seed Permission** is defined by the module that owns its namespace and is provided via `group:"permission_seeds"`.
- A **Global Role** aggregates **Permission Codes** through **Role-Permission Assignments**.
- A **System Role** is seeded at startup; a **Custom Role** is created at runtime.
- The **Super Admin Role** automatically receives all **Seed Permissions** from all modules; the **IAM Admin Role** receives only `iam.*` permissions.
- A **Role Bypass** is a middleware-level shortcut that avoids checking individual **Permission Codes** by verifying the account's role against an allow-list.

## Example dialogue

> **Dev:** "Should I add `community.read` as a permission on the community admin list endpoint?"
> **Domain expert:** "Yes — every module's admin endpoints should be gated by its own namespace. That way a `community_admin` role can read communities without also getting IAM read access."

## Flagged ambiguities

- "permission" was initially used to mean both the code string and the database entity; resolved: **Permission Code** refers to the string constant, **Seed Permission** refers to the seed-time registration.
- "role" was initially considered module-scoped; resolved: roles are **Global Roles** — a single role can authorize actions across library, community, and IAM modules.
- "super admin bypass" was initially treated as specific to IAM routes; resolved: **Role Bypass** is a pattern used uniformly across all modules, passing `["super_admin"]` as the allowed bypass roles.

### AI and agentic RAG terms

**Agentic RAG**:
A retrieval-augmented generation architecture where the LLM autonomously decides if, when, and which tools to invoke during a query. Uses a ReAct (Reason-Act-Observe) loop with a bounded iteration limit. Contrasts with simple RAG, which always retrieves and generates in one pass.
_Avoid_: Tool-calling RAG, function-calling RAG

**ReAct Loop**:
A reasoning cycle where the LLM alternates between reasoning about what to do next and acting by calling AI Tools, observing results, until it produces a final answer. Capped at a configurable maximum number of iterations (default: 5) with forced finalization if the cap is reached.
_Avoid_: Agent loop, tool loop, reasoning cycle

**Ask Strategy**:
The approach used to process a user query. Two strategies exist: **Simple Ask** (basic RAG with hybrid search, no LLM-driven tool autonomy) and **Agentic Ask** (ReAct loop with LLM-driven tool selection). Selected per-request via the API, falling back to simple when the configured LLM provider does not support tool calling.
_Avoid_: Mode, pipeline type, inference mode

**AI Tool**:
A callable function that the LLM can invoke as part of a ReAct loop. Each tool has a name, description, and JSON parameter schema consumed by the LLM for function calling. Tools are either **local** (defined and executed in the AI service, e.g., knowledge base search, trusted web search) or **remote** (defined and executed in core-backend Go modules and called via gRPC, e.g., guide search, profile lookup, compliance check).
_Avoid_: Function, plugin, capability, skill

**Tool Registry**:
A unified registry in the AI service that merges local tool definitions with remote tool definitions discovered from core-backend via `AIToolGrpcClient.ListTools()`. Exposes a combined tool list to the LLM and dispatches execution to the correct handler (local function or remote gRPC call). Remote tools are fetched at startup and refreshed on a TTL.
_Avoid_: Tool manager, function registry, tool directory

**Intent Classifier**:
A lightweight pre-processing step that classifies a user query into one of three intent categories — `knowledge` (factual information), `personal` (user-specific status), or `mixed` (both) — using cosine similarity against pre-computed embedding centroids. Determines which tools are pre-fetched before the first LLM call.
_Avoid_: Query router, intent detector, query classifier

**Tool Pre-Fetch**:
A performance optimization where certain tool results (e.g., knowledge base search) are computed before the first LLM call and injected into the initial prompt. Eliminates one full ReAct iteration for common queries where the LLM would have called that tool as its first action. Only one round of pre-fetching is performed per query.
_Avoid_: Speculative execution, eager evaluation, pre-emptive tool call

**Thinking Chunk**:
A streaming event containing the LLM's internal reasoning or chain-of-thought text. Visible only in admin/debug streaming mode to provide transparency into the agent's decision-making process. Not exposed on the user-facing endpoint.
_Avoid_: Reasoning event, chain-of-thought chunk, plan chunk

**Trusted Web Search**:
An AI Tool that fetches content only from a curated registry of official Ethiopian government pages, organized by topic area, over a whitelist of verified official domains — with no external search API dependency. The registry is strict: the model may fetch only registered URLs (per locale), each carrying a freshness timestamp ("As of") and a per-topic fallback for when the source is unreachable. Runs locally in the AI service via httpx.
_Avoid_: Web search, internet lookup, external search, free-form URL fetch

**Tool Call Record**:
A structured entry stored inside an AI Response message that captures each tool invocation during a ReAct loop: tool name, arguments, result summary, success/failure status, execution time, and iteration number. Stored as JSONB and loaded as summarized context for multi-turn conversations.
_Avoid_: Tool invocation, function call log

**Query Broadening**:
A follow-up knowledge-base search within one turn whose query widens the original question (e.g., "PLC registration process" to "Ethiopian government business registration process") while preserving the turn's entity and intent. Justified broadenings execute; their hits join the turn's citations.
_Avoid_: Query expansion, search widening

**Query Drift**:
A follow-up knowledge-base search whose query abandons the turn's entity and intent for a different topic. Drifted calls are suppressed and the model is nudged to answer the original question and invite a new message for the drifted topic.
_Avoid_: Topic shift, off-topic search

**Tool Call Suppression**:
The system-level skip of a duplicate or drifted knowledge-base search during a ReAct loop, decided by a deterministic guard (embedding cosine similarity against the turn's prompt and prior KB queries, with a token-overlap tiebreak in an ambiguous band). Suppressed calls are not executed, consume no ReAct iteration, persist in the message's tool calls with a `suppressed` flag and reason, and emit a `TOOL_SUPPRESSED` event on the debug stream only.
_Avoid_: Dedup skip, tool blocking

**Debug Streaming**:
An admin-gated variant of the Ask streaming endpoint that exposes the full ReAct loop internals: raw reasoning text (thinking chunks), complete tool arguments and results, and per-iteration latency. Regular user streaming shows only status-level events (tool call started, tool call completed).
_Avoid_: Verbose mode, developer mode

**Prompt Template**:
A Jinja2 file loaded at application startup that defines the system prompt for the LLM. Two top-level templates exist — one for the Agentic Ask strategy (including tool instructions) and one for the Simple Ask strategy — sharing common sections (persona, guardrails, locale rules) via includes. Provider-specific formatting is handled by each LLM adapter.
_Avoid_: System prompt, prompt config, prompt string

### AI quality evaluation terms

**Golden Question**:
A versioned evaluation query with an intended locale, intent, difficulty tier, expected evidence, expected tools, and a reference answer or claims. It is a measurement case, not a user-facing test question.
_Avoid_: Test case, sample prompt

**Evaluation Cell**:
The six Golden Questions sharing one intent and one locale. Results are reported per Evaluation Cell so English/Amharic and knowledge/personal/mixed performance cannot disappear inside one overall average.
_Avoid_: Cohort, bucket

**Difficulty Tier**:
The evidence and reasoning complexity of a Golden Question: easy is direct single-source evidence, medium adds multiple facts or disambiguation, and hard requires multi-step/multi-source reasoning or handles sparse context. Intent does not determine difficulty by itself.
_Avoid_: Severity, priority

**Evaluation Fixture**:
The frozen knowledge corpus and account state used to make personal and mixed Golden Questions reproducible. It is distinct from live or demo data.
_Avoid_: Demo account, production snapshot

**Pass Bar**:
A calibrated quality threshold a candidate AI change must meet, together with the no-regression rule, before it is accepted. Pass Bars are calibrated against the current system and a deliberately broken baseline rather than treated as universal metric constants.
_Avoid_: Score target, benchmark score

## Relationships (AI)

- An **Agentic RAG** query executes a **ReAct Loop** using one **Ask Strategy** (agentic), with tool execution managed by the **Tool Registry**.
- A **ReAct Loop** may invoke multiple **AI Tools**, each producing a **Tool Call Record** stored on the AI Response message.
- An **Intent Classifier** determines the **Tool Pre-Fetch** set for a query before the first LLM call.
- The **Tool Registry** merges local **AI Tools** (knowledge base search, trusted web search) with remote **AI Tools** (guide search, profile lookup, compliance check, template find, guide progress).
- **Debug Streaming** exposes **Thinking Chunks** and full tool details; user-facing streaming exposes only status-level tool events.
- **Prompt Templates** are loaded at startup and rendered per-request with the available tool list, locale, and pre-fetched context.

## Example dialogue

> **Dev:** "When the agent calls search_knowledge_base and get_user_profile in the same ReAct step, do we run them sequentially or in parallel?"
> **Domain expert:** "In parallel — they're independent. The Tool Registry executes them concurrently and feeds all results back to the LLM in the next reasoning step."

> **Dev:** "What happens if the intent classifier misclassifies a query as 'knowledge' when the user actually needs their compliance status?"
> **Domain expert:** "The LLM still has access to all tools via explicit tool calling. The pre-fetch just gives it KB results upfront. If it needs compliance, it calls check_compliance_status during the ReAct loop — there's no penalty for pre-fetching the wrong thing, just a small amount of wasted context."

> **Dev:** "Should the user-facing stream show the LLM's reasoning text?"
> **Domain expert:** "No — that's what debug streaming is for. The user sees 'Searching knowledge base...' as a status pill, not the raw chain-of-thought."

## Flagged ambiguities

- "tool calling" was initially used to mean both the programmatic auto-search pattern (regex-triggered search_guides before the LLM call) and LLM-driven function calling; resolved: programmatic auto-search is removed in favor of genuine LLM-driven tool selection via the ReAct loop.
- "agent" was initially used loosely to mean any AI enhancement; resolved: **Agentic RAG** specifically refers to the ReAct loop architecture with LLM-driven tool selection, distinct from simple RAG.
- "streaming" was overloaded between text streaming and tool event streaming; resolved: text chunks, tool use chunks, tool result chunks, and thinking chunks are distinct event types within the same SSE stream, with visibility controlled by debug mode.

## Uploads

**Upload Intent**:
A grant the backend issues that lets a client transfer a file directly to object storage without routing bytes through the backend. It names the object's storage key, the HTTP method, the headers to send, and an expiry. The URL is not cryptographically signed in the current SeaweedFS implementation — expiry is advisory only.
_Avoid_: Presigned URL (implies a signature the implementation does not produce)

**Direct-to-Storage Upload**:
An upload performed by the client itself: it obtains an **Upload Intent**, PUTs the bytes straight to storage, then calls **Finalize Upload** to register the object. Used for AI knowledge-base documents and library templates.
_Avoid_: Pre-upload

**Backend-Proxied Upload**:
An upload the client performs against the backend (multipart), which stores the bytes itself. Used for community attachments, guide cover images, and IAM avatars/business images.
_Avoid_: Pre-upload attachment (ambiguous with the intent flow)

**Finalize Upload**:
The step that validates an uploaded object exists and records its metadata — for ingestion, atomically with the outbox event that starts the pipeline.
_Avoid_: Complete upload

**Storage Key**:
The object's key in object storage. Module-scoped namespaces: `community/attachments`, `library/templates`, `guides/images`, `business` (logo/banner), `avatars`; ingestion documents use a bare key without a namespace.

**Storage Base URL — Public vs Filer**:
Every stored object has an internal base URL (the filer, used for backend-side operations) and a public-facing base URL used in URLs handed to clients. The public base URL must be absolute and reachable from client devices; when unset it falls back to the internal filer URL.
_Avoid_: Storage URL (ambiguous about which side of the boundary)

## Relationships

- A **Direct-to-Storage Upload** consists of an **Upload Intent** followed by a **Finalize Upload**; the bytes travel client → storage, never through the backend.
- A **Backend-Proxied Upload** stores bytes via the filer base URL and returns a **Storage Key**; clients later reference the object by its public base URL + key.
- **Finalize Upload** succeeds only if the object at the **Storage Key** already exists.
