# Notification Module — Repository Interfaces

All repository interfaces live in `domain/repository/` and are implemented in `infrastructure/repository/`. Each interface extends `sharedrepo.GenericRepository[T]` for standard CRUD operations, then adds domain-specific methods.

## Shared Pattern

Every repository implementation follows these conventions:

- Private struct embedding `sharedrepo.GenericRepository[T]`
- Constructor `New*Repository(db *core.Database, logger core.Logger)` returning the domain interface
- `getDB(ctx)` resolves a transaction from context (`core.TxFromContext`) or falls back to the raw DB
- `applyPaginationAndSorting()` helper for query options (shared across repos in this module)

---

## 1. NotificationTemplateRepository

**Entity:** `NotificationTemplate`

**Standard CRUD:** inherited from `GenericRepository[NotificationTemplate]`

**Domain-specific methods:**

| Method | Signature | Description |
|---|---|---|
| `GetByType` | `(ctx context.Context, notificationType NotificationType) (*NotificationTemplate, error)` | Lookup template by unique notification type |
| `GetByCategory` | `(ctx context.Context, category NotificationCategory, q query.QueryOptions) ([]*NotificationTemplate, error)` | List all templates in a category |
| `GetTranslations` | `(ctx context.Context, templateID uuid.UUID) ([]*NotificationTemplateTranslation, error)` | Fetch all translations for a template |
| `UpsertTranslation` | `(ctx context.Context, translation *NotificationTemplateTranslation) error` | Create or update a translation (unique on templateID + language) |
| `DeleteTranslation` | `(ctx context.Context, templateID uuid.UUID, language string) error` | Remove a specific translation |

---

## 2. UserNotificationPreferenceRepository

**Entity:** `UserNotificationPreference`

**Standard CRUD:** inherited from `GenericRepository[UserNotificationPreference]`

**Domain-specific methods:**

| Method | Signature | Description |
|---|---|---|
| `GetByAccountAndTypeAndChannel` | `(ctx context.Context, accountID uuid.UUID, notificationType NotificationType, channel Channel) (*UserNotificationPreference, error)` | Lookup a specific override |
| `ListByAccount` | `(ctx context.Context, accountID uuid.UUID) ([]*UserNotificationPreference, error)` | All explicit overrides for an account |
| `Upsert` | `(ctx context.Context, pref *UserNotificationPreference) error` | Create or update override (sparse model — only stores explicit deviations) |

**Upsert logic:** Uses `clause.OnConflict` on `(AccountID, NotificationType, Channel)` to update `IsEnabled`, `QuietHoursStart`, `QuietHoursEnd` on conflict.

---

## 3. MutedAccountRepository

**Entity:** `MutedAccount`

**Standard CRUD:** inherited from `GenericRepository[MutedAccount]`

**Domain-specific methods:**

| Method | Signature | Description |
|---|---|---|
| `IsMuted` | `(ctx context.Context, accountID uuid.UUID, mutedAccountID uuid.UUID) (bool, error)` | Check if accountA has muted accountB. Checks `MuteUntil` — if set and expired, returns false. |
| `ListByAccount` | `(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*MutedAccount, error)` | List all accounts this user has muted |
| `Delete` | `(ctx context.Context, accountID uuid.UUID, mutedAccountID uuid.UUID) error` | Unmute a specific account |

---

## 4. UserDeviceRepository

**Entity:** `UserDevice`

**Standard CRUD:** inherited from `GenericRepository[UserDevice]`

**Domain-specific methods:**

| Method | Signature | Description |
|---|---|---|
| `ListByAccount` | `(ctx context.Context, accountID uuid.UUID) ([]*UserDevice, error)` | All active devices for an account |
| `GetByDeviceToken` | `(ctx context.Context, deviceToken string) (*UserDevice, error)` | Deduplication lookup — check if device already registered |
| `DeactivateByAccount` | `(ctx context.Context, accountID uuid.UUID) error` | Set `IsActive = false` on all devices for an account (used on account suspension) |
| `UpdatePushToken` | `(ctx context.Context, id uuid.UUID, pushToken string) error` | Update push token when it changes (FCM token refresh) |

---

## 5. NotificationQueueRepository

**Entity:** `NotificationQueue`

**Standard CRUD:** inherited from `GenericRepository[NotificationQueue]`

**Domain-specific methods:**

| Method | Signature | Description |
|---|---|---|
| `FetchPending` | `(ctx context.Context, limit int) ([]*NotificationQueue, error)` | Worker fetches items where `Status = 'pending'` AND `ScheduledFor <= now()`, ordered by `Priority DESC, ScheduledFor ASC` |
| `MarkProcessing` | `(ctx context.Context, id uuid.UUID) error` | Set `Status = 'processing'` (locks the row for delivery) |
| `MarkDelivered` | `(ctx context.Context, id uuid.UUID) error` | Set `Status = 'delivered'` on success |
| `MarkFailed` | `(ctx context.Context, id uuid.UUID, errMsg string) error` | Set `Status = 'failed'` and `ErrorMessage` when `RetryCount >= MaxRetries` |
| `IncrementRetry` | `(ctx context.Context, id uuid.UUID, nextScheduledFor time.Time) error` | Increment `RetryCount`, update `ScheduledFor` with exponential backoff, set `Status = 'pending'` |
| `CancelByAccount` | `(ctx context.Context, accountID uuid.UUID) error` | Set `Status = 'cancelled'` for all pending items for an account |
| `CancelByCampaign` | `(ctx context.Context, campaignID uuid.UUID) error` | Set `Status = 'cancelled'` for all pending items from a campaign |
| `CountByStatus` | `(ctx context.Context, status NotificationStatus) (int64, error)` | Monitoring/metrics — count items by status |

---

## 6. NotificationHistoryRepository

**Entity:** `NotificationHistory`

**Standard CRUD:** inherited from `GenericRepository[NotificationHistory]`

**Domain-specific methods:**

| Method | Signature | Description |
|---|---|---|
| `ListByAccount` | `(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*NotificationHistory, error)` | Paginated delivery audit for a user |
| `UpdateDeliveryStatus` | `(ctx context.Context, id uuid.UUID, status DeliveryStatus, deliveredAt *time.Time) error` | Update delivery outcome from webhook |
| `MarkRead` | `(ctx context.Context, id uuid.UUID) error` | Set `ReadAt` to now |
| `MarkClicked` | `(ctx context.Context, id uuid.UUID) error` | Set `ClickedAt` to now |

---

## 7. UserNotificationInboxRepository

**Entity:** `UserNotificationInbox`

**Standard CRUD:** inherited from `GenericRepository[UserNotificationInbox]`

**Domain-specific methods:**

| Method | Signature | Description |
|---|---|---|
| `ListByAccount` | `(ctx context.Context, accountID uuid.UUID, category *NotificationCategory, q query.QueryOptions) ([]*UserNotificationInbox, error)` | Paginated inbox with optional category filter. Excludes archived and expired entries. |
| `GetUnreadCount` | `(ctx context.Context, accountID uuid.UUID) (int64, error)` | Count of unread, non-archived, non-expired entries (for badge) |
| `MarkAsRead` | `(ctx context.Context, id uuid.UUID) error` | Set `IsRead = true`. Also updates `NotificationHistory.ReadAt`. |
| `MarkAllAsRead` | `(ctx context.Context, accountID uuid.UUID) error` | Set `IsRead = true` for all unread entries. Also updates corresponding history records. |
| `Archive` | `(ctx context.Context, id uuid.UUID) error` | Set `IsArchived = true` |
| `ExpireOld` | `(ctx context.Context, before time.Time) error` | Delete entries where `ExpiresAt <= before`. Cleanup job. |
| `MarkAllReadByCategory` | `(ctx context.Context, accountID uuid.UUID, category NotificationCategory) error` | Mark all unread in a category as read |

---

## 8. NotificationCampaignRepository

**Entity:** `NotificationCampaign`

**Standard CRUD:** inherited from `GenericRepository[NotificationCampaign]`

**Domain-specific methods:**

| Method | Signature | Description |
|---|---|---|
| `ListByStatus` | `(ctx context.Context, status CampaignStatus, q query.QueryOptions) ([]*NotificationCampaign, error)` | Filter campaigns by status |
| `UpdateStatus` | `(ctx context.Context, id uuid.UUID, status CampaignStatus) error` | Transition campaign status |
| `ListScheduled` | `(ctx context.Context) ([]*NotificationCampaign, error)` | Fetch campaigns with `Status = 'scheduled'` AND `ScheduledFor <= now()` |
| `GetByCreator` | `(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*NotificationCampaign, error)` | List campaigns created by a specific admin |

---

## 9. EmailDeliveryLogRepository

**Entity:** `EmailDeliveryLog`

**Standard CRUD:** inherited from `GenericRepository[EmailDeliveryLog]`

**Domain-specific methods:**

| Method | Signature | Description |
|---|---|---|
| `GetByProviderMessageID` | `(ctx context.Context, providerMessageID string) (*EmailDeliveryLog, error)` | Lookup by provider message ID — used when Resend webhook arrives |
| `UpdateDeliveryEvent` | `(ctx context.Context, id uuid.UUID, eventType string, occurredAt time.Time, metadata map[string]interface{}) error` | Update the appropriate timestamp field based on event type: `delivered` → DeliveredAt, `opened` → OpenedAt, `clicked` → ClickedAt, `bounced` → BounceReason, `complained` → Complaint=true |
| `GetByNotificationHistoryID` | `(ctx context.Context, historyID uuid.UUID) (*EmailDeliveryLog, error)` | Link back from notification history to email log |

---

## Cross-Module Interface: MuteResolver

The notification module defines a `MuteResolver` interface that source modules implement. This enables the delivery pipeline to check module-specific mute state (e.g., community thread mutes) without coupling to those modules' data models.

```go
type MuteResolver interface {
    IsMuted(ctx context.Context, accountID uuid.UUID, itemType string, itemID uuid.UUID) (bool, error)
}
```

**Implementors:**

| Module | Resolves |
|---|---|
| Community | Thread mutes (`UserThreadSettings.IsMuted`), category mutes (`UserCategorySettings.IsMuted`) |
| Guide | (future) Guide-level notification mutes |

The `MuteResolver` is injected into the delivery usecase via the DI container. Source modules register their implementation during module initialization.
