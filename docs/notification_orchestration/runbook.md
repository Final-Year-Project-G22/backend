# Notification Orchestration Runbook

## Overview

This runbook documents the notification orchestration system migration acceptance criteria, observability, and operational procedures.

## Acceptance Criteria (Verified)

### 1. No Direct IAM Sends for Migrated Intents

**Status:** PASS

**Verification:**
- `auth_service.go` no longer contains `messageBus.Publish` calls
- `internal/events/subscribers.go` no longer handles `account.registered` or `user.email_otp_requested`
- Legacy event structs (`AccountRegisteredEvent`, `UserEmailOTPRequestedEvent`) deleted
- Only `admin.created` retains direct bus publish (out of scope)

**Evidence:**
```bash
grep -r "messageBus.Publish" internal/modules/iam/application/service/
# Only admin_service.go:214 remains (out of scope)
```

### 2. All Migrated Intents Publish Canonical Events Through Notification Outbox

**Status:** PASS

**Verification:**
- `Register()` writes `account_verification` (single/email) to outbox inside transaction
- `VerifyEmailOTP()` writes `welcome_message` (all_enabled) to outbox inside transaction
- `ResendEmailOTP()` writes `account_verification` (single/email) to outbox inside transaction
- `UpdateAccountPassword()` writes `account_alert_critical` (all_enabled) to outbox inside transaction

**Evidence:**
```bash
grep -n "writeWelcomeNotificationOutbox\|writeOTPNotificationOutbox\|writeAccountAlertOutbox" \
  internal/modules/iam/application/service/auth_service.go
```

### 3. Zero Canonical Validation Rejects

**Status:** Requires Runtime Observation

**How to Verify:**
Monitor logs for `Canonical event rejected` messages:
```bash
# During testing, grep logs for:
"Canonical event rejected"
```

**Expected:** Zero occurrences during normal operation.

**If found:** Check the `reason` field for specific validation errors:
- `missing_required_field:*` — Publisher missing required envelope field
- `invalid_schema_version:*` — Version mismatch
- `missing_single_channel` — Single policy without channel specified
- `forbidden_channel_for_all_enabled` — Channel provided with all_enabled policy

### 4. Security Flows Meet Delivery SLO (No Stale Sends)

**Status:** Requires Runtime Observation

**TTL Protection:**
- `account_verification`: No explicit TTL on template (OTP expiry handled by IAM)
- `password_reset`: Not yet implemented (future work)
- `account_alert_critical`: No TTL (immediate delivery)

**How to Verify:**
1. Trigger OTP send → verify email arrives within seconds
2. Wait 5+ minutes → verify OTP email link/code is rejected by IAM
3. Verify no "stale" OTP emails are queued/delivered after expiry

### 5. Zero Duplicate Sends

**Status:** PASS (Design), Requires Runtime Observation

**Idempotency Keys:**
- Welcome: `welcome:{accountId}`
- OTP: `verify-email:{accountId}:{otpRecordId}`
- Alert: `account-alert:{accountId}:{alertCode}:{uuid}`

**Protection:**
- Outbox table has `UNIQUE INDEX` on `idempotency_key`
- Duplicate writes to outbox will fail with unique constraint violation
- Dispatcher only processes `pending` rows

**How to Verify:**
1. Trigger same OTP resend twice rapidly
2. Check `notification_outbox` table — should have 2 rows with different idempotency keys
3. Verify only 2 emails sent (one per record, not 4)

### 6. Dead-Letter and Retry Runbook

**Status:** Documented

#### Retry Behavior

| Attempt | Backoff Delay |
|---------|--------------|
| 1 | 1 minute |
| 2 | 2 minutes |
| 3 | 4 minutes |
| 4 | 8 minutes |
| 5 | 16 minutes |
| 6-8 | 30 minutes each |

After 8 attempts: Row moves to `dead_letter` status.

#### Monitoring Queries

```sql
-- Pending rows (should be brief spikes, not growing)
SELECT COUNT(*) FROM notification_outbox WHERE status = 'pending';

-- Retry scheduled rows
SELECT COUNT(*) FROM notification_outbox 
WHERE status = 'pending' AND next_attempt_at > NOW();

-- Dead letter rows (requires investigation)
SELECT id, event_type, idempotency_key, attempt_count, last_error, created_at 
FROM notification_outbox 
WHERE status = 'dead_letter'
ORDER BY created_at DESC;

-- Recent failures by type
SELECT event_type, COUNT(*) as fail_count, MAX(attempt_count) as max_attempts
FROM notification_outbox
WHERE status = 'dead_letter'
GROUP BY event_type;
```

#### Resolving Dead Letters

1. **Check root cause:**
   ```sql
   SELECT last_error FROM notification_outbox WHERE id = 'dead-letter-uuid';
   ```

2. **Common causes:**
   - RabbitMQ connection failure → Restart dispatcher
   - Invalid payload → Check publisher code
   - Broker topic not declared → Verify event type matches subscription

3. **Manual retry (if root cause fixed):**
   ```sql
   UPDATE notification_outbox 
   SET status = 'pending', 
       next_attempt_at = NOW(), 
       attempt_count = 0,
       last_error = NULL
   WHERE id = 'dead-letter-uuid';
   ```

#### Alerting Recommendations

- **Warning:** `notification_outbox` pending rows > 100 for > 5 minutes
- **Critical:** `notification_outbox` dead_letter rows created in last hour
- **Critical:** Security notification (`account_verification`, `account_alert_critical`) delivery failure rate > 5%

## Migration Impact Summary

### What Changed

1. **Canonical Event Contract** — All notification events now use typed envelope with strict validation
2. **Notification Outbox** — Reliable at-least-once delivery with idempotency
3. **Channel Policy** — Explicit `single` vs `all_enabled` semantics
4. **IAM Integration** — Auth flows now write to notification outbox instead of direct email
5. **Legacy Removal** — Direct email subscribers removed for migrated intents

### What Did Not Change

- `admin.created` email flow (out of scope)
- Payment module integration (future work)
- AI module outbox (separate domain)
- Non-IAM notification events (community, guide, AI) — still use same event bus, but now parsed canonically

## Rollback Notes

If critical issues are found:

1. **Restore direct email sends** by reverting `auth_service.go` to use `messageBus.Publish` goroutines
2. **Restore old subscribers** in `internal/events/subscribers.go`
3. **Disable dispatcher** by commenting out the fx.Invoke in `notification/module.go`

All changes are isolated to specific functions and can be reverted individually.

---

## Developer Migration Guide

### How It Used to Work (Legacy Approach)

Previously, modules sent notifications by publishing ad-hoc events directly to RabbitMQ:

```go
// OLD: Direct bus publish with module-specific event structs
go s.messageBus.Publish(ctx, "account.registered", event.AccountRegisteredEvent{
    ID:        user.ID.String(),
    Email:     account.Email,
    FirstName: user.FirstName,
    LastName:  user.LastName,
})
```

Problems with this approach:
- **No retry guarantee** — if the bus was down, the notification was lost
- **No idempotency** — retries could cause duplicate emails
- **Fragmented contracts** — each module invented its own event shape
- **Direct coupling** — publishers knew about email templates and sending logic
- **No queueing/delivery tracking** — impossible to know if a notification was actually sent

### How It Works Now (New Approach)

Modules now write a **canonical notification envelope** to the shared `notification_outbox` table inside the same database transaction as the business change:

```go
// NEW: Write canonical envelope to outbox (inside transaction)
env := notificationevent.Envelope{
    SchemaVersion:    "1.0.0",
    EventType:        "account.registered",
    OccurredAt:       time.Now().UTC(),
    SourceModule:     "iam",
    AccountID:        account.ID,
    NotificationType: "welcome_message",
    ChannelPolicy:    "all_enabled",  // or "single"
    Variables: map[string]string{
        "platformName":      "MyApp",
        "accountName":       user.FirstName,
        "gettingStartedUrl": "/guides",
    },
    Metadata: notificationevent.Metadata{
        IdempotencyKey: "welcome:" + account.ID.String(),
        Locale:         nil,
    },
}

// This helper marshals the envelope and writes to notification_outbox
return s.writeEnvelopeToOutbox(ctx, &env)
```

The **Notification Module** then:
1. **Dispatches** the envelope from outbox to RabbitMQ reliably (with retry + dead-letter)
2. **Subscribes** to the event and parses the canonical envelope
3. **Renders** templates using the publisher-provided variables
4. **Enqueues** per-channel delivery jobs (email, in-app, push, SMS)
5. **Tracks** delivery status in `notification_queue` and `notification_history`

### How to Add Notifications from a New Module

#### Step 1: Add the Notification Outbox Repository

Inject `notifrepo.NotificationOutboxRepository` into your service:

```go
import (
    notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
    "github.com/Final-Year-Project-G22/backend/core/internal/shared/notificationevent"
    notifentity "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
    "gorm.io/datatypes"
    "encoding/json"
)

type myService struct {
    notifOutboxRepo notifrepo.NotificationOutboxRepository
    cfg             *core.Config
}
```

#### Step 2: Write the Envelope Inside Your Transaction

Build and write the canonical envelope as part of your business transaction:

```go
func (s *myService) DoSomething(ctx context.Context, accountID uuid.UUID) error {
    return s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
        // 1. Do your business logic
        if err := s.repo.DoBusinessThing(txCtx, accountID); err != nil {
            return err
        }

        // 2. Write notification outbox row (atomic with business change)
        if err := s.writeMyNotificationOutbox(txCtx, accountID); err != nil {
            s.logger.Error("Failed to write notification outbox", core.Error(err))
            // Non-fatal: don't fail the business transaction
        }

        return nil
    })
}
```

#### Step 3: Implement Your Outbox Helper

```go
func (s *myService) writeMyNotificationOutbox(ctx context.Context, accountID uuid.UUID) error {
    env := notificationevent.Envelope{
        SchemaVersion:    notificationevent.SchemaVersionV1,
        EventType:        "my_module.my_event",     // Your event name
        OccurredAt:       time.Now().UTC(),
        SourceModule:     "my_module",              // Your module name
        AccountID:        accountID,
        NotificationType: "my_notification_type",   // Must match a template in DB
        ChannelPolicy:    notificationevent.ChannelPolicyAllEnabled, // or ChannelPolicySingle
        Variables: map[string]string{
            "key1": "value1",
            "key2": "value2",
        },
        Metadata: notificationevent.Metadata{
            IdempotencyKey: "my-module:" + accountID.String() + ":" + uuid.New().String(),
            Locale:         nil, // or &"en"
        },
    }
    return s.writeEnvelopeToOutbox(ctx, &env)
}

// Reusable helper (copy from auth_service.go)
func (s *myService) writeEnvelopeToOutbox(ctx context.Context, envelope *notificationevent.Envelope) error {
    data, err := json.Marshal(envelope)
    if err != nil {
        return fmt.Errorf("failed to marshal envelope: %w", err)
    }

    var payload datatypes.JSONMap
    if err := json.Unmarshal(data, &payload); err != nil {
        return fmt.Errorf("failed to convert envelope to JSONMap: %w", err)
    }

    outbox := &notifentity.NotificationOutbox{
        EventType:      envelope.EventType,
        SchemaVersion:  envelope.SchemaVersion,
        SourceModule:   envelope.SourceModule,
        AccountID:      envelope.AccountID,
        IdempotencyKey: envelope.Metadata.IdempotencyKey,
        Payload:        payload,
        Status:         notifentity.NotificationOutboxStatusPending,
        AttemptCount:   0,
    }

    return s.notifOutboxRepo.Create(ctx, outbox)
}
```

#### Step 4: Create the Notification Template

Add a system-managed template via migration:

```sql
INSERT INTO notification_templates (id, name, description, notification_type, category, priority, is_system_managed, default_content, variables_schema)
VALUES (
  gen_random_uuid(),
  'My Notification',
  'Description of what this notification does.',
  'my_notification_type', 'my_category', 1, true,
  '{"email":{"subject":"Hello {{name}}","body":"<p>Your {{thing}} is ready</p>"},"inapp":{"title":"{{thing}} Ready","body":"Your {{thing}} is ready","actionUrl":"/things/{{thingId}}"}}'::jsonb,
  '{"required":["name","thing","thingId"]}'::jsonb
)
ON CONFLICT (notification_type) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  default_content = EXCLUDED.default_content,
  variables_schema = EXCLUDED.variables_schema,
  is_system_managed = true;
```

#### Step 5: Add Your Event to Notification Subscriptions

In `internal/modules/notification/subscription.go`, add your event to the subscription list:

```go
events := []string{
    // ... existing events ...
    "my_module.my_event",  // Add yours
}
```

No other changes needed — the canonical parser handles all events uniformly.

### Channel Policy Quick Reference

| Policy | Use When | Channel Field | Example |
|--------|----------|---------------|---------|
| `all_enabled` | Send to all channels configured for the template | Must be `nil` | Welcome messages, alerts |
| `single` | Send to exactly one channel | Required | OTP (email only), password reset (email only) |

### Idempotency Key Patterns

| Intent | Recommended Key Pattern |
|--------|------------------------|
| Welcome | `welcome:{accountId}` |
| OTP | `verify-email:{accountId}:{otpRecordId}` |
| Password Reset | `password-reset:{accountId}:{tokenId}` |
| Alerts | `account-alert:{accountId}:{alertCode}:{uuid}` |
| General | `{module}:{accountId}:{intentId}:{uuid}` |

### Common Mistakes to Avoid

1. **Don't pass `channel` when using `all_enabled`** — the validator will reject it
2. **Don't forget `idempotencyKey`** — required for deduplication
3. **Don't fail the business transaction on outbox errors** — log and continue
4. **Don't use free-form event payloads** — always use the canonical envelope
5. **Don't forget to create the template** — the notification module will fail to find it

### Testing Your Integration

1. Trigger your business flow
2. Check `notification_outbox` table:
   ```sql
   SELECT * FROM notification_outbox WHERE source_module = 'my_module' ORDER BY created_at DESC LIMIT 5;
   ```
3. Verify the dispatcher picks it up (check logs for "Notification outbox row published")
4. Check `notification_queue` for enqueued channel items
5. Check `notification_histories` for delivery records
