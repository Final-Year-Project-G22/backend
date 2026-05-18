# API Contract

All endpoints require `AuthMiddleware` + `AccountStatusMiddleware` (bearer JWT).

Base path: `/api/v1/notifications`

## Scheduled Alerts

### List Scheduled Alerts

```
GET /scheduled
```

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid",
      "templateSlug": "tax_filing",
      "title": "Tax Filing Due",
      "body": "Your tax filing deadline is approaching.",
      "channel": "in_app",
      "scheduledFor": "2026-06-15T08:00:00Z",
      "status": "pending",
      "rescheduledFrom": null,
      "sentAt": null,
      "cancelledAt": null,
      "createdAt": "2026-05-18T10:00:00Z"
    }
  ]
}
```

### Create Scheduled Alert

```
POST /scheduled
```

**Request body:**
```json
{
  "templateSlug": "tax_filing",
  "title": "Tax Filing Due",
  "body": "Your tax filing deadline is approaching.",
  "channel": "in_app",
  "scheduledFor": "2026-06-15T08:00:00Z"
}
```

**Response 201:**
```json
{
  "id": "uuid",
  "message": "Scheduled alert created"
}
```

**Error 403 (pro limit):**
```json
{
  "error": "max_scheduled_alerts_reached",
  "message": "Upgrade to Pro to create more than 3 scheduled alerts"
}
```

### Cancel Scheduled Alert

```
PATCH /scheduled/{id}/cancel
```

**Response 200:**
```json
{
  "message": "Scheduled alert cancelled"
}
```

### Reschedule Scheduled Alert

```
PATCH /scheduled/{id}/reschedule
```

**Request body:**
```json
{
  "scheduledFor": "2026-07-01T08:00:00Z"
}
```

**Response 200:**
```json
{
  "message": "Scheduled alert rescheduled",
  "previousScheduledFor": "2026-06-15T08:00:00Z"
}
```

### List Scheduled Alert Templates

```
GET /scheduled/templates
```

**Response 200:**
```json
{
  "data": [
    {
      "slug": "custom",
      "name": "Custom",
      "defaultTitle": "",
      "defaultBody": "",
      "defaultChannel": null
    },
    {
      "slug": "tax_filing",
      "name": "Tax Filing Reminder",
      "defaultTitle": "Tax Filing Due",
      "defaultBody": "Your tax filing deadline is approaching.",
      "defaultChannel": "in_app"
    }
  ]
}
```

---

Base path: `/api/v1/compliance`

## Compliance Entries

### List Compliance Entries

```
GET /entries?businessProfileId=uuid
```

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid",
      "businessProfileId": "uuid",
      "complianceType": "tax_registration",
      "referenceNumber": "1234567890",
      "issuedDate": "2025-01-15",
      "expiryDate": "2026-01-15T00:00:00Z",
      "reminderDaysBefore": 30,
      "status": "active",
      "lastNotifiedAt": null
    }
  ]
}
```

### Create Compliance Entry

```
POST /entries
```

**Request body:**
```json
{
  "businessProfileId": "uuid",
  "complianceType": "tax_registration",
  "referenceNumber": "1234567890",
  "issuedDate": "2025-01-15",
  "expiryDate": "2026-01-15T00:00:00Z",
  "reminderDaysBefore": 30
}
```

**Response 201:**
```json
{
  "id": "uuid",
  "message": "Compliance entry created"
}
```

### Update Compliance Entry

```
PATCH /entries/{id}
```

**Request body (partial):**
```json
{
  "referenceNumber": "0987654321",
  "expiryDate": "2027-01-15T00:00:00Z",
  "reminderDaysBefore": 45
}
```

**Response 200:**
```json
{
  "message": "Compliance entry updated"
}
```

### Delete Compliance Entry

```
DELETE /entries/{id}
```

**Response 200:**
```json
{
  "message": "Compliance entry deleted"
}
```

## Compliance Calendar

### Get Calendar

```
GET /calendar
```

**Response 200:**
```json
{
  "entries": [
    {
      "id": "uuid",
      "type": "compliance",
      "title": "Tax Registration Expiry",
      "referenceNumber": "1234567890",
      "date": "2026-01-15T00:00:00Z",
      "daysRemaining": 45,
      "status": "active"
    },
    {
      "id": "uuid",
      "type": "scheduled_alert",
      "title": "Tax Filing Due",
      "date": "2026-06-15T08:00:00Z",
      "daysRemaining": 28,
      "status": "pending"
    }
  ]
}
```
