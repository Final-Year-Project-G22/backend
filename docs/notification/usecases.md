# Notification Module — Use Case Interfaces & Input DTOs

All use case interfaces live in `domain/usecase/` and are implemented in `application/usecase/`. Input DTOs are defined alongside the interfaces. Application-layer services (`application/service/`) compose use cases and add infrastructure concerns.

---

## Input DTOs

Defined in `domain/usecase/inputs.go`. Follows the codebase pattern of separate Create vs Update inputs (Create uses required fields, Update uses pointer fields for partial updates).

### Template Inputs

```
CreateTemplateInput {
    Name:                 string
    Description:          *string
    NotificationType:     NotificationType
    Category:             NotificationCategory
    Priority:             NotificationPriority
    DefaultContent:       map[string]interface{}   // multi-channel JSONB
    VariablesSchema:      *map[string]interface{}   // nullable
    DefaultTTL:           *int                      // seconds, nullable
}

UpdateTemplateInput {
    Name:                 *string
    Description:          *string
    Priority:             *NotificationPriority
    DefaultContent:       *map[string]interface{}
    VariablesSchema:      *map[string]interface{}
    DefaultTTL:           *int
}

CreateTemplateTranslationInput {
    TemplateID:           uuid.UUID
    Language:             string
    Subject:              string
    Content:              map[string]interface{}    // multi-channel JSONB
}

UpdateTemplateTranslationInput {
    Subject:              *string
    Content:              *map[string]interface{}
}
```

### Preference Inputs

```
SetPreferenceInput {
    NotificationType:     NotificationType
    Channel:              Channel
    IsEnabled:            bool
    QuietHoursStart:      *time.Time
    QuietHoursEnd:        *time.Time
}
```

### Mute Inputs

```
MuteAccountInput {
    MutedAccountID:       uuid.UUID
    MuteUntil:            *time.Time
    Reason:               *string
}
```

### Device Inputs

```
RegisterDeviceInput {
    DeviceType:           DeviceType
    DeviceToken:          string
    PushToken:            *string
    DeviceName:           *string
    DeviceModel:          *string
    OSVersion:            *string
    AppVersion:           *string
}

UpdateDeviceInput {
    PushToken:            *string
    DeviceName:           *string
    OSVersion:            *string
    AppVersion:           *string
    IsActive:             *bool
}
```

### Ingest Inputs

```
ProcessEventInput {
    SourceModule:         string
    SourceEvent:          string
    NotificationType:     NotificationType
    AccountID:            uuid.UUID
    Variables:            map[string]string         // template rendering variables
    Metadata:             map[string]interface{}    // source entity references
    ExpiresAt:            *time.Time                // override template DefaultTTL
}

SendNotificationInput {
    NotificationType:     NotificationType
    AccountID:            uuid.UUID
    Channel:              Channel
    Variables:            map[string]string
    Metadata:             map[string]interface{}
    ScheduledFor:         *time.Time                // nullable = immediate
    ExpiresAt:            *time.Time
}
```

### Campaign Inputs

```
CreateCampaignInput {
    Name:                 string
    Description:          *string
    CampaignType:         CampaignType
    TargetSegment:        *map[string]interface{}   // nullable for Broadcast
    TemplateID:           uuid.UUID
    CustomSubject:        *string
    CustomContent:        *map[string]interface{}
    ScheduledFor:         *time.Time
}

UpdateCampaignInput {
    Name:                 *string
    Description:          *string
    TargetSegment:        *map[string]interface{}
    CustomSubject:        *string
    CustomContent:        *map[string]interface{}
    ScheduledFor:         *time.Time
}

ScheduleCampaignInput {
    CampaignID:           uuid.UUID
}
```

### Email Webhook Input

```
ResendWebhookEvent {
    EventType:            string       // "delivered", "opened", "clicked", "bounced", "complained"
    EmailID:              string       // provider message ID
    RecipientEmail:       string
    OccurredAt:           time.Time
    BounceReason:         *string
}
```

---

## Use Case Interfaces

### 1. NotificationTemplateUsecase

Manages notification templates and their translations. Admin-only operations.

| Method | Signature | Description |
|---|---|---|
| `CreateTemplate` | `(ctx, input CreateTemplateInput) (*NotificationTemplate, error)` | Create a new template. Validates NotificationType uniqueness. |
| `GetTemplate` | `(ctx, id uuid.UUID) (*NotificationTemplate, error)` | Get template by ID |
| `GetTemplateByType` | `(ctx, notificationType NotificationType) (*NotificationTemplate, error)` | Get template by notification type |
| `ListTemplates` | `(ctx, category *NotificationCategory, q query.QueryOptions) ([]*NotificationTemplate, error)` | List templates, optionally filtered by category |
| `UpdateTemplate` | `(ctx, id uuid.UUID, input UpdateTemplateInput) (*NotificationTemplate, error)` | Update template. System-managed templates have restricted updates. |
| `DeleteTemplate` | `(ctx, id uuid.UUID) error` | Soft-delete. System-managed templates cannot be deleted. |
| `AddTranslation` | `(ctx, input CreateTemplateTranslationInput) (*NotificationTemplateTranslation, error)` | Add a translation for a language |
| `UpdateTranslation` | `(ctx, templateID uuid.UUID, language string, input UpdateTemplateTranslationInput) (*NotificationTemplateTranslation, error)` | Update existing translation |
| `DeleteTranslation` | `(ctx, templateID uuid.UUID, language string) error` | Remove a translation |
| `GetTranslations` | `(ctx, templateID uuid.UUID) ([]*NotificationTemplateTranslation, error)` | List all translations for a template |

---

### 2. NotificationPreferenceUsecase

Manages per-type, per-channel notification preferences (sparse override model).

| Method | Signature | Description |
|---|---|---|
| `SetPreference` | `(ctx, accountID uuid.UUID, input SetPreferenceInput) error` | Create or update a preference override |
| `GetPreferences` | `(ctx, accountID uuid.UUID) ([]*UserNotificationPreference, error)` | List all explicit overrides for an account |
| `GetEffectivePreference` | `(ctx, accountID uuid.UUID, notificationType NotificationType, channel Channel) (bool, error)` | Resolve the effective preference: check override, then fall back to template default, then check IAM global toggle |
| `IsQuietHours` | `(ctx, accountID uuid.UUID, notificationType NotificationType, channel Channel) (bool, error)` | Check if current time falls within quiet hours for this preference |
| `DeletePreference` | `(ctx, accountID uuid.UUID, notificationType NotificationType, channel Channel) error` | Remove override — reverts to default |

---

### 3. NotificationMuteUsecase

Manages account-level muting. Category-level mutes are handled by `NotificationPreferenceUsecase` (disable all types in a category). Module-specific mutes are checked via `MuteResolver` interface.

| Method | Signature | Description |
|---|---|---|
| `MuteAccount` | `(ctx, accountID uuid.UUID, input MuteAccountInput) error` | Mute another account |
| `UnmuteAccount` | `(ctx, accountID uuid.UUID, mutedAccountID uuid.UUID) error` | Remove mute |
| `IsMuted` | `(ctx, accountID uuid.UUID, mutedAccountID uuid.UUID) (bool, error)` | Check if muted (checks MuteUntil expiry) |
| `ListMutedAccounts` | `(ctx, accountID uuid.UUID, q query.QueryOptions) ([]*MutedAccount, error)` | List all muted accounts for a user |

---

### 4. NotificationDeviceUsecase

Manages device registration for push notifications.

| Method | Signature | Description |
|---|---|---|
| `RegisterDevice` | `(ctx, accountID uuid.UUID, input RegisterDeviceInput) (*UserDevice, error)` | Register a new device. If DeviceToken already exists, updates push token instead of creating duplicate. |
| `UpdateDevice` | `(ctx, accountID uuid.UUID, deviceID uuid.UUID, input UpdateDeviceInput) (*UserDevice, error)` | Update device info (push token refresh, app version update) |
| `DeactivateDevice` | `(ctx, accountID uuid.UUID, deviceID uuid.UUID) error` | Deactivate a specific device |
| `ListDevices` | `(ctx, accountID uuid.UUID) ([]*UserDevice, error)` | List all active devices for an account |
| `DeactivateAllDevices` | `(ctx, accountID uuid.UUID) error` | Deactivate all devices (on account suspension) |

---

### 5. NotificationIngestUsecase

The ingestion half of the delivery pipeline. Triggered by RabbitMQ events from other modules. Responsible for: event processing, preference resolution, template rendering, and enqueueing.

| Method | Signature | Description |
|---|---|---|
| `ProcessEvent` | `(ctx, input ProcessEventInput) error` | Entry point for domain events. Resolves template, checks preferences, renders content, enqueues for delivery. |
| `SendNotification` | `(ctx, input SendNotificationInput) error` | Direct send (for system-triggered notifications that bypass events). Renders and enqueues. |
| `SendMultiChannel` | `(ctx, accountID uuid.UUID, notificationType NotificationType, variables map[string]string, metadata map[string]interface{}, channels []Channel, expiresAt *time.Time) error` | Convenience method to send to multiple channels at once (e.g., InApp + Email) |

**`ProcessEvent` internal flow:**

1. Look up `NotificationTemplate` by `NotificationType`
2. Resolve user's locale → select translation or fall back to `DefaultContent`
3. Check `NotificationPreferenceUsecase.GetEffectivePreference` for each channel
4. Check `NotificationMuteUsecase.IsMuted` if metadata contains an `authorAccountId`
5. Check `MuteResolver.IsMuted` for module-specific mutes (e.g., thread mutes)
6. Render template: replace `{{variables}}` in content for each enabled channel
7. Calculate `ExpiresAt` — event override → template DefaultTTL → null
8. For each enabled channel: create `NotificationQueue` entry with rendered `Payload`

---

### 6. NotificationDeliveryUsecase

The delivery half of the pipeline. Triggered by a worker/cron. Responsible for: fetching pending queue items, attempting delivery, handling results, creating history + inbox entries.

| Method | Signature | Description |
|---|---|---|
| `ProcessQueue` | `(ctx, batchSize int) error` | Worker entry point. Fetches pending items, delivers them, handles results. |
| `DeliverItem` | `(ctx, queueID uuid.UUID) error` | Deliver a single queue item. Marks processing, sends via channel, handles result. |
| `HandleDeliveryResult` | `(ctx, queueID uuid.UUID, success bool, errMsg *string) error` | On success: create NotificationHistory + UserNotificationInbox, mark queue delivered. On failure: increment retry or mark failed. |
| `RetryFailed` | `(ctx, batchSize int) error` | Manual retry trigger for failed items (admin action) |
| `CancelPendingForAccount` | `(ctx, accountID uuid.UUID) error` | Cancel all pending notifications for an account (on suspension) |

**`ProcessQueue` internal flow:**

1. `FetchPending(batchSize)` — items where `Status = 'pending'` AND `ScheduledFor <= now()`
2. For each item: `MarkProcessing` → `DeliverViaChannel` → `HandleDeliveryResult`
3. On success: create `NotificationHistory` row, create `UserNotificationInbox` row (InApp channel), update `NotificationQueue.Status = 'delivered'`
4. On failure: `IncrementRetry` with exponential backoff, or `MarkFailed` if max retries exceeded

**`DeliverViaChannel` dispatch:**

| Channel | Handler |
|---|---|
| `InApp` | Direct — create NotificationHistory + UserNotificationInbox. No external provider needed. |
| `Email` | Call Resend API via `infrastructure/email/resend_provider.go` |
| `Push` | Call FCM/APNs (future — placeholder for Phase 1) |
| `SMS` | Future — placeholder |

---

### 7. NotificationInboxUsecase

User-facing inbox operations. The most-read path in the module.

| Method | Signature | Description |
|---|---|---|
| `ListInbox` | `(ctx, accountID uuid.UUID, category *NotificationCategory, q query.QueryOptions) ([]*UserNotificationInbox, error)` | Paginated inbox. Excludes archived and expired. Optional category filter. |
| `GetUnreadCount` | `(ctx, accountID uuid.UUID) (int64, error)` | Badge count for UI |
| `MarkAsRead` | `(ctx, accountID uuid.UUID, inboxID uuid.UUID) error` | Mark single notification as read |
| `MarkAllAsRead` | `(ctx, accountID uuid.UUID) error` | Mark all unread as read |
| `MarkCategoryAsRead` | `(ctx, accountID uuid.UUID, category NotificationCategory) error` | Mark all unread in a category as read |
| `ArchiveNotification` | `(ctx, accountID uuid.UUID, inboxID uuid.UUID) error` | Archive a notification |
| `DeleteNotification` | `(ctx, accountID uuid.UUID, inboxID uuid.UUID) error` | Hard-delete from inbox (audit record in NotificationHistory persists) |

---

### 8. NotificationHistoryUsecase

Read-only delivery audit trail. Used by admins for monitoring and by the system for webhook updates.

| Method | Signature | Description |
|---|---|---|
| `ListByAccount` | `(ctx, accountID uuid.UUID, q query.QueryOptions) ([]*NotificationHistory, error)` | Paginated history for a user |
| `GetByID` | `(ctx, id uuid.UUID) (*NotificationHistory, error)` | Single history entry |
| `MarkRead` | `(ctx, id uuid.UUID) error` | Update ReadAt (called from inbox usecase) |
| `MarkClicked` | `(ctx, id uuid.UUID) error` | Update ClickedAt (called from email/push click tracking) |
| `UpdateDeliveryStatus` | `(ctx, id uuid.UUID, status DeliveryStatus, deliveredAt *time.Time) error` | Webhook-driven status update |

---

### 9. NotificationCampaignUsecase

Admin campaign management. Segments are resolved to static AccountID lists at schedule time.

| Method | Signature | Description |
|---|---|---|
| `CreateCampaign` | `(ctx, createdBy uuid.UUID, input CreateCampaignInput) (*NotificationCampaign, error)` | Create a new campaign in Draft status |
| `GetCampaign` | `(ctx, id uuid.UUID) (*NotificationCampaign, error)` | Get campaign by ID |
| `ListCampaigns` | `(ctx, status *CampaignStatus, q query.QueryOptions) ([]*NotificationCampaign, error)` | List campaigns with optional status filter |
| `UpdateCampaign` | `(ctx, id uuid.UUID, input UpdateCampaignInput) (*NotificationCampaign, error)` | Update campaign (only in Draft status) |
| `ScheduleCampaign` | `(ctx, input ScheduleCampaignInput) error` | Transition to Scheduled. Resolves target segment to static AccountID list. Validates template exists. |
| `CancelCampaign` | `(ctx, id uuid.UUID) error` | Transition to Cancelled. Cancels all pending queue items for this campaign. |
| `ProcessScheduledCampaigns` | `(ctx) error` | Worker entry point. Fetches campaigns where Status = 'scheduled' AND ScheduledFor <= now(). Creates queue entries for all resolved recipients. Transitions status to Sending, then Completed. |

---

### 10. EmailDeliveryUsecase

Handles Resend webhook events and updates delivery logs + history.

| Method | Signature | Description |
|---|---|---|
| `HandleWebhookEvent` | `(ctx, event ResendWebhookEvent) error` | Process a single webhook event from Resend. Looks up EmailDeliveryLog by ProviderMessageID, updates the appropriate field. Also updates NotificationHistory.DeliveryStatus if needed. |
| `GetDeliveryLog` | `(ctx, historyID uuid.UUID) (*EmailDeliveryLog, error)` | Get email delivery details for a notification |
| `GetDeliveryLogByProviderID` | `(ctx, providerMessageID string) (*EmailDeliveryLog, error)` | Lookup by provider message ID |

**`HandleWebhookEvent` internal flow:**

1. Look up `EmailDeliveryLog` by `ProviderMessageID`
2. If not found — log warning and return (stale or unknown event)
3. Update the appropriate field based on `EventType`:

| EventType | Field Updated | Also Updates |
|---|---|---|
| `delivered` | DeliveredAt | NotificationHistory.DeliveryStatus → Delivered |
| `opened` | OpenedAt | NotificationHistory.ReadAt |
| `clicked` | ClickedAt | NotificationHistory.ClickedAt |
| `bounced` | BounceReason | NotificationHistory.DeliveryStatus → Bounced |
| `complained` | Complaint = true | — |

4. Run within a transaction to keep EmailDeliveryLog + NotificationHistory in sync

---

## Application Services

Services in `application/service/` compose use cases and add infrastructure concerns, following the `CommunityService` pattern.

### NotificationService (facade)

Composes: `NotificationIngestUsecase`, `NotificationDeliveryUsecase`, `NotificationInboxUsecase`

Exposes higher-level operations that coordinate multiple use cases:

- `SendAndDeliver` — Ingest + immediately process (for urgent notifications)
- `GetInboxSummary` — Unread count + recent notifications (for notification bell dropdown)

### TemplateRenderer

Renders `DefaultContent` or translation JSONB with variable substitution.

- `Render` — Takes template content + variables, returns rendered content per channel
- Variable format: `{{variableName}}` (consistent with existing `pkg/email` pattern)
- Validates all required variables are present before rendering

### CampaignProcessor

Resolves campaign target segments into AccountID lists.

- `ResolveSegment` — Takes TargetSegment filters, queries accounts matching criteria, returns []AccountID
- `ProcessCampaign` — For a campaign: resolve segment → render template per recipient → create queue entries
