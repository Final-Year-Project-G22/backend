# Payment Module — Domain Model & Data Design

## Entity Overview

### Payment

Represents a single payment transaction attempt against Chapa.

| Field | Type | Description |
|---|---|---|
| `id` | `uuid.UUID` (PK) | Auto-generated, `gen_random_uuid()` |
| `account_id` | `uuid.UUID` (FK → accounts) | The account making the payment |
| `subscription_id` | `uuid.UUID?` (FK → subscriptions) | Set after subscription is created/updated |
| `tx_ref` | `string` | Backend-generated deterministic reference |
| `chapa_ref` | `string?` | Chapa's internal reference (from init or verify response) |
| `amount` | `int64` | Amount in minor unit (satcker) |
| `currency` | `string` | Always "ETB" for v1 |
| `plan_name` | `string` | Snapshot: "Basic" or "Pro" |
| `plan_period` | `string` | Snapshot: "monthly" or "yearly" |
| `status` | `string` | `pending`, `success`, `failed` |
| `payment_method` | `string?` | "telebirr", "cbebirr", "mpesa", "ebirr" |
| `checked_out_at` | `time?` | When Chapa checkout URL was generated |
| `verified_at` | `time?` | When payment was verified (success or failure confirmed) |
| `failed_at` | `time?` | When payment was marked failed |
| `metadata` | `jsonb` | Raw request/response payloads from Chapa for debugging |
| `created_at` | `time` | Auto |
| `updated_at` | `time` | Auto |
| `deleted_at` | `time?` | Soft delete |

**Indices:**
- `idx_payments_tx_ref` (unique) — `tx_ref`
- `idx_payments_account_id` — `account_id`
- `idx_payments_status` — `status`
- `idx_payments_chapa_ref` — `chapa_ref` (unique, sparse)

**Rules:**
- `tx_ref` is globally unique and deterministic: `tx_{account_id}_{plan_name}_{period}_{timestamp}`
- Once `status` transitions to `success` or `failed`, it is immutable (append-only)
- `metadata` stores the full Chapa initialize request, initialize response, verify response, and webhook payload

### Subscription

Represents the current active (or most recent) subscription for an account.

| Field | Type | Description |
|---|---|---|
| `id` | `uuid.UUID` (PK) | Auto-generated |
| `account_id` | `uuid.UUID` (FK → accounts, unique active) | One active subscription per account |
| `plan_name` | `string` | "Basic" or "Pro" |
| `plan_period` | `string` | "monthly" or "yearly" |
| `amount` | `int64` | Amount paid (minor unit) |
| `currency` | `string` | "ETB" |
| `status` | `string` | `active`, `expired`, `cancelled` |
| `current_period_start` | `time` | When current period began |
| `current_period_end` | `time` | When current period ends |
| `cancelled_at` | `time?` | When user cancelled (access continues until period end) |
| `renewal_count` | `int` | Number of successful renewals (for analytics) |
| `created_at` | `time` | Auto |
| `updated_at` | `time` | Auto |
| `deleted_at` | `time?` | Soft delete |

**Indices:**
- `idx_subscriptions_account_id` (unique where `deleted_at IS NULL AND status = 'active'`) — ensures single active per account
- `idx_subscriptions_status` — for expiry queries
- `idx_subscriptions_account_id_status` — composite for lookups

**Rules:**
- An account can have only **one active subscription** at a time
- Historical subscriptions (expired/cancelled) are retained with `deleted_at` or `status` as archive
- When a new payment succeeds, if there's already an active subscription: extend period (for same plan) or create new subscription entry (for plan change at renewal)

### Plan (Seeded)

Static plans seeded via DB migration. Not a full CRUD entity — read-only via API.

| Field | Type | Description |
|---|---|---|
| `id` | `uuid.UUID` (PK) | |
| `name` | `string` | "Basic" or "Pro" |
| `period` | `string` | "monthly" or "yearly" |
| `amount` | `int64` | Price in minor unit |
| `currency` | `string` | "ETB" |
| `is_active` | `bool` | Whether available for purchase |
| `created_at` | `time` | |
| `updated_at` | `time` | |
| `deleted_at` | `time?` | Soft delete |

**Seed data:**

| name | period | amount | Currency |
|---|---|---|---|
| Basic | monthly | 9900 | ETB |
| Basic | yearly | 95000 | ETB |
| Pro | monthly | 19900 | ETB |
| Pro | yearly | 190000 | ETB |

**Indices:**
- `idx_plans_name_period` (unique where `deleted_at IS NULL`)

## State Diagrams

### Payment Lifecycle

```
                ┌────────┐
                │ CREATED│ (tx_ref generated, DB record inserted)
                └────┬───┘
                     │ initializeTransaction()
                     ▼
                ┌────────┐
                │PENDING │ (Payment link sent to Chapa)
                └───┬────┘
                    │
        ┌───────────┼───────────┐
        ▼           ▼           ▼
   ┌──────────┐ ┌──────────┐ ┌──────────┐
   │ SUCCESS  │ │  FAILED  │ │ CANCELLED│
   │ (verified│ │ (verify  │ │ (status  │
   │  by API) │ │  failed) │ │ returned)│
   └──────────┘ └──────────┘ └──────────┘
        │           │           │
        ▼           ▼           ▼
   [Subscription  [Payment    [Payment
    created/       record     record
    updated]       updated]   updated]
```

### Subscription Lifecycle

```
  ┌───────────┐
  │  NONE     │ (account has no subscription)
  └─────┬─────┘
        │ First successful payment
        ▼
  ┌───────────┐
  │  ACTIVE   │ ──► Expiry/cancellation ──► EXPIRED
  └───────────┘                            │
        │                                    │
        │ New payment                        │
        └────────────────────────────────────┘

  ┌───────────┐
  │ CANCELLED │ (user cancelled, still active until period end)
  └───────────┘
        │
        │ Period ends
        ▼
  ┌───────────┐
  │ EXPIRED   │
  └───────────┘
```

## Entitlement Check

The `getSubscriptionByAccount` repository method returns the current active subscription (if any). Higher-level modules query this to gate feature access:

```
subscription := subscriptionRepo.GetActiveByAccount(ctx, accountID)
if subscription == nil || subscription.Status != "active" || subscription.CurrentPeriodEnd.Before(now) {
    return ErrNotSubscribed
}
planName := subscription.PlanName // "Basic" or "Pro"
```

## Data Storage Decisions

- All monetary values as `INT8` / `int64` (satcker — 1/100 ETB)
- JSONB `metadata` on Payment for audit and debugging
- Timestamps in UTC
- Soft deletes on all entities