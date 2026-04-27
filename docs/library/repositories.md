# Library Module — Repository Interfaces

All repository interfaces live in `domain/repository/` and are implemented in `infrastructure/repository/`. Each interface extends `sharedrepo.GenericRepository[T]` for standard CRUD operations, then adds domain-specific methods.

## Shared Pattern

Every repository implementation follows these conventions:

- Private struct embedding `sharedrepo.GenericRepository[T]`
- Constructor `New*Repository(db *core.Database, logger core.Logger)` returning the domain interface
- `getDB(ctx)` resolves a transaction from context (`core.TxFromContext`) or falls back to the raw DB
- `applyPaginationAndSorting()` helper for query options

---

## 1. LibraryCategoryRepository

**Entity:** `LibraryCategory`

**Standard CRUD:** inherited from `GenericRepository[LibraryCategory]`

**Domain-specific methods:**

| Method | Signature | Description |
|---|---|---|
| `GetBySlug` | `(ctx context.Context, parentID *uuid.UUID, slug string) (*LibraryCategory, error)` | Lookup by slug within a parent. parentID=nil for root categories. |
| `ListTree` | `(ctx context.Context, includeInactive bool) ([]*LibraryCategory, error)` | Returns all categories flat, sorted by SortOrder ASC. Used for client-side tree building. |
| `ListActive` | `(ctx context.Context, q query.QueryOptions) ([]*LibraryCategory, error)` | List only active categories with pagination. |
| `GetTranslations` | `(ctx context.Context, categoryID uuid.UUID) ([]*LibraryCategoryTranslation, error)` | Fetch all translations for a category. |
| `UpsertTranslation` | `(ctx context.Context, translation *LibraryCategoryTranslation) error` | Create or update a translation (unique on categoryID + language). |
| `DeleteTranslation` | `(ctx context.Context, categoryID uuid.UUID, language string) error` | Remove a specific translation. |

---

## 2. LibraryTemplateGroupRepository

**Entity:** `LibraryTemplateGroup`

**Standard CRUD:** inherited from `GenericRepository[LibraryTemplateGroup]`

**Domain-specific methods:**

| Method | Signature | Description |
|---|---|---|
| `GetBySlug` | `(ctx context.Context, categoryID *uuid.UUID, slug string) (*LibraryTemplateGroup, error)` | Lookup by slug within a category. categoryID=nil for cross-category lookup. |
| `ListByCategory` | `(ctx context.Context, categoryID uuid.UUID, q query.QueryOptions) ([]*LibraryTemplateGroup, error)` | Paginated groups in a category. Applies format filter, search, and SortOrder. |
| `ListByFormat` | `(ctx context.Context, format TemplateFormat, q query.QueryOptions) ([]*LibraryTemplateGroup, error)` | Filter groups by format type (e.g., all PDF templates). |
| `IncrementDownloadCount` | `(ctx context.Context, id uuid.UUID) error` | Atomic increment of `DownloadCount` by 1. Uses `UPDATE ... SET download_count = download_count + 1 WHERE id = ?` |

**Custom query notes:**
- `ListByCategory` and `ListByFormat` should join with category and preload templates for language availability info
- All list methods should exclude inactive groups by default (unless `IncludeArchived` is set)

---

## 3. LibraryTemplateRepository

**Entity:** `LibraryTemplate`

**Standard CRUD:** inherited from `GenericRepository[LibraryTemplate]`

**Domain-specific methods:**

| Method | Signature | Description |
|---|---|---|
| `GetByGroupAndLanguage` | `(ctx context.Context, groupID uuid.UUID, language string) (*LibraryTemplate, error)` | Find a specific language variant for a group. Returns nil if variant doesn't exist. |
| `ListByGroup` | `(ctx context.Context, groupID uuid.UUID) ([]*LibraryTemplate, error)` | All language variants for a group. Used to populate the language picker. |
| `FindActiveByGroup` | `(ctx context.Context, groupID uuid.UUID) ([]*LibraryTemplate, error)` | Only active variants. Used for user-facing queries. |

---

## 4. LibraryInteractiveFormRepository

**Entity:** `LibraryInteractiveForm`

**Standard CRUD:** inherited from `GenericRepository[LibraryInteractiveForm]`

**Domain-specific methods:**

| Method | Signature | Description |
|---|---|---|
| `GetByTemplateID` | `(ctx context.Context, templateID uuid.UUID) (*LibraryInteractiveForm, error)` | Get the interactive form attached to a template. Returns nil for non-interactive templates. |

---

## 5. LibraryTemplateDownloadRepository

**Entity:** `LibraryTemplateDownload`

**Standard CRUD:** inherited from `GenericRepository[LibraryTemplateDownload]`

**Domain-specific methods:**

| Method | Signature | Description |
|---|---|---|
| `ListByAccount` | `(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*LibraryTemplateDownload, error)` | Paginated download history for a user, ordered by CreatedAt DESC. |
| `CountByGroup` | `(ctx context.Context, groupID uuid.UUID) (int64, error)` | Count downloads for a specific group (for admin analytics). |
| `ListAll` | `(ctx context.Context, q query.QueryOptions) ([]*LibraryTemplateDownload, error)` | Admin: list all downloads with pagination, ordered by CreatedAt DESC. |
