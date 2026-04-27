# PRD: Template & Tools Library Module

## Problem Statement

MSMEs (Micro, Small, and Medium Enterprises) need practical business documents — business plans, invoices, record-keeping forms — but lack the expertise or resources to create them from scratch. Currently the platform offers no centralized, searchable library of downloadable business templates. Users have no way to find, preview, or download standardized business documents. Content admins have no CMS to upload and manage template files, organize them into categories with multi-language support, or gate advanced interactive tools behind the Pro tier.

## Solution

Build a Library module (`internal/modules/library/`) that provides:

- **Categorized template library**: Hierarchical categories with icon, sort order, and multi-language translations for the browsing UI
- **Template groups with language variants**: Each template concept (e.g., "Business Plan") is a group containing PDF/DOCX/XLSX files per language (English, Amharic, etc.), with a fallback default language
- **Tier-gated access**: Templates marked as `basic` (free) or `pro` (subscribers only) via a `TierService` interface (stub implementation in Phase 1)
- **File upload + validation**: Admin uploads PDF, DOCX, or XLSX files (max 10MB) via server-side upload, validated by content sniffing
- **Thumbnail preview**: Each template group can have a preview image uploaded to storage
- **Download tracking**: Presigned URL generation for file downloads, with per-group count and per-user download log (append-only)
- **Interactive forms** (Pro): Fillable digital forms with a structured JSONB layout (sections + fields) rendered by the frontend. Form definition stored on the backend; submission handled client-side in Phase 1.
- **Public templates**: `requiresAuth` flag allows certain templates to be downloadable without authentication
- **Admin CMS**: Full CRUD for categories, template groups, templates, interactive forms, and download log viewing

## User Stories

### Browsing & Discovery

1. As a user, I want to browse template categories in a hierarchy, so that I can find relevant templates by drilling down
2. As a user, I want to see localized category names and descriptions, so that I can browse in my preferred language
3. As a user, I want to search templates by title and description, so that I can quickly find specific documents
4. As a user, I want to filter templates by category and format (PDF, DOCX, XLSX), so that I can narrow results
5. As a user, I want to see template results sorted by admin-defined priority, so that important templates appear first
6. As a user, I want to see template details (description, format, file size, available languages, thumbnail), so that I know what I'm downloading
7. As a user, I want to see which language variants are available for a template group, so that I can choose my language
8. As a user, I want templates to fall back to the default language if my preferred variant doesn't exist, so that I can still access content
9. As a user, I want to see download counts for templates, so that I can find popular resources
10. As a user, I want to see a thumbnail preview of a template before downloading, so that I can visually identify it

### Downloading

11. As a user, I want to download a template file in my chosen language with one click, so that I can use it immediately
12. As a user, I want the download to start via a presigned URL redirect, so that the file downloads quickly without proxying through the backend
13. As a user, I want to access public templates without logging in, so that I can evaluate the platform
14. As a user, I want to be prompted to authenticate if I try to download a template that requires auth, so that I can proceed if I'm a member
15. As a user, I want my download to be blocked with a clear message if my subscription tier doesn't support it, so that I know to upgrade to Pro
16. As a user, I want to see my previous download history, so that I can re-download templates I've used

### Interactive Forms

17. As a Pro user, I want to access interactive fillable forms for selected templates, so that I can fill out business documents digitally
18. As a Pro user, I want to see a form rendered with sections and fields matching the template, so that I can complete it step-by-step
19. As a Pro user, I want the form to support various field types (text, select, date, textarea, etc.), so that complex documents can be handled
20. As a Pro user, I want to export my completed form data locally (frontend), so that I have a filled copy of the template

### Admin — Category Management

21. As an admin, I want to create hierarchical categories (parent + child), so that templates are organized intuitively
22. As an admin, I want to set a slug, icon name, and sort order per category, so that I control the navigation experience
23. As an admin, I want to reorder categories, so that important categories appear first
24. As an admin, I want to deactivate a category without deleting it, so that I can hide it temporarily
25. As an admin, I want to delete a category (soft-delete), so that I can remove obsolete groupings
26. As an admin, I want to add translations for categories, so that non-English users can browse in their language

### Admin — Template Group Management

27. As an admin, I want to create template groups with metadata (name, slug, description, category, format, tier, sort order, requiresAuth, default language), so that templates are findable
28. As an admin, I want to upload a thumbnail image for a template group, so that users see a visual preview
29. As an admin, I want to update template group metadata, so that I can correct mistakes or adjust settings
30. As an admin, I want to deactivate a template group, so that all language variants become unavailable at once
31. As an admin, I want to set tier access (Basic or Pro) per group, so that Pro features are gated appropriately

### Admin — Template File Management

32. As an admin, I want to upload a template file (PDF, DOCX, XLSX) for a specific language variant, so that users can download it
33. As an admin, I want uploaded files to be validated for type and size, so that only supported formats are available
34. As an admin, I want to replace a template's file and have the version number increment automatically, so that users know it's updated
35. As an admin, I want to update template metadata (title, description) independently from the file, so that I can fix content without replacing files
36. As an admin, I want to deactivate a specific language variant, so that I can hide a broken translation without affecting other languages

### Admin — Interactive Form Management

37. As an admin, I want to create an interactive form for an `interactive_form` template, so that Pro users have fillable documents
38. As an admin, I want to define the form layout as structured JSON (sections with typed fields), so that the frontend can render it
39. As an admin, I want to update the form layout and have the version increment, so that users are aware of updates
40. As an admin, I want to deactivate a form, so that it's no longer available to users

### Admin — Monitoring

41. As an admin, I want to view download logs (who downloaded what, when), so that I can track template usage
42. As an admin, I want to filter download logs by template group, so that I can analyze popularity per template
43. As an admin, I want to see total download counts per template group, so that I can identify the most popular resources

## Implementation Decisions

### Module Architecture

- **Module name**: `library` (avoids collision with IAM's existing `TemplatePreference` entity and clearly signals "library of resources")
- **Clean Architecture**: Domain → Application → Infrastructure → Delivery, following the guide/community module patterns
- **6 entities**: LibraryCategory, LibraryCategoryTranslation, LibraryTemplateGroup, LibraryTemplate, LibraryInteractiveForm, LibraryTemplateDownload
- **2 enums**: TemplateFormat (pdf, docx, xlsx, interactive_form), TierAccess (basic, pro)
- **Entity registration**: EntityProvider for SchemaManager/Atlas migration generation

### Category Design

- **Hierarchical**: Self-referential `ParentCategoryID` (nullable for root categories), matching guide/community patterns
- **Slug uniqueness**: Composite unique index on `(ParentCategoryID, Slug)` — unique within a parent
- **Translations**: `LibraryCategoryTranslation` table matching the `GuideCategoryTranslation` pattern (no BaseModel, separate PK + timestamps)
- **Icon**: Icon name string (maps to frontend icon library), not an uploaded image
- **Sort order**: Admin-controlled `SortOrder` integer on categories and template groups

### Template Group & Variant Model

- **LibraryTemplateGroup**: Shared metadata (slug, category, format, tier, requiresAuth, sortOrder, defaultLanguage, thumbnailURL, downloadCount, createdBy)
- **LibraryTemplate**: Per-language variant with title, description, fileKey, fileURL, fileSize, contentType, version, isActive
- **One row per language**: English PDF = one row, Amharic PDF = another row, same group
- **No separate translation table**: Since each LibraryTemplate row IS a language variant, the title and description are on the template itself
- **Language fallback**: Group.DefaultLanguage specifies which variant to show when user's preferred language is unavailable

### File Storage

- **Storage backend**: Existing `storage.Storage` (SeaweedFS/MinIO) via `pkg/storage`
- **Key convention**: `library/templates/{templateID}.{ext}` for template files, `library/thumbnails/{groupID}.{ext}` for thumbnails
- **Upload flow**: Server-side via `storage.Upload` (admin POSTs multipart form → backend validates → uploads to storage → stores key+URL on entity)
- **Download flow**: Presigned URL via `storage.GetPresignedURL(ctx, key, 5*time.Minute)` — client receives a temporary URL and redirects
- **File replacement**: New file uploaded with same key pattern, old key deleted from storage, version incremented
- **File validation**: TemplateFileValidator checks MIME type (PDF, DOCX, XLSX) via content sniffing + magic bytes, max 10MB

### Download Tracking

- **Group-level count**: `DownloadCount` on LibraryTemplateGroup, incremented atomically on each download attempt (URL generation)
- **Per-user log**: `LibraryTemplateDownload` append-only table (AccountID, TemplateID, GroupID, CreatedAt=DownloadedAt)
- **Anonymous downloads**: Increment count only, no log row (no AccountID available)

### Tier Gating

- **TierService interface**: `HasAccess(ctx, accountID, requiredTier) (bool, error)` — defined in the library module
- **Phase 1 stub**: BASIC always returns true; PRO returns true for all authenticated users (placeholder until subscription system)
- **Enforced at download time**: View usecase checks tier before generating presigned URL

### Interactive Forms

- **Attached to template**: `LibraryInteractiveForm` has FK to `LibraryTemplate` (one form per template), not to the group
- **Form layout**: Structured JSONB with sections → fields (field types: text, textarea, number, email, phone, date, select, multiselect, checkbox, radio)
- **Phase 1 scope**: Form definition stored + served via API. User fills and exports on the frontend. No backend submission storage.
- **Frontend responsibility**: Render form from schema, validate user input, handle export/download of filled content

### API Design

- **Route prefix**: `/api/v1/library/...` for user-facing, `/api/v1/admin/library/...` for admin
- **Framework**: Huma v2 (consistent with all existing modules)
- **Auth middleware**: IAM's `AuthMiddleware` + `AccountStatusMiddleware` on user routes
- **Permission middleware**: IAM's `PermissionMiddleware` on admin routes for write operations
- **requiresAuth**: Checked at the usecase level, not the middleware level — the middleware always runs, but the handler can still serve unauthenticated users for public templates

### Admin Permissions

- The library module does NOT define custom permission codes in Phase 1
- Admin routes use the standard `ReadAccess`/`WriteAccess`/`UpdateAccess`/`DeleteAccess` permissions from `pkg/permissions`
- Admins are assigned these permissions via the IAM role/permission system
- Specific permission codes (e.g., `library.write`) can be added later when the permission model matures

### Event Publishing

The module publishes `library.template.downloaded` events. The notification module can subscribe to these in a future phase for admin alerts or user notifications.

### Error Codes

Module-scoped error codes in `domain/error/errors.go`:
- `library.errors.templateNotFound`
- `library.errors.templateGroupNotFound`
- `library.errors.categoryNotFound`
- `library.errors.invalidFileType`
- `library.errors.fileTooLarge`
- `library.errors.tierAccessDenied`
- `library.errors.authRequired`
- `library.errors.slugAlreadyExists`
- `library.errors.categoryHasActiveGroups`

## Testing Decisions

### Testing Philosophy

Good tests verify external behavior, not implementation details. Tests should confirm that given certain inputs and state, the correct outputs and state changes occur — without asserting on internal method call sequences, private field values, or specific SQL queries.

### Modules to Test

| Module | Test Type | Priority | Rationale |
|---|---|---|---|
| `LibraryViewUsecase` | Unit (mocked repos + tier) | High | Core user flow — must correctly resolve language, check auth+tier, generate download, log |
| `LibraryAdminUsecase` | Unit (mocked repos + storage) | High | CRUD correctness, version bump logic, validation |
| `TemplateFileValidator` | Unit (no mocks needed) | High | Pure function — must correctly identify valid/invalid files by magic bytes |
| `LibraryService` | Unit (mocked usecases + storage) | Medium | Orchestration logic — uploads file, creates DB record, rolls back on failure |
| Repository implementations | Integration (test DB) | Low | Standard GORM patterns. Test only custom queries (IncrementDownloadCount, GetBySlug) |

### Prior Art

- Community module tests in `internal/modules/community/` — table-driven unit tests with mocked repository interfaces
- Uses `testify/suite` and `testify/mock`
- Integration tests use a test database with transaction rollback

### Test Cases to Prioritize

1. **Download flow**: Given valid slug → presigned URL returned, downloadCount incremented, log created. Given invalid slug → 404. Given requiresAuth with unauthenticated user → 403.
2. **Tier gating**: Given PRO template with basic-tier user → access denied. Given BASIC template → always allowed.
3. **Language fallback**: Given user requests language "am" but only "en" variant exists → returns "en" variant (fallback to DefaultLanguage).
4. **File validation**: Given PDF with correct magic bytes → accepted. Given text file renamed to .pdf → rejected. Given file > 10MB → rejected.
5. **Version bump**: Given template file replaced → version incremented by 1. Given metadata-only update → version unchanged.
6. **Category slug uniqueness**: Given two root categories with same slug → second rejected. Given same slug under different parents → both accepted.
7. **Group deletion with active templates**: Given category has active template groups → delete rejected with error.

## Out of Scope

- **Form submission storage**: Interactive form data is handled client-side only. Backend stores the form definition but not filled data.
- **Subscription/tier system**: TierService stub in Phase 1. Real subscription management is a separate module.
- **Template preview/rendering**: The frontend is responsible for rendering templates and forms. No server-side PDF generation.
- **Batch upload**: Admins upload one template at a time. Bulk upload is Phase 2.
- **Template version history**: Old file versions are deleted on replacement. No version history retention.
- **User ratings/reviews**: No rating or review system for templates in Phase 1.
- **Template analytics dashboard**: Download logs are available via API. A dedicated analytics dashboard is Phase 2.
- **RabbitMQ event publishing**: Stubbed in Phase 1. Can be enabled when the notification module is ready to subscribe.
- **Custom permission codes**: Uses generic Read/Write/Update/Delete permissions. Module-specific codes deferred.
- **Template recommendations**: No "related templates" or ML-based recommendations.
- **Search indexing**: Uses basic ILIKE search. Full-text search (PostgreSQL tsvector or Elasticsearch) is Phase 2.

## Further Notes

### Seed Data

Initial categories should be seeded via migration:
- Business Plans
- Invoices
- Record Keeping
- Financial Statements
- Contracts & Agreements
- Marketing Materials

### Future Phases

**Phase 2 possibilities:**
- Full-text search (PostgreSQL tsvector)
- Template version history
- User ratings and reviews
- Bulk upload for admins
- Template recommendations
- Form submission storage (save/resume/export)
- Notification integration (library.template.downloaded → user alert)
- Dedicated analytics dashboard
- Real subscription/tier system replacing stub

### Configuration

The module requires no additional configuration beyond the existing `Storage` config. Thumbnail validation (max 2MB, JPEG/PNG/WebP only) is hardcoded in the `TemplateFileValidator`.

### Naming Convention

All enum values use lowercase snake_case to match the community module's convention:
- `"pdf"`, `"docx"`, `"xlsx"`, `"interactive_form"`
- `"basic"`, `"pro"`

### Migration Strategy

All 6 entities are new tables — no existing schema changes required. The library module's `EntityProvider` registers all entities with `SchemaManager`, and Atlas generates the migration files. FK references to `accounts` table (in IAM module) are cross-module references that Atlas handles.

### Storage Cleanup

When a template is hard-deleted (via `HardDelete` or database-level cleanup), the associated file key in storage should also be deleted. This is handled in the admin usecase implementation — not automatic via GORM hooks, to avoid unintended storage deletes during soft-delete operations.
