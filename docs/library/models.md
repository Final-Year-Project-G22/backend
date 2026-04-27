# Library Module — Entities & Models

## Enums

### TemplateFormat

| Value | Description |
|---|---|
| `pdf` | Adobe PDF document |
| `docx` | Microsoft Word document |
| `xlsx` | Microsoft Excel spreadsheet |
| `interactive_form` | Fillable digital form (Pro tier) |

### TierAccess

| Value | Description |
|---|---|
| `basic` | Available to all users (Free tier) |
| `pro` | Available to Pro tier subscribers only |

---

## Entities

### Shared Patterns

All entities (except `LibraryCategoryTranslation`) embed `model.BaseModel` which provides:

| Field | Type | Description |
|---|---|---|
| `ID` | uuid | PK, auto-generated via `gen_random_uuid()` |
| `CreatedAt` | timestamptz | Not null, default: `CURRENT_TIMESTAMP` |
| `UpdatedAt` | timestamptz | Not null, default: `CURRENT_TIMESTAMP` |
| `DeletedAt` | timestamptz | Nullable, indexed — enables GORM soft-delete |

All foreign keys to accounts use `AccountID` (consistent with the codebase convention).

---

### 1. LibraryCategory

**Table:** `library_categories`

Hierarchical category tree for organizing template groups. Supports nesting via `ParentCategoryID`.

| Field | Type | Constraints | Description |
|---|---|---|---|
| `BaseModel` | embedded | — | ID, CreatedAt, UpdatedAt, DeletedAt |
| `Name` | varchar(200) | not null | Human-readable category name |
| `Slug` | varchar(200) | not null, uniqueIndex(priority:2) | URL-friendly identifier, unique within parent |
| `Icon` | varchar(100) | nullable | Icon name mapping to frontend icon set |
| `SortOrder` | int | not null, default:0 | Admin-controlled display order (ASC) |
| `ParentCategoryID` | uuid | nullable, uniqueIndex(priority:1), index | Self-referential FK for nesting. Null = root category |
| `ParentCategory` | LibraryCategory | FK self, OnUpdate:CASCADE, OnDelete:SET NULL | Parent relationship |
| `ChildCategories` | []LibraryCategory | FK: ParentCategoryID | Children relationship |
| `TemplateGroups` | []LibraryTemplateGroup | FK: CategoryID | Template groups in this category |
| `Translations` | []LibraryCategoryTranslation | FK: LibraryCategoryID | Per-language translations |
| `IsActive` | bool | not null, default:true | Temporarily hide without deleting |

**Unique index:** `idx_library_categories_slug_per_parent` on `(ParentCategoryID, Slug)`. Root categories have `ParentCategoryID IS NULL` — slug uniqueness among roots enforced by the DB (multiple NULLs in unique index allowed by PostgreSQL).

**GORM relationships:**

- `ParentCategory` — self-referential FK, OnDelete:SET NULL
- `ChildCategories — self-referential FK, OnDelete:SET NULL`  (wait I already listed it)
- `TemplateGroups []LibraryTemplateGroup` — FK: `CategoryID`, OnDelete:RESTRICT
- `Translations []LibraryCategoryTranslation` — FK: `LibraryCategoryID`, OnDelete:CASCADE

---

### 2. LibraryCategoryTranslation

**Table:** `library_category_translations`

Per-language translation of a category's display name and description.

> Does NOT embed `BaseModel`. Uses its own simple PK + timestamps pattern, consistent with `GuideCategoryTranslation` in the codebase.

| Field | Type | Constraints | Description |
|---|---|---|---|
| `ID` | uuid | PK, auto-generated | Primary key |
| `LibraryCategoryID` | uuid | FK → library_categories, not null, uniqueIndex(priority:1) | Parent category |
| `Language` | varchar(10) | not null, uniqueIndex(priority:2) | ISO 639-1 locale code (e.g., "en", "am") |
| `Name` | varchar(200) | not null | Translated category name |
| `Description` | text | nullable | Translated description |
| `CreatedAt` | timestamptz | not null, default:CURRENT_TIMESTAMP | |
| `UpdatedAt` | timestamptz | not null, default:CURRENT_TIMESTAMP | |

**BeforeCreate hook:** Auto-generates UUID if ID is `uuid.Nil`.

**Unique index:** `idx_library_cat_trans` on `(LibraryCategoryID, Language)`

---

### 3. LibraryTemplateGroup

**Table:** `library_template_groups`

Groups language variants of the same template together. Holds shared metadata (format, tier, slug, category). One group = one template concept (e.g., "Business Plan") with multiple language-specific file variants.

| Field | Type | Constraints | Description |
|---|---|---|---|
| `BaseModel` | embedded | — | ID, CreatedAt, UpdatedAt, DeletedAt |
| `Name` | varchar(200) | not null | Admin-facing group name |
| `Description` | text | nullable | Admin-facing description |
| `Slug` | varchar(200) | not null, uniqueIndex(priority:2) | URL-friendly identifier, unique within category |
| `CategoryID` | uuid | FK → library_categories, not null, uniqueIndex(priority:1), index | Parent category |
| `Category` | LibraryCategory | FK: CategoryID, OnDelete:RESTRICT | Category relationship |
| `Format` | varchar(20) | not null | pdf, docx, xlsx, or interactive_form |
| `TierAccess` | varchar(10) | not null, default:'basic' | basic or pro |
| `RequiresAuth` | bool | not null, default:true | If false, downloadable by unauthenticated users |
| `IsActive` | bool | not null, default:true | Temporarily hide the entire group |
| `SortOrder` | int | not null, default:0 | Admin-controlled display order within category |
| `DefaultLanguage` | varchar(10) | not null, default:'en' | Fallback language when user's preferred variant is missing |
| `ThumbnailURL` | varchar(512) | nullable | Preview image URL from storage |
| `DownloadCount` | int | not null, default:0 | Total downloads across all language variants |
| `CreatedBy` | uuid | FK → accounts, not null, index | Admin who created the group |
| `Templates` | []LibraryTemplate | FK: GroupID | Language variants |

**Unique index:** `idx_library_template_groups_slug_per_category` on `(CategoryID, Slug)`

---

### 4. LibraryTemplate

**Table:** `library_templates`

A downloadable template file for a specific language. Each row is one language variant of a template group. The file is stored in SeaweedFS/MinIO with the key `library/templates/{id}.{ext}`.

| Field | Type | Constraints | Description |
|---|---|---|---|
| `BaseModel` | embedded | — | ID, CreatedAt, UpdatedAt, DeletedAt |
| `GroupID` | uuid | FK → library_template_groups, not null, uniqueIndex(priority:1), index | Parent group |
| `Group` | LibraryTemplateGroup | FK: GroupID, OnDelete:CASCADE | Group relationship |
| `Language` | varchar(10) | not null, uniqueIndex(priority:2) | ISO 639-1 locale code |
| `Title` | varchar(200) | not null | User-facing template title |
| `Description` | text | nullable | User-facing description |
| `FileKey` | varchar(512) | not null | Storage key for the file |
| `FileURL` | varchar(512) | nullable | Direct or proxied access URL |
| `FileSize` | bigint | not null | File size in bytes |
| `ContentType` | varchar(100) | not null | MIME type (e.g., "application/pdf") |
| `Version` | int | not null, default:1 | Incremented on file replacement |
| `IsActive` | bool | not null, default:true | Temporarily hide this language variant |
| `InteractiveForm` | LibraryInteractiveForm | FK: TemplateID | One-to-one relationship (only for interactive_form format) |

**Unique index:** `idx_library_templates_group_lang` on `(GroupID, Language)`

**Storage key convention:** `library/templates/{templateID}.{ext}` (e.g., `library/templates/a1b2c3d4.pdf`)

---

### 5. LibraryInteractiveForm

**Table:** `library_interactive_forms`

Fillable digital form attached to a template of format `interactive_form`. Only exists for Pro-tier templates. The `FormLayout` JSONB defines the form structure that the frontend renders.

| Field | Type | Constraints | Description |
|---|---|---|---|
| `BaseModel` | embedded | — | ID, CreatedAt, UpdatedAt, DeletedAt |
| `TemplateID` | uuid | FK → library_templates, not null, uniqueIndex | Parent template |
| `Template` | LibraryTemplate | FK: TemplateID, OnDelete:CASCADE | Template relationship |
| `Name` | varchar(100) | not null | Form name |
| `Description` | text | nullable | Form description |
| `FormLayout` | jsonb | not null | Structured form layout (see below) |
| `Version` | int | not null, default:1 | Incremented on formLayout change |
| `IsActive` | bool | not null, default:true | Must match parent template status |

**`FormLayout` JSONB structure:**

```json
{
  "sections": [
    {
      "id": "company_info",
      "title": "Company Information",
      "description": "Enter your company details",
      "fields": [
        {
          "id": "company_name",
          "type": "text",
          "label": "Company Name",
          "placeholder": "e.g., ABC Corp",
          "required": true,
          "maxLength": 200
        },
        {
          "id": "registration_number",
          "type": "text",
          "label": "Registration Number",
          "required": false
        },
        {
          "id": "business_type",
          "type": "select",
          "label": "Business Type",
          "required": true,
          "options": [
            { "value": "sole_proprietorship", "label": "Sole Proprietorship" },
            { "value": "partnership", "label": "Partnership" },
            { "value": "llc", "label": "Limited Liability Company" }
          ]
        },
        {
          "id": "founded_date",
          "type": "date",
          "label": "Date Founded",
          "required": false
        },
        {
          "id": "description",
          "type": "textarea",
          "label": "Business Description",
          "required": true,
          "maxLength": 2000
        }
      ]
    }
  ]
}
```

**Supported field types:** `text`, `textarea`, `number`, `email`, `phone`, `date`, `select`, `multiselect`, `checkbox`, `radio`

The frontend renders the form based on this schema. Form submission data is processed on the frontend only (Phase 1) — no backend storage of filled form data.

---

### 6. LibraryTemplateDownload

**Table:** `library_template_downloads`

Append-only log tracking which users downloaded which templates. Used for analytics, user download history, and download counting.

| Field | Type | Constraints | Description |
|---|---|---|---|
| `BaseModel` | embedded | — | ID, CreatedAt (= downloaded timestamp), UpdatedAt, DeletedAt |
| `AccountID` | uuid | FK → accounts, not null, index | Who downloaded |
| `TemplateID` | uuid | FK → library_templates, not null, index | Which language variant was downloaded |
| `GroupID` | uuid | FK → library_template_groups, not null, index | Denormalized for efficient group-level counting |

> Anonymous downloads (requiresAuth=false, unauthenticated user) only increment the group's `DownloadCount` — no log row is created since there is no AccountID.

---

## Entity Relationship Summary

```
LibraryCategory
  ├── LibraryCategoryTranslation (1:N, CASCADE)
  ├── ParentCategory (self, 1:N, SET NULL)
  └── LibraryTemplateGroup (1:N, RESTRICT)

LibraryTemplateGroup
  ├── LibraryTemplate (1:N, CASCADE)
  └── LibraryTemplateDownload (1:N, CASCADE)

LibraryTemplate
  ├── LibraryInteractiveForm (1:1, CASCADE)
  └── LibraryTemplateDownload (1:N, CASCADE)

Account
  ├── LibraryTemplateGroup (as CreatedBy, 1:N, RESTRICT)
  └── LibraryTemplateDownload (1:N, CASCADE)
```

## IAM Module Integration

The IAM module already owns `TemplatePreference` with `DefaultTemplate` and `EditorMode` fields. The library module reads this preference to highlight the user's default template. The preference is managed through IAM's `PreferenceUsecase.UpdateTemplatePreference`.

**Integration points:**
- Library view usecase reads IAM's `TemplatePreferenceRepository.GetByAccountID` to surface the default template
- Library admin routes use IAM's `AuthMiddleware`, `AccountStatusMiddleware`, and `PermissionMiddleware` for access control
- The library module's `CreatedBy` FK references the IAM `Account` entity
