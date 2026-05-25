# Localization Implementation Plan

## Overview

Full localization coverage across all modules: all user-facing error and success messages resolved through the i18n system with locale-aware translation files (English + Amharic).

## Glossary of key naming

```
module.errors.descriptiveKey      → error messages
module.successes.descriptiveKey   → success messages
```

Example: `notification.errors.channelRequired`, `guide.successes.stepCompleted`

The shared `errors.*` namespace is reserved for cross-cutting infrastructure errors: `databaseError`, `cacheError`, `networkError`, `unknownError`.

---

## Phase 1 — Foundation

**Goal**: Establish the infrastructure that all subsequent phases depend on. No behavioral change yet.

### 1.1 — `i18n.T()` helper

**File**: `pkg/i18n/context.go`

Add a convenience function that wraps the common locale-from-context → resolve pattern:

```go
func T(ctx context.Context, key string, params ...map[string]string) string {
    return Resolve(key, LocaleFromContext(ctx), params...)
}
```

This becomes the canonical way to resolve any i18n key from a handler context.

### 1.2 — Huma locale middleware

**File**: `pkg/middleware/locale.go` (new)

Write a Huma middleware that:
1. Reads `Accept-Language` header from `ctx.Header("Accept-Language")`
2. Normalizes: split on `-`, lowercase, match `"en"` → `constants.LocaleEnglish`, `"am"` → `constants.LocaleAmharic`, default `constants.LocaleEnglish`
3. Wires into Go context via `ctx = huma.WithContext(ctx, i18n.WithLocale(ctx.Context(), locale))`
4. Calls `next(ctx)`

The normalisation logic should match the existing `NormalizeLocale` helper in `internal/shared/middleware/locale.go` (which will be removed in Phase 6).

### 1.3 — Register middleware globally

**File**: `internal/core/api.go`

Add `api.UseMiddleware(middleware.LocaleResolver(...))` in `NewHumaAPI()` so every Huma route receives locale context.

**Files to create**: 1 (`pkg/middleware/locale.go`)
**Files to modify**: 2 (`pkg/i18n/context.go`, `internal/core/api.go`)

---

## Phase 2 — Translation Content

**Goal**: Add all new i18n keys to both translation files. No code changes in module handlers yet.

### 2.1 — Fix existing am.json inconsistencies

| Issue | Fix |
|---|---|
| `am.json:125` truncated `networkError`: `"የ�arm"` | Replace with correct Amharic: `"የኔትወርክ ስህተት ተከስቷል"` |
| `am.json` has `otpResendTooSoon` (key missing from `en.json`) | Remove from `am.json` or add to `en.json` — confirm with domain expert |
| `am.json` uses `uploadFileFailed` but `en.json` uses `uploadFailed` | Align to `uploadFailed` in both files |
| `am.json` missing keys present in `en.json` | Add missing translations for: `accountStatusPendingVerification`, `accountStatusLocked`, `accountStatusSuspended`, `accountStatusDisabled`, `emailAlreadyExists`, `usernameAlreadyExists`, `otpResendLimitExceeded`, `updateFailed` |

### 2.2 — Notification error keys (new)

Add to `notification.errors.*` in both `en.json` and `am.json`:

| Key | English value |
|---|---|
| `notification.errors.channelRequired` | "At least one notification channel is required" |
| `notification.errors.titleAndBodyRequired` | "Title and body are required" |
| `notification.errors.scheduledTimeMustBeFuture` | "Scheduled time must be in the future" |
| `notification.errors.cancelOnlyPending` | "Only pending scheduled alerts can be cancelled" |
| `notification.errors.rescheduleOnlyPending` | "Only pending scheduled alerts can be rescheduled" |
| `notification.errors.rescheduleTimeMustBeFuture` | "Rescheduled time must be in the future" |
| `notification.errors.complianceExpiryMustBeFuture` | "Expiry date must be in the future" |

(Note: `maxScheduledAlertsReached` already exists — no new key needed.)

### 2.3 — Payment error keys (new namespace)

Move from `errors.*` to `payment.errors.*`. Add to both translation files:

| Key | English value |
|---|---|
| `payment.errors.planNotFound` | "Plan not found" |
| `payment.errors.paymentNotFound` | "Payment not found" |
| `payment.errors.alreadyPaid` | "Payment already completed" |
| `payment.errors.invalidSignature` | "Invalid payment signature" |
| `payment.errors.invalidWebhookPayload` | "Invalid payment webhook payload" |
| `payment.errors.paymentInitFailed` | "Payment initialisation failed" |
| `payment.errors.paymentVerifyFailed` | "Payment verification failed" |
| `payment.errors.transactionFailed` | "Transaction failed" |

### 2.4 — Notification success keys (new)

Add to `notification.successes.*`:

| Key | English value |
|---|---|
| `notification.successes.scheduledAlertCreated` | "Scheduled alert created" |
| `notification.successes.scheduledAlertCancelled` | "Scheduled alert cancelled" |
| `notification.successes.scheduledAlertRescheduled` | "Scheduled alert rescheduled" |
| `notification.successes.complianceEntryCreated` | "Compliance entry created" |
| `notification.successes.complianceEntryUpdated` | "Compliance entry updated" |
| `notification.successes.complianceEntryDeleted` | "Compliance entry deleted" |
| `notification.successes.markedAsRead` | "Marked as read" |
| `notification.successes.markedAsUnread` | "Marked as unread" |
| `notification.successes.markedAllAsRead` | "Marked all as read" |
| `notification.successes.deleted` | "Deleted" |
| `notification.successes.muteAdded` | "Mute added" |
| `notification.successes.muteRemoved` | "Mute removed" |
| `notification.successes.preferencesUpdated` | "Preferences updated" |
| `notification.successes.deliveryChannelUpdated` | "Delivery channel updated" |

### 2.5 — IAM admin success keys (new)

Add to `iam.successes.*`:

| Key | English value |
|---|---|
| `iam.successes.adminAccountCreated` | "Admin account created" |
| `iam.successes.adminRolesUpdated` | "Admin roles updated" |
| `iam.successes.adminStatusUpdated` | "Admin status updated" |
| `iam.successes.passwordResetLinkSent` | "Password reset link sent" |
| `iam.successes.passwordResetSuccessfully` | "Password reset successfully" |
| `iam.successes.roleDeleted` | "Role deleted" |

### 2.6 — Guide success keys (add missing, keep existing)

Existing keys (already in `en.json`, never wired): `stepCompleted`, `stepStarted`, `bookmarkAdded`, `bookmarkRemoved`, `bookmarkUpdated`.

New keys to add:

| Key | English value |
|---|---|
| `guide.successes.stepMarkedIncomplete` | "Step marked as incomplete" |
| `guide.successes.stepSkipped` | "Step skipped" |
| `guide.successes.progressUpdated` | "Progress updated" |
| `guide.successes.guideUpdated` | "Guide updated" |
| `guide.successes.guideDeleted` | "Guide deleted" |
| `guide.successes.conditionAdded` | "Condition added" |
| `guide.successes.conditionRemoved` | "Condition removed" |
| `guide.successes.translationsUpdated` | "Translations updated" |
| `guide.successes.stepUpdated` | "Step updated" |
| `guide.successes.stepDeleted` | "Step deleted" |
| `guide.successes.stepsReordered` | "Steps reordered" |
| `guide.successes.dependencyAdded` | "Dependency added" |
| `guide.successes.dependencyRemoved` | "Dependency removed" |
| `guide.successes.stepReverted` | "Step reverted to version" |
| `guide.successes.journeyInvalidated` | "User journey invalidated" |
| `guide.successes.allJourneysInvalidated` | "All journeys invalidated" |

### 2.7 — Library success keys (new)

Add to `library.successes.*`:

| Key | English value |
|---|---|
| `library.successes.categoryDeleted` | "Category deleted" |
| `library.successes.translationUpdated` | "Translation updated" |
| `library.successes.translationDeleted` | "Translation deleted" |
| `library.successes.templateGroupDeleted` | "Template group deleted" |
| `library.successes.templateDeleted` | "Template deleted" |
| `library.successes.interactiveFormDeleted` | "Interactive form deleted" |

### 2.8 — Community success keys (new)

Add to `community.successes.*`:

| Key | English value |
|---|---|
| `community.successes.threadUpdated` | "Thread updated" |
| `community.successes.threadDeleted` | "Thread deleted" |
| `community.successes.postUpdated` | "Post updated" |
| `community.successes.postDeleted` | "Post deleted" |
| `community.successes.solutionMarked` | "Solution marked" |
| `community.successes.threadFollowed` | "Thread followed" |
| `community.successes.threadUnfollowed` | "Thread unfollowed" |
| `community.successes.threadMarkedAsRead` | "Thread marked as read" |
| `community.successes.categoryFollowed` | "Category followed" |
| `community.successes.categoryUnfollowed` | "Category unfollowed" |
| `community.successes.userBlocked` | "User blocked" |
| `community.successes.userUnblocked` | "User unblocked" |
| `community.successes.attachmentDeleted` | "Attachment deleted" |
| `community.successes.reportResolved` | "Report resolved" |

**Files to modify**: 2 (`pkg/i18n/messages/en.json`, `pkg/i18n/messages/am.json`)

---

## Phase 3 — Notification Module Errors

**Goal**: Replace 7 hardcoded `errors.New(...)` in usecases with `apperrors.XxxError()` using i18n keys, and fix the 1 handler-level i18n bypass.

### 3.1 — Usecase error conversion

**File**: `internal/modules/notification/application/usecase/user_scheduled_notification_usecase.go`

Replace `errors.New(...)` calls:

| Line | Current code | Replacement |
|---|---|---|
| 38 | `errors.New("notification: at least one channel is required")` | `apperrors.BadRequestError("notification.errors.channelRequired")` |
| 48 | `errors.New("notification: title and body are required")` | `apperrors.BadRequestError("notification.errors.titleAndBodyRequired")` |
| 52 | `errors.New("notification: scheduled time must be in the future")` | `apperrors.BadRequestError("notification.errors.scheduledTimeMustBeFuture")` |
| 115 | `errors.New("notification: can only cancel pending scheduled alerts")` | `apperrors.BadRequestError("notification.errors.cancelOnlyPending")` |
| 129 | `errors.New("notification: can only reschedule pending scheduled alerts")` | `apperrors.BadRequestError("notification.errors.rescheduleOnlyPending")` |
| 132 | `errors.New("notification: rescheduled time must be in the future")` | `apperrors.BadRequestError("notification.errors.rescheduleTimeMustBeFuture")` |

**File**: `internal/modules/notification/application/usecase/compliance_entry_usecase.go`

| Line | Current code | Replacement |
|---|---|---|
| 35 | `errors.New("notification: expiry date must be in the future")` | `apperrors.BadRequestError("notification.errors.complianceExpiryMustBeFuture")` |

### 3.2 — Handler error fix

**File**: `internal/modules/notification/delivery/handler/scheduled_alert_handler.go`

| Line | Current code | Replacement |
|---|---|---|
| 39 | `return nil, huma.NewError(403, "Upgrade to Pro...")` | `return nil, apperrors.ForbiddenError("notification.errors.maxScheduledAlertsReached")` |

**Note**: The `errors.Is(err, notiferror.ErrMaxScheduledAlertsReached)` check stays as the conditional. The response changes from `huma.NewError` to `apperrors.ForbiddenError`, which then flows through the existing `return nil, apperrors.ToHumaError(ctx, err)` path — but this requires restructuring: either return the `apperrors.ForbiddenError` directly, or handle it inside `ToHumaError`.

**Recommended approach**: Return the `apperrors.ForbiddenError` directly after the `if errors.Is(...)` block, letting the existing `ToHumaError` call at the end of the handler catch everything else.

**Files to modify**: 3

---

## Phase 4 — Payment + IAM Admin

### 4.1 — Payment key rename

**File**: `internal/modules/payment/application/usecase/payment_usecase.go`

Rename all `"errors.Xxx"` keys under `payment.*`:

| Key references in code (7 total) | New key |
|---|---|
| `"errors.planNotFound"` | `"payment.errors.planNotFound"` |
| `"errors.paymentNotFound"` (generated by `NotFoundError("payment", ...)`) | `"payment.errors.paymentNotFound"` (update `NotFoundError` call from `"payment"` to `"payment.payment"`) |
| `"errors.alreadyPaid"` | `"payment.errors.alreadyPaid"` |
| `"errors.invalidSignature"` | `"payment.errors.invalidSignature"` |
| `"errors.invalidWebhookPayload"` | `"payment.errors.invalidWebhookPayload"` |
| `"errors.paymentInitFailed"` | `"payment.errors.paymentInitFailed"` |
| `"errors.paymentVerifyFailed"` | `"payment.errors.paymentVerifyFailed"` |
| `"errors.transactionFailed"` | `"payment.errors.transactionFailed"` |

**Note**: `NotFoundError("payment", txRef)` at line 173 generates key `errors.paymentNotFound`. To generate `payment.errors.paymentNotFound`, use `NotFoundErrorWithKey("payment.errors.paymentNotFound")` instead.

### 4.2 — IAM admin success messages

**File**: `internal/modules/iam/delivery/handler/admin_handler.go`

Replace hardcoded strings with `i18n.T(ctx, "iam.successes.key")` at these lines:

| Line | Current value | Replacement |
|---|---|---|
| 44 | `"Admin account created"` | `i18n.T(ctx, "iam.successes.adminAccountCreated")` |
| 67 | `"Admin roles updated"` | `i18n.T(ctx, "iam.successes.adminRolesUpdated")` |
| 145 | `"Admin status updated"` | `i18n.T(ctx, "iam.successes.adminStatusUpdated")` |
| 167 | `"Password reset link sent"` | `i18n.T(ctx, "iam.successes.passwordResetLinkSent")` |
| 181 | `"Password reset successfully"` | `i18n.T(ctx, "iam.successes.passwordResetSuccessfully")` |

**File**: `internal/modules/iam/delivery/handler/role_handler.go`

| Line | Current value | Replacement |
|---|---|---|
| 146 | `"Role deleted"` | `i18n.T(ctx, "iam.successes.roleDeleted")` |

**Files to modify**: 3

---

## Phase 5 — All Remaining Module Success Messages

### 5.1 — Guide handler wiring

**Files**: 
- `internal/modules/guide/delivery/handler/guide_view_handler.go` (8 occurrences)
- `internal/modules/guide/delivery/handler/guide_admin_handler.go` (16 occurrences)

Replace each hardcoded `Message: "..."` with `Message: i18n.T(ctx, "guide.successes.{key}")`.

Key mapping:

| Hardcoded string | i18n key |
|---|---|
| "Step started" | `guide.successes.stepStarted` (exists) |
| "Step completed" | `guide.successes.stepCompleted` (exists) |
| "Step marked as incomplete" | `guide.successes.stepMarkedIncomplete` |
| "Step skipped" | `guide.successes.stepSkipped` |
| "Progress updated" | `guide.successes.progressUpdated` |
| "Bookmark added" | `guide.successes.bookmarkAdded` (exists) |
| "Bookmark updated" | `guide.successes.bookmarkUpdated` (exists) |
| "Bookmark removed" | `guide.successes.bookmarkRemoved` (exists) |
| "Guide updated" | `guide.successes.guideUpdated` |
| "Guide deleted" | `guide.successes.guideDeleted` |
| "Condition added" | `guide.successes.conditionAdded` (appears twice) |
| "Condition removed" | `guide.successes.conditionRemoved` (appears twice) |
| "Translations updated" | `guide.successes.translationsUpdated` (appears twice) |
| "Step updated" | `guide.successes.stepUpdated` |
| "Step deleted" | `guide.successes.stepDeleted` |
| "Steps reordered" | `guide.successes.stepsReordered` |
| "Dependency added" | `guide.successes.dependencyAdded` |
| "Dependency removed" | `guide.successes.dependencyRemoved` |
| "Step reverted to version" | `guide.successes.stepReverted` |
| "User journey invalidated" | `guide.successes.journeyInvalidated` |
| "All journeys invalidated" | `guide.successes.allJourneysInvalidated` |

### 5.2 — Community handler wiring

**File**: 
- `internal/modules/community/delivery/handler/community_handler.go` (13 occurrences)
- `internal/modules/community/delivery/handler/community_admin_handler.go` (7 occurrences)

Replace each hardcoded `Message: "..."` with `Message: i18n.T(ctx, "community.successes.{key}")`.

Key mapping:

| Hardcoded string | i18n key |
|---|---|
| "Thread updated" | `community.successes.threadUpdated` |
| "Thread deleted" | `community.successes.threadDeleted` |
| "Post updated" | `community.successes.postUpdated` |
| "Post deleted" | `community.successes.postDeleted` |
| "Solution marked" | `community.successes.solutionMarked` |
| "Thread followed" | `community.successes.threadFollowed` |
| "Thread unfollowed" | `community.successes.threadUnfollowed` |
| "Thread marked as read" | `community.successes.threadMarkedAsRead` |
| "Category followed" | `community.successes.categoryFollowed` |
| "Category unfollowed" | `community.successes.categoryUnfollowed` |
| "User blocked" | `community.successes.userBlocked` (appears in both handlers) |
| "User unblocked" | `community.successes.userUnblocked` (appears in both handlers) |
| "Attachment deleted" | `community.successes.attachmentDeleted` |
| "Category updated" | `community.successes.categoryUpdated` |
| "Category deleted" | `community.successes.categoryDeleted` |
| "Thread deleted and report resolved" | `community.successes.reportThreadDeleted` |
| "Post deleted and report resolved" | `community.successes.reportPostDeleted` |
| "User blocked and report resolved" | `community.successes.reportUserBlocked` |

### 5.3 — Library handler wiring

**File**: `internal/modules/library/delivery/handler/library_admin_handler.go` (6 occurrences)

Replace each hardcoded `Message: "..."` with `Message: i18n.T(ctx, "library.successes.{key}")`.

Key mapping:

| Hardcoded string | i18n key |
|---|---|
| "Category deleted" | `library.successes.categoryDeleted` |
| "Translation updated" | `library.successes.translationUpdated` |
| "Translation deleted" | `library.successes.translationDeleted` |
| "Template group deleted" | `library.successes.templateGroupDeleted` |
| "Template deleted" | `library.successes.templateDeleted` |
| "Interactive form deleted" | `library.successes.interactiveFormDeleted` |

### 5.4 — Notification success message wiring

**File**: 
- `internal/modules/notification/delivery/handler/scheduled_alert_handler.go` (3 occurrences)
- `internal/modules/notification/delivery/handler/compliance_handler.go` (3 occurrences)
- `internal/modules/notification/delivery/handler/notification_handler.go` (8 occurrences)

Replace hardcoded strings with `i18n.T(ctx, "notification.successes.{key}")`.

Key mapping matches section 2.4.

**Files to modify**: ~9 handler files across 4 modules

---

## Phase 6 — Cleanup

### 6.1 — Remove Gin LocaleResolver

**File**: `internal/shared/middleware/locale.go` (to delete)

**File**: `internal/core/server.go`

Remove `middleware.LocaleResolver()` from the Gin middleware stack in `NewGinEngine()`.

### 6.2 — Remove dead response package

**File**: `pkg/response/response.go` (to delete)

The entire `pkg/response/` directory. No code references it.

### 6.3 — Remove unused constants

**File**: `internal/shared/constants/locale.go`

Check if `Locale` type and `LocaleEnglish`/`LocaleAmharic` constants are still used after removing the Gin middleware. If they're only used by the Gin middleware, remove them. If referenced elsewhere (e.g., by the new Huma middleware), keep them.

### 6.4 — Update error_handler.go getLocale()

**File**: `pkg/middleware/error_handler.go`

Replace `getLocale()` which reads `Accept-Language` header directly. Since this is a Gin middleware for panic recovery:
- Option A: Replace `getLocale(c)` with `i18n.LocaleFromContext(c.Request.Context())` — works if any upstream middleware sets locale on context
- Option B: Keep reading `Accept-Language` as is (simpler, adequate for panic recovery edge case)

**Recommended**: Keep `Accept-Language` reading in the error handler. It's a safety net for unhandled panics and doesn't warrant refactoring.

**Files to delete**: 2 (`internal/shared/middleware/locale.go`, `pkg/response/response.go`)
**Files to modify**: 2 (`internal/core/server.go`, `internal/shared/constants/locale.go` — maybe)

---

## Summary of files changed

| Phase | Create | Modify | Delete |
|---|---|---|---|
| 1 | `pkg/middleware/locale.go` | `pkg/i18n/context.go`, `internal/core/api.go` | — |
| 2 | — | `pkg/i18n/messages/en.json`, `pkg/i18n/messages/am.json` | — |
| 3 | — | 3 notification files | — |
| 4 | — | 3 files (payment + IAM) | — |
| 5 | — | ~9 handler files | — |
| 6 | — | 1 (`server.go`) | 2 (`locale.go`, `response.go`) |

## Total new i18n keys

| Module | Error keys (new) | Success keys (new) | Key renames |
|---|---|---|---|
| notification | 7 | 14 | 0 |
| payment | 0 | 0 | 7 |
| iam | 0 | 6 | 0 |
| guide | 0 | 12 (+5 existing) | 0 |
| library | 0 | 6 | 0 |
| community | 0 | 14 | 0 |
| **Total** | **7** | **52** | **7** |
