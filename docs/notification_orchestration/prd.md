# Notification Orchestration PRD

## Problem Statement

As the backend nears completion, notification behavior is fragmented across modules. Some user-facing flows still send notifications directly (especially IAM-driven email flows), while the new notification module already supports queueing, templates, delivery tracking, preferences, and multi-channel behavior. This split causes inconsistent contracts, duplicated logic, and higher risk of missed or duplicate notifications. The team needs one reliable, canonical orchestration approach so all current and future modules can publish notification intents consistently and have the notification module handle delivery.

## Solution

Adopt a canonical notification event envelope and route notification intents from publisher modules (starting with IAM) through a shared notification outbox into the notification module. Enforce hard contract validation, deterministic idempotency, and explicit channel policy semantics. Migrate existing IAM notification flows (welcome, verification OTP, password reset, account alerts) to this model and retire direct send paths for these intents. Keep scope pre-production pragmatic: canonical-only integration, minimal invalid-event handling (logs + metrics), and staged migration by notification type.

## User Stories

1. As a platform user, I want welcome notifications delivered consistently, so that onboarding communications are reliable.
2. As a platform user, I want account verification notifications to arrive promptly, so that I can complete registration without friction.
3. As a platform user, I want password reset notifications to be time-valid and not stale, so that account recovery is trustworthy.
4. As a platform user, I want critical account alerts to bypass quiet hours and opt-out, so that I do not miss urgent security events.
5. As a platform user, I want informational account alerts to respect my preferences and quiet hours, so that I can control non-critical noise.
6. As a platform user, I want in-app and email channels to behave consistently with my settings, so that my notification experience feels predictable.
7. As a module developer, I want one canonical event contract, so that I can publish notification intents without custom payload mapping.
8. As a module developer, I want strict validation feedback when payloads are invalid, so that I can fix integration issues quickly.
9. As a module developer, I want notification type and channel policy semantics to be explicit, so that behavior is deterministic at runtime.
10. As a module developer, I want publisher-owned template variables, so that business context remains in the source module.
11. As a module developer, I want notification orchestration to resolve channel delivery and user eligibility centrally, so that I do not duplicate delivery logic.
12. As a module developer, I want a shared builder for canonical events, so that integration errors are caught earlier.
13. As a security stakeholder, I want auth-critical notifications handled through reliable outbox publishing, so that event loss risk is minimized.
14. As a security stakeholder, I want no raw secrets leaked in notification logs/history, so that sensitive data exposure is reduced.
15. As an operator, I want channel-level outcomes (delivered/skipped/failed), so that I can diagnose notification behavior accurately.
16. As an operator, I want standardized reject reason codes for invalid events, so that alerting and triage are consistent.
17. As an operator, I want bounded retry behavior with dead-letter outcomes, so that failed publishes are observable and controllable.
18. As a product owner, I want account alerts split into critical and informational types, so that policy behavior matches severity.
19. As a product owner, I want to migrate notification intents step-by-step without production flags, so that implementation remains simple pre-launch.
20. As a future module owner (e.g., payment), I want reusable orchestration contracts, so that onboarding new notification intents is straightforward.
21. As an API consumer, I want channel policy rules (`single` vs `all_enabled`) to be unambiguous, so that requests do not produce conflicting behavior.
22. As a QA engineer, I want deterministic idempotency behavior, so that retries do not create duplicate user sends.
23. As a QA engineer, I want canonical field naming conventions (camelCase), so that payload compatibility is easier to validate.
24. As an architect, I want notification orchestration concerns isolated in deep modules, so that business modules remain simple and stable.

## Implementation Decisions

- Canonical notification envelope is required for migrated intents and uses strict camelCase naming.
- Canonical envelope includes required fields: schema version, event identity, source module, account identity, notification type, channel policy, variables, and idempotency metadata.
- Channel policy supports two values only: `single` and `all_enabled`.
- `single` requires an explicit channel; `all_enabled` forbids explicit channel.
- Notification type is explicit in publisher payloads; subscriber inference is deprecated for migrated flows.
- Variables remain string-based for rendering consistency; metadata has required typed keys with optional extension fields.
- Hard validation at ingest rejects invalid payloads and prevents partial enqueue.
- Invalid-event handling is minimal for now: structured logs and metrics only.
- Shared notification outbox is introduced for notification-producing modules, scoped separately from AI ingestion outbox.
- Outbox write is atomic with source module business transaction.
- Outbox payload stores the full canonical envelope.
- Dispatcher publishes canonical envelope as-is, with bounded retries and dead-letter state.
- Deterministic idempotency keys are required and intent-scoped.
- Channel outcomes are independent; no default cross-channel escalation on failure.
- Eligibility checks are enforced per channel; ineligible channels are skipped under multi-channel policy.
- Locale resolution uses publisher-provided locale first, then account default, then system fallback.
- IAM is first migration target; payment integration is intentionally deferred.
- IAM migration order: welcome, account alerts, account verification, password reset.
- Account alerts are split into two notification types:
  - `account_alert_critical` (always-on, bypass quiet hours)
  - `account_alert_info` (preference-controlled, quiet-hours aware, initial channels in-app + email)
- IAM owns severity/type selection for account alert events.
- IAM owns controlled `alertCode` taxonomy for account alerts.
- Auth-critical notification templates remain system-managed and fully immutable for now.
- If system templates are insufficient, new non-system templates are created and selected via explicit notification type changes.
- Legacy direct-send IAM paths for migrated intents are retired after acceptance criteria pass.

## Testing Decisions

- Good tests validate externally observable behavior and contracts, not internal implementation details.
- Contract tests validate canonical envelope acceptance/rejection rules, including channel policy constraints and required fields.
- Integration tests validate end-to-end flow: publisher transaction -> outbox row -> dispatcher publish -> notification ingest -> queueing.
- Usecase tests validate channel eligibility, preference handling, quiet hours behavior, and independent channel outcomes.
- Idempotency tests validate no duplicate sends for retries of same intent key.
- Security-flow tests validate expiry/TTL behavior for verification and reset notifications.
- Alert policy tests validate critical vs informational behavior differences.
- Modules prioritized for tests:
  - Shared notification event contract/builder
  - Shared notification outbox and dispatcher
  - IAM notification publisher adapters
  - Notification ingest validation and policy enforcement
  - Notification delivery outcome handling
- Prior art should follow existing repository style for usecase-level behavior tests, repository integration tests, and outbox/retry patterns used elsewhere in the codebase.

## Out of Scope

- Payment module integration and payment-specific notification intents.
- Full invalid-event replay tooling and dedicated reject persistence tables.
- Broad cross-domain outbox unification beyond notification-producing flows.
- Push/SMS hardening beyond currently agreed channel defaults.
- New user-facing preference UI changes.

## Further Notes

- This PRD is based on the completed notification design interview and reflects intentional pre-production simplifications.
- Design prioritizes deterministic behavior, explicit contracts, and low migration ambiguity over early flexibility.
- Future enhancements (e.g., replay tooling, broader critical-business classification, shared outbox generalization) are intentionally deferred until core migration is stable.
