# Library Module — Use Case Interfaces & Input DTOs

All use case interfaces live in `domain/usecase/` and are implemented in `application/usecase/`. Input DTOs are defined alongside the interfaces. An application service (`application/service/`) composes use cases with storage operations for file uploads.

---

## Input DTOs

Defined in `domain/usecase/inputs.go`. Follows the codebase pattern of separate Create vs Update inputs.

### Category Inputs

```
CreateCategoryInput {
    Name:               string
    Slug:               string
    Icon:               *string
    SortOrder:          int
    ParentCategoryID:   *uuid.UUID
}

UpdateCategoryInput {
    Name:               *string
    Slug:               *string
    Icon:               *string
    SortOrder:          *int
    ParentCategoryID:   *uuid.UUID
    IsActive:           *bool
}

CreateCategoryTranslationInput {
    CategoryID:         uuid.UUID
    Language:           string
    Name:               string
    Description:        *string
}

UpdateCategoryTranslationInput {
    Name:               *string
    Description:        *string
}
```

### Template Group Inputs

```
CreateTemplateGroupInput {
    Name:               string
    Description:        *string
    Slug:               string
    CategoryID:         uuid.UUID
    Format:             TemplateFormat
    TierAccess:         TierAccess
    RequiresAuth:       bool
    SortOrder:          int
    DefaultLanguage:    string
    ThumbnailBytes:     []byte         // image bytes for thumbnail upload
    ThumbnailFilename:  *string        // original filename for extension detection
}

UpdateTemplateGroupInput {
    Name:               *string
    Description:        *string
    Slug:               *string
    CategoryID:         *uuid.UUID
    Format:             *TemplateFormat
    TierAccess:         *TierAccess
    RequiresAuth:       *bool
    SortOrder:          *int
    DefaultLanguage:    *string
    IsActive:           *bool
    ThumbnailBytes:     []byte
    ThumbnailFilename:  *string
}
```

### Template Inputs

```
CreateTemplateInput {
    GroupID:            uuid.UUID
    Language:           string
    Title:              string
    Description:        *string
    FileBytes:          []byte         // uploaded file content
    Filename:           string         // original filename for extension detection
}

UpdateTemplateInput {
    Title:              *string
    Description:        *string
    FileBytes:          []byte         // if provided, replaces the file and bumps version
    Filename:           *string
    IsActive:           *bool
}
```

### Interactive Form Inputs

```
CreateInteractiveFormInput {
    TemplateID:         uuid.UUID
    Name:               string
    Description:        *string
    FormLayout:         map[string]interface{}   // structured sections+fields JSON
}

UpdateInteractiveFormInput {
    Name:               *string
    Description:        *string
    FormLayout:         *map[string]interface{}
}
```

### Download Input

```
DownloadInput {
    Slug:               string
    Language:           *string       // nil = use group's DefaultLanguage
    AccountID:          *uuid.UUID    // nil = anonymous (no log created)
}

DownloadOutput {
    PresignedURL:       string
    ExpiresAt:          time.Time
    Filename:           string        // derived from title + extension
}
```

---

## Cross-Module Interface: TierService

```go
type TierService interface {
    HasAccess(ctx context.Context, accountID uuid.UUID, requiredTier TierAccess) (bool, error)
}
```

**Phase 1 stub behavior:**
- `BASIC` → returns `true` for all accounts
- `PRO` → returns `true` for all authenticated accounts (placeholder until subscription system exists)

---

## Use Case Interfaces

### 1. LibraryViewUsecase

User-facing operations: browsing categories, searching templates, downloading files.

| Method | Signature | Description |
|---|---|---|
| `ListCategories` | `(ctx, locale *string) ([]*LibraryCategory, error)` | List active categories as a tree (flat list sorted by SortOrder). If locale provided, returns translated name/description. |
| `ListTemplateGroups` | `(ctx, categoryID *uuid.UUID, format *TemplateFormat, q query.QueryOptions) ([]*LibraryTemplateGroup, error)` | Paginated list of active template groups. Filters by category, format. Each result includes available language codes. |
| `GetTemplateGroup` | `(ctx, slug string, locale *string) (*LibraryTemplateGroup, []*LibraryTemplate, error)` | Get a group by slug with its active language variants. If locale provided, returns the matching variant. Falls back to DefaultLanguage if locale variant doesn't exist. |
| `DownloadTemplate` | `(ctx, input DownloadInput) (*DownloadOutput, error)` | Generate presigned URL for download. Checks requiresAuth, tierAccess, isActive. Increments downloadCount. Creates download log (if authenticated). |

**`DownloadTemplate` internal flow:**

1. Look up `LibraryTemplateGroup` by slug
2. Check `IsActive` — if false, return 404
3. Check `RequiresAuth` — if true, ensure `AccountID` is present
4. Check `TierAccess` — if PRO, call `TierService.HasAccess`
5. Resolve language: if `input.Language` provided → find that variant; else → use `DefaultLanguage`. If variant not found → error.
6. Check `Template.IsActive`
7. Generate presigned URL via `storage.GetPresignedURL` with 5-minute expiry
8. Increment `Group.DownloadCount` via repository
9. If `AccountID` is not nil → create `LibraryTemplateDownload` row
10. Return `DownloadOutput`

---

### 2. LibraryAdminUsecase

Admin operations: CRUD for categories, groups, templates, and interactive forms.

| Method | Signature | Description |
|---|---|---|
| `CreateCategory` | `(ctx, input CreateCategoryInput) (*LibraryCategory, error)` | Create a category. Validates slug uniqueness within parent. |
| `GetCategory` | `(ctx, id uuid.UUID) (*LibraryCategory, error)` | Get category with translations. |
| `UpdateCategory` | `(ctx, id uuid.UUID, input UpdateCategoryInput) (*LibraryCategory, error)` | Update category. Includes soft-toggle via IsActive. |
| `DeleteCategory` | `(ctx, id uuid.UUID) error` | Soft-delete. Fails with error if category has active template groups (must reassign or delete groups first). |
| `ListAllCategories` | `(ctx, includeInactive bool) ([]*LibraryCategory, error)` | List categories including inactive. For admin management UI. |
| `AddCategoryTranslation` | `(ctx, input CreateCategoryTranslationInput) (*LibraryCategoryTranslation, error)` | Add a translation for a language. |
| `UpdateCategoryTranslation` | `(ctx, categoryID uuid.UUID, language string, input UpdateCategoryTranslationInput) (*LibraryCategoryTranslation, error)` | Update existing translation. |
| `DeleteCategoryTranslation` | `(ctx, categoryID uuid.UUID, language string) error` | Remove a translation. |
| `CreateTemplateGroup` | `(ctx, createdBy uuid.UUID, input CreateTemplateGroupInput) (*LibraryTemplateGroup, error)` | Create group with optional thumbnail upload. Validates slug uniqueness within category. |
| `GetTemplateGroup` | `(ctx, id uuid.UUID) (*LibraryTemplateGroup, error)` | Get group with all templates (including inactive). |
| `UpdateTemplateGroup` | `(ctx, id uuid.UUID, input UpdateTemplateGroupInput) (*LibraryTemplateGroup, error)` | Update group metadata. Replace thumbnail if provided. |
| `DeleteTemplateGroup` | `(ctx, id uuid.UUID) error` | Soft-delete. Cascades to templates and downloads. |
| `ListAllTemplateGroups` | `(ctx, categoryID *uuid.UUID, status *bool, q query.QueryOptions) ([]*LibraryTemplateGroup, error)` | List all groups (including inactive) with admin filters. |
| `CreateTemplate` | `(ctx, input CreateTemplateInput) (*LibraryTemplate, error)` | Upload file + create language variant. Validates format+size. Stores file via storage. |
| `UpdateTemplate` | `(ctx, id uuid.UUID, input UpdateTemplateInput) (*LibraryTemplate, error)` | Update metadata. If FileBytes provided, replaces file in storage and bumps version. |
| `DeleteTemplate` | `(ctx, id uuid.UUID) error` | Soft-delete. Deletes file from storage on hard delete. |
| `CreateInteractiveForm` | `(ctx, input CreateInteractiveFormInput) (*LibraryInteractiveForm, error)` | Create form for an interactive template. Validates parent template format is `interactive_form`. |
| `UpdateInteractiveForm` | `(ctx, id uuid.UUID, input UpdateInteractiveFormInput) (*LibraryInteractiveForm, error)` | Update form. Bumps version on FormLayout change. |
| `DeleteInteractiveForm` | `(ctx, id uuid.UUID) error` | Soft-delete. |
| `GetDownloadLogs` | `(ctx, groupID *uuid.UUID, q query.QueryOptions) ([]*LibraryTemplateDownload, error)` | View download history. Optionally filtered by group. |

---

## Application Service

### LibraryService

Composes admin use cases with storage operations (file uploads, thumbnail uploads). Follows the `CommunityService` pattern.

```go
type LibraryService interface {
    CreateTemplateGroup(ctx context.Context, accountID uuid.UUID, input CreateTemplateGroupInput) (*entity.LibraryTemplateGroup, error)
    UpdateTemplateGroup(ctx context.Context, id uuid.UUID, input UpdateTemplateGroupInput) (*entity.LibraryTemplateGroup, error)
    CreateTemplate(ctx context.Context, input CreateTemplateInput) (*entity.LibraryTemplate, error)
    UpdateTemplate(ctx context.Context, id uuid.UUID, input UpdateTemplateInput) (*entity.LibraryTemplate, error)
}
```

**File handling logic:**
- `CreateTemplateGroup`: if `ThumbnailBytes` provided → validates as image (JPEG/PNG/WebP, max 2MB) → uploads to `library/thumbnails/{groupID}.{ext}` → stores URL
- `UpdateTemplateGroup`: if new thumbnail provided → uploads, deletes old thumbnail
- `CreateTemplate`: validates file via `TemplateFileValidator` → uploads to `library/templates/{templateID}.{ext}` → creates DB record
- `UpdateTemplate`: if new file provided → uploads, deletes old file from storage, bumps version

### TemplateFileValidator

Validates uploaded template files for type and size.

- Allowed MIME types: `application/pdf` (.pdf), `application/vnd.openxmlformats-officedocument.wordprocessingml.document` (.docx), `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` (.xlsx)
- Type detection via content sniffing (magic bytes)
- Max file size: 10MB
- Additional format-specific validation: PDF magic bytes `%PDF-`, DOCX ZIP magic bytes `PK\x03\x04`
