# PRD: Multiple Attachments Support in Community System

## Problem Statement
The current community system only supports a single attachment per post, thread, or reply via two columns (`attachment_url` and `attachment_type`) on the `discussion_posts` table. Users cannot upload multiple files to a single post, cannot mix images and documents in one post, and must upload files inline with post content, which complicates request handling. There is no way to add or remove specific attachments when updating a post, and no support for pre-uploading files before finalizing post content.

## Solution
Implement multiple attachment support using a dedicated `attachments` table and a pre-upload pattern. Clients will upload attachments to a dedicated endpoint first, receive full attachment metadata, then link attachments to posts/threads/replies using attachment IDs. The solution includes proper limits (count, size, file types), ownership enforcement, lifecycle management for orphaned files, and follows all existing project patterns (BaseModel embedding, custom enums, time.NewTicker background jobs).

## User Stories

1. As a community member, I want to pre-upload multiple attachments (up to 10) before creating a thread, so that I can prepare my files before composing the thread content.
2. As a community member, I want to pre-upload multiple attachments (up to 10) before creating a post in a thread, so that I can attach all relevant files upfront.
3. As a community member, I want to pre-upload multiple attachments (up to 10) before replying to a post, so that my reply includes all necessary files.
4. As a community member, I want to upload attachments of types image/jpeg, image/png, image/gif, image/webp, and application/pdf, so that I can share both images and documents.
5. As a community member, I want each attachment to have a unique identifier, so that I can reference specific attachments when updating my posts.
6. As a community member, I want to link pre-uploaded attachments to a new thread's initial post, so that the thread has all necessary files from creation.
7. As a community member, I want to link pre-uploaded attachments to a new post, so that the post includes all relevant files.
8. As a community member, I want to link pre-uploaded attachments to a reply post, so that the reply includes all relevant files.
9. As a community member, I want to add new pre-uploaded attachments to an existing post, so that I can include additional files when updating my post.
10. As a community member, I want to remove specific attachments from my post using their unique IDs, so that I can delete only the files I no longer want.
11. As a community member, I want to remove all attachments from my post at once, so that I can quickly clear all files when restructuring my post content.
12. As a community member, I want to see all attachments (with URLs, types, filenames, and sizes) when viewing a thread or post, so that I can access shared files easily.
13. As a community member, I want each attachment upload to be limited to 10MB, so that I don't accidentally upload excessively large files.
14. As a community member, I want the total size of attachments per post to be limited to 50MB, so that posts don't become too large to load.
15. As a community member, I want to only link attachments that I uploaded myself, so that I cannot use other users' files without permission.
16. As a community member, I want to only link attachments that haven't been used in other posts, so that each attachment is unique to one post.
17. As a community member, I want the pre-upload endpoint to return full attachment metadata (ID, URL, type, filename, size), so that I can preview files before creating a post.
18. As a community member, I want to delete a pending attachment I uploaded before linking it to a post, so that I can remove files I no longer want.
19. As a community member, I want soft-deleted posts to retain their attachments, so that if a post is accidentally deleted and restored, the files are still present.
20. As a system background worker, I want to periodically clean up pending attachments older than 1 hour, so that orphaned files don't accumulate and waste storage space.
21. As a developer, I want attachment entities to embed the project's BaseModel, so that the codebase follows consistent entity conventions.
22. As a developer, I want attachment status to follow the project's custom string enum pattern, so that the codebase is consistent across all modules.
23. As a community member, I want the old single-attachment fields to be removed from the database, so that the schema is clean and up-to-date.

## Implementation Decisions

### Data Storage
- Create a dedicated `attachments` table with a nullable `post_id` foreign key to `discussion_posts`, rather than using a JSONB column on `discussion_posts` (revised decision to support pre-upload pattern where attachments exist independently before linking).
- Attachment fields include: storage key (internal, for file deletion), public URL (client access), MIME type, original filename, file size, nullable post ID, uploader account ID (ownership), and status enum (pending/linked).

### Pre-Upload Pattern
- Clients upload attachments to a dedicated endpoint first, then link them to posts using attachment IDs.
- Pre-uploaded attachments are stored with `post_id = NULL` and `status = pending`.
- The `attachments` table includes an `uploaded_by` foreign key to `accounts` to enforce ownership.
- When linking attachments to posts, validate that the attachment exists, has status `pending`, and the uploader matches the authenticated user.

### Limits & Validation
- Maximum 10 attachments per post.
- Allowed MIME types match existing system: image/jpeg, image/png, image/gif, image/webp, application/pdf.
- Per-file maximum size: 10MB (enforced at pre-upload endpoint).
- Total post attachment size maximum: 50MB (enforced at post create/update endpoint).

### Update Flow
- Post updates support adding new attachments (via IDs), removing specific attachments (via IDs), and removing all attachments at once.
- Removed attachments are permanently deleted (database row and storage file).
- Old single-attachment columns (`attachment_url`, `attachment_type`) are dropped immediately from `discussion_posts` (development stage, no production data).

### Entity & Architectural Patterns
- `Attachment` entity embeds `model.BaseModel` to follow project convention.
- `AttachmentStatus` uses the project's custom string enum pattern (constants `pending` and `linked` in `enums.go`).
- Usecase input structs are updated to include `AttachmentIds`, `RemoveAttachmentIds` fields.
- Existing community service, handlers, and DTOs are updated to support the new flow.

### Lifecycle Management
- Soft-deleted posts retain linked attachments (GORM filters soft-deleted posts by default, so attachments are not fetched).
- Orphaned pending attachments are cleaned up via a background job using the existing `time.NewTicker` pattern.
- Cleanup job runs every 15 minutes, deletes attachments where `status=pending` and `created_at` is older than 1 hour.

### Module Changes
- New database migration to create `attachments` table and drop old columns.
- New `Attachment` entity in domain layer.
- New `AttachmentUsecase` interface and `AttachmentService` implementation.
- New pre-upload and orphan-delete handlers, updated existing community handlers.
- New `AttachmentCleanupWorker` background job.
- Updated DTOs to replace single attachment fields with attachment arrays.

## Testing Decisions

### What Makes a Good Test
- Focus on external behavior of modules, not implementation details.
- Use mock dependencies to isolate modules (following existing project test patterns).
- Test success and error cases (invalid inputs, ownership violations, limit enforcement).

### Modules to Test
1. **Attachment Usecase**: Test upload, link, unlink, delete orphan operations. Use mock storage and repository interfaces.
2. **Attachment Service**: Test validation logic (file size, type, ownership, pending status), storage integration, error handling.
3. **Attachment Cleanup Worker**: Test cleanup of old pending attachments, mock ticker and storage dependencies.

### Prior Art
- Follow existing test patterns in `internal/modules/ai/application/service/*_test.go`, which use mock structs for dependencies and the standard Go testing package.
- Example: `ask_service_test.go` uses `mockInferencePort` to isolate service behavior from external dependencies.

## Out of Scope
- Supporting attachment types beyond the existing allowed list (image/jpeg, image/png, image/gif, image/webp, application/pdf).
- Reusing attachments across multiple posts (one-attachment-to-one-post relationship is enforced).
- Client-side implementation of the pre-upload flow (only backend changes are in scope).
- Image processing (thumbnails, resizing) for attachments.
- Storage-level TTL configuration (relies on background cleanup job instead).
- Production data migration (system is in development stage).

## Further Notes
- All implementation decisions were validated via a 22-question grill session with the user, where all recommended options were accepted.
- The associated ADR is stored at `docs/adrs/001-multiple-attachments-support.md`.
- The system is in development stage, so no production data migration is required.
- The solution follows all existing project patterns: BaseModel embedding, custom string enums, time.NewTicker for background jobs, and GORM with soft delete support.
