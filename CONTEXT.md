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

### Scheduled alert terms

**Scheduled Alert**:
A user-created notification that fires at a future time, with user-defined title, body, and delivery channel. Supports cancellation and rescheduling. Bound by a pro-tier limit of 3 pending items for non-pro users.
_Avoid_: Reminder, custom notification, personal alert

**Scheduled Alert Template**:
A seeded template that pre-fills the title and body of a Scheduled Alert. Users pick a template (e.g., tax filing, license renewal, custom) and may override the content.
_Avoid_: Preset, example

### Compliance terms

**Compliance Entry**:
A tracked deadline tied to a Business Profile, representing an official registration or license with an expiry date and a user-set reminder window. Examples: tax registration (TIN), trade license, business registration.
_Avoid_: Compliance record, license entry, deadline item

**Compliance Type**:
A seeded classification of compliance entries (e.g., `tax_registration`, `trade_license`, `business_registration`). Extensible by adding new seed rows.
_Avoid_: Category, kind

**Business Alert**:
A system-generated notification triggered when a Compliance Entry's expiry date falls within its configured reminder window. Delivered via the standard notification pipeline (queue → history + inbox).
_Avoid_: Compliance notification, auto-reminder

**Compliance Calendar**:
A read-only view showing upcoming Compliance Entry deadlines and active Scheduled Alerts on a timeline. Displayed as a widget on the Home dashboard and as a full view inside the Notifications tab.
_Avoid_: Deadline dashboard, compliance timeline

## Relationships (additions)

- A **Compliance Entry** belongs to exactly one **Business Profile**.
- A **Compliance Entry** has one **Compliance Type**.
- A **Business Alert** is triggered by a **Compliance Entry** reaching its reminder window.
- A **Scheduled Alert** is optionally based on a **Scheduled Alert Template**.
- A **Scheduled Alert** targets exactly one **Channel**.
- A **Compliance Calendar** aggregates **Compliance Entries** and **Scheduled Alerts** into a unified timeline.

### Localization terms

**User Locale**:
A user's language preference stored on their account profile. Controls the language of all API responses (error messages, success messages) and notification delivery (email, in-app, push). The system supports English (`en`) and Amharic (`am`).
_Avoid_: Language, region setting

**Request Locale**:
The locale detected from an incoming API request, derived from the `Accept-Language` header. Used to resolve error and success messages for that request when no user session is available.
_Avoid_: Request language, locale param

**Canonical Locale Resolution**:
A single middleware layer that extracts the request locale, stores it in the request context, and makes it available to all downstream handlers and error adapters. Eliminates ad-hoc locale extraction spread across packages.
_Avoid_: Multiple getLocale() paths

**Localized Message**:
A user-facing string (error or success) resolved at request time against the user's locale, using a dot-notated key (e.g., `notification.errors.scheduledAlertPastDue`) and the translation file for the detected locale. Falls back to English when a locale or key is missing.
_Avoid_: Hardcoded English string, raw message

**Notification Locale**:
The locale carried inside a **Canonical Notification Event Envelope**, set by the publisher from the target user's locale at the time of event creation. Ensures that asynchronously delivered notifications reach the user in their preferred language.
_Avoid_: Language field, locale metadata

## Relationships (localization)

- A **User** has one **User Locale**.
- A **Request Locale** is determined per API call; it may differ from the **User Locale**.
- All API responses use **Localized Messages** resolved against the **Request Locale**.
- A **Canonical Notification Event Envelope** carries a **Notification Locale** set from the target user's **User Locale**.
- The **Canonical Locale Resolution** middleware populates the **Request Locale** and supersedes all prior ad-hoc extraction points.

## Example dialogue

> **Dev:** "When we send a scheduled alert notification, what locale should the email use?"
> **Domain expert:** "The notification locale in the envelope — which was set from the user's stored locale at the time the event was published. If the user changed their preference afterward, it takes effect on the next notification."

## Flagged ambiguities

- "language" was initially used interchangeably with "locale"; resolved: **locale** is the canonical term, encompassing both language and regional formatting conventions.
- "notification language" was used to mean both the user's display preference and the template output language; resolved: **User Locale** governs account-level preference, **Notification Locale** is the resolved locale carried in the event envelope.

### Permission and authorization terms

**Permission Code**:
A dot-notated string that identifies one discrete action in the system, such as `library.read`, `iam.admin.list`, or `guide.write`. Permission codes follow the pattern `module.action` — where `module` is the owning module's namespace. Permissions are the atomic unit of authorization.
_Avoid_: Right, scope, privilege

**Module Permission Namespace**:
The prefix of a permission code that identifies the owning module (e.g., `library.*` for the library module, `iam.*` for the IAM module). Each module defines its own namespace, keeping authorization concerns co-located with the code they protect.
_Avoid_: Permission category, permission group

**Seed Permission**:
A permission registered at application startup by a module through the FX group injection mechanism (`group:"permission_seeds"`). Upserted into the permissions table by the Role Permission Seeder on every start. Exists for the lifetime of the application.
_Avoid_: Dynamic permission, runtime permission

**Global Role**:
A role entity stored in the IAM module that can hold **Permission Codes** from any module. Roles are not scoped to a single module. Assigning a **Global Role** to an account grants all linked permissions across all modules.
_Avoid_: Module role, scoped role

**System Role**:
A non-mutable **Global Role** seeded at startup, such as `super_admin` or `iam_admin`. Cannot be deleted or renamed through the API.
_Avoid_: Default role, built-in role

**Custom Role**:
A mutable **Global Role** created through the admin API. Can be assigned any subset of seeded permissions from any module.
_Avoid_: User role, custom permission set

**Role-Permission Assignment**:
The mapping between a **Global Role** and a **Permission Code** via the `role_permissions` join table. Determines what permissions are granted to any account that holds the role.
_Avoid_: Role permission link, grant record

**Role Bypass**:
A middleware optimization that skips the permission lookup for accounts holding an enumerated set of role codes (e.g., `super_admin`). Used when the role implicitly grants all applicable permissions.
_Avoid_: Super admin bypass, role skip

**Idempotent Permission Seed**:
A seed operation that upserts a **Seed Permission** only when its code does not already exist in the database. Prevents duplicate permission rows across restarts.
_Avoid_: Create-or-skip, first-run seed

**Super Admin Role**:
The **System Role** with code `super_admin`. Receives every **Seed Permission** from every module automatically via the seeder. Uses **Role Bypass** in all middleware instances to avoid per-request permission checks.
_Avoid_: Root, admin, global admin

**IAM Admin Role**:
The **System Role** with code `iam_admin`. Receives an explicit subset of `iam.*` permissions. Does not receive permissions from other modules (library, community, guide, notification).
_Avoid_: Identity admin, user admin

## Relationships (permissions)

- A **Permission Code** belongs to exactly one **Module Permission Namespace**.
- A **Seed Permission** is defined by the module that owns its namespace and is provided via `group:"permission_seeds"`.
- A **Global Role** aggregates **Permission Codes** through **Role-Permission Assignments**.
- A **System Role** is seeded at startup; a **Custom Role** is created at runtime.
- The **Super Admin Role** automatically receives all **Seed Permissions** from all modules; the **IAM Admin Role** receives only `iam.*` permissions.
- A **Role Bypass** is a middleware-level shortcut that avoids checking individual **Permission Codes** by verifying the account's role against an allow-list.

## Example dialogue

> **Dev:** "Should I add `community.read` as a permission on the community admin list endpoint?"
> **Domain expert:** "Yes — every module's admin endpoints should be gated by its own namespace. That way a `community_admin` role can read communities without also getting IAM read access."

## Flagged ambiguities

- "permission" was initially used to mean both the code string and the database entity; resolved: **Permission Code** refers to the string constant, **Seed Permission** refers to the seed-time registration.
- "role" was initially considered module-scoped; resolved: roles are **Global Roles** — a single role can authorize actions across library, community, and IAM modules.
- "super admin bypass" was initially treated as specific to IAM routes; resolved: **Role Bypass** is a pattern used uniformly across all modules, passing `["super_admin"]` as the allowed bypass roles.
