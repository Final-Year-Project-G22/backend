# Library Module — PR Breakdown Plan

## 1) PR Decomposition Summary

The library module is broken into 4 pull requests, sequenced by dependency order. Foundation work (entities, interfaces, repos, validators) comes first, then feature work splits into admin CRUD and user-facing browsing/download, capped by the integration PR that wires everything together.

| PR | Title | Dependencies | Risk |
|----|-------|--------------|------|
| PR-L1 | Domain + Infrastructure Foundation | None | Low |
| PR-L2 | Admin CRUD + Storage Orchestration | PR-L1 | Medium |
| PR-L3 | User Browsing + Download | PR-L1 | Medium |
| PR-L4 | Module Integration | PR-L1, PR-L2, PR-L3 | Low |

---

## 2) Clarifications Needed

All ambiguities resolved in prior conversation. Key decisions:
- IAM exposes a thin usecase method for reading `TemplatePreference` — library module calls it for default template hinting
- `TierService` stub returns `true` for all authenticated accounts (BASIC and PRO both allowed in Phase 1)
- Seed data for initial categories created via migration file
- Event publishing (`library.template.downloaded`) is stub/no-op in Phase 1
- Module name is `library` in code, route prefix is `/api/v1/library/...`

---

## 3) PR Plan

---

### PR-L1 — Domain + Infrastructure Foundation

**Goal:** Establish all domain contracts and data layer implementations — entities, enums, repository interfaces + implementations, use case interfaces, input DTOs, TierService interface, TemplateFileValidator, error codes, and EntityProvider.

**Scope:**
- 2 enum types: `TemplateFormat`, `TierAccess`
- 6 entity structs with GORM tags, `TableName()`, embedded `BaseModel`:
  - `LibraryCategory`, `LibraryCategoryTranslation`, `LibraryTemplateGroup`, `LibraryTemplate`, `LibraryInteractiveForm`, `LibraryTemplateDownload`
- 5 repository interfaces extending `GenericRepository[T]`:
  - `LibraryCategoryRepository`, `LibraryTemplateGroupRepository`, `LibraryTemplateRepository`, `LibraryInteractiveFormRepository`, `LibraryTemplateDownloadRepository`
- 2 use case interfaces with complete method signatures:
  - `LibraryViewUsecase`, `LibraryAdminUsecase`
- All input DTOs in `domain/usecase/inputs.go`
- `TierService` interface: `HasAccess(ctx, accountID, requiredTier) (bool, error)`
- `TemplateFileValidator` in `application/service/template_file_validator.go`
  - Validate MIME type via content sniffing (PDF, DOCX, XLSX)
  - Validate file size (max 10MB)
  - Extension + magic byte detection
- 5 repository implementations in `infrastructure/repository/`
- Shared `helpers.go` with `getDB(ctx)` and `applyPaginationAndSorting()`
- `EntityProvider` in `internal/modules/library/entities.go` registering all 6 entities
- Domain error codes in `domain/error/errors.go`
- Empty directory scaffolding: `application/usecase/`, `application/service/`, `infrastructure/repository/`, `infrastructure/tier/`, `delivery/handler/`, `delivery/dto/`, `delivery/routes/`

**Out of scope:** Any use case implementations, handlers, routes, module wiring, seed data, storage upload orchestration

**Dependencies:** None — pure definitions + data layer implementations

**Files/modules likely touched:**
```
internal/modules/library/domain/entity/enums.go
internal/modules/library/domain/entity/library_category.go
internal/modules/library/domain/entity/library_category_translation.go
internal/modules/library/domain/entity/library_template_group.go
internal/modules/library/domain/entity/library_template.go
internal/modules/library/domain/entity/library_interactive_form.go
internal/modules/library/domain/entity/library_template_download.go
internal/modules/library/domain/repository/category_repository.go
internal/modules/library/domain/repository/template_group_repository.go
internal/modules/library/domain/repository/template_repository.go
internal/modules/library/domain/repository/interactive_form_repository.go
internal/modules/library/domain/repository/template_download_repository.go
internal/modules/library/domain/usecase/inputs.go
internal/modules/library/domain/usecase/library_view_usecase.go
internal/modules/library/domain/usecase/library_admin_usecase.go
internal/modules/library/domain/usecase/tier_service.go
internal/modules/library/domain/error/errors.go
internal/modules/library/entities.go
internal/modules/library/application/service/template_file_validator.go
internal/modules/library/infrastructure/repository/helpers.go
internal/modules/library/infrastructure/repository/category_repository.go
internal/modules/library/infrastructure/repository/template_group_repository.go
internal/modules/library/infrastructure/repository/template_repository.go
internal/modules/library/infrastructure/repository/interactive_form_repository.go
internal/modules/library/infrastructure/repository/template_download_repository.go
```

**Commit checkpoints:**
1. Enums + entity structs (6 entities, 2 enums)
2. Repository interfaces (all 5)
3. Use case interfaces + input DTOs (2 + inputs.go)
4. TierService interface
5. Domain error codes
6. Repository helpers (getDB, applyPaginationAndSorting)
7. Repository implementations (all 5)
8. TemplateFileValidator
9. EntityProvider
10. Empty directory scaffolding

**Tests:**
- **Unit:**
  - `TemplateFileValidator`: Given PDF with `%PDF-` magic bytes → accepted. Given DOCX with `PK\x03\x04` magic bytes → accepted. Given text file renamed to `.pdf` → rejected. Given file > 10MB → rejected. Given unsupported MIME type (e.g., image/png) → rejected.

**Acceptance criteria:**
- `go build ./...` passes cleanly
- All entity structs embed `model.BaseModel` (except `LibraryCategoryTranslation` which follows `GuideCategoryTranslation` pattern)
- `LibraryCategoryTranslation` has proper `BeforeCreate` hook for UUID generation
- All GORM tags follow codebase conventions
- FK references use `AccountID` for account references (`CreatedBy` on group, `AccountID` on download)
- `LibraryTemplateGroup` unique index on `(CategoryID, Slug)`
- `LibraryTemplate` unique index on `(GroupID, Language)`
- `LibraryCategory` unique index on `(ParentCategoryID, Slug)`
- Repository implementations match the guide/community patterns (embed `GenericRepository`, constructor returning interface)
- `TemplateFileValidator` correctly distinguishes PDF, DOCX, XLSX from other file types via magic bytes
- `EntityProvider.Entities()` returns all 6 entity pointers
- `EntityProvider.ModuleName()` returns `"library"`
- Error codes use module-scoped format: `library.errors.*`

**Risk level:** Low — pure definitions + standard repository implementations

**Notes:** The `LibraryCategoryTranslation` entity follows the exact same pattern as `GuideCategoryTranslation` (no BaseModel, custom PK + timestamps, BeforeCreate hook). The `TemplateFileValidator` is a pure function — it operates on `[]byte` content and doesn't need the actual file on disk. It supports three formats: PDF (magic bytes `%PDF-`), DOCX (ZIP magic bytes `PK\x03\x04`), XLSX (also ZIP magic bytes). Future formats can be added by extending the allowed types map.

---

### PR-L2 — Admin CRUD + Storage Orchestration

**Goal:** Implement the admin use case, the LibraryService orchestrator (file uploads + DB operations), the TierService stub, and the admin HTTP routes + handler.

**Scope:**
- `LibraryAdminUsecase` implementation (`application/usecase/library_admin_usecase.go`):
  - **Categories:** Create, Get, Update, Delete (soft), ListAll
  - **Category Translations:** Add, Update, Delete
  - **Template Groups:** Create (with optional thumbnail), Get, Update (with optional thumbnail replace), Delete (soft), ListAll
  - **Templates:** Create (file upload), Get, Update (metadata + optional file replace), Delete (soft)
  - **Interactive Forms:** Create (validates parent format is `interactive_form`), Get, Update (bumps version on FormLayout change), Delete
  - **Download Logs:** GetDownloadLogs with group filter
- `LibraryService` (`application/service/library_service.go`):
  - `CreateTemplateGroup` — validates thumbnail → uploads to storage → delegates to admin usecase
  - `UpdateTemplateGroup` — replaces thumbnail if provided → delegates to admin usecase
  - `CreateTemplate` — validates file → uploads to storage (`library/templates/{templateID}.{ext}`) → delegates to admin usecase
  - `UpdateTemplate` — replaces file if provided → deletes old file from storage → delegates to admin usecase
- `TierService` stub (`infrastructure/tier/tier_service_stub.go`):
  - `HasAccess(ctx, accountID, tierAccess)` → returns `true` for all inputs (Phase 1 placeholder)
- Admin DTOs (`delivery/dto/library_admin_dto.go`)
- Admin handler (`delivery/handler/library_admin_handler.go`)
- Admin routes (`delivery/routes/library_admin_routes.go`)

**Out of scope:** User-facing browse/download (PR-L3), presigned URL generation (PR-L3), download tracking (PR-L3)

**Dependencies:** PR-L1 (entities, repos, interfaces, validator)

**Files/modules likely touched:**
```
internal/modules/library/application/usecase/library_admin_usecase.go
internal/modules/library/application/service/library_service.go
internal/modules/library/infrastructure/tier/tier_service_stub.go
internal/modules/library/delivery/handler/library_admin_handler.go
internal/modules/library/delivery/dto/library_admin_dto.go
internal/modules/library/delivery/routes/library_admin_routes.go
```

**Commit checkpoints:**
1. `TierService` stub implementation
2. `LibraryAdminUsecase` — category CRUD + translations
3. `LibraryAdminUsecase` — template group CRUD
4. `LibraryAdminUsecase` — template CRUD (file upload logic in next commit)
5. `LibraryService` — file upload + thumbnail orchestration
6. `LibraryAdminUsecase` — interactive form CRUD
7. `LibraryAdminUsecase` — download log listing
8. Admin DTOs
9. Admin handler + routes

**Tests:**
- **Unit (High priority):**
  - **Category CRUD:** CreateCategory → validates slug uniqueness. UpdateCategory → updates fields. DeleteCategory with active groups → error. DeleteCategory without groups → soft-delete.
  - **Template Group CRUD:** Create → creates with metadata. Update → updates fields. Delete → soft-delete, cascades to templates.
  - **Template CRUD:** Create with file → file stored, DB record created. Update file → old file deleted, version bumped. Update metadata only → version unchanged.
  - **Interactive Form CRUD:** Create on non-interactive format → error. Create on interactive format → success. Update FormLayout → version bumped.
  - **LibraryService:** CreateTemplate → file uploaded to storage, DB record created. On DB failure → file deleted from storage (rollback).

**Acceptance criteria:**
- Admin can create category via `POST /api/v1/admin/library/categories`
- Admin can get category with translations via `GET /api/v1/admin/library/categories/{id}`
- Admin can update category via `PATCH /api/v1/admin/library/categories/{id}`
- Admin can delete category via `DELETE /api/v1/admin/library/categories/{id}` (blocked if has active groups)
- Admin can list categories (including inactive) via `GET /api/v1/admin/library/categories`
- Admin can add/update/delete category translations
- Admin can create template group with thumbnail via `POST /api/v1/admin/library/template-groups`
- Admin can upload template file via `POST /api/v1/admin/library/template-groups/{groupId}/templates`
- Admin can replace template file via `PATCH /api/v1/admin/library/templates/{id}` → version increments
- Admin can create/update/delete interactive forms
- Admin can view download logs via `GET /api/v1/admin/library/downloads`
- File validation: uploading a `.txt` renamed to `.pdf` returns 400 with `library.errors.invalidFileType`
- File size > 10MB returns 413 with `library.errors.fileTooLarge`
- TierService stub returns `true` for all accounts

**Risk level:** Medium — file upload orchestration needs careful error handling (rollback storage on DB failure). Version bump logic must use atomic increment (`SET version = version + 1`), not read-then-write.

**Notes:** The `LibraryService` follows the same pattern as `CommunityService` — it wraps usecase calls with storage operations. Error handling: if the DB insert fails after a successful storage upload, the storage file must be deleted (cleanup). The `TierServiceStub` is intentionally trivial for Phase 1 — it returns `true` unconditionally. When the subscription module exists, this stub is replaced with a real implementation that checks account tier.

---

### PR-L3 — User Browsing + Download

**Goal:** Implement the user-facing template library — browse categories, list/search template groups, view details, and download with presigned URL generation + download tracking.

**Scope:**
- `LibraryViewUsecase` implementation (`application/usecase/library_view_usecase.go`):
  - `ListCategories` — active categories as tree, optionally localized (reads translations)
  - `ListTemplateGroups` — paginated, filterable by category/format/search, sorted by SortOrder
  - `GetTemplateGroup` — group details with active language variants, language fallback support
  - `DownloadTemplate` — the full download flow:
    1. Lookup group by slug, check IsActive
    2. Check RequiresAuth → if true, AccountID must be present
    3. Check TierAccess → call TierService
    4. Resolve language: requested → DefaultLanguage fallback
    5. Check Template.IsActive
    6. Generate presigned URL via `storage.GetPresignedURL(ctx, fileKey, 5*time.Minute)`
    7. Increment Group.DownloadCount atomically
    8. If AccountID ≠ nil → create LibraryTemplateDownload row
    9. Return DownloadOutput with presigned URL + expiresAt + filename
- User-facing DTOs (`delivery/dto/library_dto.go`)
- User-facing handler (`delivery/handler/library_handler.go`)
- User-facing routes (`delivery/routes/library_routes.go`)

**Out of scope:** Admin CRUD (PR-L2), interactive form serving (form data returned as JSON, frontend renders it), module wiring (PR-L4), seed data (PR-L4)

**Dependencies:** PR-L1 (entities, repos, interfaces, TierService interface, storage.Storage)

**Files/modules likely touched:**
```
internal/modules/library/application/usecase/library_view_usecase.go
internal/modules/library/delivery/handler/library_handler.go
internal/modules/library/delivery/dto/library_dto.go
internal/modules/library/delivery/routes/library_routes.go
```

**Commit checkpoints:**
1. `LibraryViewUsecase` — ListCategories (with locale support)
2. `LibraryViewUsecase` — ListTemplateGroups (pagination, filters, search)
3. `LibraryViewUsecase` — GetTemplateGroup (with language fallback)
4. `LibraryViewUsecase` — DownloadTemplate (full flow: auth, tier, language, presigned URL, count, log)
5. User-facing DTOs
6. User-facing handler + routes

**Tests:**
- **Unit (High priority):**
  - **Download flow:** Given valid slug → presigned URL returned, downloadCount incremented, log created. Given invalid slug → `library.errors.templateNotFound`. Given requiresAuth with anonymous user → `library.errors.authRequired`. Given PRO template → TierService called. Given language "am" with only "en" variant → returns "en" variant (fallback). Given inactive group → 404.
  - **Browse/list:** ListCategories → returns tree sorted by SortOrder, with translations if locale provided. ListTemplateGroups → paginated correctly, search matches title+description. GetTemplateGroup → returns group + active variants.
  - **Download log:** Authenticated download creates log row. Anonymous download does NOT create log row.
  - **TierService integration:** `HasAccess` called with correct tierAccess value.

**Acceptance criteria:**
- User can browse categories via `GET /api/v1/library/categories` (tree, sorted by SortOrder)
- User can list template groups via `GET /api/v1/library/templates?category={id}&format=pdf&search=business&page=1&pageSize=20`
- User can view group details via `GET /api/v1/library/templates/{slug}?locale=en` with language variants
- User can download via `GET /api/v1/library/templates/{slug}/download?language=en` → receives presigned URL
- Downloading increments the group's `DownloadCount`
- Download creates a `LibraryTemplateDownload` log row (authenticated only)
- Anonymous download (requiresAuth=false) skips auth middleware and log row
- Accessing PRO template without appropriate tier returns 403
- Language fallback: requesting "am" when only "en" exists returns the "en" variant
- Inactive templates and groups return 404

**Risk level:** Medium — the download flow has multiple decision points (auth, tier, language resolution, presigned URL, count increment, log creation). Each must be correct and gracefully handle errors.

**Notes:** The `DownloadTemplate` flow should use a transaction for the count increment + log creation to ensure consistency. The presigned URL generation is idempotent and can be retried if the transaction fails. For the language resolution: the user's locale comes from the query parameter `language`, not from account preferences (that can be added in Phase 2). The `TierService` call should not throw an error — the stub returns true, but when real implementation exists, it must not crash the download flow.

---

### PR-L4 — Module Integration

**Goal:** Wire everything together — module.go, route registration, seed data migration, modules.go. This is the capstone PR that activates the library module.

**Scope:**
- `internal/modules/library/module.go`:
  - All `fx.Provide` for repos, use cases, services, handlers
  - All `fx.Annotate` bindings (interface → implementation)
  - `fx.Invoke` for SchemaManager registration
  - `fx.Invoke` for route registration (user + admin routes)
- Route dependencies struct:
  - AuthMiddleware, AccountStatusMiddleware, PermissionMiddleware (from IAM)
  - LibraryHandler, LibraryAdminHandler
- `internal/modules/modules.go`:
  - Add `library.Module` to `Modules`
- Seed data migration file
  - Creates initial categories: Business Plans, Invoices, Record Keeping, Financial Statements, Contracts & Agreements, Marketing Materials
  - Each category seeded with proper slug, sort order, and English translation

**Out of scope:** Any business logic changes — pure integration work

**Dependencies:** PR-L1, PR-L2, PR-L3 (all implementations must exist)

**Files/modules likely touched:**
```
internal/modules/library/module.go
internal/modules/modules.go
internal/core/migration/xxx_seed_library_categories.go
```

**Commit checkpoints:**
1. `module.go` — all fx.Provide and fx.Annotate bindings
2. Route registration fx.Invoke — user routes + admin routes
3. Seed data migration — initial categories
4. modules.go registration

**Tests:**
- **Integration:** Verify module starts correctly with fx (no missing dependencies). Verify routes are accessible after registration.

**Acceptance criteria:**
- `go build ./...` passes with library module fully wired
- Application starts without fx panics
- User-facing routes (`/api/v1/library/...`) are registered and accessible
- Admin routes (`/api/v1/admin/library/...`) are registered and accessible
- `modules.go` includes `library.Module`
- Seed migration creates 6 initial categories with translations
- Seed categories are queryable via GET /api/v1/library/categories after migration runs

**Risk level:** Low — standard module wiring following existing patterns (guide, community)

**Notes:** The module.go follows the exact same pattern as `guide/module.go`. The EntityProvider registration with SchemaManager is identical. Route dependencies use IAM's AuthMiddleware, AccountStatusMiddleware, and PermissionMiddleware (from `iammiddleware` package). Seed data migration uses Atlas migration format consistent with existing seed migrations in the codebase.

---

## 4) Recommended Order

```
PR-L1 (Domain + Infrastructure Foundation)
   ↓                    ↓
PR-L2 (Admin CRUD)   PR-L3 (User Browsing + Download)   ← can be parallel
   ↓                    ↓
PR-L4 (Module Integration)
```

**Merge order:** PR-L1 → PR-L2 + PR-L3 (parallel) → PR-L4

**Parallel work:** After PR-L1 merges, PR-L2 and PR-L3 can be developed in parallel by separate developers since they have no dependency on each other — they only depend on PR-L1.

---

## 5) Test Strategy Across the Full Epic

| Test Area | PR Introduced | Test Type | Notes |
|-----------|--------------|-----------|-------|
| TemplateFileValidator | PR-L1 | Unit (pure function) | No mocks, test magic byte detection |
| Category CRUD | PR-L2 | Unit (mocked repos) | Slug uniqueness, active groups constraint |
| Template Group CRUD | PR-L2 | Unit (mocked repos) | Thumbnail upload, lifecycle |
| Template CRUD | PR-L2 | Unit (mocked repos + storage) | File upload, replace, version bump |
| Interactive Form CRUD | PR-L2 | Unit (mocked repos) | Format validation, version on layout change |
| LibraryService orchestration | PR-L2 | Unit (mocked usecase + storage) | Upload → create DB, rollback on failure |
| Download flow | PR-L3 | Unit (mocked repos + storage + tier) | Auth, tier, language, presigned URL, count, log |
| Browse/list/search | PR-L3 | Unit (mocked repos) | Categories tree, pagination, filters, search |
| Language fallback | PR-L3 | Unit (mocked repos) | Fallback to DefaultLanguage |
| Anonymous download | PR-L3 | Unit (mocked repos + storage) | No log row, count only |
| Module startup | PR-L4 | Integration | DI wiring, route registration, seed data |

**Integration tests:** (deferred, low priority)
- Repository implementations with test DB
- Full download flow end-to-end (handler → usecase → storage → response)

---

## 6) Open Questions Remaining

- **IAM TemplatePreference read**: The library module reads IAM's `TemplatePreference`. IAM exposes a thin usecase method (e.g., `GetTemplatePreference(ctx, accountID)`). The library module calls this from `LibraryViewUsecase.GetTemplateGroup` to highlight the user's default template. **Recommendation:** IAM adds `GetTemplatePreference(ctx, accountID uuid.UUID) (*entity.TemplatePreference, error)` to its `PreferenceUsecase`. The library module injects IAM's `PreferenceUsecase` (or a thin interface) via DI.
- **LibraryTemplateDownload cleanup**: Download logs grow indefinitely. A cleanup strategy (e.g., delete logs older than 90 days) should be planned for Phase 2.
