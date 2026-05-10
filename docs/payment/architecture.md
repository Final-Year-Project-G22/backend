# Payment Module — Architecture

## Overview

This document describes the architecture for integrating **Chapa** (https://developer.chapa.co/) as the payment gateway for the Adisu platform. The payment module enables subscription-based access to premium features and content.

## Subscription Model Summary

| Aspect | Decision |
|---|---|
| Tiers | **Basic** (free, no payment required) and **Pro** (paid) |
| Periods | Monthly and Yearly |
| Pricing (minor units) | Basic monthly: 9900, Basic yearly: 95000, Pro monthly: 19900, Pro yearly: 190000 |
| Currency | ETB only |
| Renewal | Manual (user must pay again each period) |
| Changes | Apply at next renewal, no proration |
| Cancellation | Access continues until current period ends |
| Grace period | 3 days after expiry, then access cut immediately |
| Checkout | Native checkout (ETB wallets) primary, web checkout fallback |
| Ownership | Account-level (single active subscription per account) |
| Trial | No free trial for v1 |

## Module Directory Structure

```
payment/
├── architecture.md               ← this file
├── payment-model.md              ← domain model and data design
├── api-contract.md               ← API endpoints and request/response contracts
├── chapa-integration.md          ← pkg/chapa client design and integration
├── database-migration-and-seed.md ← schema and seed data
├── flutter-integration.md        ← mobile app integration design
└── prd.md                        ← Product Requirements Document
```

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Flutter Mobile App                        │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────────┐  │
│  │  GoRouter    │  │  Payment UI   │  │  Chapa Flutter SDK    │  │
│  │  (deep links)│  │  (plan list,  │  │  (native/web checkout)│  │
│  │              │  │   checkout)   │  │                       │  │
│  └──────┬───────┘  └──────┬───────┘  └───────────┬───────────┘  │
│         │                 │                      │              │
└─────────┼─────────────────┼──────────────────────┼──────────────┘
          │                 │                      │
          ▼                 ▼                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Go Backend (core-backend)                    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  HTTP Layer (Huma + Gin)                                │    │
│  │  ┌───────────────────────────────────────────────────┐  │    │
│  │  │  /api/v1/payments/plans          GET    (public)   │  │    │
│  │  │  /api/v1/payments/initiate       POST   (auth)     │  │    │
│  │  │  /api/v1/payments/verify         POST   (auth)     │  │    │
│  │  │  /api/v1/me/subscription          GET    (auth)     │  │    │
│  │  │  /api/v1/webhooks/chapa          POST   (public)   │  │    │
│  │  └───────────────────────────────────────────────────┘  │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Payment Module (internal/modules/payment/)              │    │
│  │  ┌───────────────────────────────────────────────────┐  │    │
│  │  │  Domain Layer                                     │  │    │
│  │  │  ├─ entity/Payment          (GORM model)          │  │    │
│  │  │  ├─ entity/Subscription     (GORM model)          │  │    │
│  │  │  ├─ entity/enums.go         (status constants)     │  │    │
│  │  │  ├─ repository/             (interfaces/ports)     │  │    │
│  │  │  └─ usecase/                (interfaces/ports)     │  │    │
│  │  ├─ Application Layer                                   │  │    │
│  │  │  └─ usecase/               (orchestration)         │  │    │
│  │  ├─ Infrastructure Layer                                │  │    │
│  │  │  └─ repository/             (GORM implementations)  │  │    │
│  │  └─ Delivery Layer                                      │  │    │
│  │     ├─ handler/               (Huma handlers)          │  │    │
│  │     ├─ dto/                   (request/response)       │  │    │
│  │     └─ routes/                (route registration)     │  │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Chapa Client (pkg/chapa/)                               │    │
│  │  ├─ client.go          (Client interface + impl)        │  │    │
│  │  ├─ types.go           (request/response types)         │  │    │
│  │  └─ webhook.go         (signature verification)         │  │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Notification Module (event consumer)                    │    │
│  │  Receives payment.confirmation events, sends             │    │
│  │  in-app + email notifications                            │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────┐
│  Chapa API (https://api.chapa.co/v1)                            │
│  ├─ POST /v1/transaction/initialize   → checkout_url            │
│  ├─ GET  /v1/transaction/verify/<ref> → transaction details     │
│  └─ POST webhook                     → event notifications      │
└─────────────────────────────────────────────────────────────────┘
```

## Payment Flow

### 1. User Initiates Subscription (Mobile → Backend)

```
[Flutter] GET /api/v1/payments/plans
         → Returns available plans with prices

[Flutter] POST /api/v1/payments/initiate { planName, period }
         → Backend generates tx_ref
         → Creates Payment record (status: pending)
         → Calls Chapa /transaction/initialize
         → Returns { checkout_url, tx_ref }

[Flutter] Opens Chapa checkout via SDK (native or web)
         → User completes payment
```

### 2. Chapa Notifies Backend (Webhook)

```
[Chapa] POST /api/v1/webhooks/chapa { event: "charge.success", tx_ref: "..." }
        → Verify x-chapa-signature header
        → Find Payment by tx_ref
        → Call Chapa /transaction/verify/<tx_ref> (double-confirm)
        → Update Payment status → "success"
        → Create/Update Subscription record
        → Emit "payment.confirmation" event to notification module
        → Return 200 OK
```

### 3. Client Confirms (Fallback / Immediate UX)

```
[Flutter] User redirected to adisu://payment/result
         → App calls POST /api/v1/payments/verify { tx_ref }
         → Backend checks Payment status from DB
         → If still pending, calls Chapa verify API
         → Returns current subscription state
         → App navigates to success/failure screen
         → App periodically polls GET /api/v1/me/subscription until active
```

### 4. Renewal Flow

```
[Backend cron/poller] — NOT implemented in v1
  Future: scheduled job checks expiring subscriptions,
  notifies user via push/email, generates new payment link
```

> In v1, renewal is **manual**: user must re-purchase. The app surfaces a "Renew" prompt when subscription nears expiry.

## Cross-Module Dependencies

### Payment Module Depends On

| Dependency | Purpose |
|---|---|
| `*core.Database` | GORM DB access, Transactor for transactions |
| `core.Logger` | Logging |
| `core.Config` | Chapa config (keys, URLs) |
| `rabbitmq.Bus` | Publish `payment.confirmation` events |
| `shared/notificationevent.Envelope` | Canonical event format |

### Other Modules Depend On Payment

| Module | Dependency | Purpose |
|---|---|---|
| Notification | Subscribes to `payment.confirmation` event | Send payment confirmation notifications |
| AI (future) | Reads subscription status | Gate AI usage by Pro tier |
| All content modules (future) | Reads subscription status | Gate premium content |

## Configuration

### ChapaConfig (`core.Config`)

```
CHAPA_ENABLED=true
CHAPA_SECRET_KEY=CHASECK-test_xxxxxxxxxxxxxxxx
CHAPA_PUBLIC_KEY=CHAPUBK-test_xxxxxxxxxxxxxxxx
CHAPA_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxxxxx
CHAPA_BASE_URL=https://api.chapa.co/v1
```

### Plan Pricing (DB Seed)

| Plan | Period | Amount (minor unit) | Amount (display) |
|---|---|---|---|
| Basic | monthly | 9900 | 99.00 ETB |
| Basic | yearly | 95000 | 950.00 ETB |
| Pro | monthly | 19900 | 199.00 ETB |
| Pro | yearly | 190000 | 1,900.00 ETB |

## Key Design Decisions

1. **Custom `pkg/chapa/` client** — not the unofficial Go SDK. Our config system, error handling, and webhook verification patterns require a custom implementation.
2. **Webhook-first, client-verify backup** — webhooks are authoritative for payment success/failure; client verify is for immediate UX feedback.
3. **Server-generated `tx_ref`** — deterministic, includes account + plan + period + timestamp. Prevents replay and enables idempotency.
4. **Integer minor units** — all monetary values stored as `int64` (ETB minor unit = satcker, 1/100 ETB).
5. **Account-level subscriptions** — single active subscription per account, shared across users on that account.
6. **DB seed for plans** — static plans in DB, queryable via API, easy to update in future.
7. **Canonical notification events** — reuses existing `payment.confirmation` event from notification module.
8. **No custom webhook auth path** — webhook endpoint is public but signature-verified using `x-chapa-signature` header. This follows Chapa's standard pattern.