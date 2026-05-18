# Implementation Plan — Backend

Organized as sequential PRs. Each PR is independently deployable.

---

## PR 1: Domain Entities + Enums

**Files to create:**
- `internal/modules/notification/domain/entity/user_scheduled_notification.go`
- `internal/modules/notification/domain/entity/scheduled_alert_template.go`
- `internal/modules/notification/domain/entity/compliance_entry.go`

**Files to modify:**
- `internal/modules/notification/domain/entity/enums.go` — add `ScheduleStatus`, `ComplianceType`, `ComplianceEntryStatus`, `NotificationTypeUserScheduled`

**Tests:** Entity creation, enum validation, GORM tags.

---

## PR 2: Repository Interfaces + Implementations

**Files to create:**
- `internal/modules/notification/domain/repository/user_scheduled_notification_repository.go` — interface
- `internal/modules/notification/domain/repository/scheduled_alert_template_repository.go` — interface
- `internal/modules/notification/domain/repository/compliance_entry_repository.go` — interface
- `internal/modules/notification/domain/repository/subscription_reader.go` — interface
- `internal/modules/notification/infrastructure/repository/user_scheduled_notification_repository.go`
- `internal/modules/notification/infrastructure/repository/scheduled_alert_template_repository.go`
- `internal/modules/notification/infrastructure/repository/compliance_entry_repository.go`

**Tests:** CRUD operations, `FetchDue`, `CountPendingByAccount`, `FetchExpiringSoon`, edge cases.

---

## PR 3: Entity Registration + DI Setup

**Files to modify:**
- `internal/modules/notification/entities.go` — register all 3 new entities
- `internal/modules/notification/module.go` — register repository providers, `SubscriptionReader` default

This PR ensures auto-migration creates the tables in the database.

**Verification:** Run app, confirm `user_scheduled_notifications`, `scheduled_alert_templates`, `compliance_entries` tables exist.

---

## PR 4: Seed Data — Scheduled Alert Templates

**Files to create:**
- `migrations/YYYYMMDDHHMMSS_seed_scheduled_alert_templates.sql`

**Data:**
- 6 templates: custom, tax_filing, license_renewal, registration_renewal, meeting, deadline

**Verification:** Query `scheduled_alert_templates` table after migration.

---

## PR 5: Use Case Interfaces + Input DTOs

**Files to create:**
- `internal/modules/notification/domain/usecase/user_scheduled_notification_usecase.go`
- `internal/modules/notification/domain/usecase/compliance_entry_usecase.go`

**Files to modify:**
- `internal/modules/notification/domain/usecase/inputs.go` — add `ScheduleUserNotificationInput`, `CreateComplianceEntryInput`, `UpdateComplianceEntryInput`, `RescheduleUserNotificationInput`, `CalendarEntry`, `ComplianceCalendar`

---

## PR 6: Application Use Case Implementations

**Files to create:**
- `internal/modules/notification/application/usecase/user_scheduled_notification_usecase.go`
- `internal/modules/notification/application/usecase/compliance_entry_usecase.go`

**Logic covered:**
- `Schedule`: threshold check (3 pending for non-pro), channel validation, entity creation
- `Cancel`: ownership + status verification
- `Reschedule`: ownership + rescheduled_from tracking
- `List`: scoped to account, ordered
- `Create/Update/Delete ComplianceEntry`: business profile ownership validation
- `GetCalendar`: merge compliance entries + scheduled alerts

**Registration:** Add both to `module.go` with `fx.As`.

**Tests:** Full use case tests with mocked repositories.

---

## PR 7: API Handlers + DTOs + Routes

**Files to create:**
- `internal/modules/notification/delivery/handler/scheduled_alert_handler.go`
- `internal/modules/notification/delivery/handler/compliance_handler.go`
- `internal/modules/notification/delivery/dto/scheduled_alert_dto.go`
- `internal/modules/notification/delivery/dto/compliance_dto.go`
- `internal/modules/notification/delivery/routes/scheduled_alert_routes.go`
- `internal/modules/notification/delivery/routes/compliance_routes.go`

**Files to modify:**
- `internal/modules/notification/delivery/routes/routes.go` — register new route groups

**Endpoints:**
- `GET/POST /api/v1/notifications/scheduled`
- `PATCH /api/v1/notifications/scheduled/{id}/cancel`
- `PATCH /api/v1/notifications/scheduled/{id}/reschedule`
- `GET /api/v1/notifications/scheduled/templates`
- `GET/POST /api/v1/compliance/entries`
- `PATCH/DELETE /api/v1/compliance/entries/{id}`
- `GET /api/v1/compliance/calendar`

**Verification:** Call each endpoint with curl/Postman.

---

## PR 8: Background Schedulers

**Files to create:**
- `internal/modules/notification/application/service/user_notification_scheduler.go`
- `internal/modules/notification/application/service/business_alert_scheduler.go`

**Files to modify:**
- `internal/modules/notification/module.go` — register providers + lifecycle hooks

**Scheduler 1 — UserNotificationScheduler:**
- Poll every 10s
- `FetchDue`: `WHERE status=pending AND scheduled_for ≤ now`
- For each: create `NotificationQueue` record, mark as `sent`
- Handle partial failures (continue on error per item)

**Scheduler 2 — BusinessAlertScheduler:**
- Poll every 1h
- `FetchExpiringSoon`: `WHERE status=active AND expiry_date - reminder_days_before ≤ now AND (last_notified_at IS NULL OR ...)`
- For each: render template, enqueue, update `last_notified_at`

**Tests:** Integration test with time mocking.

---

## PR 9: SubscriptionReader Adapter + Compositor Override

**Files to create:**
- `internal/modules/payment/infrastructure/notification/subscription_reader.go` — adapter wrapping `paymentrepo.SubscriptionRepository`

**Files to modify:**
- `internal/modules/modules.go` — add `fx.Decorate` for `SubscriptionReader`

**Logic:**
- Adapter calls `subscriptionRepo.GetActiveByAccount(ctx, accountID)`
- If subscription exists and `PlanName` indicates pro → return `true`
- Otherwise → return `false`

---

## PR 10: Integration + E2E Verification

**Tasks:**
1. Run full test suite
2. Manual E2E flow:
   - Create compliance entry → wait for Business Alert → verify inbox
   - Create 3 pending Scheduled Alerts → attempt 4th → verify 403
   - Create scheduled alert → wait for delivery → verify inbox
   - Reschedule and cancel flows
3. Update OpenAPI spec document

---

## File Change Summary

| PR | New Files | Modified Files |
|----|-----------|----------------|
| PR 1 | 3 | 1 |
| PR 2 | 7 | 0 |
| PR 3 | 0 | 2 |
| PR 4 | 1 | 0 |
| PR 5 | 2 | 1 |
| PR 6 | 2 | 0 |
| PR 7 | 6 | 1 |
| PR 8 | 2 | 1 |
| PR 9 | 1 | 1 |
| PR 10 | 0 | 1 |
| **Total** | **24** | **8** |
