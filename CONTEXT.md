# Core Backend Notifications Context

This context defines the shared business language for user-facing notifications emitted by backend modules and delivered through the notification module. It exists to keep event contracts, delivery intent, and migration decisions consistent across modules.

## Language

### Core notification terms

**Notification Type**:
The business meaning of a notification intent, such as welcome, account verification, password reset, or account alert.
_Avoid_: Channel, medium

**Channel**:
The transport used to deliver a notification to a user: in-app, email, push, or SMS.
_Avoid_: Type

**Channel Policy**:
A rule describing whether a notification type targets one explicit channel (`single`) or all enabled channels for a user (`all_enabled`).
_Avoid_: Type policy

**All Enabled Channels**:
Delivery to every channel that is both configured for the notification type and enabled/eligible for the target user.
_Avoid_: Broadcast to everything

### Event and delivery terms

**Canonical Notification Event Envelope**:
A versioned event shape used by publishers that includes event identity, account identity, notification type, channel policy, variables, and metadata.
_Avoid_: Ad-hoc event payload

**Publisher-Owned Variables**:
Business variables required by templates are assembled by the source module that emits the event.
_Avoid_: Notification-enriched business context

**Idempotency Key**:
A deterministic key that uniquely identifies one notification intent so retries do not cause duplicate sends.
_Avoid_: Random dedupe token

**Skipped Delivery**:
A non-error channel outcome where delivery is intentionally not attempted due to policy, preference, eligibility, or expiry.
_Avoid_: Failure

**Failed Delivery**:
A channel outcome where delivery was attempted but did not succeed after retry policy.
_Avoid_: Skipped

### Security notification terms

**Critical Account Alert**:
An account security notification that is mandatory and not user-disableable.
_Avoid_: Optional alert

**Informational Account Alert**:
An account security-related notification that users can control with preferences.
_Avoid_: Mandatory alert

## Relationships

- A **Notification Type** can target one or more **Channels** according to its **Channel Policy**.
- **All Enabled Channels** includes **in-app** by default and filters by user preferences and channel eligibility.
- A **Canonical Notification Event Envelope** includes one **Notification Type**, one target account, and publisher-owned variables.
- One notification intent is protected by one **Idempotency Key**, and each channel can produce independent outcomes: delivered, skipped, or failed.
- **Critical Account Alert** uses always-on delivery policy and bypasses user opt-out preferences.
- **Informational Account Alert** uses standard preference-controlled delivery.

## Example dialogue

> **Dev:** "For password reset, should we send email and push if both are enabled?"
> **Domain expert:** "Yes, because the channel policy is all enabled channels, but only if each channel is eligible for that account and the reset notification is still within TTL."

## Flagged ambiguities

- "notification type" was initially used to mean delivery transports (email/push/SMS); resolved: transports are **Channels**, while **Notification Type** means business intent.
- "account alert" was initially treated as one class of message; resolved into **Critical Account Alert** and **Informational Account Alert** to separate mandatory vs preference-controlled behavior.
