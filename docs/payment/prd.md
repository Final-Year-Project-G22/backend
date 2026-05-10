# Payment Integration — Product Requirements Document

## Problem Statement

The Adisu platform currently has no payment or subscription functionality. Users have access to all content regardless of tier. To create a sustainable revenue model and differentiate between free and premium users, we need a complete payment and subscription system integrated with the Chapa payment gateway (https://developer.chapa.co/).

The mobile app (Flutter) and backend (Go) need to work together to:
1. Present subscription plans to users
2. Process payments via Chapa
3. Track subscription status and entitlements
4. Gate premium content based on subscription status
5. Handle payment webhooks for async confirmation

## Solution

Implement a full-stack payment and subscription module in the backend, with corresponding integration in the Flutter mobile app using the Chapa Flutter SDK for checkout.

### Key design choices:
- **Two tiers**: Basic (free) and Pro (paid)
- **Two periods**: Monthly and yearly
- **Manual renewal**: User re-purchases each period (no auto-renew in v1)
- **Server-driven**: Backend is the single source of truth for plans, payments, and subscriptions
- **Webhook-first**: Chapa webhooks are the primary payment confirmation mechanism; client-side verify is a fallback
- **Account-level**: One active subscription per account

## User Stories

### Plan Browsing

1. As an **unregistered user**, I want to see the available subscription plans, so that I can understand what's offered.
2. As a **logged-in user**, I want to see the available subscription plans with prices, so that I can decide which to purchase.
3. As a **Pro subscriber**, I want to see my current plan details (period, renewal date, price), so that I know what I're paying for.

### Payment

4. As a **Basic user**, I want to select a Pro plan (monthly or yearly) and pay via Chapa, so that I get access to premium features.
5. As a **user mid-payment**, I want a seamless checkout experience in the app (native wallet or web fallback), so that I complete the purchase without friction.
6. As a **user**, I want to receive a confirmation when my payment succeeds, so that I know the purchase worked.
7. As a **user**, I want to be notified if my payment fails, so that I can retry.

### Subscription Management

8. As a **Pro subscriber**, I want my subscription to remain active until the end of the paid period, even if I close the app, so that I don't lose access mid-cycle.
9. As a **user**, I want the app to detect my subscription status on launch, so that I see the correct tier immediately.
10. As a **user near expiry**, I want to be prompted to renew before my subscription expires, so that I don't lose access unexpectedly.
11. As a **user who cancelled**, I want to retain access until my current period ends, so that I'm not penalized immediately.

### Access Control

12. As a **backend system**, I want to gate premium content to Pro subscribers only, so that free users can't access paid features.
13. As a **backend system**, I want to gate AI rate limits by subscription tier, so that Pro users get more usage.
14. As a **Pro user**, I want to access premium tools that are unavailable to Basic users, so that I get value from my subscription.

### Admin / Operations

15. As an **admin**, I want to see all payment transactions for audit purposes, so that I can reconcile financial records.
16. As an **admin**, I want to see subscription history per account, so that I can support user inquiries.

## Implementation Decisions

### 1. Chapa Integration
- Custom `pkg/chapa/` client package (not the unofficial Go SDK)
- Integrates with existing config system
- Handles transaction initialization, verification, and webhook signature verification

### 2. Data Model
- Three DB tables: `plans`, `payments`, `subscriptions`
- All monetary values stored as integer minor units (satcker)
- `payments` table stores every attempt (pending, success, failed) for audit
- `subscriptions` table holds one active record per account; expired/cancelled retained for history
- Price snapshots stored per payment (preserves history if prices change)

### 3. API Design
- REST endpoints following existing Huma pattern
- Public plan listing, authenticated payment/subscription endpoints
- Public webhook endpoint with signature verification
- All endpoints use existing `AuthMiddleware` + `AccountStatusMiddleware` where appropriate

### 4. Subscription Lifecycle
- No auto-renew in v1 (manual purchase each period)
- 3-day grace period after expiry before access is cut
- Plan changes take effect at next renewal (no mid-cycle proration)
- Cancellation preserves access until end of current period

### 5. Mobile Integration
- Chapa Flutter SDK for native checkout (ETB only)
- Web checkout fallback for future USD support
- Deep link return from Chapa + polling fallback
- Riverpod state management for subscription status
- Entitlement checked server-side; app reflects server state

### 6. Notifications
- Reuses existing `payment.confirmation` notification type
- Triggers in-app and email notifications on successful payment
- Uses existing canonical notification event envelope pattern

### 7. Security
- Webhook signature verification (HMAC-SHA256)
- Backend-generated deterministic `tx_ref` as idempotency key
- Duplicate webhook handling (idempotent processing)
- All sensitive keys (Chapa secret, webhook secret) in environment variables

### 8. Plan Management
- Plans seeded via DB migration (not admin UI)
- Static for v1; admin CRUD can be added later
- Plans served via public API endpoint

## Testing Decisions

### What makes a good test here
- Test external behavior (API responses, DB state changes), not implementation details
- Mock Chapa API in tests; don't make real HTTP calls
- Each layer testable independently via interface mocking

### Modules to test
- `pkg/chapa/` — Mock HTTP responses; test request construction and response parsing
- Payment use cases — Test state transitions (pending → success, pending → failed)
- Payment handlers — Test API endpoint responses with various inputs
- Webhook handler — Test signature verification, idempotency, event routing
- Subscription entitlement — Test that correct plan/period is returned for various states

### Prior art
- Existing tests in `internal/modules/iam/` for handler + usecase patterns
- `pkg/email/` for third-party integration mock patterns
- Notification module tests for event publishing patterns

## Out of Scope

- Auto-renew / billing tokens (Chapa recurring payments)
- Promo codes / coupons
- Refund processing (acknowledged but not implemented)
- Transfer/payout functionality (Chapa transfer API)
- Split payments
- Multi-currency support (USD support in v1)
- Admin dashboard UI for payments
- Invoice generation PDF
- Apple/Google platform-specific billing (not applicable to web-based checkout)

## Further Notes

- The Chapa Go SDK (github.com/Chapa-Et/chapa-go) exists but is unofficial and minimal. A custom client gives more control and better integration with existing patterns.
- The `payment.confirmation` event type is already registered in the notification module — the payment module just needs to emit it.
- The accounts table FK constraint depends on the IAM module's actual table name and structure — verify during implementation.
- Plan pricing (99/199 ETB monthly, 950/1900 ETB yearly) is placeholder for v1 — adjust before launch.
- The module is designed so future auto-renew can be added on top of the existing payment/subscription tables without schema changes.