# Ingestion Event Contract v1

Status: Active
Owner: core-backend (producer), ai-service (consumer)

## 1) Ownership and Boundary

- **core-backend** owns the ingestion request entrypoint and publishes
  `document.ingestion.requested.v1` events.
- **ai-service** owns consumption, orchestration, and publishes
  `document.ingestion.status.updated.v1` events.
- **core-backend** consumes status events to maintain a client-facing
  projection (PRD 3).
- This contract is the single source of truth for schema compatibility
  between producer and consumer.

## 2) Transport Binding

- **Protocol**: RabbitMQ topic exchange.
- **Routing key**: `event_type` dot-notation string (e.g.
  `document.ingestion.requested.v1`).
- **Content type**: `application/json` — envelope serialised as JSON-encoded
  proto3.
- **Delivery semantics**: at-least-once.  Consumers MUST be idempotent.

## 3) Envelope

Every ingestion event MUST carry an `IngestionEventEnvelope` wrapper.
Required fields:

| Field               | Type     | Required | Description                                    |
|---------------------|----------|----------|------------------------------------------------|
| event_id            | UUID     | yes      | Unique per event                               |
| event_type          | string   | yes      | Versioned dot-notation                         |
| schema_version      | string   | yes      | Envelope schema version, e.g. "1.0.0"          |
| occurred_at         | RFC3339  | yes      | When the event happened                        |
| producer            | string   | yes      | Service identity, e.g. "core-backend"          |
| key_id              | string   | yes      | HMAC key identifier                            |
| signature           | bytes    | yes      | HMAC-SHA256 of canonical payload               |
| idempotency_key     | string   | yes      | Stable deduplication key                       |
| account_id          | UUID     | yes      | Tenant identity                                |
| user_id             | UUID     | yes      | Initiating user identity                       |
| batch_id            | UUID     | no       | Optional grouping for batch uploads            |
| replay_count        | int32    | yes      | 0 on first delivery, increments on outbox retry |
| payload             | oneof    | yes      | Exactly one payload variant MUST be set         |

### Canonical Signing Order

The signature covers all envelope fields **except** `signature` itself.
Fields are signed in proto field-number order, concatenated as JSON values
with no whitespace, separated by `|`:

```
event_id|event_type|schema_version|occurred_at|producer|key_id|
idempotency_key|account_id|user_id|batch_id|replay_count|<payload_canonical>
```

Where `<payload_canonical>` is the JSON-serialised payload message in
field-number order, values separated by `|`.

The HMAC is computed using SHA-256 with the secret corresponding to `key_id`.

### Key Rotation

- Keys are identified by `key_id`.
- At any time, one key is **active** (used for signing new events).
- The **previous** key remains valid for verification until rotated out.
- Consumers MUST accept signatures from both active and previous keys.
- If `key_id` is unknown, the event MUST be rejected.

## 4) Payload Types

### document.ingestion.requested.v1

Pointer-only payload — no file bytes.

| Field               | Type     | Required | Description                                    |
|---------------------|----------|----------|------------------------------------------------|
| document_id         | UUID     | yes      | Ingestion job identifier                      |
| storage_key         | string   | yes      | Object storage key/path                        |
| content_type        | string   | yes      | MIME type (e.g. "application/pdf")             |
| size_bytes          | int64    | yes      | Verified object size in bytes                  |
| checksum_sha256     | string   | yes      | Hex SHA-256 digest                             |
| source_filename     | string   | no       | Original filename                              |
| declared_language   | string   | no       | ISO 639-1 language hint                        |
| payload_schema_version | string| yes      | Payload contract version                       |

### document.ingestion.status.updated.v1 (reserved)

Emitted by AI workers.  Fields are defined in the proto but will be
expanded incrementally in later PRDs.

### document.lifecycle.archived.v1 / removed.v1 (reserved)

Stubs for future lifecycle events.  Defined now for build-time type
safety but not yet produced or consumed.

## 5) Schema Evolution Rules

1. **Field numbers never reused.**  Once a tag is assigned, it is retired
   forever after deprecation.
2. **New fields are always additive** with new tag numbers.
3. **No field deletion** — use `deprecated = true` annotation.
4. **No semantic change** to existing fields without a `schema_version`
   bump on the envelope or `payload_schema_version` bump on the payload.
5. **Oneof variants are additive** — new payload types are added as new
   oneof options with new field numbers.
6. **buf breaking-change detection** must pass in CI before merge.

## 6) Event Name Registry

| Event Type                                      | Producer     | Consumer     | Proto Payload                  |
|-------------------------------------------------|--------------|--------------|--------------------------------|
| document.ingestion.requested.v1                  | core-backend | ai-service   | DocumentIngestionRequested     |
| document.ingestion.status.updated.v1            | ai-service   | core-backend | DocumentIngestionStatusUpdated |
| document.lifecycle.archived.v1                   | core-backend | (reserved)   | DocumentLifecycleArchived      |
| document.lifecycle.removed.v1                   | core-backend | (reserved)   | DocumentLifecycleRemoved       |

## 7) Security

- Events are HMAC-SHA256 signed.  Signature covers all fields except the
  `signature` field itself, using canonical ordering described in §3.
- Payload contains **pointer metadata only** — never file bytes.
- No raw document content or text snippets in events or logs.
- AI workers fetch content through short-lived signed URLs (15-minute
  expiry) via internal core gRPC (not part of this commit).

## 8) Verification

- Contract tests MUST prove envelope required fields are present.
- Signature tests MUST prove active/previous key verification works.
- Schema compatibility tests MUST run in CI via buf breaking-change
  detection.
- Integration tests MUST cover: finalise → outbox → publish → consume.