# Payment Integration — Implementation Phases

This document defines the step-by-step implementation plan. Each phase is independently reviewable, testable, and committable. We follow the rule: **implement → review → commit → next phase**.

---

## Phase 1: Foundation — Config & Chapa Client

**Goal:** Standalone Chapa HTTP client package + config wiring. No payment module yet.

### Files to Create

| File | Purpose |
|---|---|
| `pkg/chapa/types.go` | Request/response DTOs for Chapa API |
| `pkg/chapa/client.go` | Client interface + HTTP implementation |
| `pkg/chapa/webhook.go` | HMAC-SHA256 signature verification |
| `pkg/chapa/errors.go` | Chapa-specific error type |
| `pkg/chapa/doc.go` | Package documentation |

### Files to Modify

| File | Change |
|---|---|
| `internal/core/types.go` | Add `ChapaConfig` to `Config` struct |
| `internal/core/config.go` | Bind `CHAPA_*` env vars |
| `core-backend/.env` | Add Chapa placeholder config |

### Review Checklist

- [ ] `ChapaConfig` struct has all required fields with `mapstructure` tags
- [ ] Env vars properly bound (no unmarshaling errors)
- [ ] `Client` interface is clean and mockable
- [ ] `InitializeTransaction` constructs correct Chapa request body
- [ ] `VerifyTransaction` parses Chapa response correctly
- [ ] `VerifySignature` computes HMAC-SHA256 correctly
- [ ] Error type includes HTTP status and raw response for debugging
- [ ] All methods accept `context.Context` for timeouts

---

## Phase 2: Domain Layer — Models & DB Schema

**Goal:** Pure domain code + database schema. No business logic yet.

### Files to Create

| File | Purpose |
|---|---|
| `internal/modules/payment/domain/entity/plan.go` | GORM entity |
| `internal/modules/payment/domain/entity/payment.go` | GORM entity |
| `internal/modules/payment/domain/entity/subscription.go` | GORM entity |
| `internal/modules/payment/domain/entity/enums.go` | Status constants |
| `internal/modules/payment/domain/repository/plan_repository.go` | Interface |
| `internal/modules/payment/domain/repository/payment_repository.go` | Interface |
| `internal/modules/payment/domain/repository/subscription_repository.go` | Interface |
| `internal/modules/payment/domain/usecase/payment_usecase.go` | Interface |

### Files to Create (Migration)

| File | Purpose |
|---|---|
| `migrations/XXX_create_plans_table.up.sql` | Plans table + enum |
| `migrations/XXX_create_payments_table.up.sql` | Payments table + enum |
| `migrations/XXX_create_subscriptions_table.up.sql` | Subscriptions table + enum |
| Seed SQL (in migration or seed script) | Insert 4 static plans |

### Review Checklist

- [ ] All entities embed `model.BaseModel`
- [ ] `TableName()` methods present
- [ ] `amount` is `int64` (minor units)
- [ ] `tx_ref` has unique index
- [ ] Single active subscription per account enforced via partial unique index
- [ ] Soft delete (`deleted_at`) on all entities
- [ ] Repository interfaces use `context.Context`
- [ ] FK references correct table name (`accounts`)

---

## Phase 3: Infrastructure & Application — Business Logic

**Goal:** Repository implementations, usecase orchestration, and Chapa integration.

### Files to Create

| File | Purpose |
|---|---|
| `internal/modules/payment/infrastructure/repository/plan_repository.go` | GORM impl |
| `internal/modules/payment/infrastructure/repository/payment_repository.go` | GORM impl |
| `internal/modules/payment/infrastructure/repository/subscription_repository.go` | GORM impl |
| `internal/modules/payment/application/usecase/payment_usecase.go` | Business logic |

### Business Logic Methods

| Method | Logic |
|---|---|
| `InitiatePayment` | Validate plan → generate tx_ref → create Payment (pending) → call Chapa init → return checkout URL |
| `VerifyPayment` | Find Payment → call Chapa verify → update status → if success, create/update Subscription → write outbox |
| `HandleWebhook` | Verify signature → find Payment → check idempotency → call Chapa verify → update records → write outbox |
| `GetActiveSubscription` | Query by account + status=active |
| `ListPlans` | Query active plans |

### Review Checklist

- [ ] All DB operations wrapped in transactions where needed
- [ ] `tx_ref` generation is deterministic and unique
- [ ] Webhook handler is idempotent (skips already-terminal payments)
- [ ] On success: Payment updated, Subscription created/updated, outbox entry written atomically
- [ ] On failure: Payment marked failed, no subscription created
- [ ] Error mapping is clear (Chapa error → domain error → HTTP error)
- [ ] No direct RabbitMQ publish — uses outbox pattern

---

## Phase 4: Delivery — API Surface

**Goal:** HTTP handlers, DTOs, routes, and module wiring.

### Files to Create

| File | Purpose |
|---|---|
| `internal/modules/payment/delivery/dto/payment_dto.go` | Request/response structs |
| `internal/modules/payment/delivery/handler/payment_handler.go` | Plan listing, initiate, verify, subscription |
| `internal/modules/payment/delivery/handler/webhook_handler.go` | Chapa webhook receiver |
| `internal/modules/payment/delivery/routes/payment_routes.go` | Route registration |
| `internal/modules/payment/module.go` | fx.Module wiring |
| `internal/modules/payment/entities.go` | EntityProvider for SchemaManager |

### Files to Modify

| File | Change |
|---|---|
| `internal/modules/modules.go` | Add `payment.Module` to `fx.Options` |

### Routes

| Method | Path | Auth | Handler |
|---|---|---|---|
| GET | `/api/v1/payments/plans` | None | ListPlans |
| POST | `/api/v1/payments/initiate` | JWT + Account | InitiatePayment |
| POST | `/api/v1/payments/verify` | JWT + Account | VerifyPayment |
| GET | `/api/v1/me/subscription` | JWT + Account | GetMySubscription |
| POST | `/api/v1/webhooks/chapa` | Signature | ChapaWebhook |

### Review Checklist

- [ ] All Huma operation definitions include tags, summaries, responses
- [ ] Auth routes use `AuthMiddleware` + `AccountStatusMiddleware`
- [ ] Webhook route is **not** behind auth middleware
- [ ] DTOs use pointer fields for optional values
- [ ] Response shapes match API contract in `api-contract.md`
- [ ] Error responses follow existing `pkg/response` patterns
- [ ] Module wiring follows existing module pattern (guide, IAM)

---

## Phase 5: Tests & OpenAPI

**Goal:** Quality assurance, test coverage, and API spec completeness.

### Files to Create

| File | Purpose |
|---|---|
| `pkg/chapa/client_test.go` | Mock HTTP server tests for Chapa client |
| `pkg/chapa/webhook_test.go` | Signature verification tests |
| `internal/modules/payment/application/usecase/payment_usecase_test.go` | Business logic tests |
| `internal/modules/payment/delivery/handler/payment_handler_test.go` | Handler tests |

### Files to Regenerate / Update

| File | Action |
|---|---|
| `mobile/api_client/openapi/openapi.json` | Copy auto-generated spec from backend |
| `mobile/api_client/lib/api_client.dart` | Export new clients/models |

### Review Checklist

- [ ] Chapa client tests cover success, failure, and network error paths
- [ ] Webhook tests verify correct and incorrect signatures
- [ ] Usecase tests cover: pending→success, pending→failed, duplicate webhook
- [ ] Handler tests verify HTTP status codes and response shapes
- [ ] All tests use `httptest` or mocks (no real Chapa calls)
- [ ] OpenAPI spec includes all 5 endpoints with correct schemas
- [ ] Backend boots successfully with payment module enabled

---

## Phase 6: Mobile Wiring Documentation

**Goal:** Document the Flutter-side integration (no mobile code changes in backend repo).

### Files to Create

| File | Purpose |
|---|---|
| `docs/payment/mobile-wiring.md` | Step-by-step Flutter integration guide |

### Scope

- How to regenerate `api_client` from OpenAPI spec
- How to add Riverpod providers for subscription state
- How to integrate Chapa Flutter SDK with backend-generated `tx_ref`
- Deep link + polling strategy for payment confirmation
- Entitlement gating patterns in the mobile app

---

## Commit Convention

Each phase is a single commit using conventional commits:

```
feat(payment): Phase 1 — add Chapa client and config wiring
feat(payment): Phase 2 — add domain models and DB schema
feat(payment): Phase 3 — add business logic and Chapa integration
feat(payment): Phase 4 — add HTTP handlers and API routes
test(payment): Phase 5 — add tests and OpenAPI spec
docs(payment): Phase 6 — add mobile wiring documentation
```

## Rollback Strategy

Each phase can be rolled back independently by reverting its commit. The schema migrations include corresponding `.down.sql` files. If a phase needs rework, revert and re-implement without affecting other phases.

## Dependencies Between Phases

```
Phase 1 (Config+Client)
    │
    ▼
Phase 2 (Domain+Schema)
    │
    ▼
Phase 3 (Business Logic) ── depends on Phase 1 (uses Chapa client)
    │                        depends on Phase 2 (uses entities)
    ▼
Phase 4 (API Surface) ───── depends on Phase 3 (uses usecases)
    │
    ▼
Phase 5 (Tests+Spec) ────── depends on Phase 4 (tests handlers)
    │
    ▼
Phase 6 (Mobile Docs) ───── depends on Phase 5 (needs final API spec)
```

No phase depends on a later phase. Each phase is self-contained.