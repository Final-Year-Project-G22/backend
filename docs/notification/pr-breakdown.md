# Notification Module — PR Breakdown Plan

## 1) PR Decomposition Summary

The notification module is broken into 8 pull requests, sequenced by dependency order. Foundation work (entities, interfaces, repos) comes first, then feature work in vertical slices (templates, pipeline, settings, email, campaigns), capped by the integration PR that wires everything together.

| PR | Title | Dependencies | Risk |
|----|-------|--------------|------|
| PR-N1 | Domain Foundation | None | Low |
| PR-N2 | Data Layer Implementation | PR-N1 | Low |
| PR-N3 | Template Management | PR-N2 | Low |
| PR-N4 | Core Delivery Pipeline | PR-N2, PR-N3 | High |
| PR-N5 | User Settings | PR-N2 | Medium |
| PR-N6 | Email Delivery + Webhooks | PR-N2, PR-N4 | Medium |
| PR-N7 | Campaign System | PR-N2, PR-N3 | Medium |
| PR-N8 | Module Integration | All above | Medium |

---

## 2) Clarifications Needed

All ambiguities resolved in prior conversation. Key decisions:
- Resend replaces SMTP and lives as a shared `pkg/email` upgrade alongside existing `gomail.v2` SMTP, coexisting initially
- System-managed templates seeded via migration file (not startup hook)
- Cross-module events: notification module defines expected event contracts; source modules add publishing when they're ready
- Event handler registration follows existing `internal/handlers/handlers.go` pattern

---

## 3) PR Plan

---

### PR-N1 — Domain Foundation

**Goal:** Establish all domain contracts — entities, enums, repository interfaces, use case interfaces, input DTOs, error codes, and the `MuteResolver` cross-module interface. Pure definitions with no implementations.

**Scope:**
- 9 enum types: `NotificationCategory`, `NotificationType`, `NotificationPriority`, `Channel`, `NotificationStatus`, `DeliveryStatus`, `CampaignType`, `CampaignStatus`, `DeviceType`
- 10 entity structs with GORM tags, `TableName()`, embedded `BaseModel`:
  - `NotificationTemplate`, `NotificationTemplateTranslation`, `UserNotificationPreference`, `MutedAccount`, `UserDevice`, `NotificationQueue`, `NotificationHistory`, `UserNotificationInbox`, `NotificationCampaign`, `EmailDeliveryLog`
- 9 repository interfaces extending `GenericRepository[T]`:
  - `NotificationTemplateRepository`, `UserNotificationPreferenceRepository`, `MutedAccountRepository`, `UserDeviceRepository`, `NotificationQueueRepository`, `NotificationHistoryRepository`, `UserNotificationInboxRepository`, `NotificationCampaignRepository`, `EmailDeliveryLogRepository`
- 10 use case interfaces with complete method signatures:
  - `NotificationTemplateUsecase`, `NotificationPreferenceUsecase`, `NotificationMuteUsecase`, `NotificationDeviceUsecase`, `NotificationIngestUsecase`, `NotificationDeliveryUsecase`, `NotificationInboxUsecase`, `NotificationHistoryUsecase`, `NotificationCampaignUsecase`, `EmailDeliveryUsecase`
- All input DTOs in `domain/usecase/inputs.go`
- `MuteResolver` interface: `IsMuted(ctx, accountID, itemType, itemID) (bool, error)`
- Domain error codes in `domain/error/errors.go`
- Empty directory scaffolding: `application/usecase/`, `application/service/`, `infrastructure/repository/`, `infrastructure/email/`, `delivery/handler/`, `delivery/dto/`, `delivery/routes/`

**Out of scope:** Any implementations, module wiring, config, migration files

**Dependencies:** None

**Files/modules likely touched:**
```
internal/modules/notification/domain/entity/enums.go
internal/modules/notification/domain/entity/notification_template.go
internal/modules/notification/domain/entity/notification_template_translation.go
internal/modules/notification/domain/entity/user_notification_preference.go
internal/modules/notification/domain/entity/muted_account.go
internal/modules/notification/domain/entity/user_device.go
internal/modules/notification/domain/entity/notification_queue.go
internal/modules/notification/domain/entity/notification_history.go
internal/modules/notification/domain/entity/user_notification_inbox.go
internal/modules/notification/domain/entity/notification_campaign.go
internal/modules/notification/domain/entity/email_delivery_log.go
internal/modules/notification/domain/repository/notification_template_repository.go
internal/modules/notification/domain/repository/user_notification_preference_repository.go
internal/modules/notification/domain/repository/muted_account_repository.go
internal/modules/notification/domain/repository/user_device_repository.go
internal/modules/notification/domain/repository/notification_queue_repository.go
internal/modules/notification/domain/repository/notification_history_repository.go
internal/modules/notification/domain/repository/user_notification_inbox_repository.go
internal/modules/notification/domain/repository/notification_campaign_repository.go
internal/modules/notification/domain/repository/email_delivery_log_repository.go
internal/modules/notification/domain/repository/mute_resolver.go
internal/modules/notification/domain/usecase/inputs.go
internal/modules/notification/domain/usecase/notification_template_usecase.go
internal/modules/notification/domain/usecase/notification_preference_usecase.go
internal/modules/notification/domain/usecase/notification_mute_usecase.go
internal/modules/notification/domain/usecase/notification_device_usecase.go
internal/modules/notification/domain/usecase/notification_ingest_usecase.go
internal/modules/notification/domain/usecase/notification_delivery_usecase.go
internal/modules/notification/domain/usecase/notification_inbox_usecase.go
internal/modules/notification/domain/usecase/notification_history_usecase.go
internal/modules/notification/domain/usecase/notification_campaign_usecase.go
internal/modules/notification/domain/usecase/email_delivery_usecase.go
internal/modules/notification/domain/error/errors.go
```

**Commit checkpoints:**
1. Enums + entity structs (all 10 entities, 9 enums)
2. Repository interfaces (all 9, each extends GenericRepository[T])
3. Use case interfaces + input DTOs (all 10 + inputs.go)
4. MuteResolver interface + domain error codes
5. Empty directory scaffolding for application, infrastructure, delivery layers

**Tests:** None — pure definitions, no runtime behavior

**Acceptance criteria:**
- `go build ./...` passes cleanly
- All entity structs embed `model.BaseModel` (except `NotificationTemplateTranslation`)
- All GORM tags follow codebase conventions (`type:varchar(N)`, `not null`, `default:`, `uniqueIndex:`, `index:`)
- FK references use `AccountID uuid.UUID` (not UserID)
- All repository interfaces embed `sharedrepo.GenericRepository[T]`
- All use case interfaces specify full method signatures matching `docs/notification/usecases.md`
- Input DTOs use pointer fields for optional updates (`*string`, `*bool`, `*time.Time`)
- Enum values use lowercase snake_case `string` constants
- Error codes use module-scoped format: `notification.errors.*`

**Risk level:** Low — pure definitions, no runtime behavior, no impact on existing code

**Notes:** This PR establishes the complete type system. Every subsequent PR references these types. Review should focus on naming consistency with existing codebase conventions (community, guide modules) and alignment with the design docs.

---

### PR-N2 — Data Layer Implementation

**Goal:** Implement all 9 repository interfaces, register entities with SchemaManager, add Resend configuration to `core.Config`.

**Scope:**
- 9 repository implementations in `infrastructure/repository/`
- Shared `helpers.go` with `getDB(ctx)` and `applyPaginationAndSorting()` utilities
- `EntityProvider` in `internal/modules/notification/entities.go` registering all 10 entities
- Config additions to `internal/core/config.go` for Resend (Enabled, APIKey, WebhookSecret, FromEmail, FromName)
- All repository constructors follow `New*Repository(db *core.Database, logger core.Logger) → domainInterface` pattern
- Proper use of `clause.OnConflict` for `UserNotificationPreferenceRepository.Upsert`
- Proper `FetchPending` query in `NotificationQueueRepository` (ordered by Priority DESC, ScheduledFor ASC)
- Proper `UpdateDeliveryEvent` in `EmailDeliveryLogRepository` mapping Resend event types to columns

**Out of scope:** Use case implementations, handlers, routes, workers, DI wiring

**Dependencies:** PR-N1 (domain interfaces must exist)

**Files/modules likely touched:**
```
internal/core/config.go                                    # Resend config struct addition
internal/modules/notification/entities.go                  # EntityProvider + SchemaManager registration
internal/modules/notification/infrastructure/repository/helpers.go
internal/modules/notification/infrastructure/repository/notification_template_repository.go
internal/modules/notification/infrastructure/repository/user_notification_preference_repository.go
internal/modules/notification/infrastructure/repository/muted_account_repository.go
internal/modules/notification/infrastructure/repository/user_device_repository.go
internal/modules/notification/infrastructure/repository/notification_queue_repository.go
internal/modules/notification/infrastructure/repository/notification_history_repository.go
internal/modules/notification/infrastructure/repository/user_notification_inbox_repository.go
internal/modules/notification/infrastructure/repository/notification_campaign_repository.go
internal/modules/notification/infrastructure/repository/email_delivery_log_repository.go
```

**Commit checkpoints:**
1. `core.Config` Resend struct addition
2. Repository helpers (`getDB`, `applyPaginationAndSorting`)
3. Infrastructure repository implementations (9 files — one commit per 2-3 repos or all together in logical groups)
4. `EntityProvider` registration (`entities.go`)

**Tests:**
- **Unit:** None for this PR (repository testing deferred or done via integration tests)
- **Integration (deferred):** `FetchPending` query correctness, `Upsert` OnConflict behavior, `IncrementRetry` atomic update. Can be tested with test DB + transaction rollback. These tests are **low priority** and can be added in a follow-up or during PR-N4 acceptance.

**Acceptance criteria:**
- `go build ./...` passes
- All 9 repository constructors return the domain interface type via `fx.Annotate` compatible signature
- Every custom query method uses `getDB(ctx)` for transaction-aware DB access
- `EntityProvider.Entities()` returns all 10 entity pointers
- `EntityProvider.ModuleName()` returns `"notification"`
- Resend config struct has proper `mapstructure` tags
- `IncrementDownloadCount` (or whatever similar atomic increment logic in queue) uses SQL `SET column = column + 1` to avoid race conditions
- `FetchPending` correctly filters `Status = 'pending' AND ScheduledFor <= now()` and orders by `Priority DESC, ScheduledFor ASC`

**Risk level:** Low — standard GORM repository implementations matching established patterns

**Notes:** Repository implementations follow the guide/community module implementations exactly. Each embeds `GenericRepository[T]` via `NewBaseRepository[T]`. The `FetchPending` method in `NotificationQueueRepository` is the most complex query and should be carefully reviewed.

---

### PR-N3 — Template Management (Admin)

**Goal:** Implement template CRUD with translations, the template renderer service, and create the seed migration for 15 system-managed templates.

**Scope:**
- `NotificationTemplateUsecase` implementation (`application/usecase/notification_template_usecase.go`)
  - `CreateTemplate` — validates type uniqueness, sets category from type mapping
  - `GetTemplate`, `GetTemplateByType` — single lookups
  - `ListTemplates` — optional category filter, pagination
  - `UpdateTemplate` — restricted updates for system-managed templates
  - `DeleteTemplate` — soft-delete, blocked for system-managed
  - `AddTranslation`, `UpdateTranslation`, `DeleteTranslation`, `GetTranslations`
- `TemplateRenderer` service (`application/service/template_renderer.go`)
  - `Render(content, variables)` — replaces `{{variable}}` placeholders
  - `RenderMultiChannel(defaultContent, variables, channels)` — renders per channel
  - `ValidateVariables(variablesSchema, variables)` — checks required variables present
- Seed migration file (Atlas migration) inserting 15 system-managed templates with default content for each NotificationType
- Admin handler + routes for template CRUD (`delivery/handler/notification_admin_handler.go`, `delivery/routes/notification_admin_routes.go`)
- Admin DTOs for request/response mapping

**Out of scope:** Ingest pipeline (next PR), preferences, mutes, delivery

**Dependencies:** PR-N2 (repositories must exist for DB operations)

**Files/modules likely touched:**
```
internal/modules/notification/application/usecase/notification_template_usecase.go
internal/modules/notification/application/service/template_renderer.go
internal/modules/notification/delivery/handler/notification_admin_handler.go
internal/modules/notification/delivery/dto/notification_admin_dto.go
internal/modules/notification/delivery/routes/notification_admin_routes.go
internal/modules/notification/domain/event/events.go           # Event name constants (if needed)
internal/core/migration/xxx_seed_notification_templates.go      # Seed migration file
```

**Commit checkpoints:**
1. `TemplateRenderer` service — pure function, no dependencies
2. `NotificationTemplateUsecase` implementation
3. Admin DTOs (request/response types for template CRUD)
4. Admin handler wiring (HTTP → usecase mapping)
5. Admin route registration
6. Seed migration for 15 system-managed templates

**Tests:**
- **Unit:**
  - `TemplateRenderer`: Given variables map → correct substitution. Given missing required variable → error. Given empty variables → unchanged template. Given multi-channel content → all channels rendered.
  - `NotificationTemplateUsecase`: Create → template created with correct fields. Update system-managed → restricted fields blocked. Delete system-managed → error. GetByType → correct template returned.

**Acceptance criteria:**
- Admin can create a template via `POST /api/v1/admin/notifications/templates`
- Admin can list templates via `GET /api/v1/admin/notifications/templates` with category filter
- Admin can update a template (non-system-managed) via `PATCH /api/v1/admin/notifications/templates/{id}`
- Attempting to delete a system-managed template returns 403
- Admin can add translations via `POST /api/v1/admin/notifications/templates/{id}/translations`
- `TemplateRenderer.Render` correctly substitutes `{{variableName}}` and returns error for missing required variables
- Seed migration creates 15 rows in `notification_templates` with correct type, category, content

**Risk level:** Low — standard CRUD use case, pure function renderer

**Notes:** The TemplateRenderer uses simple string replacement (`strings.ReplaceAll` or similar) for `{{variable}}` patterns — no need for a full template engine library. The notification template variable format should be consistent with `{{variableName}}` (matching existing `pkg/email` patterns).

---

### PR-N4 — Core Delivery Pipeline

**Goal:** Implement the two-stage ingest → deliver pipeline for InApp channel, including queue processing, history creation, inbox creation, and the delivery worker.

**Scope:**
- `NotificationIngestUsecase` implementation
  - `ProcessEvent` — resolves template, checks preferences, checks mutes, renders content, enqueues
  - `SendNotification` — direct send bypassing events
  - `SendMultiChannel` — convenience for multiple channel sends
- `NotificationDeliveryUsecase` implementation
  - `ProcessQueue` — fetches pending items, dispatches per channel, handles results
  - `DeliverItem` — marks processing, calls channel handler, handles result
  - `HandleDeliveryResult` — success: creates history + inbox; failure: retry with backoff or mark failed
- `NotificationInboxUsecase` implementation
  - `ListInbox` — paginated with category filter, excludes archived/expired
  - `GetUnreadCount` — badge count
  - `MarkAsRead`, `MarkAllAsRead`, `MarkCategoryAsRead` — read status
  - `ArchiveNotification`, `DeleteNotification` — lifecycle
  - `ExpireOld` — cleanup expired entries
- `NotificationHistoryUsecase` implementation
  - `ListByAccount` — paginated history for user
  - `GetByID`, `MarkRead`, `MarkClicked`, `UpdateDeliveryStatus` — webhook callbacks
- Delivery worker goroutine (`fx.Lifecycle.OnStart`)
  - Poll interval: 5 seconds
  - Batch size: 50
  - Calls `ProcessQueue(ctx, 50)`
- InApp channel delivery (direct: create history + inbox, no external provider)
- Inbox expiry cleanup worker (runs every 1 hour)
- User-facing routes + handler for inbox and history
- Admin monitoring routes: queue status, retry-failed

**Out of scope:** Email delivery (PR-N6), Push delivery (Phase 2), SMS (Phase 2), Campaign integration (PR-N7)

**Dependencies:** PR-N2 (repositories), PR-N3 (templates + renderer for ingest)

**Files/modules likely touched:**
```
internal/modules/notification/application/usecase/notification_ingest_usecase.go
internal/modules/notification/application/usecase/notification_delivery_usecase.go
internal/modules/notification/application/usecase/notification_inbox_usecase.go
internal/modules/notification/application/usecase/notification_history_usecase.go
internal/modules/notification/delivery/handler/notification_handler.go
internal/modules/notification/delivery/dto/notification_dto.go
internal/modules/notification/delivery/routes/notification_routes.go
```

**Commit checkpoints:**
1. `NotificationIngestUsecase` — ProcessEvent, SendNotification, SendMultiChannel
2. `NotificationDeliveryUsecase` — ProcessQueue, DeliverItem, HandleDeliveryResult, RetryFailed, CancelPendingForAccount
3. `NotificationInboxUsecase` — ListInbox, GetUnreadCount, MarkAsRead, MarkAllAsRead, MarkCategoryAsRead, Archive, Delete, ExpireOld
4. `NotificationHistoryUsecase` — ListByAccount, GetByID, MarkRead, MarkClicked, UpdateDeliveryStatus
5. User-facing DTOs for inbox + history
6. Handler + routes for inbox (list, read, archive, delete, unread-count, read-all) and history
7. Admin monitoring handler + routes (queue status, retry-failed)
8. Delivery worker goroutine + inbox expiry cleanup worker

**Tests:**
- **Unit (High priority):**
  - **Ingest:** Given valid event → template loaded, preferences checked, content rendered, queue entry created. Given muted author → no queue entry. Given IAM toggle off → no queue entry. Given all channels disabled → no queue entry.
  - **Delivery:** Given pending queue items → delivered successfully → history + inbox created. Given delivery failure → retryCount incremented, ScheduledFor updated. Given max retries exceeded → Status=Failed.
  - **Inbox:** Given inbox → ListInbox returns paginated, filtered correctly. Given MarkAsRead → IsRead=true + NotificationHistory.ReadAt updated. Given expired entry → ExpireOld deletes it.

**Acceptance criteria:**
- User can list inbox via `GET /api/v1/notifications/inbox` with pagination and category filter
- User can see unread count via `GET /api/v1/notifications/inbox/unread-count`
- User can mark single notification as read via `PATCH /api/v1/notifications/inbox/{id}/read`
- User can mark all as read via `POST /api/v1/notifications/inbox/read-all`
- User can archive via `PATCH /api/v1/notifications/inbox/{id}/archive`
- User can delete via `DELETE /api/v1/notifications/inbox/{id}`
- Admin can see queue status counts
- Admin can retry failed notifications
- Delivery worker polls queue every 5 seconds and processes items
- Inbox expiry worker removes expired entries
- Muted notifications create history but NO inbox entry
- Queue retry: 1st retry → 1min, 2nd → 2min, 3rd → 4min, then Failed

**Risk level:** High — this is the core of the notification module. The ingest/delivery pipeline is complex with many edge cases (preference resolution, mute checks, retry logic, concurrent worker).

**Notes:** The ingest use case needs to read IAM's `NotificationPreference` global toggles. This is done by injecting IAM's `NotificationPreferenceRepository` or defining a thin interface in the notification domain. The `MuteResolver` interface is checked during ingest — source modules register their implementations. For Phase 1, if no `MuteResolver` is registered, mutes are skipped (not an error).

---

### PR-N5 — User Settings (Preferences + Mutes + Devices)

**Goal:** Implement user-facing preference management, account muting, and device registration endpoints.

**Scope:**
- `NotificationPreferenceUsecase` implementation
  - `SetPreference` — upsert with OnConflict on (AccountID, NotificationType, Channel)
  - `GetPreferences` — list all overrides for an account
  - `GetEffectivePreference` — three-layer resolution: IAM global → per-type override → template default
  - `IsQuietHours` — check if current time falls within quiet window
  - `DeletePreference` — remove override, revert to default
- `NotificationMuteUsecase` implementation
  - `MuteAccount` — create mute, optional MuteUntil
  - `UnmuteAccount` — delete mute
  - `IsMuted` — check with MuteUntil expiry
  - `ListMutedAccounts` — paginated
- `NotificationDeviceUsecase` implementation
  - `RegisterDevice` — create or update (dedup by DeviceToken)
  - `UpdateDevice` — update push token, metadata
  - `DeactivateDevice` — set IsActive = false
  - `ListDevices` — active devices only
  - `DeactivateAllDevices` — on account suspension
- User-facing routes + handlers for preferences, mutes, devices

**Out of scope:** Push notification delivery (Phase 2), batch preference update (Phase 2), IAM global toggle editing (owned by IAM)

**Dependencies:** PR-N2 (repositories)

**Files/modules likely touched:**
```
internal/modules/notification/application/usecase/notification_preference_usecase.go
internal/modules/notification/application/usecase/notification_mute_usecase.go
internal/modules/notification/application/usecase/notification_device_usecase.go
internal/modules/notification/delivery/handler/notification_handler.go   # Extend existing
internal/modules/notification/delivery/dto/notification_dto.go            # Extend existing
internal/modules/notification/delivery/routes/notification_routes.go      # Extend existing
```

**Commit checkpoints:**
1. `NotificationPreferenceUsecase` implementation
2. `NotificationMuteUsecase` implementation
3. `NotificationDeviceUsecase` implementation
4. Preference routes + handler (list, set, delete)
5. Mute routes + handler (list, create, delete)
6. Device routes + handler (list, register, update, deactivate)
7. DTOs for preference, mute, and device requests/responses

**Tests:**
- **Unit (High priority):**
  - **Preference resolution:** IAM toggle OFF → GetEffectivePreference returns false. IAM toggle ON + per-type override OFF → false. IAM toggle ON + no override → template default. IAM toggle ON + override ENABLED → true. Quiet hours active → IsQuietHours returns true.
  - **Mutes:** MuteAccount → IsMuted returns true. UnmuteAccount → IsMuted returns false. Expired MuteUntil → IsMuted returns false.
  - **Devices:** RegisterDevice → creates. Same DeviceToken again → updates (no duplicate). DeactivateDevice → IsActive=false.

**Acceptance criteria:**
- User can list preferences via `GET /api/v1/notifications/preferences`
- User can set preference via `PUT /api/v1/notifications/preferences` with notificationType + channel + isEnabled + quietHours
- User can delete preference via `DELETE /api/v1/notifications/preferences/{type}/{channel}`
- User can mute account via `POST /api/v1/notifications/mutes` with mutedAccountID + optional MuteUntil
- User can list mutes via `GET /api/v1/notifications/mutes`
- User can unmute via `DELETE /api/v1/notifications/mutes/{accountId}`
- User can register device via `POST /api/v1/notifications/devices`
- User can list devices via `GET /api/v1/notifications/devices`
- User can deactivate device via `DELETE /api/v1/notifications/devices/{id}`
- Re-registering same device token updates push token (no duplicate)

**Risk level:** Medium — preference resolution logic is nuanced (three layers + quiet hours). Device upsert dedup logic needs care.

**Notes:** The `GetEffectivePreference` method reads IAM's `NotificationPreference` entity. This requires injecting IAM's repository or a thin interface. A thin `IAMNotificationPreferenceReader` interface with a single `GetByAccountID(ctx, accountID) (*IAMNotificationPreference, error)` method is cleaner than importing the full IAM module.

---

### PR-N6 — Email Delivery + Resend Integration

**Goal:** Implement Resend as the email delivery provider, wire email into the delivery pipeline, and handle Resend webhook events for delivery tracking.

**Scope:**
- `ResendProvider` (`infrastructure/email/resend_provider.go`)
  - `Send(ctx, to, subject, body, metadata)` → sends via Resend REST API
  - `VerifyWebhookSignature(ctx, payload, signature)` → validates Resend webhook
  - Configurable via Resend config section in `core.Config`
- `EmailDeliveryLogRepository` implementation (already wired in PR-N2, now used)
- `EmailDeliveryUsecase` implementation
  - `HandleWebhookEvent` — maps Resend events to field updates
  - `GetDeliveryLog`, `GetDeliveryLogByProviderID` — log lookups
- Email channel integration in `NotificationDeliveryUsecase.DeliverItem`:
  - If channel is Email → call `ResendProvider.Send` → create `EmailDeliveryLog` → on success create history
- Webhook handler + route (`POST /api/v1/webhooks/resend`)
  - Public endpoint (no auth middleware)
  - Signature verification via `ResendProvider.VerifyWebhookSignature`
  - Routes event to `EmailDeliveryUsecase.HandleWebhookEvent`
- Event type mapping:
  - `email.delivered` → `DeliveredAt`, `DeliveryStatus = Delivered`
  - `email.opened` → `OpenedAt`, `NotificationHistory.ReadAt`
  - `email.clicked` → `ClickedAt`, `NotificationHistory.ClickedAt`
  - `email.bounced` → `BounceReason`, `DeliveryStatus = Bounced`
  - `email.complained` → `Complaint = true`

**Out of scope:** Replacing existing `pkg/email` SMTP (coexists), push delivery (Phase 2), SMS (Phase 2), repeat engagement tracking (Phase 2)

**Dependencies:** PR-N2 (repositories, config), PR-N4 (delivery pipeline for integration)

**Files/modules likely touched:**
```
internal/modules/notification/infrastructure/email/resend_provider.go
internal/modules/notification/application/usecase/email_delivery_usecase.go
internal/modules/notification/delivery/handler/webhook_handler.go
internal/modules/notification/delivery/dto/notification_dto.go      # Webhook event DTO
internal/modules/notification/delivery/routes/notification_routes.go # Webhook route
internal/modules/notification/application/usecase/notification_delivery_usecase.go  # Email channel case
```

**Commit checkpoints:**
1. `ResendProvider` — send + webhook verification
2. `EmailDeliveryUsecase` — HandleWebhookEvent, GetDeliveryLog
3. Email channel branch in delivery pipeline
4. Webhook handler + route (signature verification, event routing)
5. Webhook request/response DTOs

**Tests:**
- **Unit (Medium priority):**
  - `EmailDeliveryUsecase.HandleWebhookEvent`: Given `delivered` event → `DeliveredAt` and `DeliveryStatus` updated. Given `bounced` event → `BounceReason` set. Given unknown event type → logged and skipped.
  - `ResendProvider.VerifyWebhookSignature`: Given valid signature → passes. Given invalid signature → fails.
- **Integration (Medium priority):**
  - Webhook handler: Given valid Resend event → correct fields updated. Given invalid signature → 401.

**Acceptance criteria:**
- Email notifications are sent via Resend when delivery pipeline processes an Email-channel queue item
- `EmailDeliveryLog` row created with `SentAt` on dispatch
- Resend webhook endpoint receives delivery events and updates EmailDeliveryLog fields
- Webhook signature verification rejects invalid requests with 401
- Bounced emails update `DeliveryStatus = Bounced` and `BounceReason`
- Opened emails update `OpenedAt` and `NotificationHistory.ReadAt`

**Risk level:** Medium — external API dependency (Resend). Webhook signature verification is security-critical. Email delivery failures must not crash the delivery pipeline (graceful error handling).

**Notes:** The Resend REST API sends emails via `POST https://api.resend.com/emails` with JSON body. The provider should return the provider message ID for webhook correlation. The `ResendProvider` implements an interface that can be mocked in delivery usecase tests — define a local `EmailProvider` interface in the infrastructure layer.

---

### PR-N7 — Campaign System

**Goal:** Implement admin campaign management (Broadcast + Segmented) and the campaign scheduler worker.

**Scope:**
- `NotificationCampaignUsecase` implementation
  - `CreateCampaign` — Draft status, optional TargetSegment for Segmented
  - `GetCampaign`, `ListCampaigns` — with status filter
  - `UpdateCampaign` — Draft-only updates
  - `ScheduleCampaign` — transition to Scheduled, resolve segment to static AccountID list, validate template
  - `CancelCampaign` — transition to Cancelled, cancel pending queue items
  - `ProcessScheduledCampaigns` — worker entry point: fetch due campaigns, create queue entries for all recipients
- `CampaignProcessor` service (`application/service/campaign_processor.go`)
  - `ResolveSegment(filters)` — converts TargetSegment filters to []AccountID
  - `ProcessCampaign(campaign)` — resolve segment → render template per recipient → create queue entries
- Campaign scheduler worker (`fx.Lifecycle.OnStart`)
  - Poll interval: 30 seconds
  - Calls `ProcessScheduledCampaigns(ctx)`
- Admin routes + handler for campaign CRUD

**Out of scope:** Triggered campaigns (Phase 2), A/B testing (Phase 2), campaign analytics dashboard (Phase 2)

**Dependencies:** PR-N2 (repositories), PR-N3 (templates + renderer)

**Files/modules likely touched:**
```
internal/modules/notification/application/usecase/notification_campaign_usecase.go
internal/modules/notification/application/service/campaign_processor.go
internal/modules/notification/delivery/handler/notification_admin_handler.go    # Extend
internal/modules/notification/delivery/dto/notification_admin_dto.go             # Extend
internal/modules/notification/delivery/routes/notification_admin_routes.go       # Extend
```

**Commit checkpoints:**
1. `CampaignProcessor` — segment resolution logic
2. `NotificationCampaignUsecase` — CRUD + schedule + cancel + process
3. Campaign admin DTOs
4. Campaign admin handler + routes
5. Campaign scheduler worker goroutine

**Tests:**
- **Unit (Medium priority):**
  - `CampaignProcessor.ResolveSegment`: Given Broadcast filters → all accounts. Given Segmented filters → filtered account list matching criteria.
  - `NotificationCampaignUsecase.ScheduleCampaign`: Draft → Scheduled transition works. Invalid template → error. Segment resolved and stored.
  - `NotificationCampaignUsecase.CancelCampaign`: Pending queue items cancelled. Status transitions to Cancelled.
  - `NotificationCampaignUsecase.ProcessScheduledCampaigns`: Due campaigns processed. Queue entries created for all recipients.

**Acceptance criteria:**
- Admin can create Broadcast campaign via `POST /api/v1/admin/notifications/campaigns`
- Admin can create Segmented campaign with filters
- Admin can schedule campaign via `POST /api/v1/admin/notifications/campaigns/{id}/schedule` → filters resolved to static AccountID list
- Admin can cancel scheduled campaign via `POST /api/v1/admin/notifications/campaigns/{id}/cancel`
- Scheduled campaigns due for sending are processed by scheduler worker
- Campaign status transitions: Draft → Scheduled → Sending → Completed (or Cancelled)
- Campaign content override (CustomSubject, CustomContent) respected during rendering

**Risk level:** Medium — segment resolution logic could be complex with multiple filter types. The scheduler worker needs to handle large campaigns (thousands of recipients) without overwhelming the system.

**Notes:** Phase 1 segment resolution is basic — filters on account type, role, registration date range. The `resolvedAccountIDs` snapshot is stored in the `TargetSegment` JSONB on schedule. For very large segments, consider batch processing (1000 recipients per batch) rather than creating all queue entries in one transaction.

---

### PR-N8 — Module Integration

**Goal:** Wire everything together — module.go, route registration, event subscriptions, worker lifecycle, modules.go. This is the capstone PR that activates the entire notification module.

**Scope:**
- `internal/modules/notification/module.go`:
  - All `fx.Provide` for repos, use cases, services, handlers
  - All `fx.Annotate` bindings (interface → implementation)
  - `fx.Invoke` for SchemaManager registration
  - `fx.Invoke` for route registration (user + admin + webhook routes)
  - `fx.Invoke` for event handler registration
  - `fx.Lifecycle.OnStart` hooks for: delivery worker, campaign scheduler, inbox expiry worker
- Route dependencies struct:
  - AuthMiddleware, AccountStatusMiddleware, PermissionMiddleware (from IAM)
  - All handlers
- Event subscription registration in `internal/handlers/handlers.go`:
  - Subscribe to 14 events from source modules
  - Wire to `NotificationIngestUsecase.ProcessEvent`
- `internal/modules/modules.go`:
  - Add `notification.Module` to `Modules`
- `internal/modules/notification/domain/event/events.go`:
  - Event name constants for subscribed events + `notification.failed` published event
- `internal/modules/notification/application/service/notification_service.go`:
  - Facade composing ingest + delivery + inbox (optional, for convenience)

**Out of scope:** Any business logic changes — pure integration work

**Dependencies:** PR-N1 through PR-N7 (all implementations must exist)

**Files/modules likely touched:**
```
internal/modules/notification/module.go
internal/modules/notification/domain/event/events.go
internal/modules/notification/application/service/notification_service.go
internal/handlers/handlers.go
internal/modules/modules.go
```

**Commit checkpoints:**
1. Event name constants (`domain/event/events.go`)
2. Module module.go — all fx.Provide fx.Annotate bindings
3. Route registration fx.Invoke — user routes, admin routes, webhook routes
4. Event handler registration — subscribe to each source module event
5. Worker lifecycle hooks — delivery worker, campaign scheduler, inbox expiry
6. modules.go registration
7. NotificationService facade (optional convenience layer)

**Tests:**
- **Integration:** Verify module starts correctly with fx (no missing dependencies). Verify event handler registration doesn't panic. Verify routes are accessible after registration.

**Acceptance criteria:**
- `go build ./...` passes with notification module fully wired
- Application starts with all notification workers running (delivery, campaign, expiry)
- Event handlers are subscribed to all 14 events
- All routes (user-facing + admin + webhook) are registered and accessible
- `modules.go` includes `notification.Module`
- Graceful shutdown: workers stop on context cancellation
- Missing `MuteResolver` implementations don't crash the ingest pipeline (graceful skip)

**Risk level:** Medium — DI wiring errors could cause startup panics. Missing event handlers could silently drop events. Each dependency must be correctly annotated with `fx.As`.

**Notes:** This PR should be tested carefully in a local environment. The most common failure mode is missing DI bindings causing startup panics. Test by running the application and verifying: (1) no fx panics on startup, (2) routes respond, (3) delivery worker starts polling, (4) events can be published and consumed.

---

## 4) Recommended Order

```
PR-N1 (Domain Foundation)
   ↓
PR-N2 (Data Layer Implementation)
   ↓                      ↓
PR-N3 (Template Mgmt)  PR-N5 (User Settings)     ← can start in parallel after PR-N2
   ↓                      ↓
PR-N4 (Core Pipeline)     ↓                      ← depends on PR-N3
   ↓                      ↓
PR-N6 (Email Delivery)    ↓                      ← depends on PR-N4
   ↓
PR-N7 (Campaign System)                          ← depends on PR-N4
   ↓
PR-N8 (Module Integration)                       ← depends on all above
```

**Merge order:** PR-N1 → PR-N2 → PR-N3 → PR-N5 → PR-N4 → PR-N6 → PR-N7 → PR-N8

**Parallel work:** After PR-N2 merges, PR-N3 (templates) and PR-N5 (settings) can be developed in parallel by separate developers.

---

## 5) Test Strategy Across the Full Epic

| Test Area | PR Introduced | Test Type | Notes |
|-----------|--------------|-----------|-------|
| TemplateRenderer | PR-N3 | Unit (pure function) | No mocks needed |
| Template CRUD | PR-N3 | Unit (mocked repos) | Standard CRUD tests |
| Ingest pipeline | PR-N4 | Unit (mocked repos) | High priority — complex logic |
| Delivery pipeline | PR-N4 | Unit (mocked repos + email) | High priority — retry, failure, cancel |
| Inbox operations | PR-N4 | Unit (mocked repos) | Pagination, filtering, read/archive |
| History operations | PR-N4 | Unit (mocked repos) | Audit trail queries |
| Preference resolution | PR-N5 | Unit (mocked repos + IAM reader) | Three-layer logic + quiet hours |
| Mute operations | PR-N5 | Unit (mocked repos) | Mute/unmute/expiry |
| Device operations | PR-N5 | Unit (mocked repos) | Register/upsert/deactivate |
| Email webhook mapping | PR-N6 | Unit (mocked repos) | Event → field mapping |
| Webhook signature | PR-N6 | Unit | Security critical |
| Campaign ops | PR-N7 | Unit (mocked repos) | Status transitions, segment resolution |
| Segment resolution | PR-N7 | Unit (mocked accounts) | Filter logic |
| Module startup | PR-N8 | Integration | DI wiring, route registration, event handlers |

**Contract tests** (cross-module): Verify that `MuteResolver` implementations from community module conform to the interface. Done when community module implements `MuteResolver`.

**Integration tests** (test DB): Low priority — test only custom queries like `FetchPending`, `IncrementRetry`, `Upsert`. Can be done as a follow-up or during PR-N8 validation.

---

## 6) Open Questions Remaining

- **IAM PreferenceReader interface**: Should the thin `IAMNotificationPreferenceReader` interface live in the notification module (domain/repository/) or in a shared location? **Recommendation:** Define it in the notification module's `domain/repository/` as a separate file (e.g., `iam_preference_reader.go`) since it's a consumer-defined interface.
- **MuteResolver implementation in community**: Does the community module implement `MuteResolver` as part of this epic or as a separate PR? **Assumption:** Separate community module PR, not part of notification module PRs.
- **Event publishing from source modules**: Are the 14 events added to source modules as part of this epic? **Assumption:** Source module event publishing is tracked separately. Notification module defines the contracts and subscribes; source modules publish when ready.
