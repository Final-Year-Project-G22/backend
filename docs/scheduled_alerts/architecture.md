# Architecture

## Module Structure

The feature follows the same Clean Architecture layers as the existing notification module:

```
notification/
├── domain/          # Entities, repository interfaces, use case interfaces
├── application/     # Use case implementations, background services
├── infrastructure/  # Repository implementations (GORM), external adapters
└── delivery/        # HTTP handlers, DTOs, routes
```

## Data Flow

### Scheduled Alert Flow

```
User (mobile app)
    │
    ▼
POST /api/v1/notifications/scheduled
    │  Schedule(ctx, accountID, input)
    │    ├─ Check pro limit (count pending ≤ 3 for non-pro)
    │    ├─ Validate channel (in_app, email, push)
    │    └─ Create user_scheduled_notification (status=pending)
    ▼
[user_scheduled_notifications table]
    │
    ▼  (every 10s)
[UserNotificationScheduler]
    │  FetchDue(ctx) → enqueue to NotificationQueue
    │    ├─ Create NotificationQueue entry
    │    │   notification_type = "user_scheduled"
    │    │   payload = {title, body, channel}
    │    │   scheduled_for = now (immediate delivery)
    │    └─ Mark user_scheduled_notification.status = "sent"
    ▼
[NotificationQueue] ──→ [DeliveryWorker] ──→ History + Inbox
```

### Business Alert Flow

```
User (mobile app)
    │
    ▼
POST /api/v1/compliance/entries
    │  Create(user enters TIN, expiry date, reminder_days_before)
    ▼
[compliance_entries table]  (status=active)
    │
    ▼  (every 1h)
[BusinessAlertScheduler]
    │  FetchExpiringSoon(ctx)
    │    WHERE status=active
    │      AND expiry_date - reminder_days_before ≤ now
    │      AND (last_notified_at IS NULL OR last_notified_at < computed_alert_time)
    │
    │  For each:
    │    ├─ Load BusinessProfile + Account
    │    ├─ Resolve Account email/name
    │    ├─ Create NotificationQueue entry
    │    │   notification_type = "account_alert_info"
    │    │   channel = in_app + email
    │    │   payload = rendered from template with business data
    │    └─ Update compliance_entry.last_notified_at = now
    ▼
[NotificationQueue] ──→ [DeliveryWorker] ──→ History + Inbox
```

### Compliance Calendar Query Flow

```
GET /api/v1/compliance/calendar
    │
    ├─ Query compliance_entries WHERE expiry_date > now
    │   ORDER BY expiry_date ASC
    │   (for the user's business profile)
    │
    └─ Query user_scheduled_notifications WHERE status=pending
        ORDER BY scheduled_for ASC
        (for the same account)

Returns unified timeline sorted by date.
```

## Dependency Injection

### module.go additions

```go
// --- New Repositories ---
fx.Provide(fx.Annotate(infrarepo.NewUserScheduledNotificationRepository,
    fx.As(new(repository.UserScheduledNotificationRepository)))),
fx.Provide(fx.Annotate(infrarepo.NewScheduledAlertTemplateRepository,
    fx.As(new(repository.ScheduledAlertTemplateRepository)))),
fx.Provide(fx.Annotate(infrarepo.NewComplianceEntryRepository,
    fx.As(new(repository.ComplianceEntryRepository)))),

// --- SubscriptionReader (default: not pro) ---
fx.Provide(fx.Annotate(func() repository.SubscriptionReader {
    return &defaultSubscriptionReader{}
}, fx.As(new(repository.SubscriptionReader)))),

// --- Use Cases ---
fx.Provide(fx.Annotate(appusecase.NewUserScheduledNotificationUsecase,
    fx.As(new(usecase.UserScheduledNotificationUsecase)))),
fx.Provide(fx.Annotate(appusecase.NewComplianceEntryUsecase,
    fx.As(new(usecase.ComplianceEntryUsecase)))),

// --- Schedulers ---
fx.Provide(appservice.NewUserNotificationScheduler),
fx.Provide(appservice.NewBusinessAlertScheduler),

// --- Handlers ---
fx.Provide(handler.NewScheduledAlertHandler),
fx.Provide(handler.NewComplianceHandler),
```

### modules.go (compositor) addition

```go
fx.Decorate(func(
    _ notifrepo.SubscriptionReader,
    subscriptionRepo paymentrepo.SubscriptionRepository,
) notifrepo.SubscriptionReader {
    return paymentnotification.NewSubscriptionReaderAdapter(subscriptionRepo)
}),
```

### Scheduler Lifecycle

Both schedulers follow the same pattern as `DeliveryWorker`:

```go
fx.Invoke(func(lc fx.Lifecycle, scheduler *appservice.UserNotificationScheduler) {
    ctx, cancel := context.WithCancel(context.Background())
    lc.Append(fx.Hook{
        OnStart: func(context.Context) error {
            scheduler.Start(ctx)
            return nil
        },
        OnStop: func(context.Context) error {
            cancel()
            return nil
        },
    })
}),
```

## Background Workers

| Worker | Poll Interval | Batch Size | Purpose |
|--------|--------------|------------|---------|
| `UserNotificationScheduler` | 10s | 50 | Enqueues due Scheduled Alerts to NotificationQueue |
| `BusinessAlertScheduler` | 1h | 100 | Triggers Business Alerts for expiring Compliance Entries |

## Pro Limit Enforcement

The `SubscriptionReader` interface abstracts the pro status check:

```go
type SubscriptionReader interface {
    HasActiveProSubscription(ctx context.Context, accountID uuid.UUID) (bool, error)
}
```

- Default implementation returns `(false, nil)` — everyone is non-pro.
- The compositor (`modules.go`) overrides with a real adapter that queries `paymentrepo.SubscriptionRepository.GetActiveByAccount()`.
- Pro is determined by checking if the active subscription's `PlanName` indicates a pro tier.

In the `Schedule` use case:

```
1. Count pending user_scheduled_notifications for accountID
2. If count >= 3:
   a. Check HasActiveProSubscription
   b. If not pro → return ErrMaxScheduledAlertsReached
3. Otherwise → create
```
