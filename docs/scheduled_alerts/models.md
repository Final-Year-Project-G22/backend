# Data Models

## New Tables

### user_scheduled_notifications

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK, default gen_random_uuid() | |
| account_id | uuid | NOT NULL, FK → accounts.id | |
| template_slug | varchar(64) | NULLABLE | Which template was picked (for reference) |
| title | varchar(255) | NOT NULL | User-editable |
| body | text | NOT NULL | User-editable |
| channel | varchar(20) | NOT NULL | One of: in_app, email, push |
| scheduled_for | timestamptz | NOT NULL, INDEX | When to fire |
| status | varchar(20) | NOT NULL, DEFAULT 'pending', INDEX | pending, sent, cancelled |
| rescheduled_from | timestamptz | NULLABLE | Original scheduled time if rescheduled |
| sent_at | timestamptz | NULLABLE | When scheduler processed it |
| cancelled_at | timestamptz | NULLABLE | When user cancelled |
| created_at | timestamptz | NOT NULL, DEFAULT CURRENT_TIMESTAMP | |
| updated_at | timestamptz | NOT NULL, DEFAULT CURRENT_TIMESTAMP | |
| deleted_at | timestamptz | NULLABLE | Soft delete |

**Indexes:**
- `idx_user_scheduled_account` on `account_id`
- `idx_user_scheduled_time` on `scheduled_for`
- `idx_user_scheduled_status` on `status`
- Composite: `(account_id, status)` for pro limit count query

### scheduled_alert_templates

Seeded table — users read, never write.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK | |
| slug | varchar(64) | NOT NULL, UNIQUE | e.g., "custom", "tax_filing", "license_renewal" |
| name | varchar(255) | NOT NULL | Display name |
| default_title | varchar(255) | NOT NULL | Pre-fill value |
| default_body | text | NOT NULL | Pre-fill value |
| default_channel | varchar(20) | NULLABLE | Optional channel suggestion |
| is_active | boolean | NOT NULL, DEFAULT true | |
| created_at | timestamptz | NOT NULL | |
| updated_at | timestamptz | NOT NULL | |

**Seed data:**

| slug | name | default_title | default_body | default_channel |
|------|------|---------------|--------------|-----------------|
| `custom` | Custom | (empty) | (empty) | NULL |
| `tax_filing` | Tax Filing Reminder | Tax Filing Due | Your tax filing deadline is approaching. | in_app |
| `license_renewal` | License Renewal | License Expiring | Your trade license renewal is due soon. | in_app |
| `registration_renewal` | Registration Renewal | Registration Expiring | Your business registration renewal is approaching. | in_app |
| `meeting` | Meeting Reminder | Meeting Today | Don't forget your scheduled meeting. | push |
| `deadline` | Custom Deadline | Deadline Approaching | A deadline you set is coming up. | in_app |

### compliance_entries

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| id | uuid | PK | |
| business_profile_id | uuid | NOT NULL, FK → business_profiles.id | |
| compliance_type | varchar(64) | NOT NULL | One of seeded types |
| reference_number | varchar(255) | NULLABLE | TIN, license number, etc. |
| issued_date | date | NULLABLE | |
| expiry_date | timestamptz | NOT NULL | |
| reminder_days_before | int | NOT NULL, DEFAULT 30 | e.g., 30 days before expiry |
| status | varchar(20) | NOT NULL, DEFAULT 'active' | active, expired, renewed |
| last_notified_at | timestamptz | NULLABLE | Prevents duplicate alerts |
| created_at | timestamptz | NOT NULL | |
| updated_at | timestamptz | NOT NULL | |
| deleted_at | timestamptz | NULLABLE | |

**Indexes:**
- `idx_compliance_profile` on `business_profile_id`
- `idx_compliance_expiry` on `expiry_date`
- `idx_compliance_status` on `status`

## New Enums

```go
type ScheduleStatus string
const (
    ScheduleStatusPending   ScheduleStatus = "pending"
    ScheduleStatusSent      ScheduleStatus = "sent"
    ScheduleStatusCancelled ScheduleStatus = "cancelled"
)

type ComplianceType string
const (
    ComplianceTypeTaxRegistration     ComplianceType = "tax_registration"
    ComplianceTypeTradeLicense        ComplianceType = "trade_license"
    ComplianceTypeBusinessRegistration ComplianceType = "business_registration"
)

type ComplianceEntryStatus string
const (
    ComplianceEntryStatusActive  ComplianceEntryStatus = "active"
    ComplianceEntryStatusExpired ComplianceEntryStatus = "expired"
    ComplianceEntryStatusRenewed ComplianceEntryStatus = "renewed"
)
```

## Existing Table Additions

### NotificationType enum (enums.go)

```go
const (
    // ... existing types ...
    NotificationTypeUserScheduled NotificationType = "user_scheduled"
)
```

This type allows the NotificationQueue to carry user-scheduled notifications through the delivery pipeline.
