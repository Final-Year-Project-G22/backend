# Unified Business Taxonomy — Change Log

> Generated: 2025-05-06  
> Covers: Phase 0 (Planning) through Phase 10 (Cleanup)  
> Target audiences: Mobile, Web, Backend

---

## Overview

Replaced category-based content classification with a unified **sector + tag + region + stage** taxonomy across all modules. Categories were dropped from the database and removed from code. Five distinct taxonomy dimensions now power content targeting:

| Dimension | Description | Storage | Example |
|-----------|-------------|---------|---------|
| Sector | Hierarchical business sector (with subtree matching) | `uuid[]` on content, `uuid` on profile | Agriculture > Crop Farming |
| Tag | Flat, grouped tags with single/multi-select rules | `uuid[]` on content and profile | LicenseRequired, ImportExport |
| Region | Geographic region | `varchar(50)` nullable | addis_ababa |
| Stage | Business maturity stage | `varchar(50)` nullable | idea, startup, growth, established |

---

## Module-by-Module Changes

### 1. Shared Taxonomy Package (`internal/shared/taxonomy`)

**Added:**  
`types.go` — `TaxonomyFilter`, `ProfileTaxonomy` structs  
`matcher.go` — `MatchSector()`, `MatchTags()`, `MatchRegion()`, `MatchStage()`, `MatchAll()`  
`matcher_test.go` — 28 unit tests

**Design rules:**
- Empty `sector_ids`/`tag_ids` arrays on content = "match everyone" (no restriction)
- `nil` profile sector/region/stage = "exclude restricted-only content"
- Sector subtree matching via `ancestor_ids` overlap
- Tag matching: any-of within group, all-of across groups

---

### 2. IAM Module (`internal/modules/iam`)

#### Schema & Entities

| Entity | Added | Removed | Notes |
|--------|-------|---------|-------|
| `Sector` | `ParentID *uuid.UUID`, `AncestorIDs []uuid.UUID` | — | Self-referential hierarchy; `AncestorIDs` enables subtree overlap queries |
| `Tag` | `Group TagGroup`, `IsMultiSelect bool` | — | Five groups: REGION, BUSINESS_TYPE, LEGAL_FORM, REGULATORY, SIZE_STAGE |
| `BusinessProfile` | `SectorID *uuid.UUID`, `Tags []Tag`, `Region *Region`, `Stage *BusinessStage` | — | Single-sector, multi-tag profile |

#### New Tables

| Table | Purpose |
|-------|---------|
| `sector_translations` | EN/AM translations for sectors |
| `tag_translations` | EN/AM translations for tags |

#### Seed Data

- 18 sectors (6 roots, 12 children) with EN/AM translations
- 26 tags across 5 groups with EN/AM translations

#### Business Logic

| Feature | Description |
|---------|-------------|
| **Tag group validation** | `validateTags()` enforces: single-select groups allow only 1 tag; multi-select unlimited |
| **Profile CRUD** | `BusinessProfileUsecase` with Create/Get/Update/UpdateSocialLinks, tag validation |

#### Admin API (NEW)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/v1/admin/sectors` | List sectors with EN/AM translations |
| GET | `/api/v1/admin/sectors/{id}` | Get single sector |
| POST | `/api/v1/admin/sectors` | Create sector |
| PUT | `/api/v1/admin/sectors/{id}` | Update sector |
| DELETE | `/api/v1/admin/sectors/{id}` | Soft-delete sector |
| GET | `/api/v1/admin/tags` | List tags with EN/AM translations |
| GET | `/api/v1/admin/tags/{id}` | Get single tag |
| POST | `/api/v1/admin/tags` | Create tag |
| PUT | `/api/v1/admin/tags/{id}` | Update tag |
| DELETE | `/api/v1/admin/tags/{id}` | Soft-delete tag |

**NOT DONE:**
- `AncestorIDs` are not auto-computed when creating/updating sectors (parent-child hierarchy traversal needed)
- No validation that sector/tag IDs exist when referenced by content modules
- No resolved sector/tag names in guide/thread responses (only admin endpoints resolve names)

---

### 3. Guides Module (`internal/modules/guide`)

#### Entity Changes

| Entity | Added | Removed |
|--------|-------|---------|
| `Guide` | `SectorIDs []uuid.UUID`, `TagIDs []uuid.UUID` | `CategoryID uuid.UUID` |

#### Removed (Categories)

- `GuideCategory` entity file
- `GuideCategoryCondition` entity file
- `GuideCategoryTranslation` entity file
- `CategoryRepository` interface + implementation
- All category admin endpoints (7 routes)
- All category admin handler methods (7 methods)
- All category usecase methods (`CreateCategory`, `UpdateCategory`, `DeleteCategory`, etc.)
- `CreateCategoryInput`, `UpdateCategoryInput` types

#### Repository

| Change | Old | New |
|--------|-----|-----|
| Method rename | `ListByCategory(ctx, categoryID, q, locale)` | `ListByTaxonomy(ctx, sectorIDs, tagIDs, q, locale)` |
| Query filter | `WHERE category_id = ?` | `WHERE sector_ids && ? AND tag_ids && ?` (PostgreSQL array overlap) |
| `GetBySlug` | `(ctx, categoryID, slug, locale)` | `(ctx, slug, locale)` — no category filter |

#### API Changes

| Method | Old Endpoint | New Endpoint | Auth |
|--------|-------------|-------------|------|
| GET | `/api/v1/guides/categories/tree` | `/api/v1/guides` | Required |
| GET | `/api/v1/admin/guides/categories/tree` | — (removed) | — |
| POST | `/api/v1/admin/guides/categories` | — (removed) | — |
| PUT | `/api/v1/admin/guides/categories/{id}` | — (removed) | — |
| DELETE | `/api/v1/admin/guides/categories/{id}` | — (removed) | — |

**Key behavior change:** `GET /api/v1/guides` now filters by the authenticated user's BusinessProfile taxonomy. If no profile exists, all guides are returned.

**Guide creation DTOs** already accept `sectorIds` and `tagIds` (JSON fields unchanged from earlier schema migration).

**NOT DONE:**
- Resolved sector/tag names not included in guide list responses (only raw UUID arrays)
- No validation that referenced sector/tag IDs exist

---

### 4. Community Module (`internal/modules/community`)

#### Entity Changes

| Entity | Added | Removed |
|--------|-------|---------|
| `DiscussionThread` | `SectorIDs []uuid.UUID`, `TagIDs []uuid.UUID` | `CategoryID uuid.UUID` |

#### Repository

| Change | Old | New |
|--------|-----|-----|
| Method rename | `ListByCategory(ctx, categoryID, q)` | `ListByTaxonomy(ctx, sectorIDs, tagIDs, q)` |
| Query filter | `WHERE category_id = ?` | `WHERE sector_ids && ? AND tag_ids && ?` |
| `Search` param | `(ctx, keyword, categoryID, q)` | `(ctx, keyword, sectorIDs, tagIDs, q)` |

#### API Changes

| Method | Old Endpoint | New Endpoint | Auth |
|--------|-------------|-------------|------|
| GET | `/api/v1/community/categories/{id}/threads` | — (removed) | — |
| GET | `/api/v1/community/threads` (with `categoryId` query) | `/api/v1/community/threads` (no categoryId) | Required |

**Key behavior change:** Thread listing now filters by the user's BusinessProfile taxonomy.

**DTO changes:** Removed `categoryId` from `ListThreadsInput` and `SearchThreadsInput`.

**NOT DONE:**
- `CommunityCategory` entity, repository, usecase, handlers, and routes **retained** (used by follow/unfollow category feature)
- Category follow/unfollow endpoints still work against `community_categories` table
- Thread creation already accepts `sectorIds`/`tagIds`

---

### 5. AI Module (`internal/modules/ai`)

#### Entity Changes

| Entity | Added | Removed |
|--------|-------|---------|
| `IngestionDocument` | `SectorIDs []uuid.UUID`, `TagIDs []uuid.UUID`, `Region *string`, `Stage *string` | — |

#### Ingestion (Upload) API

| Field | Type | Description |
|-------|------|-------------|
| `sectorIds` | `[]uuid` | Target sector IDs for document taxonomy |
| `tagIds` | `[]uuid` | Target tag IDs for document taxonomy |
| `region` | `string?` | Target region |
| `stage` | `string?` | Target business stage |

All added to `FinalizeUploadRequest` DTO. Stored on `IngestionDocument` at finalize.

#### AI Retrieval (Ask)

**Backend changes:**
- `AskRequest` port struct extended with `SectorIDs`, `TagIDs`, `Region`, `Stage`
- `AskService` fetches user's `BusinessProfile` and builds taxonomy filter
- gRPC client passes taxonomy as **metadata headers**:

| Metadata Header | Format |
|----------------|--------|
| `x-taxonomy-sector-ids` | Comma-separated UUIDs |
| `x-taxonomy-tag-ids` | Comma-separated UUIDs |
| `x-taxonomy-region` | String |
| `x-taxonomy-stage` | String |

**NOT DONE:**
- Proto file (`service.proto`) not yet updated with taxonomy fields
- Inference (Python/gRPC) service needs to read metadata and apply vector search filters
- No user-override on AI query taxonomy (always auto-filtered by profile)
- Chunk propagation: taxonomy from document to chunks is handled by inference service, not backend

---

### 6. Notifications Module (`internal/modules/notification`)

#### Entity Changes

| Entity | Added | Removed |
|--------|-------|---------|
| `NotificationTemplate` | `TemplateGroup string` | `Category NotificationCategory` |
| `NotificationCampaign` | `SectorIDs []uuid.UUID`, `TagIDs []uuid.UUID`, `Region *string`, `Stage *string` | — |
| `UserNotificationInbox` | — | `Category` column |

#### Removed

| Item | Description |
|------|-------------|
| `NotificationCategory` enum | All 7 category constants deleted |
| `MarkCategoryAsRead` | Usecase, repository, handler, DTO, route all removed |
| `ListTemplates` category filter | Replaced with `templateGroup` query param |

#### Campaign API

| Field | Type | Endpoints |
|-------|------|-----------|
| `sectorIds` | `[]uuid` | CreateCampaign, UpdateCampaign, CampaignDetail |
| `tagIds` | `[]uuid` | CreateCampaign, UpdateCampaign, CampaignDetail |
| `region` | `string?` | CreateCampaign, UpdateCampaign, CampaignDetail |
| `stage` | `string?` | CreateCampaign, UpdateCampaign, CampaignDetail |

Populated at campaign creation, updatable in draft state.

#### Template API

| Field | Old | New |
|-------|-----|-----|
| Filter query | `category` | `templateGroup` |

**NOT DONE:**
- Delivery usecase: campaign dispatch-time taxonomy matching not implemented (AccountReader needs taxonomy extension)
- Campaign processor: `ResolveSegment` still uses JSON `TargetSegment`, not taxonomy fields

---

## Database Migrations

### Schema Migration (`20260506135016`)

**Added columns:**

| Table | New Column | Index |
|-------|-----------|-------|
| `sectors` | `ancestor_ids uuid[]` | GIN |
| `guides` | `sector_ids uuid[]`, `tag_ids uuid[]` | GIN on both |
| `discussion_threads` | `sector_ids uuid[]`, `tag_ids uuid[]` | GIN on both |
| `ingestion_documents` | `sector_ids uuid[]`, `tag_ids uuid[]`, `region`, `stage` | GIN on arrays |
| `notification_campaigns` | `sector_ids uuid[]`, `tag_ids uuid[]`, `region`, `stage` | GIN on arrays |
| `notification_templates` | `template_group` | B-tree |

**Dropped columns:**

| Table | Dropped Column | Indexes Dropped |
|-------|---------------|-----------------|
| `guides` | `category_id` | `idx_guides_category`, `idx_guides_slug_per_category` |
| `discussion_threads` | `category_id` | `idx_discussion_threads_category`, `idx_discussion_threads_slug_per_category` |
| `notification_templates` | `category` | `idx_notif_templates_category` |
| `user_notification_inboxes` | `category` | `idx_notif_inbox_category` |

### Translation Tables Migration (`20260506142106`)

**New tables:** `sector_translations`, `tag_translations` with GIN indexes (fixed from Atlas-generated B-tree)

### Seed Migration (`20260506142200`)

18 sectors, 26 tags, EN/AM translations for all.

---

## API Contract Summary for Mobile/Web

### New Endpoints

| Method | Endpoint | Module |
|--------|----------|--------|
| GET | `/api/v1/admin/sectors` | IAM |
| GET | `/api/v1/admin/sectors/{id}` | IAM |
| POST | `/api/v1/admin/sectors` | IAM |
| PUT | `/api/v1/admin/sectors/{id}` | IAM |
| DELETE | `/api/v1/admin/sectors/{id}` | IAM |
| GET | `/api/v1/admin/tags` | IAM |
| GET | `/api/v1/admin/tags/{id}` | IAM |
| POST | `/api/v1/admin/tags` | IAM |
| PUT | `/api/v1/admin/tags/{id}` | IAM |
| DELETE | `/api/v1/admin/tags/{id}` | IAM |

### Removed Endpoints

| Method | Old Endpoint | Module |
|--------|-------------|--------|
| GET | `/api/v1/guides/categories/tree` | Guide |
| GET | `/api/v1/admin/guides/categories/tree` | Guide |
| POST | `/api/v1/admin/guides/categories` | Guide |
| PUT | `/api/v1/admin/guides/categories/{id}` | Guide |
| DELETE | `/api/v1/admin/guides/categories/{id}` | Guide |
| POST/PUT/DELETE | `/api/v1/admin/guides/categories/*` (all category routes) | Guide |
| GET | `/api/v1/community/categories/{id}/threads` | Community |
| POST | `/api/v1/notification/inbox/category/{category}/read` | Notification |

### Changed Endpoints

| Method | Endpoint | Change | Module |
|--------|----------|--------|--------|
| GET | `/api/v1/guides` | Now returns taxonomy-filtered guides (auth required); previously was `/api/v1/guides/categories/tree` (unauth, category tree) | Guide |
| GET | `/api/v1/community/threads` | No longer accepts `categoryId` query param; auto-filters by profile taxonomy | Community |
| GET | `/api/v1/community/threads/search` | No longer accepts `categoryId` query param; auto-filters by profile taxonomy | Community |
| POST | `/api/v1/ai/ingestion/finalize` | New optional body fields: `sectorIds`, `tagIds`, `region`, `stage` | AI |
| GET | `/api/v1/admin/notifications/templates` | Query param changed from `category` to `templateGroup` | Notification |
| POST/PUT | `/api/v1/admin/notifications/campaigns` | New optional body fields: `sectorIds`, `tagIds`, `region`, `stage` | Notification |

### New Response Fields

All responses for guides, threads, campaigns now include:

| Field | Type | Example |
|-------|------|---------|
| `sectorIds` | `array<uuid>` | `["uuid1", "uuid2"]` |
| `tagIds` | `array<uuid>` | `["uuid1"]` |

Campaign responses additionally include `region` (string?) and `stage` (string?).

---

## Deferred / Not Done

### Phase-by-Phase

| Phase | Item | Reason |
|-------|------|--------|
| 4 (IAM) | Sector slug uniqueness validation | Deferred to admin endpoints |
| 4 (IAM) | Resolved sector/tag names in profile responses | No profile handler/DTOs yet |
| 5 (Guides) | Resolved sector/tag names in guide list | Requires cross-module lookup service |
| 5 (Guides) | Validation of sector/tag IDs on guide create | Requires cross-module ID validation |
| 6 (Community) | Remove `CommunityCategory` entity | Still used by follow/unfollow |
| 6 (Community) | Resolved sector/tag names | Same as guides |
| 7 (AI) | Proto file taxonomy fields | Needs coordination with ML team |
| 7 (AI) | Chunk taxonomy propagation | Inference service responsibility |
| 7 (AI) | User taxonomy overrides in AI queries | Feature not designed yet |
| 8 (Notifications) | Campaign dispatch-time taxonomy matching | AccountReader needs taxonomy extension |
| 8 (Notifications) | Delivery worker taxonomy evaluation | Needs AccountReader changes |
| 9 (Admin) | Sector `ancestor_ids` auto-computation | Parent hierarchy traversal needed |
| 9 (Admin) | OpenAPI spec regeneration | Huma auto-generates on startup |
| All | Integration/E2E tests | No test infrastructure built yet |

### Known Gaps

1. **OpenAPI spec** (`docs/openapi.json`) not regenerated after Phase 10 category removals — will automatically update on next Huma startup
2. **Mobile/Web mapping**: Mobile must map old category IDs to new sector/tag concepts (backward compatibility not maintained)
3. **Migration rollback**: No down migration prepared — this is a one-way change
4. **Ancestor auto-fill**: Creating a child sector does not auto-populate `ancestor_ids` from parent

---

## UI Migration Guide

### How to Migrate from Categories to Sectors/Tags

**No backward compatibility.** All category endpoints are removed. UI must transition to the new taxonomy model.

---

### Step 1: Fetch Seed Taxonomy Data

UI should call the admin sector/tag APIs at startup (or periodically) to populate local picker UIs. UUIDs are generated by the database — do not hardcode.

```
GET /api/v1/admin/sectors      → use response for sector picker
GET /api/v1/admin/tags          → use response for tag picker
```

Response shapes:

<details><summary>Sector Response (click to expand)</summary>

```json
{
  "id": "uuid",
  "slug": "trade",
  "parentId": "uuid-or-null",
  "nameEn": "Trade",
  "nameAm": "ንግድ",
  "descEn": "Retail and wholesale distribution of goods",
  "descAm": "የእቃ መሸጥ እና ማከፋፈል ስራዎች",
  "createdAt": "2025-05-06T...",
  "updatedAt": "2025-05-06T..."
}
```

Parent-child logic: sector has a parent if `parentId` is non-null. Root sectors (`trade`, `manufacturing`, `services`, `agriculture`, `construction`) have `parentId: null`.

</details>

<details><summary>Tag Response (click to expand)</summary>

```json
{
  "id": "uuid",
  "slug": "sole-proprietor",
  "group": "LEGAL_STRUCTURE",
  "isMultiSelect": false,
  "nameEn": "Sole Proprietor",
  "nameAm": "የግል ማህበር",
  "createdAt": "2025-05-06T...",
  "updatedAt": "2025-05-06T..."
}
```

Tag selection rules:
- `isMultiSelect: true` → user can pick any number from this group
- `isMultiSelect: false` → user MUST pick exactly zero or one from this group
</details>

---

### Step 2: Reference Data Tables

#### Sector Hierarchy

| Slug | Parent | Name EN | Name AM |
|------|--------|---------|---------|
| `trade` | — | Trade | ንግድ |
| `retail` | trade | Retail | ችርቻሮ |
| `wholesale` | trade | Wholesale | ሙሉ ሻጭ |
| `manufacturing` | — | Manufacturing | ማምረቻ |
| `food-beverage` | manufacturing | Food & Beverage | ምግብና መጠጥ |
| `textiles-apparel` | manufacturing | Textiles & Apparel | ጨርቃ ጨርቅ |
| `leather` | manufacturing | Leather | ቆዳ ሥራ |
| `wood-metal` | manufacturing | Wood & Metal | እንጨትና ብረታ ብረት |
| `services` | — | Services | አገልግሎት |
| `it-tech` | services | IT & Tech | ኢንፎርሜሽን ቴክኖሎጂ |
| `hospitality` | services | Hospitality | መስተንግዶ |
| `consulting` | services | Consulting | አማካሪ |
| `transport-logistics` | services | Transport & Logistics | መጓጓዣ |
| `agriculture` | — | Agriculture | ግብርና |
| `crop-farming` | agriculture | Crop Farming | የሰብል እርሻ |
| `livestock-poultry` | agriculture | Livestock & Poultry | እንስሳት |
| `construction` | — | Construction | ኮንስትራክሽን |
| `contracting` | construction | Contracting | ተቋራጭነት |
| `construction-mat` | construction | Construction Materials | የግንባታ እቃዎች |

#### Tag Groups

| Group | Select | Tags (slug → EN name) |
|-------|--------|----------------------|
| `LEGAL_STRUCTURE` | Single | `sole-proprietor` → Sole Proprietor, `plc` → PLC, `share-company` → Share Company, `partnership` → Partnership, `cooperative` → Cooperative |
| `TAX_STATUS` | Single | `tax-vat` → VAT Payer, `tax-tot` → TOT Payer, `tax-excise` → Excise Tax, `tax-exempt` → Tax Exempt |
| `GENERAL_OPERATIONS` | Multi | `op-importer` → Importer, `op-exporter` → Exporter, `op-tender` → Tender Participant, `op-food-handling` → Food Handling, `op-vehicles` → Vehicles, `op-hazardous` → Hazardous Materials, `op-ecommerce` → E-Commerce, `op-home-based` → Home-Based, `op-creates-ip` → Creates IP |
| `EMPLOYMENT` | Single | `has-employees` → Has Employees, `no-employees` → No Employees |
| `DEMOGRAPHICS` | Multi | `demo-women-owned` → Women-Owned, `demo-youth` → Youth Enterprise, `demo-investor` → Investor |

#### Region Values

| Value | Display |
|-------|---------|
| `ADDIS_ABABA` | Addis Ababa |
| `DIRE_DAWA` | Dire Dawa |
| `OROMIA` | Oromia |
| `AMHARA` | Amhara |
| `SIDAMA` | Sidama |
| `SOMALI` | Somali |
| `TIGRAY` | Tigray |
| `AFAR` | Afar |
| `HARARI` | Harari |
| `BENISHANGUL_GUMUZ` | Benishangul-Gumuz |
| `SWEPR` | South West Ethiopia |
| `CENTRAL_ETHIOPIA` | Central Ethiopia |
| `SOUTH_ETHIOPIA` | South Ethiopia |
| `FEDERAL` | Federal |

#### Business Stage Values

| Value | Display |
|-------|---------|
| `IDEA` | Idea Stage |
| `REGISTRATION` | Registration |
| `OPERATIONAL` | Operational |
| `SCALING` | Scaling |

---

### Step 3: Before/After JSON Examples

#### Guides — Listing

**Before (category tree):**
```
GET /api/v1/guides/categories/tree
```
```json
{
  "categories": [{
    "id": "uuid",
    "slug": "legal",
    "name": "Legal Guides",
    "children": [...],
    "guides": [{
      "id": "uuid",
      "slug": "...",
      "name": "...",
      "icon": "..."
    }]
  }]
}
```

**After (flat, taxonomy-filtered, auth required):**
```
GET /api/v1/guides
Authorization: Bearer <token>
```
```json
{
  "guides": [{
    "id": "uuid",
    "slug": "register-your-business",
    "name": "Register Your Business",
    "description": "Step-by-step guide...",
    "icon": "govt-office",
    "sectorIds": ["uuid1", "uuid2"],
    "tagIds": ["uuid3"]
  }]
}
```

**Key changes:**
- Tree structure → flat list (UI must build tree from sector hierarchy locally)
- No auth → auth required (user's profile determines which guides appear)
- Category nesting removed → guides carry `sectorIds` for filtering

---

#### Guides — Admin Create

**Before:**
```json
{
  "slug": "my-guide",
  "sortOrder": 0,
  "translations": [...],
  "conditions": [...]
}
```

**After:**
```json
{
  "sectorIds": ["uuid-sector-1", "uuid-sector-2"],
  "tagIds": ["uuid-tag-1"],
  "slug": "my-guide",
  "sortOrder": 0,
  "translations": [...],
  "conditions": [...]
}
```

---

#### Community — Thread List

**Before:**
```
GET /api/v1/community/threads?categoryId=<uuid>
```

**After:**
```
GET /api/v1/community/threads
Authorization: Bearer <token>
```
(Filters by user's BusinessProfile taxonomy automatically — no query param needed)

Thread response already includes `sectorIds` and `tagIds`:
```json
{
  "threads": [{
    "id": "uuid",
    "title": "How to register VAT?",
    "sectorIds": ["uuid"],
    "tagIds": ["uuid"],
    "authorDisplayName": "...",
    "replyCount": 5,
    ...
  }]
}
```

---

#### Community — Thread Create

**Already updated in schema migration — no changes from UI perspective:**
```json
{
  "sectorIds": "uuid1,uuid2",
  "tagIds": "uuid3",
  "title": "...",
  "slug": "...",
  "description": "...",
  "initialPostContent": "...",
  "attachmentIds": "..."
}
```

---

#### AI — Ingestion Finalize

**Before:**
```json
{
  "storageKey": "uploads/abc.pdf",
  "contentType": "application/pdf",
  "sizeBytes": 12345,
  "checksumSha256": "abc...",
  "idempotencyKey": "key-123",
  "sourceFilename": "mydoc.pdf",
  "declaredLanguage": "en"
}
```

**After (new fields added, all optional):**
```json
{
  "storageKey": "uploads/abc.pdf",
  "contentType": "application/pdf",
  "sizeBytes": 12345,
  "checksumSha256": "abc...",
  "idempotencyKey": "key-123",
  "sourceFilename": "mydoc.pdf",
  "declaredLanguage": "en",
  "sectorIds": ["uuid-sector"],
  "tagIds": ["uuid-tag"],
  "region": "ADDIS_ABABA",
  "stage": "OPERATIONAL"
}
```

---

#### Notifications — Campaign Create

**Before:**
```json
{
  "name": "New VAT deadline",
  "campaignType": "segmented",
  "targetSegment": {"region": "addis_ababa"},
  "campaignTemplateId": "uuid",
  "scheduledFor": null
}
```

**After (new taxonomy fields):**
```json
{
  "name": "New VAT deadline",
  "campaignType": "segmented",
  "targetSegment": {"region": "addis_ababa"},
  "campaignTemplateId": "uuid",
  "scheduledFor": null,
  "sectorIds": ["uuid-sector"],
  "tagIds": ["uuid-tag"],
  "region": "ADDIS_ABABA",
  "stage": "OPERATIONAL"
}
```

Campaign detail response now includes:
```json
{
  ...
  "sectorIds": ["uuid"],
  "tagIds": ["uuid"],
  "region": "ADDIS_ABABA",
  "stage": "OPERATIONAL"
}
```

---

#### BusinessProfile — Response Shape

When the user has a business profile, the response includes:

```json
{
  "id": "uuid",
  "accountId": "uuid",
  "companyName": "Acme Corp",
  "companyEmail": "...",
  "companyPhoneNumber": "...",
  "sectorId": "uuid-sector",        // NEW - single sector
  "tags": [                           // NEW - array of tag objects
    {"id": "uuid", "slug": "tax-vat", "group": "TAX_STATUS", "isMultiSelect": false}
  ],
  "region": "ADDIS_ABABA",           // NEW
  "stage": "OPERATIONAL",            // NEW
  "socialLinks": {...},
  "registrationNumber": "...",
  ...
}
```

**Key:** `sectorId` is a single UUID (not array). `tags` is embedded tag objects.

---

### Step 4: UI Implementation Checklist

- [ ] Fetch sectors from `GET /api/v1/admin/sectors` and build tree structure using `parentId`
- [ ] Fetch tags from `GET /api/v1/admin/tags` and group by `group` field
- [ ] Implement tag picker respecting `isMultiSelect` (single-select groups = radio/picker; multi-select = checkboxes)
- [ ] Profile screen: add sector picker (single), tag picker (multi, group-aware), region dropdown, stage dropdown
- [ ] Guide list: replace category tree UI with flat list filtered by user's profile
- [ ] Thread list: remove category filter UI (auto-filtered by backend)
- [ ] Campaign create: add sector/tag/region/stage fields alongside existing `targetSegment`
- [ ] Document upload: add sector/tag/region/stage fields to finalize screen
- [ ] Remove all hardcoded category IDs/names — use sector/tag UUIDs from API
- [ ] Cache taxonomy data locally (rarely changes) to avoid repeated API calls
