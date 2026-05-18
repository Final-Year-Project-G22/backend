# Scheduled Alerts, Compliance, and Business Alerts

This module adds three connected features to the notification system:

| Feature | Type | Description |
|---------|------|-------------|
| **Scheduled Alert** | User-defined | Users create notifications that fire at a future time with custom title, body, and channel. Limited to 3 pending for non-pro users. |
| **Compliance Entry** | User-managed | Tracked deadlines tied to a Business Profile (tax registration, trade license, business registration) with expiry dates and reminder windows. |
| **Business Alert** | System-generated | Auto-triggered when a Compliance Entry's expiry date enters its reminder window. Delivered through the standard notification pipeline. |
| **Compliance Calendar** | Read-only view | Aggregates Compliance Entry deadlines and active Scheduled Alerts on a timeline. Widget on Home dashboard + full view in Notifications tab. |

## Relationship to existing notification module

All three features integrate with the existing notification delivery pipeline (`NotificationQueue` → `DeliveryWorker` → History + Inbox). Business Alerts use the existing `account_alert_info` notification type. Scheduled Alerts bypass the template system since content is user-provided.

## Directory Structure

```
internal/modules/notification/
├── ... (existing files)
├── domain/entity/
│   ├── user_scheduled_notification.go      (NEW)
│   ├── scheduled_alert_template.go         (NEW)
│   ├── compliance_entry.go                 (NEW)
│   └── enums.go                            (MODIFIED: new ScheduleStatus, ComplianceType, ComplianceEntryStatus)
├── domain/repository/
│   ├── user_scheduled_notification_repository.go   (NEW)
│   ├── scheduled_alert_template_repository.go      (NEW)
│   ├── compliance_entry_repository.go              (NEW)
│   └── subscription_reader.go                      (NEW)
├── domain/usecase/
│   ├── inputs.go                                   (MODIFIED: new input DTOs)
│   ├── user_scheduled_notification_usecase.go      (NEW)
│   └── compliance_entry_usecase.go                 (NEW)
├── application/usecase/
│   ├── user_scheduled_notification_usecase.go      (NEW)
│   └── compliance_entry_usecase.go                 (NEW)
├── application/service/
│   ├── user_notification_scheduler.go              (NEW)
│   └── business_alert_scheduler.go                 (NEW)
├── infrastructure/repository/
│   ├── user_scheduled_notification_repository.go   (NEW)
│   ├── scheduled_alert_template_repository.go      (NEW)
│   └── compliance_entry_repository.go              (NEW)
├── delivery/handler/
│   ├── scheduled_alert_handler.go                  (NEW)
│   └── compliance_handler.go                       (NEW)
├── delivery/dto/
│   ├── scheduled_alert_dto.go                      (NEW)
│   └── compliance_dto.go                           (NEW)
├── delivery/routes/
│   ├── scheduled_alert_routes.go                   (NEW)
│   └── compliance_routes.go                        (NEW)
├── entities.go                                     (MODIFIED: register new entities)
├── module.go                                       (MODIFIED: DI wiring)
└── delivery/routes/routes.go                       (MODIFIED: register new routes)

internal/modules/
├── modules.go                                      (MODIFIED: SubscriptionReader override)

internal/modules/payment/infrastructure/notification/
└── subscription_reader.go                          (NEW)
```
