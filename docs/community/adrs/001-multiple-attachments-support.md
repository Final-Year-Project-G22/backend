# ADR 001: Multiple Attachments Support in Community System

## Status
Accepted

## Date
2026-05-02

## Context
The existing community system supports only a single attachment per post/thread/reply via `attachment_url` (text) and `attachment_type` (varchar(50)) columns on `discussion_posts`. The user requested support for:
- Multiple attachments per post/thread/reply
- Both images and general files together
- A pre-upload pattern where attachments are stored independently first, then mapped to posts

The system is in development stage (no production data), so backward compatibility with existing production data is not required. Existing project patterns include extensive JSONB usage, `time.NewTicker` for background jobs, `model.BaseModel` embedding for entities, and custom string enums in `enums.go`.

All decisions below were validated via a 22-question grill session where the recommended option was accepted for each question.

## Decision
We will implement multiple attachment support using a dedicated `attachments` table and pre-upload pattern. The following specific decisions were made:

### Data Storage (Revised Q1, Q11)
**Initial recommendation**: Use JSONB column on `discussion_posts` to store attachment array (aligned with existing project patterns like `uploaded_documents` in `user_guide_progresses`).
**Revised decision**: Create a dedicated `attachments` table with a nullable `post_id` FK to `discussion_posts`. This revision was necessary to support the pre-upload pattern (attachments must exist independently before being linked to a post).

### Attachment Structure (Q2, Q3, Q17)
Each attachment will include:
- `id` (UUID): Unique identifier (from `model.BaseModel` embedding)
- `storage_key` (text, not null): Internal key for storage operations (delete)
- `file_url` (text, not null): Public URL for client access
- `file_type` (varchar(50), not null): MIME type
- `file_name` (varchar(255), not null): Original upload filename
- `file_size` (bigint): File size in bytes
- `post_id` (uuid, nullable): FK to `discussion_posts`, null for pending pre-uploads
- `uploaded_by` (uuid, not null): FK to `accounts` for ownership enforcement
- `status` (varchar(20), not null, default 'pending'): Enum with values `pending`/`linked`

### Limits & Validation (Q4, Q7, Q21)
- Max 10 attachments per post
- Allowed MIME types: `image/jpeg`, `image/png`, `image/gif`, `image/webp`, `application/pdf` (existing allowed types)
- Per-file max size: 10MB (enforced at pre-upload endpoint)
- Total post attachment size: 50MB max (enforced at post create/update endpoint)

### Pre-Upload Pattern (Q9, Q10, Q12, Q16)
- Clients upload attachments to a dedicated endpoint first (`POST /api/v1/community/attachments`)
- Pre-upload endpoint returns full attachment objects (id, fileUrl, fileType, fileName, fileSize)
- Attachments are stored with `post_id = NULL` and `status = 'pending'`
- `attachments` table includes `uploaded_by` FK to `accounts` to enforce ownership
- When linking attachments to posts (create/update), validate:
  1. Attachment ID exists
  2. Status is `pending`
  3. `uploaded_by` matches the authenticated user's account ID

### Update Flow (Q5, Q14, Q15, Q22)
- Update operations support:
  1. Adding new attachments via `AttachmentIds []uuid.UUID`
  2. Removing specific attachments via `RemoveAttachmentIds []uuid.UUID`
  3. Removing all attachments via `RemoveAllAttachments bool`
- Removed attachments are permanently deleted (DB row + storage file)
- Old `attachment_url` and `attachment_type` columns on `discussion_posts` are dropped immediately (dev stage, no production data)

### Entity & Enum Design (Q18, Q19)
- `Attachment` entity embeds `model.BaseModel` (follows project convention per `DiscussionPost`)
- `AttachmentStatus` string type added to `enums.go` with constants `AttachmentStatusPending = "pending"` and `AttachmentStatusLinked = "linked"` (follows project enum pattern)
- Usecase input structs updated to include `AttachmentIds`, `RemoveAttachmentIds` fields

### Lifecycle Management (Q6, Q13, Q20)
- Soft-deleted posts retain linked attachments (GORM filters soft-deleted posts by default, so attachments are not fetched)
- Orphaned pending attachments (uploaded but never linked) are cleaned up via a background job using the existing `time.NewTicker` pattern (matches `delivery_worker.go`, `state_manager.go`)
- Cleanup job runs every 15 minutes, deletes attachments where `status='pending'` and `created_at < now() - 1 hour`

## Consequences

### Positive
- Pre-upload pattern decouples file upload from post creation, simplifying request handling
- Dedicated table allows independent attachment management and ownership enforcement
- Follows all existing project patterns (BaseModel, enums, ticker jobs)
- No production data migration needed (dev stage)
- Clear ownership and validation prevents cross-user attachment abuse

### Negative
- Additional database table join when fetching posts with attachments (resolved by proper indexing)
- Background cleanup job adds minor operational complexity
- Pre-upload requires clients to make two requests (upload + create post) instead of one multipart request

### Dependencies
- Relies on existing `storage.Storage` interface (SeaweedFS/MinIO) for file operations
- Relies on existing `time.NewTicker` pattern for background cleanup
- Requires GORM with soft delete support (already configured)

## References
- Interview grill session with 22 questions, all recommendations accepted
- Project enum pattern: `internal/modules/community/domain/entity/enums.go`
- Project background job pattern: `internal/modules/notification/application/service/delivery_worker.go`
- Existing JSONB usage examples: `user_guide_progresses.uploaded_documents`
