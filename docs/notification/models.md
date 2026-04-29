# Notification Module — Entities & Models

## Enums

### NotificationCategory

Domain-aligned categories mapping to existing platform modules.

| Value | Description |
|---|---|
| `System` | Platform-level announcements, policy updates, welcome messages |
| `Community` | Community replies, solutions, mentions |
| `Guide` | Guide step completion, deadlines, updates |
| `AI` | AI quota limits, response readiness |
| `Security` | Account alerts, verification, password resets |
| `Payment` | Payment confirmations |
| `Marketing` | Reserved for campaign-driven notifications |

### NotificationType

Each type maps to exactly one category.

| Value | Category | Description |
|---|---|---|
| `SystemAnnouncement` | System | Platform-wide announcements |
| `PolicyUpdate` | System | Terms/policy change notifications |
| `WelcomeMessage` | System | New account welcome message |
| `CommunityReply` | Community | Someone replied to your thread |
| `CommunitySolution` | Community | Your thread received a solution |
| `CommunityMention` | Community | You were mentioned in a post |
| `GuideStepCompleted` | Guide | A guide step was marked complete |
| `GuideDeadline` | Guide | A compliance deadline is approaching |
| `GuideUpdate` | Guide | A guide you follow was updated |
| `AIQuotaLimit` | AI | AI usage quota reached or near limit |
| `AIResponseReady` | AI | AI async response is ready |
| `AccountAlert` | Security | Security alert on account |
| `AccountVerification` | Security | Account verification required |
| `PasswordReset` | Security | Password reset requested |
| `PaymentConfirmation` | Payment | Payment transaction confirmed |

### NotificationPriority

Stored as `smallint` in the database. Higher value = higher priority.

| Value | Integer Weight | Description |
|---|---|---|
| `Low` | 0 | Informational, no urgency |
| `Medium` | 1 | Default priority |
| `High` | 2 | Time-sensitive, should surface promptly |
| `Urgent` | 3 | Critical, must be seen immediately (e.g., security alerts) |

### Channel

| Value | Description |
|---|---|
| `InApp` | In-app notification (stored in inbox) |
| `Email` | Email delivery (via Resend) |
| `Push` | Push notification (FCM/APNs) |
| `SMS` | SMS delivery (future) |

### NotificationStatus

Lifecycle status for `NotificationQueue` entries.

| Value | Description |
|---|---|
| `Pending` | Enqueued, awaiting delivery |
| `Processing` | Currently being delivered |
| `Delivered` | Successfully delivered |
| `Failed` | Delivery failed after max retries |
| `Cancelled` | Cancelled (user mute, account deactivation, admin action) |

### DeliveryStatus

Final delivery outcome for `NotificationHistory` and `EmailDeliveryLog`.

| Value | Description |
|---|---|
| `Sent` | Dispatched to provider |
| `Delivered` | Confirmed delivered by provider |
| `Failed` | Delivery failed |
| `Bounced` | Recipient address bounced |

### CampaignType

| Value | Description |
|---|---|
| `Broadcast` | Send to all accounts |
| `Segmented` | Send to filtered segment of accounts |

> `Triggered` campaign type is deferred to Phase 2. Event-driven notifications use the RabbitMQ → ingest pipeline instead.

### CampaignStatus

| Value | Description |
|---|---|
| `Draft` | Created but not scheduled |
| `Scheduled` | Scheduled for future send |
| `Sending` | Currently being processed |
| `Completed` | All notifications sent |
| `Cancelled` | Cancelled by admin |

### DeviceType

| Value | Description |
|---|---|
| `Android` | Android mobile device |
| `iOS` | Apple mobile device |
| `Web` | Web browser |

---

## Entities

### Shared Patterns

All entities (except `NotificationTemplateTranslation`) embed `model.BaseModel` which provides:

| Field | Type | Description |
|---|---|---|
| `ID` | uuid | PK, auto-generated via `gen_random_uuid()` |
| `CreatedAt` | timestamptz | Not null, default: `CURRENT_TIMESTAMP` |
| `UpdatedAt` | timestamptz | Not null, default: `CURRENT_TIMESTAMP` |
| `DeletedAt` | timestamptz | Nullable, indexed — enables GORM soft-delete |

All foreign keys to accounts use `AccountID` (not `UserID`), consistent with the delivery-channel binding pattern used across the codebase.

---

### 1. NotificationTemplate

**Table:** `notification_templates`

Defines the rendering template for each `NotificationType`. One template per type. Stores default (English) multi-channel content and optional per-language translations via `NotificationTemplateTranslation`.

| Field | Type | Constraints | Description |
|---|---|---|---|
| `BaseModel` | embedded | — | ID, CreatedAt, UpdatedAt, DeletedAt |
| `Name` | varchar(200) | not null, uniqueIndex | Human-readable template name |
| `Description` | text | nullable | What this template is for |
| `NotificationType` | varchar(64) | not null, uniqueIndex | The notification type this template renders |
| `Category` | varchar(32) | not null, index | Derived from NotificationType, stored for query efficiency |
| `Priority` | smallint | not null, default:1 | Default priority weight (0–3) |
| `IsSystemManaged` | boolean | not null, default:false | If true, cannot be deleted by admins |
| `DefaultContent` | jsonb | not null | Multi-channel content (see structure below) |
| `VariablesSchema` | jsonb | nullable | Defines template variables and their types |
| `DefaultTTL` | integer | nullable | Default time-to-live in **seconds**. Null = no expiry. |

**`DefaultContent` JSONB structure:**

```json
{
  "email": {
    "subject": "New reply to '{{threadTitle}}'",
    "body": "<p>{{authorName}} replied to your thread...</p>"
  },
  "push": {
    "title": "New reply",
    "body": "{{authorName}} replied to {{threadTitle}}"
  },
  "sms": {
    "body": "{{authorName}} replied to {{threadTitle}}"
  },
  "inapp": {
    "title": "New reply",
    "body": "{{authorName}} replied to {{threadTitle}}",
    "actionUrl": "/community/threads/{{threadSlug}}"
  }
}
```

**`VariablesSchema` JSONB structure:**

```json
{
  "required": ["threadTitle", "authorName", "threadSlug"],
  "optional": ["authorAvatarUrl"]
}
```

**GORM relationships:**

- `Translations []NotificationTemplateTranslation` — FK: TemplateID, OnDelete:CASCADE

---

### 2. NotificationTemplateTranslation

**Table:** `notification_template_translations`

Per-language translation of a template. Mirrors the parent's multi-channel `Content` structure. Unique per `(TemplateID, Language)`.

> Does NOT embed `BaseModel`. Uses its own simple PK + timestamps pattern, consistent with `GuideTranslation` and `GuideCategoryTranslation` in the codebase.

| Field | Type | Constraints | Description |
|---|---|---|---|
| `ID` | uuid | PK, auto-generated | Primary key |
| `TemplateID` | uuid | FK → notification_templates, not null, uniqueIndex(priority:1) | Parent template |
| `Language` | varchar(10) | not null, uniqueIndex(priority:2) | ISO 639-1 locale code (e.g., "en", "am") |
| `Subject` | varchar(500) | not null | Denormalized subject for quick lookups |
| `Content` | jsonb | not null | Multi-channel content, same structure as `DefaultContent` |
| `CreatedAt` | timestamptz | not null, default:CURRENT_TIMESTAMP | |
| `UpdatedAt` | timestamptz | not null, default:CURRENT_TIMESTAMP | |

**BeforeCreate hook:** Auto-generates UUID if ID is `uuid.Nil`.

---

### 3. UserNotificationPreference

**Table:** `user_notification_preferences`

Sparse override model — only stores rows where a user has explicitly deviated from template defaults. No row = use default (derived from `NotificationTemplate`).

| Field | Type | Constraints | Description |
|---|---|---|---|
| `BaseModel` | embedded | — | ID, CreatedAt, UpdatedAt, DeletedAt |
| `AccountID` | uuid | FK → accounts, not null, uniqueIndex(priority:1) | The account this preference belongs to |
| `NotificationType` | varchar(64) | not null, uniqueIndex(priority:2) | Which notification type |
| `Channel` | varchar(20) | not null, uniqueIndex(priority:3) | Which channel this override applies to |
| `IsEnabled` | boolean | not null | True = explicitly enabled, False = explicitly disabled |
| `QuietHoursStart` | time | nullable | Start of quiet hours (no delivery) |
| `QuietHoursEnd` | time | nullable | End of quiet hours (no delivery) |

**Unique index:** `idx_user_notif_prefs_account_type_channel` on `(AccountID, NotificationType, Channel)`

**Resolution logic at delivery time:**

1. Check if override row exists for `(AccountID, NotificationType, Channel)`
2. If yes → use `IsEnabled` value
3. If no → use template default (channel presence in `DefaultContent` = enabled)
4. Check quiet hours — if current time is within `[QuietHoursStart, QuietHoursEnd]`, defer delivery

---

### 4. MutedAccount

**Table:** `muted_accounts`

Cross-cutting mute: "I don't want notifications triggered by this specific account." Category-level mutes (e.g., "mute all community notifications") are handled by `UserNotificationPreference` (disable all types in that category).

Module-specific mutes (e.g., muting a specific thread) are owned by the source module (e.g., community module's `UserThreadSettings.IsMuted`).

| Field | Type | Constraints | Description |
|---|---|---|---|
| `BaseModel` | embedded | — | ID, CreatedAt, UpdatedAt, DeletedAt |
| `AccountID` | uuid | FK → accounts, not null, uniqueIndex(priority:1), index | Who muted |
| `MutedAccountID` | uuid | FK → accounts, not null, uniqueIndex(priority:2), index | Who is muted |
| `MuteUntil` | timestamptz | nullable | When the mute expires. Null = permanent. |
| `Reason` | text | nullable | Why the account was muted |

**Unique index:** `idx_muted_accounts_account_pair` on `(AccountID, MutedAccountID)`

**Resolution logic at delivery time:**

1. Check if `MutedAccount` row exists where `AccountID = recipient` AND `MutedAccountID = eventAuthor`
2. If `MuteUntil` is null → permanently muted
3. If `MuteUntil` is set and `now() < MuteUntil` → currently muted
4. Otherwise → not muted

---

### 5. UserDevice

**Table:** `user_devices`

Device registration for push notifications. One row per physical device per account.

| Field | Type | Constraints | Description |
|---|---|---|---|
| `BaseModel` | embedded | — | ID, CreatedAt, UpdatedAt, DeletedAt |
| `AccountID` | uuid | FK → accounts, not null, index: `idx_user_devices_account` | Device owner |
| `DeviceType` | varchar(20) | not null | Android, iOS, or Web |
| `DeviceToken` | varchar(512) | not null, uniqueIndex: `idx_user_devices_token` | Platform-specific device identifier (for deduplication) |
| `PushToken` | text | nullable | FCM registration token, APNs device token, or web push subscription JSON |
| `DeviceName` | varchar(200) | nullable | User-friendly device name |
| `DeviceModel` | varchar(200) | nullable | Hardware model (e.g., "Pixel 7") |
| `OSVersion` | varchar(50) | nullable | Operating system version |
| `AppVersion` | varchar(50) | nullable | Installed app version |
| `IsActive` | boolean | not null, default:true | False = device unregistered or token invalidated |
| `LastActiveAt` | timestamptz | nullable | Last time the app was opened on this device |

**Push token by device type:**

| DeviceType | PushToken content |
|---|---|
| Android | FCM registration token string |
| iOS | APNs device token string |
| Web | JSON: `{"endpoint":"...","keys":{"p256dh":"...","auth":"..."}}` |

---

### 6. NotificationQueue

**Table:** `notification_queue`

Processing queue for notifications awaiting delivery. Template is rendered at enqueue time — `Payload` contains the final rendered content. Ephemeral by design — can be cleaned up after successful delivery.

| Field | Type | Constraints | Description |
|---|---|---|---|
| `BaseModel` | embedded | — | ID, CreatedAt, UpdatedAt, DeletedAt |
| `NotificationType` | varchar(64) | not null | Type of notification |
| `AccountID` | uuid | FK → accounts, not null, index: `idx_notif_queue_account` | Recipient |
| `Priority` | smallint | not null, default:1 | Weight 0–3, higher = delivered first |
| `TemplateID` | uuid | FK → notification_templates, nullable | Template used for rendering |
| `Channel` | varchar(20) | not null | Target delivery channel |
| `Payload` | jsonb | not null | Fully rendered channel-specific content |
| `ScheduledFor` | timestamptz | not null, index: `idx_notif_queue_scheduled` | When to attempt delivery |
| `MaxRetries` | integer | not null, default:3 | Maximum retry attempts |
| `RetryCount` | integer | not null, default:0 | Current retry count |
| `Status` | varchar(20) | not null, default:'pending', index: `idx_notif_queue_status` | Pending/Processing/Delivered/Failed/Cancelled |
| `ErrorMessage` | text | nullable | Last error message if failed |

**Retry strategy:** Exponential backoff. On each failure, `ScheduledFor` is updated:
- 1st retry: `ScheduledFor = now + 1min`
- 2nd retry: `ScheduledFor = now + 2min`
- 3rd retry: `ScheduledFor = now + 4min`
- After `RetryCount >= MaxRetries` → Status = Failed

**Cancelled triggers:**
- User mutes the notification type after enqueue
- Account is deactivated/suspended
- Admin cancels a campaign mid-send

---

### 7. NotificationHistory

**Table:** `notification_histories`

Immutable delivery audit trail. Created at first successful delivery. Records what was sent, to whom, via which channel, and delivery outcomes. Never modified by user actions.

| Field | Type | Constraints | Description |
|---|---|---|---|
| `BaseModel` | embedded | — | ID, CreatedAt, UpdatedAt, DeletedAt |
| `AccountID` | uuid | FK → accounts, not null, index: `idx_notif_history_account` | Recipient |
| `NotificationType` | varchar(64) | not null | Type of notification |
| `Channel` | varchar(20) | not null | Delivery channel |
| `Title` | varchar(500) | not null | Rendered notification title |
| `Content` | text | not null | Rendered notification body |
| `ActionUrl` | varchar(512) | nullable | Deep link for notification action |
| `SentAt` | timestamptz | not null | When the notification was dispatched |
| `DeliveredAt` | timestamptz | nullable | When provider confirmed delivery |
| `ReadAt` | timestamptz | nullable | When the notification was first read |
| `ClickedAt` | timestamptz | nullable | When a link in the notification was clicked |
| `DeliveryStatus` | varchar(20) | not null | Sent/Delivered/Failed/Bounced |
| `FailureReason` | text | nullable | Error details if delivery failed |
| `Metadata` | jsonb | nullable | Source entity references only |

**`Metadata` JSONB structure:**

```json
{
  "sourceModule": "community",
  "sourceEvent": "thread.reply",
  "threadId": "uuid...",
  "postId": "uuid...",
  "authorAccountId": "uuid..."
}
```

> Note: `IsArchived` and `ExpiresAt` are NOT on this entity. They belong to `UserNotificationInbox` only. History is immutable.

---

### 8. UserNotificationInbox

**Table:** `user_notification_inboxes`

Mutable user-facing notification feed. Only created for notifications that pass preference/mute checks. User can read, archive, or delete entries. Entries can expire based on TTL rules.

| Field | Type | Constraints | Description |
|---|---|---|---|
| `BaseModel` | embedded | — | ID, CreatedAt, UpdatedAt, DeletedAt |
| `AccountID` | uuid | FK → accounts, not null, index: `idx_notif_inbox_account` | Inbox owner |
| `NotificationHistoryID` | uuid | FK → notification_histories, not null | Link to delivery record |
| `Category` | varchar(32) | not null, index: `idx_notif_inbox_category` | Denormalized from NotificationType → Category mapping |
| `ActionUrl` | varchar(512) | nullable | Denormalized from NotificationHistory for query performance |
| `IsRead` | boolean | not null, default:false | Has the user read this notification |
| `IsArchived` | boolean | not null, default:false | Has the user archived this notification |
| `ExpiresAt` | timestamptz | nullable, index: `idx_notif_inbox_expires` | When this inbox entry expires |

**`ExpiresAt` resolution:**

1. Check if source event payload includes `expiresAt` override → use it
2. Else check `NotificationTemplate.DefaultTTL` → `expiresAt = createdAt + defaultTTL seconds`
3. Else `expiresAt = null` (no expiry)

**Inbox creation flow:**

1. Notification delivered → `NotificationHistory` row created
2. Check preference/mute rules
3. If not muted → `UserNotificationInbox` row created
4. If muted → only history exists (audit), no inbox entry

---

### 9. NotificationCampaign

**Table:** `notification_campaigns`

Admin-created notification campaigns. Supports broadcast (all users) and segmented (filtered audience) sends. Target segment is resolved to a static list of AccountIDs when the campaign status changes to `Scheduled`.

| Field | Type | Constraints | Description |
|---|---|---|---|
| `BaseModel` | embedded | — | ID, CreatedAt, UpdatedAt, DeletedAt |
| `Name` | varchar(200) | not null | Campaign name |
| `Description` | text | nullable | Campaign description |
| `CampaignType` | varchar(20) | not null | Broadcast or Segmented |
| `TargetSegment` | jsonb | nullable | Filter rules (null for Broadcast) |
| `TemplateID` | uuid | FK → notification_templates, not null | Template to render |
| `CustomSubject` | varchar(500) | nullable | Override template subject |
| `CustomContent` | jsonb | nullable | Override template content |
| `ScheduledFor` | timestamptz | nullable | When to start sending |
| `SentAt` | timestamptz | nullable | When sending actually started |
| `Status` | varchar(20) | not null, default:'draft' | Draft/Scheduled/Sending/Completed/Cancelled |
| `CreatedBy` | uuid | FK → accounts, not null, index | Admin who created the campaign |

**`TargetSegment` JSONB structure (Segmented campaigns):**

```json
{
  "filters": {
    "accountTypes": ["business", "individual"],
    "roles": ["admin"],
    "registrationDateRange": {
      "from": "2025-01-01",
      "to": "2025-12-31"
    },
    "hasCompletedGuide": "business-registration"
  },
  "resolvedAccountIDs": ["uuid1", "uuid2", "..."]
}
```

> `resolvedAccountIDs` is populated when status transitions to `Scheduled`. This snapshot ensures consistent recipients even if account data changes between scheduling and sending.

**Permission enforcement:** Only accounts with module-bound write permissions can create/manage campaigns. Enforced at application layer via `PermissionMiddleware`.

---

### 10. EmailDeliveryLog

**Table:** `email_delivery_logs`

Tracks email delivery lifecycle for notifications sent via Resend. One row per email. Updated asynchronously as Resend webhook events arrive (delivered, opened, clicked, bounced, complained).

| Field | Type | Constraints | Description |
|---|---|---|---|
| `BaseModel` | embedded | — | ID, CreatedAt, UpdatedAt, DeletedAt |
| `NotificationHistoryID` | uuid | FK → notification_histories, not null | Link to notification record |
| `Provider` | varchar(50) | not null | Email provider name (e.g., "resend") |
| `ProviderMessageID` | varchar(255) | nullable, index: `idx_email_delivery_provider_msg` | Provider-assigned message ID for webhook matching |
| `RecipientEmail` | varchar(255) | not null | Recipient email address |
| `Subject` | varchar(500) | not null | Email subject line |
| `SentAt` | timestamptz | not null | When the email was dispatched |
| `DeliveredAt` | timestamptz | nullable | When provider confirmed delivery |
| `OpenedAt` | timestamptz | nullable | First open event timestamp |
| `ClickedAt` | timestamptz | nullable | First click event timestamp |
| `BounceReason` | text | nullable | Bounce details if email bounced |
| `Complaint` | boolean | not null, default:false | True if recipient filed a spam complaint |

**Webhook event mapping (Resend → fields):**

| Resend Event | Field Updated |
|---|---|
| `email.sent` | `SentAt` |
| `email.delivered` | `DeliveredAt` |
| `email.opened` | `OpenedAt` |
| `email.clicked` | `ClickedAt` |
| `email.bounced` | `BounceReason` |
| `email.complained` | `Complaint` = true |

> Only stores first open/click timestamps. Repeat engagement tracking would require an `EmailDeliveryEvent` table (Phase 2).

---

## Entity Relationship Summary

```
NotificationTemplate
  └── NotificationTemplateTranslation (1:N, CASCADE)

Account
  ├── UserNotificationPreference (1:N, CASCADE)
  ├── MutedAccount (as AccountID, 1:N, CASCADE)
  ├── MutedAccount (as MutedAccountID, 1:N, CASCADE)
  ├── UserDevice (1:N, CASCADE)
  ├── NotificationQueue (1:N, CASCADE)
  ├── NotificationHistory (1:N, CASCADE)
  ├── UserNotificationInbox (1:N, CASCADE)
  └── NotificationCampaign (as CreatedBy, 1:N, RESTRICT)

NotificationHistory
  ├── UserNotificationInbox (1:1, CASCADE)
  └── EmailDeliveryLog (1:1, CASCADE)

NotificationTemplate
  ├── NotificationQueue (1:N, SET NULL)
  └── NotificationCampaign (1:N, RESTRICT)
```

---

## IAM Module Integration

The IAM module already owns `NotificationPreference` (global channel toggles per account). The notification module's `UserNotificationPreference` layers on top as granular per-type overrides.

**Preference resolution order:**

1. **IAM global toggle** — If `EnableEmailNotification = false` → no email notifications, regardless of per-type settings
2. **Notification module per-type override** — If override exists → use `IsEnabled` value
3. **Template default** — Channel present in `DefaultContent` → enabled
4. **Quiet hours** — If current time is within quiet window → defer delivery

This layered approach avoids cross-module migration and keeps IAM's existing entity graph intact.
