# Notification Module — Architecture

## Module Directory Structure

```
internal/modules/notification/
├── module.go                              # Uber FX DI wiring
├── entities.go                            # EntityProvider for SchemaManager
│
├── domain/
│   ├── entity/
│   │   ├── notification_template.go
│   │   ├── notification_template_translation.go
│   │   ├── user_notification_preference.go
│   │   ├── muted_account.go
│   │   ├── user_device.go
│   │   ├── notification_queue.go
│   │   ├── notification_history.go
│   │   ├── user_notification_inbox.go
│   │   ├── notification_campaign.go
│   │   ├── email_delivery_log.go
│   │   └── enums.go
│   │
│   ├── repository/
│   │   ├── notification_template_repository.go
│   │   ├── user_notification_preference_repository.go
│   │   ├── muted_account_repository.go
│   │   ├── user_device_repository.go
│   │   ├── notification_queue_repository.go
│   │   ├── notification_history_repository.go
│   │   ├── user_notification_inbox_repository.go
│   │   ├── notification_campaign_repository.go
│   │   ├── email_delivery_log_repository.go
│   │   └── mute_resolver.go              # Cross-module interface
│   │
│   ├── usecase/
│   │   ├── inputs.go                     # All input DTOs
│   │   ├── notification_template_usecase.go
│   │   ├── notification_preference_usecase.go
│   │   ├── notification_mute_usecase.go
│   │   ├── notification_device_usecase.go
│   │   ├── notification_ingest_usecase.go
│   │   ├── notification_delivery_usecase.go
│   │   ├── notification_inbox_usecase.go
│   │   ├── notification_history_usecase.go
│   │   ├── notification_campaign_usecase.go
│   │   └── email_delivery_usecase.go
│   │
│   ├── event/
│   │   └── events.go                     # Event name constants (subscription contracts)
│   │
│   └── error/
│       └── errors.go                     # Domain-specific errors
│
├── application/
│   ├── usecase/
│   │   ├── notification_template_usecase.go
│   │   ├── notification_preference_usecase.go
│   │   ├── notification_mute_usecase.go
│   │   ├── notification_device_usecase.go
│   │   ├── notification_ingest_usecase.go
│   │   ├── notification_delivery_usecase.go
│   │   ├── notification_inbox_usecase.go
│   │   ├── notification_history_usecase.go
│   │   ├── notification_campaign_usecase.go
│   │   └── email_delivery_usecase.go
│   │
│   └── service/
│       ├── notification_service.go       # Facade composing ingest + delivery + inbox
│       ├── template_renderer.go          # Template variable substitution
│       └── campaign_processor.go         # Segment resolution + bulk queue creation
│
├── infrastructure/
│   ├── repository/
│   │   ├── helpers.go                    # applyPaginationAndSorting, getDB
│   │   ├── notification_template_repository.go
│   │   ├── user_notification_preference_repository.go
│   │   ├── muted_account_repository.go
│   │   ├── user_device_repository.go
│   │   ├── notification_queue_repository.go
│   │   ├── notification_history_repository.go
│   │   ├── user_notification_inbox_repository.go
│   │   ├── notification_campaign_repository.go
│   │   └── email_delivery_log_repository.go
│   │
│   └── email/
│       └── resend_provider.go            # Resend API client + webhook signature verification
│
└── delivery/
    ├── handler/
    │   ├── notification_handler.go       # User-facing: inbox, preferences, devices, mutes
    │   ├── notification_admin_handler.go # Admin: templates, campaigns, history
    │   └── webhook_handler.go            # Resend webhook endpoint
    │
    ├── dto/
    │   ├── notification_dto.go           # User-facing request/response DTOs
    │   └── notification_admin_dto.go     # Admin request/response DTOs
    │
    └── routes/
        ├── routes.go                     # RouteDependencies + RegisterRoutes
        ├── notification_routes.go        # User-facing routes
        └── notification_admin_routes.go  # Admin routes
```

---

## Dependency Injection (module.go)

The module follows the same Uber FX pattern as IAM, community, and guide modules.

### Providers

| Layer | Provide | Bind To |
|---|---|---|
| Entity | `NewEntityProvider` | — |
| Repository | `infrarepo.NewNotificationTemplateRepository` | `repository.NotificationTemplateRepository` |
| Repository | `infrarepo.NewUserNotificationPreferenceRepository` | `repository.UserNotificationPreferenceRepository` |
| Repository | `infrarepo.NewMutedAccountRepository` | `repository.MutedAccountRepository` |
| Repository | `infrarepo.NewUserDeviceRepository` | `repository.UserDeviceRepository` |
| Repository | `infrarepo.NewNotificationQueueRepository` | `repository.NotificationQueueRepository` |
| Repository | `infrarepo.NewNotificationHistoryRepository` | `repository.NotificationHistoryRepository` |
| Repository | `infrarepo.NewUserNotificationInboxRepository` | `repository.UserNotificationInboxRepository` |
| Repository | `infrarepo.NewNotificationCampaignRepository` | `repository.NotificationCampaignRepository` |
| Repository | `infrarepo.NewEmailDeliveryLogRepository` | `repository.EmailDeliveryLogRepository` |
| Usecase | `appusecase.NewNotificationTemplateUsecase` | `usecase.NotificationTemplateUsecase` |
| Usecase | `appusecase.NewNotificationPreferenceUsecase` | `usecase.NotificationPreferenceUsecase` |
| Usecase | `appusecase.NewNotificationMuteUsecase` | `usecase.NotificationMuteUsecase` |
| Usecase | `appusecase.NewNotificationDeviceUsecase` | `usecase.NotificationDeviceUsecase` |
| Usecase | `appusecase.NewNotificationIngestUsecase` | `usecase.NotificationIngestUsecase` |
| Usecase | `appusecase.NewNotificationDeliveryUsecase` | `usecase.NotificationDeliveryUsecase` |
| Usecase | `appusecase.NewNotificationInboxUsecase` | `usecase.NotificationInboxUsecase` |
| Usecase | `appusecase.NewNotificationHistoryUsecase` | `usecase.NotificationHistoryUsecase` |
| Usecase | `appusecase.NewNotificationCampaignUsecase` | `usecase.NotificationCampaignUsecase` |
| Usecase | `appusecase.NewEmailDeliveryUsecase` | `usecase.EmailDeliveryUsecase` |
| Service | `service.NewNotificationService` | `service.NotificationService` |
| Service | `service.NewTemplateRenderer` | — |
| Service | `service.NewCampaignProcessor` | — |
| Infrastructure | `email.NewResendProvider` | — |
| Handler | `handler.NewNotificationHandler` | — |
| Handler | `handler.NewNotificationAdminHandler` | — |
| Handler | `handler.NewWebhookHandler` | — |

### Invocations

| Invoke | Purpose |
|---|---|
| Register EntityProvider with SchemaManager | Enable migration generation |
| Subscribe to RabbitMQ events | Register event handlers for notification ingestion |
| Register HTTP routes | Wire handlers with middleware |
| Start delivery worker (fx.Lifecycle OnStart) | Begin polling NotificationQueue |
| Start campaign scheduler (fx.Lifecycle OnStart) | Begin checking scheduled campaigns |
| Start inbox expiry cleanup (fx.Lifecycle OnStart) | Periodic cleanup of expired inbox entries |

---

## Event Flow

### Ingestion Pipeline (Event → Queue)

```
┌─────────────┐    ┌────────────┐    ┌──────────────────┐    ┌───────────────┐
│ Source Module│    │  RabbitMQ  │    │  Ingest Usecase  │    │  Queue Table  │
│              │───>│    Bus     │───>│                  │───>│               │
│ Publishes    │    │            │    │ 1. Load template │    │ Status:       │
│ domain event │    │ Routes to  │    │ 2. Check prefs   │    │   pending     │
│              │    │ subscriber │    │ 3. Check mutes   │    │ Payload:      │
│              │    │            │    │ 4. Render content│    │   rendered    │
│              │    │            │    │ 5. Calculate TTL │    │ ScheduledFor: │
│              │    │            │    │ 6. Enqueue       │    │   now()       │
└─────────────┘    └────────────┘    └──────────────────┘    └───────────────┘
```

### Delivery Pipeline (Queue → History + Inbox)

```
┌───────────────┐    ┌──────────────────┐    ┌────────────────┐    ┌──────────┐
│  Queue Table  │    │ Delivery Usecase │    │ Channel Handler│    │ History  │
│               │───>│                  │───>│                │───>│ + Inbox  │
│ Status:       │    │ 1. Fetch pending │    │ InApp: direct  │    │          │
│   pending     │    │ 2. Mark process  │    │ Email: Resend  │    │ Create   │
│               │    │ 3. Dispatch      │    │ Push: FCM/APNs │    │ records  │
│               │    │ 4. Handle result │    │ SMS: future    │    │          │
└───────────────┘    └──────────────────┘    └────────────────┘    └──────────┘
        │                    │
        │         On failure │
        │                    ▼
        │            ┌──────────────┐
        └───────────>│ Increment    │
                     │ retryCount   │
                     │ + backoff    │
                     └──────────────┘
```

### Webhook Pipeline (Resend → EmailDeliveryLog + History)

```
┌────────────┐    ┌────────────────┐    ┌──────────────────┐    ┌──────────────┐
│   Resend   │    │ Webhook Handler│    │ Email Delivery   │    │ History +    │
│   Webhook  │───>│                │───>│ Usecase          │───>│ EmailLog     │
│            │    │ 1. Verify sig  │    │ 1. Lookup log by │    │              │
│  delivered │    │ 2. Parse event │    │    providerMsgID │    │ Update       │
│  opened    │    │ 3. Route to UC │    │ 2. Update field  │    │ delivery     │
│  clicked   │    │                │    │ 3. Update history│    │ status       │
│  bounced   │    │                │    │                  │    │              │
│  complained│    │                │    │                  │    │              │
└────────────┘    └────────────────┘    └──────────────────┘    └──────────────┘
```

---

## Cross-Module Dependencies

### Dependencies FROM notification module TO other modules

| Dependency | Interface | Purpose |
|---|---|---|
| IAM | `NotificationPreferenceRepository` (in IAM) | Read global channel toggles |
| IAM | `token.TokenService` | Auth middleware for routes |
| IAM | `service.AuthService` | Account status middleware |
| Community | `MuteResolver` (implemented by community) | Check thread/category mute status |
| Core | `*core.Database` | DB access, Transactor |
| Core | `core.Logger` | Logging |
| Core | `core.Config` | App configuration |
| Core | `rabbitmq.Bus` | Subscribe to domain events |
| Pkg | `email.Emailer` | Email delivery (existing SMTP, to be upgraded to Resend) |
| Pkg | `storage.Storage` | (if needed for template assets) |

### Dependencies FROM other modules TO notification module

| Source Module | What it calls | Purpose |
|---|---|---|
| Community | Publishes events to RabbitMQ | Triggers CommunityReply, CommunitySolution, CommunityMention notifications |
| Guide | Publishes events to RabbitMQ | Triggers GuideStepCompleted, GuideDeadline, GuideUpdate notifications |
| IAM | Publishes events to RabbitMQ | Triggers WelcomeMessage, AccountVerification, PasswordReset, AccountAlert notifications |
| AI | Publishes events to RabbitMQ | Triggers AIQuotaLimit, AIResponseReady notifications |

All cross-module communication is **async via RabbitMQ events**. No module directly imports the notification module's use cases. This keeps modules decoupled.

---

## Event Subscription Contract

Defined in `domain/event/events.go`. These are the event names the notification module subscribes to. Event names are owned by the source module (consistent with IAM's existing pattern).

### Subscribed events:

| Event Name | Source Module | NotificationType |
|---|---|---|
| `account.registered` | IAM | WelcomeMessage |
| `account.verification` | IAM | AccountVerification |
| `password.reset` | IAM | PasswordReset |
| `account.alert` | IAM | AccountAlert |
| `thread.reply` | Community | CommunityReply |
| `thread.solution` | Community | CommunitySolution |
| `post.mention` | Community | CommunityMention |
| `step.completed` | Guide | GuideStepCompleted |
| `deadline.approaching` | Guide | GuideDeadline |
| `guide.updated` | Guide | GuideUpdate |
| `ai.quota_limit` | AI | AIQuotaLimit |
| `ai.response_ready` | AI | AIResponseReady |
| `system.announcement` | System | SystemAnnouncement |
| `policy.update` | System | PolicyUpdate |

### Published events:

| Event Name | When Published |
|---|---|
| `notification.failed` | When a notification exhausts all retries |

---

## Route Structure

### User-facing routes (`/api/v1/notifications/...`)

All require `AuthMiddleware` + `AccountStatusMiddleware`.

| Method | Route | Handler | Description |
|---|---|---|---|
| GET | `/inbox` | ListInbox | Paginated inbox with category filter |
| GET | `/inbox/unread-count` | GetUnreadCount | Badge count |
| PATCH | `/inbox/{id}/read` | MarkAsRead | Mark single as read |
| POST | `/inbox/read-all` | MarkAllAsRead | Mark all as read |
| POST | `/inbox/read-all/{category}` | MarkCategoryAsRead | Mark category as read |
| DELETE | `/inbox/{id}` | DeleteNotification | Remove from inbox |
| PATCH | `/inbox/{id}/archive` | ArchiveNotification | Archive |
| GET | `/preferences` | ListPreferences | All overrides for current user |
| PUT | `/preferences` | SetPreference | Create/update override |
| DELETE | `/preferences/{type}/{channel}` | DeletePreference | Remove override |
| GET | `/devices` | ListDevices | All registered devices |
| POST | `/devices` | RegisterDevice | Register new device |
| PATCH | `/devices/{id}` | UpdateDevice | Update device info |
| DELETE | `/devices/{id}` | DeactivateDevice | Deactivate device |
| GET | `/mutes` | ListMutedAccounts | List muted accounts |
| POST | `/mutes` | MuteAccount | Mute an account |
| DELETE | `/mutes/{accountId}` | UnmuteAccount | Unmute an account |
| GET | `/history` | ListHistory | Delivery audit trail |

### Admin routes (`/api/v1/admin/notifications/...`)

Require `AuthMiddleware` + `AccountStatusMiddleware` + permission-specific middleware.

| Method | Route | Handler | Description |
|---|---|---|---|
| GET | `/templates` | ListTemplates | List all templates |
| POST | `/templates` | CreateTemplate | Create template |
| GET | `/templates/{id}` | GetTemplate | Get template + translations |
| PATCH | `/templates/{id}` | UpdateTemplate | Update template |
| DELETE | `/templates/{id}` | DeleteTemplate | Delete template (non-system only) |
| POST | `/templates/{id}/translations` | AddTranslation | Add translation |
| PATCH | `/templates/{id}/translations/{lang}` | UpdateTranslation | Update translation |
| DELETE | `/templates/{id}/translations/{lang}` | DeleteTranslation | Delete translation |
| GET | `/campaigns` | ListCampaigns | List campaigns |
| POST | `/campaigns` | CreateCampaign | Create campaign |
| GET | `/campaigns/{id}` | GetCampaign | Get campaign |
| PATCH | `/campaigns/{id}` | UpdateCampaign | Update draft campaign |
| POST | `/campaigns/{id}/schedule` | ScheduleCampaign | Schedule campaign |
| POST | `/campaigns/{id}/cancel` | CancelCampaign | Cancel campaign |
| GET | `/queue/status` | GetQueueStatus | Queue metrics by status |
| POST | `/queue/retry-failed` | RetryFailed | Manual retry trigger |

### Webhook route (public — no auth)

| Method | Route | Handler | Description |
|---|---|---|---|
| POST | `/webhooks/resend` | HandleResendWebhook | Receive Resend delivery events |

Webhook authentication is via Resend's signature verification (not JWT auth middleware).

---

## Workers / Background Jobs

### 1. Delivery Worker

- **Trigger:** `fx.Lifecycle.OnStart` — runs as a goroutine
- **Poll interval:** 5 seconds
- **Batch size:** 50 items per poll
- **Logic:** `NotificationDeliveryUsecase.ProcessQueue(ctx, 50)`
- **Shutdown:** Respects context cancellation

### 2. Campaign Scheduler

- **Trigger:** `fx.Lifecycle.OnStart` — runs as a goroutine
- **Poll interval:** 30 seconds
- **Logic:** `NotificationCampaignUsecase.ProcessScheduledCampaigns(ctx)`
- **Shutdown:** Respects context cancellation

### 3. Inbox Expiry Cleanup

- **Trigger:** `fx.Lifecycle.OnStart` — runs as a goroutine
- **Poll interval:** 1 hour
- **Logic:** `NotificationInboxUsecase` calls `ExpireOld` with `before = now()`
- **Shutdown:** Respects context cancellation

---

## Module Registration

Add to `internal/modules/modules.go`:

```go
var Modules = fx.Options(
    ai.Module,
    iam.Module,
    guide.Module,
    community.Module,
    coregrpc.Module,
    notification.Module,  // new
)
```

---

## Resend Integration

### Email Provider Upgrade Path

The current `pkg/email` uses SMTP via `gomail.v2`. The notification module introduces Resend as a dedicated email provider for notification delivery. This coexists with the existing SMTP emailer (used for welcome emails, OTPs, etc.) and can replace it in a future iteration.

### Resend Provider (`infrastructure/email/resend_provider.go`)

Responsibilities:

- Send email via Resend REST API
- Return provider message ID for webhook correlation
- Verify webhook signatures using Resend's signing secret
- Map Resend event types to `DeliveryStatus` values

### Configuration

Add to `core.Config`:

```
Resend struct {
    Enabled       bool   `mapstructure:"enabled"`
    APIKey        string `mapstructure:"api_key"`
    WebhookSecret string `mapstructure:"webhook_secret"`
    FromEmail     string `mapstructure:"from_email"`
    FromName      string `mapstructure:"from_name"`
}
```
