# Notification Orchestration Migration Design

## Status

- Scope: design only (no implementation in this change)
- Goal: migrate existing module-level notification sends (especially IAM email flows) into the notification module using a single canonical event contract
- Out of scope: payment module implementation (planned later)

## Problem Statement

The backend currently has a completed notification module, but module integrations are inconsistent:

- Existing publishers use mixed payload shapes (`id`, `account_id`, etc.)
- Notification ingest currently relies on event-name mapping and ad-hoc parsing
- Existing email flows (registration/OTP/reset-like flows) are not yet consistently routed through notification orchestration

We need one stable, typed contract and one operational flow for user-facing notifications before expanding to remaining modules.

## Objectives

1. Use notification module as the single sender/orchestrator for migrated user-facing notifications.
2. Standardize module-to-notification integration on a canonical event envelope.
3. Preserve delivery safety with hard validation, idempotency, and reliable publishing.
4. Support both single-channel and all-enabled-channel policies in a clear way.
5. Keep migration simple for pre-production (no feature-flag matrix required).

## Canonical Language (Agreed)

- **Notification Type**: business intent (`welcome_message`, `account_verification`, etc.)
- **Channel**: transport (`in_app`, `email`, `push`, `sms`)
- **Channel Policy**: `single` or `all_enabled`
- **All Enabled Channels**: channels enabled/eligible for user and configured by template/type
- **Idempotency Key**: deterministic key for dedupe/retry safety
- **Critical Account Alert**: mandatory, bypasses opt-out and quiet hours
- **Informational Account Alert**: preference-controlled, quiet-hours-aware

## Migration Scope (Now)

### In scope

- IAM integration to notification module for:
  - `account.registered` -> `welcome_message`
  - `user.email_otp_requested` -> `account_verification`
  - `password.reset_requested` -> `password_reset` (event to be introduced if not present)
  - `account.alert` -> `account_alert_critical` or `account_alert_info`

### Out of scope

- Payment module integration
- Full DLQ replay tooling for invalid canonical payloads (minimal reject handling only)

## Channel Policy Rules

### Allowed values

- `single`
- `all_enabled`

### Contract constraints

- `single` -> `channel` is required
- `all_enabled` -> `channel` must be absent
- For `single`, the requested channel must exist in template-configured channels or ingest rejects

### Type-level channel defaults (current decisions)

- `welcome_message`: typically `all_enabled`
- `account_verification`: `single` + `email`
- `password_reset`: `single` + `email`
- `account_alert_critical`: `all_enabled`
- `account_alert_info`: start with `in_app + email`

## Canonical Event Envelope

All publishers must emit this shape (camelCase only):

```json
{
  "schemaVersion": "1.0.0",
  "eventType": "account.registered",
  "occurredAt": "2026-04-30T12:34:56Z",
  "sourceModule": "iam",
  "accountId": "uuid",
  "notificationType": "welcome_message",
  "channelPolicy": "all_enabled",
  "channel": "email",
  "variables": {
    "platformName": "Acme",
    "accountName": "Jane",
    "gettingStartedUrl": "https://..."
  },
  "metadata": {
    "idempotencyKey": "welcome:<accountId>",
    "locale": "en",
    "traceId": "..."
  }
}
```

Notes:

- `channel` appears only for `single`
- `variables` remains `map[string]string`
- `metadata` has required minimum keys plus optional extension keys

## Validation Model

### Ingest behavior

- Canonical-only ingest (legacy payload parsing removed during migration)
- Hard reject invalid payloads; no partial enqueue

### Reject handling (minimal v1)

- Structured log + metrics only
- No dedicated reject table yet
- No replay tool yet

### Standard reason codes

- `missing_required_field`
- `invalid_schema_version`
- `invalid_account_id`
- `invalid_channel_policy`
- `missing_single_channel`
- `forbidden_channel_for_all_enabled`
- `unknown_notification_type`
- `template_variable_missing`
- `invalid_idempotency_key`

## Reliability Model

### Publisher reliability

- Use shared notification outbox table (notification-specific, separate from AI outbox)
- Outbox write must be in same domain transaction as source business change
- Outbox payload stores full canonical envelope

### Dispatcher behavior

- Publish canonical envelope as-is to broker
- Retry with exponential backoff + jitter
- Baseline attempts: 8
- Backoff baseline: `1m, 2m, 4m, 8m, 16m, 30m, 30m, 30m`
- Exhausted rows move to dead-letter status for ops review

### Dedupe

- Deterministic idempotency keys by intent:
  - `welcome:<accountId>`
  - `verify-email:<accountId>:<otpRecordId>`
  - `password-reset:<accountId>:<resetTokenId>`
  - `account-alert:<accountId>:<alertId>`

## Alert Type Policy

Two separate notification types are used with one source event family (`account.alert`):

- `account_alert_critical`
  - always on (not user-disableable)
  - bypasses quiet hours
  - typically `all_enabled`
- `account_alert_info`
  - preference-controlled
  - respects quiet hours
  - initial channels: `in_app + email`

IAM chooses which type to emit; notification module does not infer severity.

## Required IAM Alert Taxonomy

`alertCode` is required and controlled by IAM enum. Initial set:

- `new_device_login`
- `password_changed`
- `email_changed`
- `mfa_enabled`
- `mfa_disabled`
- `failed_login_threshold`
- `account_locked`
- `suspicious_location_login`

## Template Governance

- System-managed templates remain fully immutable for now
- If system templates are insufficient, create new non-system templates and switch via explicit `notificationType`
- Auth-critical templates must remain strictly non-promotional

## Channel Eligibility

- `email`: require valid destination and policy eligibility; auth-sensitive flows use verified/trusted source
- `push`: require active device token
- `sms`: require normalized verified phone (when enabled later)
- `in_app`: available by account identity and policy
- Under `all_enabled`, ineligible channels are skipped, not global failure

## Ordering and Delivery Outcomes

- Global strict ordering is not required
- Delivery outcomes are channel-independent
- Must record distinct outcomes:
  - delivered
  - skipped (policy/preference/eligibility/expiry)
  - failed (attempted but not successful)

## Metrics and Observability

Notification module owns canonical metric naming for notification ingest and dispatch outcomes.

Track at minimum:

- ingest accepted/rejected counts by reason code
- outbox pending/retry/dead-letter counts
- per-type and per-channel delivered/skipped/failed counts
- queue lag and retry rates

## Migration Order

1. `welcome_message`
2. `account_alert` family
3. `account_verification`
4. `password_reset`

No feature flags are required for this pre-production migration; migrate step-by-step.

## Cutover Acceptance Criteria

Before declaring legacy direct-email paths retired:

1. No direct IAM sends remain for migrated intents.
2. All migrated intents publish canonical events through notification outbox.
3. Ingest shows zero canonical validation rejects over agreed validation window.
4. Security flows meet delivery SLO and never send stale expired content.
5. Duplicate-send rate is zero per idempotency key.
6. Operational runbook documents dead-letter/retry and reason codes.

## Implementation Checklist (Next Execution Pass)

1. Add shared canonical event package and validator.
2. Add shared notification outbox entity/repository/dispatcher.
3. Wire IAM publishers to build canonical envelope and write outbox rows atomically.
4. Remove/replace legacy notification ingest payload mapping with canonical-only parsing.
5. Add/seed alert split notification types and templates.
6. Implement/confirm channel policy validation (`single` vs `all_enabled`).
7. Add structured reject logs/metrics for invalid payloads.
8. Remove legacy direct email paths for migrated intents.
