# Core User gRPC Contract (Step 1 Spec)

Status: Draft for review (no implementation committed yet)

## 1) Ownership and boundary

- `core-backend` owns user identity, account roles, and preference data.
- `ai-service` consumes a narrow read-only profile through gRPC.
- This contract exposes only fields AI needs now: `user_id`, `account_id`, `tier`, `preferred_language`.

## 2) RPC surface (v1)

- Service: `CoreUserService`
- Method: `GetUserProfile`
- Package: `core.user.v1`
- RPC type: unary

Method signature:

```proto
rpc GetUserProfile(GetUserProfileRequest) returns (GetUserProfileResponse);
```

## 3) Message schema (v1)

Request:

```proto
message GetUserProfileRequest {
  string user_id = 1; // UUID string
}
```

Response (string values for flexibility in v1):

```proto
message GetUserProfileResponse {
  string user_id = 1;              // same UUID
  string account_id = 2;           // primary account UUID
  string tier = 3;                 // one of: basic | pro | premium
  string preferred_language = 4;   // one of: en | am
}
```

Why strings in v1:

- Avoid tight enum coupling while role/tier mapping can evolve in core.
- Keep AI mapping strict in code (reject unknown values).
- Keep wire contract minimal and easy to extend.

## 4) Behavioral contract

- `INVALID_ARGUMENT`:
  - `user_id` is missing/empty.
  - `user_id` is not a valid UUID.
- `NOT_FOUND`:
  - no core account/profile can be resolved for that user.
- `OK`:
  - response includes valid `user_id`, `account_id`, `tier`, and `preferred_language`.
- `INTERNAL`:
  - repository/db/internal failures.

## 5) Core mapping rules

Tier mapping (deterministic):

- default: `basic`
- if account has role code `iam_admin` -> `pro`
- if account has role code `super_admin` -> `premium`
- if both apply, highest wins (`premium`)

Preferred language mapping:

- from account preference language code on the resolved `account_id`
- `am` -> `am`
- `en` -> `en`
- missing or unknown -> `en` (safe fallback for v1)

## 6) AI consumer expectations

- AI adapter maps:
  - `account_id` -> `uuid.UUID` (for account-scoped follow-up use cases)
  - `tier` -> domain `Tier` enum (`basic/pro/premium`)
  - `preferred_language` -> domain `Language` enum (`en/am`)
- Unknown returned values are treated as provider/data contract errors.
- `NOT_FOUND` maps to `None` in AI adapter.

## 7) Versioning and compatibility

- Use package namespace `core.user.v1`.
- Backward-compatible evolution only:
  - additive fields with new tags only
  - no field/tag reuse
  - no semantic change to existing fields without version bump

## 8) Non-goals for this PR

- No quota or billing data in this RPC.
- No full user/account payload.
- No AI preference payload.
- No streaming APIs.

## 9) Acceptance criteria for Step 1

- Spec approved by review.
- Then proceed to Step 2:
  - implement proto file and generation config,
  - no commit before review checkpoint.
