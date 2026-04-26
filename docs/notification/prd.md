# PRD: Notification Module

## Problem Statement

The platform currently has no unified notification system. Each module (IAM, community, guide, AI) would need to build its own ad-hoc notification delivery, leading to duplicated email logic, no centralized preference management, no audit trail, and no way for users to control what they receive or for admins to manage templates and campaigns. Without a notification module, users receive no feedback when someone replies to their thread, when a guide deadline approaches, or when their AI quota is exhausted — resulting in poor engagement and missed critical alerts.

## Solution

Build a dedicated notification module following the codebase's Clean Architecture / DDD patterns that provides:

- **Centralized delivery pipeline**: A two-stage ingest → deliver pipeline that all source modules feed via RabbitMQ events, with queue-based retry and exponential backoff
- **Template engine**: Per-type templates with multi-channel default content (email, push, in-app, SMS) and per-language translations, rendered at enqueue time
- **Layered preference system**: IAM's existing global channel toggles remain the first gate; the notification module adds granular per-type, per-channel overrides with quiet hours support
- **Cross-module mute resolution**: Account-to-account muting owned by the notification module; module-specific mutes (e.g., community thread mutes) checked via a `MuteResolver` interface implemented by source modules
- **Dual storage model**: Immutable `NotificationHistory` (audit trail) + mutable `UserNotificationInbox` (user-facing feed with read/archive/expire)
- **Email delivery tracking**: Resend integration with webhook-driven delivery status updates (delivered, opened, clicked, bounced, complained)
- **Admin campaign system**: Broadcast and segmented campaigns with segment resolution at schedule time
- **Device registration**: Two-token model (DeviceToken for dedup + PushToken for delivery) supporting FCM, APNs, and web push

## User Stories

### In-app Inbox

1. As a user, I want to see a paginated list of my unread notifications, so that I can catch up on what I missed
2. As a user, I want to filter my inbox by category (Community, Guide, Security, etc.), so that I can focus on specific areas
3. As a user, I want to see an unread count badge on the notification icon, so that I know at a glance if there are new notifications
4. As a user, I want to mark a single notification as read, so that I can track which ones I've seen
5. As a user, I want to mark all notifications as read with one action, so that I can quickly clear my badge
6. As a user, I want to mark all notifications in a specific category as read, so that I can clear community notifications without affecting security alerts
7. As a user, I want to archive a notification, so that I can remove it from my main feed without deleting it
8. As a user, I want to delete a notification from my inbox, so that I can remove unwanted items (audit record preserved)
9. As a user, I want notifications to automatically expire based on their type's TTL, so that stale notifications don't clutter my inbox
10. As a user, I want to click a notification and be taken to the relevant content (thread, guide, etc.), so that I can act on it immediately

### Preferences

11. As a user, I want to see all my notification preferences, so that I understand what I'm opted into
12. As a user, I want to enable or disable a specific notification type on a specific channel, so that I can (for example) get community replies in-app but not by email
13. As a user, I want to remove a preference override, so that I revert to the default for that notification type
14. As a user, I want to set quiet hours for a notification type, so that I'm not disturbed during specific times
15. As a user, I want my IAM global channel toggle to act as a master switch, so that disabling email at the IAM level turns off all email notifications regardless of per-type settings

### Muting

16. As a user, I want to mute another account, so that I stop receiving notifications triggered by that person
17. As a user, I want to set a temporary mute with an expiry date, so that I can mute someone for a limited time
18. As a user, I want to unmute an account, so that I resume receiving their notifications
19. As a user, I want to see a list of all accounts I've muted, so that I can manage my mutes
20. As a user, I want muting a specific thread in the community to suppress notifications from that thread, so that I can control noise at the module level

### Device Management

21. As a user, I want to register my mobile device for push notifications, so that I receive alerts on my phone
22. As a user, I want to register my web browser for push notifications, so that I receive alerts on my desktop
23. As a user, I want my push token to be updated automatically when it refreshes, so that push delivery continues working
24. As a user, I want to see all my registered devices, so that I can manage them
25. As a user, I want to deactivate a specific device, so that I stop getting push notifications on a device I no longer use
26. As a user, I want duplicate device registrations to be handled gracefully (update rather than error), so that re-installing the app doesn't create multiple entries

### Notification Content

27. As a user, I want notification content to be rendered in my preferred language, so that I can understand notifications (with English as fallback)
28. As a user, I want notification templates to include relevant context (who replied, which thread, what deadline), so that I can decide whether to act
29. As a user, I want actionable notifications that link to the relevant page, so that I can navigate directly to the source

### Delivery Reliability

30. As a user, I want failed notifications to be retried automatically, so that transient failures don't cause permanent loss
31. As a user, I want notifications to respect my preference and mute settings, so that I only receive what I've opted into
32. As a user, I want muted notifications to still appear in the audit history, so that nothing is silently lost
33. As a user, I want notifications that fail after all retries to be flagged, so that I can be informed of delivery issues

### Community Notifications

34. As a community participant, I want to be notified when someone replies to my thread, so that I can respond
35. As a community participant, I want to be notified when my thread receives a solution, so that I can review it
36. As a community participant, I want to be notified when I'm mentioned in a post, so that I can engage

### Guide Notifications

37. As a guide follower, I want to be notified when a guide step is completed, so that I can track progress
38. As a guide follower, I want to be notified when a compliance deadline is approaching, so that I can take action in time
39. As a guide follower, I want to be notified when a guide I follow is updated, so that I stay informed

### AI Notifications

40. As an AI user, I want to be notified when I'm approaching my AI usage quota, so that I can adjust my usage
41. As an AI user, I want to be notified when my async AI response is ready, so that I can view the result

### Security Notifications

42. As an account holder, I want to receive security alerts about my account, so that I can respond to suspicious activity
43. As an account holder, I want to receive account verification notifications, so that I can complete verification
44. As an account holder, I want to receive password reset notifications, so that I can reset my password securely

### Payment Notifications

45. As a payer, I want to receive payment confirmation notifications, so that I have a record of my transactions

### System Notifications

46. As a user, I want to receive platform-wide announcements, so that I'm aware of important updates
47. As a user, I want to be notified of policy updates, so that I can review changes
48. As a new user, I want to receive a welcome message, so that I feel oriented on the platform

### Admin — Templates

49. As an admin, I want to create notification templates for each notification type, so that the system knows what content to send
50. As an admin, I want to define multi-channel content per template (email subject/body, push title/body, in-app title/body/action URL), so that each channel gets appropriately formatted content
51. As an admin, I want to add translations to templates for multiple languages, so that users receive notifications in their preferred language
52. As an admin, I want to update template content, so that I can fix typos or adjust messaging
53. As an admin, I want to mark templates as system-managed, so that they cannot be accidentally deleted
54. As an admin, I want to define template variables with a schema, so that the system can validate that all required variables are provided before rendering
55. As an admin, I want to set a default TTL per template, so that time-sensitive notifications (like password resets) auto-expire from the inbox
56. As an admin, I want to set notification priority per template, so that urgent notifications are delivered first

### Admin — Campaigns

57. As an admin, I want to create a broadcast campaign that sends to all accounts, so that I can announce platform-wide changes
58. As an admin, I want to create a segmented campaign with filters (account type, role, registration date, guide completion), so that I can target specific audiences
59. As an admin, I want to schedule a campaign for a future date, so that I can plan ahead
60. As an admin, I want segment filters to be resolved to a static recipient list at schedule time, so that the recipient list is consistent even if account data changes between scheduling and sending
61. As an admin, I want to cancel a scheduled campaign, so that I can stop it if plans change
62. As an admin, I want to see campaign status (Draft, Scheduled, Sending, Completed, Cancelled), so that I can track progress
63. As an admin, I want to override template subject and content for a campaign, so that I can customize the message without modifying the base template
64. As an admin, I want to see which campaigns I created, so that I can manage my own work

### Admin — Monitoring

65. As an admin, I want to see notification queue status counts (pending, processing, delivered, failed, cancelled), so that I can monitor system health
66. As an admin, I want to manually retry failed notifications, so that I can recover from system issues
67. As an admin, I want to view the full notification delivery history, so that I can audit what was sent to whom and when
68. As an admin, I want to see email delivery logs with open/click/bounce/complaint tracking, so that I can measure engagement and deliverability

### Admin — Webhooks

69. As a system, I want to receive Resend webhook events (delivered, opened, clicked, bounced, complained), so that delivery status is tracked in real-time
70. As a system, I want webhook signatures verified, so that only authentic Resend events update our records

### System — Event-Driven Ingestion

71. As a source module (community), I want to publish a RabbitMQ event when a thread receives a reply, so that the notification module can send a CommunityReply notification
72. As a source module (guide), I want to publish a RabbitMQ event when a deadline is approaching, so that the notification module can send a GuideDeadline notification
73. As a source module (IAM), I want to publish a RabbitMQ event when a user registers, so that the notification module can send a WelcomeMessage
74. As a source module (AI), I want to publish a RabbitMQ event when an async response is ready, so that the notification module can send an AIResponseReady notification
75. As the notification system, I want to publish a `notification.failed` event when a notification exhausts all retries, so that monitoring systems can alert operators

## Implementation Decisions

### Module Architecture

- **Decentralized module**: The notification module is a standalone module under `internal/modules/notification/`, following the same Clean Architecture layers (domain → application → infrastructure → delivery) as IAM, community, and guide modules
- **Entity registration**: All 10 entities registered via `EntityProvider` interface (`Entities() []any`, `ModuleName() string`) for SchemaManager/Atlas migration generation
- **DI wiring**: Uber FX module with `fx.Annotate(infraImpl, fx.As(new(domainInterface)))` pattern for all 9 repositories, 10 use cases, and 3 services
- **Module registration**: Added to `internal/modules/modules.go` as `notification.Module`

### Entity Design

- **10 entities**: NotificationTemplate, NotificationTemplateTranslation, UserNotificationPreference, MutedAccount, UserDevice, NotificationQueue, NotificationHistory, UserNotificationInbox, NotificationCampaign, EmailDeliveryLog
- **9 enums**: NotificationCategory (7 values), NotificationType (15 values), NotificationPriority (4 values), Channel (4 values), NotificationStatus (5 values), DeliveryStatus (4 values), CampaignType (2 values), CampaignStatus (5 values), DeviceType (3 values)
- **BaseModel**: All entities except `NotificationTemplateTranslation` embed `model.BaseModel` (ID, CreatedAt, UpdatedAt, DeletedAt). Translation uses its own simple PK + timestamps, consistent with `GuideTranslation` pattern
- **AccountID binding**: All FK references to accounts use `AccountID` (not `UserID`), consistent with delivery-channel binding pattern
- **Sparse override model**: `UserNotificationPreference` only stores explicit deviations from template defaults. No row = use default. Unique index on `(AccountID, NotificationType, Channel)` with `clause.OnConflict` upsert

### Template Design

- **One template per NotificationType**: `NotificationTemplate` stores default English multi-channel content in a single `DefaultContent` JSONB column (keys: email, push, sms, inapp)
- **No separate channel field**: Channel presence in `DefaultContent` determines the default enabled channels
- **DefaultSubject absorbed**: Email subject is a field within the email channel content, not a separate column
- **VariablesSchema**: Defines required and optional template variables for validation
- **DefaultTTL in seconds**: Supports sub-hour expiry (e.g., password reset = 900 seconds / 15 minutes)
- **Translation mirrors parent structure**: `NotificationTemplateTranslation.Content` uses the same JSONB structure as `DefaultContent`, with a denormalized `Subject` column for quick lookups

### Delivery Pipeline

- **Two-stage pipeline**: `NotificationIngestUsecase` (event → queue) + `NotificationDeliveryUsecase` (queue → deliver)
- **Template rendering at enqueue time**: `Payload` in `NotificationQueue` contains fully rendered content, not template + variables. This ensures content consistency even if the template is later modified
- **Queue retry**: maxRetries=3, exponential backoff (1min → 2min → 4min). `ScheduledFor` updated on each retry. After max retries → Status=Failed
- **Cancelled triggers**: User mute after enqueue, account deactivation, admin campaign cancel all set Status=Cancelled
- **Decoupled history**: `NotificationQueue` has no FK to `NotificationHistory`. History is created at delivery time, not enqueue time

### Dual Storage Model

- **NotificationHistory**: Immutable audit trail. Records what was sent, to whom, via which channel, and delivery outcomes. Never modified by user actions
- **UserNotificationInbox**: Mutable user-facing feed. Created only for notifications that pass preference/mute checks. Supports read, archive, and expiry
- **Muted notifications**: Create history (audit) but no inbox entry. User never sees them, but the record exists
- **Inbox category**: Denormalized stored field set at creation time, never changes. Used for category-filtered queries
- **Inbox ExpiresAt**: Resolved from event override → template DefaultTTL → null (no expiry). Expired entries cleaned up by periodic job

### Preference Resolution

- **Three-layer resolution**: (1) IAM global toggle (master switch per channel), (2) notification module per-type override, (3) template default (channel presence in DefaultContent)
- **Quiet hours**: Stored on `UserNotificationPreference`. If current time falls within `[QuietHoursStart, QuietHoursEnd]`, delivery is deferred by adjusting `ScheduledFor`

### Mute Architecture

- **Decentralized mutes**: The notification module owns `MutedAccount` (account-to-account muting). Community module owns thread/category mutes. The notification module accesses module-specific mutes via the `MuteResolver` interface
- **MuteResolver interface**: `IsMuted(ctx, accountID, itemType, itemID) (bool, error)` — implemented by community module (thread mutes, category mutes), extensible to guide and other modules
- **Mute expiry**: `MuteUntil` nullable — null = permanent, set value = temporary mute with auto-expiry

### Campaign System

- **Two campaign types**: Broadcast (all accounts) and Segmented (filtered audience). Triggered campaigns deferred to Phase 2
- **Hybrid segment resolution**: Filters defined at creation time, resolved to static `resolvedAccountIDs` list when status transitions to Scheduled. This snapshot ensures consistent recipients
- **Campaign permission**: `CreatedBy` references AccountID, permission enforced at application layer via `PermissionMiddleware`
- **Campaign content override**: `CustomSubject` and `CustomContent` allow per-campaign customization without modifying the base template

### Email Provider

- **Resend**: Chosen as the email provider (free tier: 3,000 emails/month). Coexists with existing SMTP `pkg/email` initially
- **ResendProvider**: Custom client in `infrastructure/email/resend_provider.go` for sending via REST API and verifying webhook signatures
- **Webhook-driven tracking**: `EmailDeliveryLog` updated by Resend webhook events. One row per email with first-only open/click timestamps
- **Webhook authentication**: Resend signature verification (not JWT), webhook endpoint is public (no auth middleware)

### Event-Driven Integration

- **Async via RabbitMQ**: All cross-module communication uses RabbitMQ events. No module directly imports the notification module's use cases
- **14 subscribed events**: Mapping from source module events to NotificationTypes (account.registered → WelcomeMessage, thread.reply → CommunityReply, etc.)
- **1 published event**: `notification.failed` published when a notification exhausts all retries
- **Event handler registration**: Follows existing pattern in `internal/handlers/handlers.go`

### Device Registration

- **Two-token model**: `DeviceToken` (platform-specific identifier, uniqueIndex for dedup) + `PushToken` (delivery token, nullable — may not be available at registration time)
- **PushToken format varies by DeviceType**: FCM token string (Android), APNs device token string (iOS), JSON web push subscription (Web)
- **Auto-upsert on re-registration**: If DeviceToken already exists, update PushToken instead of creating duplicate

### Workers

- **Delivery worker**: Polls every 5 seconds, batch size 50, processes pending queue items
- **Campaign scheduler**: Polls every 30 seconds, checks for scheduled campaigns ready to send
- **Inbox expiry cleanup**: Runs every 1 hour, deletes expired inbox entries
- **All workers**: Started via `fx.Lifecycle.OnStart`, respect context cancellation for graceful shutdown

## Testing Decisions

### Testing Philosophy

Good tests verify external behavior, not implementation details. Tests should confirm that given certain inputs and state, the correct outputs and state changes occur — without asserting on internal method call sequences, private field values, or specific SQL queries.

### Modules to Test

| Module | Test Type | Priority | Rationale |
|---|---|---|---|
| `NotificationIngestUsecase` | Unit (mocked repos) | High | Core pipeline — must correctly resolve preferences, check mutes, render templates, and enqueue |
| `NotificationDeliveryUsecase` | Unit (mocked repos + email) | High | Core pipeline — must correctly handle success, failure, retry, and cancellation paths |
| `NotificationPreferenceUsecase` | Unit (mocked repos) | High | Preference resolution logic is nuanced (three layers + quiet hours) |
| `NotificationInboxUsecase` | Unit (mocked repos) | Medium | User-facing — must correctly filter, paginate, and update read/archive state |
| `NotificationCampaignUsecase` | Unit (mocked repos) | Medium | Segment resolution and status transitions must be correct |
| `EmailDeliveryUsecase` | Unit (mocked repos) | Medium | Webhook event mapping to field updates must be accurate |
| `TemplateRenderer` | Unit (no mocks needed) | High | Pure function — variable substitution must handle missing variables, special characters, and multi-channel rendering |
| `CampaignProcessor` | Unit (mocked repos) | Medium | Segment filter resolution logic |
| Repository implementations | Integration (test DB) | Low | Standard GORM patterns — low risk, high setup cost. Test only custom queries (FetchPending, Upsert with OnConflict) |
| HTTP handlers | Integration (test server) | Low | Standard HTTP patterns — test auth middleware integration and request/response mapping |
| Resend webhook handler | Integration | Medium | Signature verification + event routing — correctness matters for security |

### Prior Art

- Existing tests in `internal/modules/community/` and `internal/modules/iam/` follow the pattern of table-driven unit tests with mocked repository interfaces
- The codebase uses `testify/suite` and `testify/mock` for test organization and mocking
- Integration tests use a test database with transaction rollback for isolation

### Test Cases to Prioritize

1. **Ingest pipeline**: Given a RabbitMQ event → verify correct template loaded, preferences checked, content rendered, queue entry created
2. **Preference resolution**: Given IAM toggle OFF → no notifications sent. Given IAM toggle ON but per-type override OFF → no notifications for that type. Given no override → template default used
3. **Mute resolution**: Given account A muted account B → notifications from B to A suppressed. Given expired mute → notifications delivered
4. **Delivery retry**: Given first failure → RetryCount=1, ScheduledFor updated with backoff. Given RetryCount >= MaxRetries → Status=Failed
5. **Campaign segment resolution**: Given scheduled campaign → TargetSegment.filters resolved to static resolvedAccountIDs
6. **Webhook event mapping**: Given "delivered" event → DeliveredAt set. Given "bounced" event → BounceReason set, DeliveryStatus updated
7. **Template rendering**: Given variables map → all `{{variable}}` placeholders replaced. Given missing required variable → error returned
8. **Inbox expiry**: Given inbox entry with ExpiresAt in the past → cleanup job deletes it

## Out of Scope

- **Push notification delivery (FCM/APNs)**: Device registration is in scope, but actual push sending is Phase 2. Phase 1 creates queue entries for Push channel but logs a warning on delivery
- **SMS delivery**: Channel enum includes SMS, but no implementation. Templates may include SMS content, but delivery is not wired
- **Triggered campaigns**: Only Broadcast and Segmented campaigns are implemented. Event-triggered campaigns are deferred to Phase 2
- **Repeat engagement tracking**: `EmailDeliveryLog` stores only first open/click timestamps. An `EmailDeliveryEvent` detail table for repeat tracking is Phase 2
- **Real-time WebSocket/SSE delivery**: In-app notifications are pull-based (API polling). WebSocket push for live inbox updates is Phase 2
- **Batch API for preference updates**: Single-preference updates only in Phase 1. Bulk preference update endpoint is Phase 2
- **Notification digest/summary emails**: Daily/weekly digest emails aggregating multiple notifications is Phase 2
- **Compliance reminders**: `ComplianceDeadline` and `ComplianceReminderLog` entities were considered but moved out — this is business domain data, not notification infrastructure
- **Replacing existing `pkg/email` SMTP**: Resend coexists with existing SMTP emailer. Migration of existing email flows (welcome, OTP) to Resend is a separate effort
- **Notification templating UI**: Admin CRUD via API is in scope. A visual template editor is out of scope
- **A/B testing for campaigns**: Campaign variants and A/B testing are deferred
- **Notification rate limiting**: Per-user rate limiting on notification delivery (beyond quiet hours) is Phase 2

## Further Notes

### Phase Plan

**Phase 1 (current PRD):**
- Domain entities, enums, repositories, use cases
- Template CRUD with translations
- Ingest pipeline (RabbitMQ event subscription → queue)
- Delivery pipeline (queue worker → email via Resend + in-app inbox)
- Inbox CRUD (list, read, archive, delete, expiry)
- Preference management (per-type overrides, quiet hours)
- Account muting
- Device registration (push token storage only, no push delivery)
- Campaign management (Broadcast + Segmented)
- Email delivery tracking via Resend webhooks
- Admin monitoring (queue status, retry)

**Phase 2 (future):**
- Push notification delivery (FCM/APNs)
- Triggered campaigns
- Repeat engagement tracking (EmailDeliveryEvent table)
- Real-time inbox updates (WebSocket/SSE)
- Batch preference updates
- Notification digest emails
- Per-user rate limiting
- Replace existing `pkg/email` SMTP with Resend

### Configuration

The module requires these configuration additions to `core.Config`:

- `Resend.Enabled` — toggle Resend integration
- `Resend.APIKey` — Resend API key
- `Resend.WebhookSecret` — for webhook signature verification
- `Resend.FromEmail` — sender email address
- `Resend.FromName` — sender display name

### Data Seeding

System-managed templates (`IsSystemManaged = true`) for all 15 notification types should be seeded via migration or startup hook. These templates define the initial content and defaults for each type. Admins can update content but cannot delete system-managed templates.

### Migration Strategy

All 10 entities are new tables — no existing schema changes required. The notification module's `EntityProvider` registers all entities with `SchemaManager`, and Atlas generates the migration files. FK references to `accounts` table (in IAM module) are cross-module references that Atlas handles.

### IAM Integration Note

The notification module reads IAM's `NotificationPreference` entity (global channel toggles) but does not modify it. The IAM module continues to own the global toggle CRUD. The notification module's `NotificationPreferenceUsecase.GetEffectivePreference` reads the IAM toggle as the first gate in the resolution chain. This avoids cross-module writes and keeps the IAM entity graph intact.
