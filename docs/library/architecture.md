# Library Module — Architecture

## Module Directory Structure

```
internal/modules/library/
├── module.go                              # Uber FX DI wiring
├── entities.go                            # EntityProvider for SchemaManager
│
├── domain/
│   ├── entity/
│   │   ├── enums.go                       # TemplateFormat, TierAccess
│   │   ├── library_category.go
│   │   ├── library_category_translation.go
│   │   ├── library_template_group.go
│   │   ├── library_template.go
│   │   ├── library_interactive_form.go
│   │   └── library_template_download.go
│   │
│   ├── repository/
│   │   ├── category_repository.go
│   │   ├── template_group_repository.go
│   │   ├── template_repository.go
│   │   ├── interactive_form_repository.go
│   │   └── template_download_repository.go
│   │
│   ├── usecase/
│   │   ├── inputs.go                     # All input DTOs
│   │   ├── library_view_usecase.go
│   │   ├── library_admin_usecase.go
│   │   └── tier_service.go              # Cross-module interface
│   │
│   └── error/
│       └── errors.go                     # Domain-specific errors
│
├── application/
│   ├── usecase/
│   │   ├── library_view_usecase.go       # Implementation
│   │   └── library_admin_usecase.go      # Implementation
│   │
│   └── service/
│       ├── library_service.go            # Orchestrates storage + use cases
│       └── template_file_validator.go    # File type/size validation
│
├── infrastructure/
│   ├── repository/
│   │   ├── helpers.go                    # applyPaginationAndSorting, getDB
│   │   ├── category_repository.go
│   │   ├── template_group_repository.go
│   │   ├── template_repository.go
│   │   ├── interactive_form_repository.go
│   │   └── template_download_repository.go
│   │
│   └── tier/
│       └── tier_service_stub.go          # Stub implementation of TierService
│
└── delivery/
    ├── handler/
    │   ├── library_handler.go            # User-facing: browse, download
    │   └── library_admin_handler.go      # Admin: CRUD, upload, logs
    │
    ├── dto/
    │   ├── library_dto.go                # User-facing request/response DTOs
    │   └── library_admin_dto.go          # Admin request/response DTOs
    │
    └── routes/
        ├── routes.go                     # RouteDependencies + RegisterRoutes
        ├── library_routes.go             # User-facing routes
        └── library_admin_routes.go       # Admin routes
```

---

## Dependency Injection (module.go)

Follows the same Uber FX pattern as IAM, community, and guide modules.

### Providers

| Layer | Provide | Bind To |
|---|---|---|
| Entity | `NewEntityProvider` | — |
| Repository | `infrarepo.NewLibraryCategoryRepository` | `repository.LibraryCategoryRepository` |
| Repository | `infrarepo.NewLibraryTemplateGroupRepository` | `repository.LibraryTemplateGroupRepository` |
| Repository | `infrarepo.NewLibraryTemplateRepository` | `repository.LibraryTemplateRepository` |
| Repository | `infrarepo.NewLibraryInteractiveFormRepository` | `repository.LibraryInteractiveFormRepository` |
| Repository | `infrarepo.NewLibraryTemplateDownloadRepository` | `repository.LibraryTemplateDownloadRepository` |
| Usecase | `appusecase.NewLibraryViewUsecase` | `usecase.LibraryViewUsecase` |
| Usecase | `appusecase.NewLibraryAdminUsecase` | `usecase.LibraryAdminUsecase` |
| Service | `service.NewLibraryService` | `service.LibraryService` |
| Service | `service.NewTemplateFileValidator` | — |
| Infrastructure | `tier.NewTierServiceStub` | `usecase.TierService` |
| Handler | `handler.NewLibraryHandler` | — |
| Handler | `handler.NewLibraryAdminHandler` | — |

### Invocations

| Invoke | Purpose |
|---|---|
| Register EntityProvider with SchemaManager | Enable migration generation |
| Register HTTP routes | Wire handlers with middleware |

---

## Cross-Module Dependencies

### Dependencies FROM library module TO other modules

| Dependency | Interface | Purpose |
|---|---|---|
| IAM | `AuthMiddleware` | JWT auth for routes |
| IAM | `AccountStatusMiddleware` | Account status check for routes |
| IAM | `PermissionMiddleware` | Admin permission enforcement |
| IAM | `TemplatePreferenceRepository` | Read user's default template preference |
| IAM | `RoleAssignmentUsecase` | Permission checks via middleware |
| Core | `*core.Database` | DB access, Transactor |
| Core | `core.Logger` | Logging |
| Core | `core.Config` | App configuration |
| Pkg | `storage.Storage` | File upload, download, delete operations |
| Pkg | `query.QueryOptions` | Pagination, search, filtering |

### Dependencies FROM other modules TO library module

None in Phase 1. The library module may publish RabbitMQ events (`library.template.downloaded`) in a future phase for the notification module to subscribe to.

---

## Route Structure

### User-facing routes (`/api/v1/library/...`)

All require `AuthMiddleware` + `AccountStatusMiddleware`, except for `requiresAuth=false` templates which skip auth. The `requiresAuth` check is done in the view usecase at download time, not at the middleware level — the middleware always runs but the handler may still serve unauthenticated users for public templates.

| Method | Route | Handler | Description |
|---|---|---|---|
| GET | `/categories` | ListCategories | Active category tree, optionally localized |
| GET | `/templates` | ListTemplateGroups | Paginated groups with filters (category, format, search) |
| GET | `/templates/{slug}` | GetTemplateGroup | Group details with available language variants |
| GET | `/templates/{slug}/download` | DownloadTemplate | Generate download URL (query: language) |
| GET | `/downloads` | ListMyDownloads | User's download history |

### Admin routes (`/api/v1/admin/library/...`)

Require `AuthMiddleware` + `AccountStatusMiddleware` + `PermissionMiddleware` with appropriate permissions.

#### Categories

| Method | Route | Handler | Description |
|---|---|---|---|
| GET | `/categories` | ListAllCategories | All categories (including inactive) |
| POST | `/categories` | CreateCategory | Create category |
| GET | `/categories/{id}` | GetCategory | Category with translations |
| PATCH | `/categories/{id}` | UpdateCategory | Update category |
| DELETE | `/categories/{id}` | DeleteCategory | Soft-delete category |
| POST | `/categories/{id}/translations` | AddCategoryTranslation | Add translation |
| PATCH | `/categories/{id}/translations/{lang}` | UpdateCategoryTranslation | Update translation |
| DELETE | `/categories/{id}/translations/{lang}` | DeleteCategoryTranslation | Remove translation |

#### Template Groups

| Method | Route | Handler | Description |
|---|---|---|---|
| GET | `/template-groups` | ListAllTemplateGroups | All groups (including inactive) |
| POST | `/template-groups` | CreateTemplateGroup | Create group with optional thumbnail |
| GET | `/template-groups/{id}` | GetTemplateGroup | Group with all templates |
| PATCH | `/template-groups/{id}` | UpdateTemplateGroup | Update group |
| DELETE | `/template-groups/{id}` | DeleteTemplateGroup | Soft-delete group |

#### Templates

| Method | Route | Handler | Description |
|---|---|---|---|
| GET | `/template-groups/{groupId}/templates` | ListTemplatesByGroup | All templates for a group |
| POST | `/template-groups/{groupId}/templates` | CreateTemplate | Upload file + create variant |
| GET | `/templates/{id}` | GetTemplate | Template details |
| PATCH | `/templates/{id}` | UpdateTemplate | Update metadata or replace file |
| DELETE | `/templates/{id}` | DeleteTemplate | Soft-delete template |

#### Interactive Forms

| Method | Route | Handler | Description |
|---|---|---|---|
| GET | `/templates/{templateId}/interactive-form` | GetInteractiveForm | Get form definition |
| POST | `/templates/{templateId}/interactive-form` | CreateInteractiveForm | Create form for interactive template |
| PATCH | `/interactive-forms/{id}` | UpdateInteractiveForm | Update form layout |
| DELETE | `/interactive-forms/{id}` | DeleteInteractiveForm | Soft-delete form |

#### Monitoring

| Method | Route | Handler | Description |
|---|---|---|---|
| GET | `/downloads` | GetDownloadLogs | Paginated download history |

---

## File Storage Convention

All files are stored via `storage.Storage` using the following key conventions:

| Content | Key Pattern | Example |
|---|---|---|
| Template file | `library/templates/{templateID}.{ext}` | `library/templates/a1b2c3d4-...pdf` |
| Group thumbnail | `library/thumbnails/{groupID}.{ext}` | `library/thumbnails/e5f6g7h8-...jpg` |

**Lifecycle:**
- **Create:** Upload with `storage.Upload`, store the returned key and URL on the entity
- **Replace:** Upload new file with same key pattern → delete old key from storage → update DB
- **Delete (hard):** Delete key from storage (only on `HardDelete`, not soft-delete)
- **Download:** Generate presigned URL via `storage.GetPresignedURL(ctx, key, 5*time.Minute)`
- **Fallback URL:** If `FileInfo.URL` is empty (MinIO), construct as `/api/v1/files/{key}`

---

## Download Flow Diagram

```
┌──────────┐    ┌───────────────┐    ┌────────────┐    ┌─────────┐
│  Client   │    │ View Usecase  │    │  Storage   │    │   DB    │
│           │───>│               │───>│            │    │         │
│ GET /down-│    │ 1. Lookup     │    │ 4. Generate│    │ 2. Load │
│ load/{slug│    │    group      │    │    presigned│    │    group│
│ }?lang=en │    │ 2. Check auth │    │    URL     │    │ 3. Load │
│           │    │    + tier     │    │            │    │    tmpl │
│           │    │ 3. Find var-  │    │            │    │ 6. Incr │
│           │    │    iant       │    │            │    │    count│
│           │    │ 5. Presigned  │    │            │    │ 7. Log  │
│           │<───│    URL        │    │            │    │    dl   │
│ Redirect  │    │               │    │            │    │         │
│ to URL    │───>│               │    │            │    │         │
└──────────┘    └───────────────┘    └────────────┘    └─────────┘
```

---

## Event Publishing

The library module can publish the following events in Phase 1:

| Event | When Published | Payload |
|---|---|---|
| `library.template.downloaded` | On successful download (authenticated users) | `{ accountID, groupID, templateID, language }` |

The notification module can subscribe to these events in a future phase to send alerts or admin notifications.

---

## Module Registration

Add to `internal/modules/modules.go`:

```go
var Modules = fx.Options(
    ai.Module,
    iam.Module,
    guide.Module,
    community.Module,
    coregrpc.Module,
    library.Module,   // new
)
```
