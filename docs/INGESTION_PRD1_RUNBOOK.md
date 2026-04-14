# Ingestion PRD1 Runbook

This runbook covers operational procedures introduced by PRD1 ingestion work.

## 1) Required Environment Variables

Core service:

- `INGESTION_ENABLED` (default `true`)
- `INGESTION_SIGNING_ACTIVE_KEY_ID`
- `INGESTION_SIGNING_ACTIVE_KEY_SECRET`
- `INGESTION_DISPATCHER_BATCH_SIZE`
- `INGESTION_DISPATCHER_INTERVAL`
- `INGESTION_DISPATCHER_RETRY_BASE_DELAY`
- `INGESTION_DISPATCHER_RETRY_MAX_DELAY`
- `INGESTION_DISPATCHER_MAX_ATTEMPTS_BEFORE_DLQ`

AI service:

- `INGESTION_SIGNING_ACTIVE_KEY_ID`
- `INGESTION_SIGNING_ACTIVE_KEY_SECRET`
- `INGESTION_SIGNING_PREVIOUS_KEYS_JSON`

`INGESTION_SIGNING_PREVIOUS_KEYS_JSON` format:

```json
{"ingestion-v0":"previous-secret"}
```

## 2) Rollout Sequence

1. Deploy both services with `INGESTION_ENABLED=false` in core.
2. Verify core starts and outbox dispatcher is idle.
3. Verify AI service has verifier keys configured.
4. Enable ingestion by setting `INGESTION_ENABLED=true`.
5. Run a smoke flow:
   - Create upload intent
   - Upload object
   - Finalize upload
   - Confirm outbox row is published

## 3) Key Rotation Procedure

1. Generate new key pair:
   - new `key_id`
   - new secret
2. On AI service:
   - keep current key as active for now
   - add new key to `INGESTION_SIGNING_PREVIOUS_KEYS_JSON`
3. On core service:
   - switch `INGESTION_SIGNING_ACTIVE_KEY_ID` and `INGESTION_SIGNING_ACTIVE_KEY_SECRET` to new values
4. Verify new events validate in AI.
5. After stability window, remove old key from AI previous key map.

## 4) Ingestion Pause / Resume

Pause new ingestion requests:

- Set `INGESTION_ENABLED=false` in core and redeploy.
- Existing outbox rows continue dispatcher behavior unless app-level handler logic gates finalize entrypoints.

Resume:

- Set `INGESTION_ENABLED=true` and redeploy.

## 5) Outbox Retry and Dead-Letter Behavior

- Failed publish attempts are retried with exponential backoff.
- Once attempts exceed `INGESTION_DISPATCHER_MAX_ATTEMPTS_BEFORE_DLQ`, row status moves to `dead`.

Recommended checks:

- Track count of `pending`, `published`, `dead` rows.
- Alert on growth of `dead` rows.
- Alert when `pending` backlog continuously increases.

## 6) Rollback

If ingestion rollout causes issues:

1. Set `INGESTION_ENABLED=false`.
2. Redeploy core.
3. Keep outbox rows intact for later replay.
4. Fix issue and re-enable ingestion.

## 7) Security Rules

- Never log ingestion signing secrets.
- Do not include raw file bytes in events.
- Keep secrets rotated periodically.
- Ensure AI verifier has active and previous key support during rotation windows.
