# Use Cases

## UserScheduledNotificationUsecase

### Schedule(ctx, accountID, input *ScheduleUserNotificationInput) (*UserScheduledNotification, error)

**Input:**
```go
type ScheduleUserNotificationInput struct {
    TemplateSlug *string           // optional template reference
    Title        string
    Body         string
    Channel      entity.Channel    // in_app, email, push
    ScheduledFor time.Time
}
```

**Logic:**
1. Validate channel is one of `in_app`, `email`, `push`
2. Validate `scheduled_for` is in the future (or within a small tolerance)
3. Count pending notifications for this account
4. If count >= 3:
   - Call `SubscriptionReader.HasActiveProSubscription(ctx, accountID)`
   - If false → return `ErrMaxScheduledAlertsReached`
5. Create `UserScheduledNotification` with `status=pending`

### List(ctx, accountID) ([]*UserScheduledNotification, error)

**Logic:**
1. Query all `UserScheduledNotification` where `account_id = ?`
2. Order by `scheduled_for ASC`
3. Return list

### GetByID(ctx, accountID, id uuid.UUID) (*UserScheduledNotification, error)

### Cancel(ctx, accountID, id uuid.UUID) error

**Logic:**
1. Fetch by ID
2. Verify `entity.AccountID == accountID` (ownership)
3. Verify `entity.Status == pending` (cannot cancel already sent/cancelled)
4. Set `Status = cancelled`, `CancelledAt = now`

### Reschedule(ctx, accountID, id uuid.UUID, newScheduledFor time.Time) error

**Logic:**
1. Fetch by ID
2. Verify ownership and pending status
3. Set `RescheduledFrom = entity.ScheduledFor` (preserve original)
4. Update `ScheduledFor = newScheduledFor`

## ScheduledAlertTemplateUsecase

### ListActive(ctx) ([]*ScheduledAlertTemplate, error)

Returns all active templates for the template picker UI.

## ComplianceEntryUsecase

### Create(ctx, accountID uuid.UUID, input *CreateComplianceEntryInput) (*ComplianceEntry, error)

**Input:**
```go
type CreateComplianceEntryInput struct {
    BusinessProfileID   uuid.UUID
    ComplianceType      entity.ComplianceType
    ReferenceNumber     *string
    IssuedDate          *time.Time
    ExpiryDate          time.Time
    ReminderDaysBefore  int           // default 30
}
```

**Logic:**
1. Validate business profile belongs to the account
2. Validate compliance type is a known seeded type
3. Set `status = active`
4. Create entry

### ListByBusinessProfile(ctx, accountID, businessProfileID uuid.UUID) ([]*ComplianceEntry, error)

Returns all entries for a business profile, ordered by `expiry_date ASC`.

### Update(ctx, accountID, id uuid.UUID, input *UpdateComplianceEntryInput) error

**Input:**
```go
type UpdateComplianceEntryInput struct {
    ReferenceNumber     *string
    IssuedDate          *time.Time
    ExpiryDate          *time.Time
    ReminderDaysBefore  *int
    Status              *entity.ComplianceEntryStatus
}
```

**Logic:**
1. Verify ownership via business profile chain
2. Partial update on provided fields

### Delete(ctx, accountID, id uuid.UUID) error

Hard-delete the compliance entry.

### GetCalendar(ctx, accountID uuid.UUID) (*ComplianceCalendar, error)

**Output:**
```go
type ComplianceCalendar struct {
    Deadlines   []CalendarEntry   // from compliance_entries
    ScheduledAlerts []CalendarEntry // from user_scheduled_notifications
}

type CalendarEntry struct {
    ID          uuid.UUID
    Type        string        // "compliance" or "scheduled_alert"
    Title       string
    Date        time.Time
    DaysRemaining int
    Status      string
}
```

**Logic:**
1. Fetch all `compliance_entries` where `expiry_date > now AND business_profile_id IN (user's profiles)`
2. Map to `CalendarEntry{Type: "compliance"}`
3. Fetch all `user_scheduled_notifications` where `status=pending AND account_id = ?`
4. Map to `CalendarEntry{Type: "scheduled_alert"}`
5. Merge both lists, sort by date ASC
