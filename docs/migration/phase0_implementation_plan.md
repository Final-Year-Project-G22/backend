# Phase 0 Implementation Plan: Unified Business Taxonomy

## Objective

Prepare a complete, actionable inventory of every structural change required to replace category-based classification with sector/tag taxonomy across guides, community, AI ingestion, and notifications, and establish canonical conventions before any code or migration is written.

---

## 1. Current State Inventory

### 1.1 Existing Taxonomy Entities (IAM)

| Entity | Table | Fields | Status |
|--------|-------|--------|--------|
| Sector | `sectors` | `id`, `slug`, `parent_id` | Canonical, needs `ancestor_ids` |
| SectorTranslation | `sector_translations` | `id`, `sector_id`, `language`, `name`, `description` | Canonical |
| Tag | `tags` | `id`, `slug`, `group`, `is_multi_select` | Canonical |
| TagTranslation | `tag_translations` | `id`, `tag_id`, `language`, `name`, `description` | Canonical |
| BusinessProfile | `business_profiles` | `sector_id`, `region`, `stage` + tags via `business_profile_tags` | Already updated |

### 1.2 Category Entities to Remove

| Module | Entity | Table | Dependencies |
|--------|--------|-------|-------------|
| Guides | GuideCategory | `guide_categories` | GuideCategoryCondition, GuideCategoryTranslation, Guide.CategoryID FK |
| Guides | GuideCategoryCondition | `guide_category_conditions` | GuideCategory |
| Guides | GuideCategoryTranslation | `guide_category_translations` | GuideCategory |
| Community | CommunityCategory | `community_categories` | DiscussionThread.CategoryID FK |
| Library | LibraryCategory | `library_categories` | LibraryTemplateGroup.CategoryID FK |
| Notifications | NotificationCategory enum | N/A | Used in NotificationTemplate.Category, UserNotificationInbox.Category |

### 1.3 Content Entities Requiring Taxonomy Fields

| Module | Entity | Table | Current FK | Needs Added |
|--------|--------|-------|-----------|------------|
| Guides | Guide | `guides` | `category_id` | `sector_ids uuid[]`, `tag_ids uuid[]` |
| Community | DiscussionThread | `discussion_threads` | `category_id` | `sector_ids uuid[]`, `tag_ids uuid[]` |
| AI | IngestionDocument | `ingestion_documents` | None | `sector_ids uuid[]`, `tag_ids uuid[]`, `region varchar(50)`, `stage varchar(50)` |
| Notifications | NotificationCampaign | `notification_campaigns` | None (JSONB target_segment) | `sector_ids uuid[]`, `tag_ids uuid[]`, `region varchar(50)`, `stage varchar(50)` |
| Notifications | NotificationTemplate | `notification_templates` | `category varchar(32)` | `template_group varchar(100)` |
| Notifications | UserNotificationInbox | `user_notification_inboxes` | `category varchar(32)` | Remove `category` |

---

## 2. Target State Definition

### 2.1 Schema Changes Summary

#### Additions

1. **`sectors.ancestor_ids`** — `uuid[]` with GIN index for subtree matching
2. **`guides.sector_ids`** — `uuid[]` with GIN index
3. **`guides.tag_ids`** — `uuid[]` with GIN index
4. **`discussion_threads.sector_ids`** — `uuid[]` with GIN index
5. **`discussion_threads.tag_ids`** — `uuid[]` with GIN index
6. **`ingestion_documents.sector_ids`** — `uuid[]` with GIN index
7. **`ingestion_documents.tag_ids`** — `uuid[]` with GIN index
8. **`ingestion_documents.region`** — `varchar(50)`
9. **`ingestion_documents.stage`** — `varchar(50)`
10. **`notification_campaigns.sector_ids`** — `uuid[]` with GIN index
11. **`notification_campaigns.tag_ids`** — `uuid[]` with GIN index
12. **`notification_campaigns.region`** — `varchar(50)`
13. **`notification_campaigns.stage`** — `varchar(50)`
14. **`notification_templates.template_group`** — `varchar(100)`

#### Removals (same migration)

1. Drop `guide_categories` table + translations + conditions
2. Drop `community_categories` table
3. Drop `library_categories` table + translations
4. Drop `notification_templates.category` column
5. Drop `user_notification_inboxes.category` column
6. Drop `guides.category_id` column + related indexes
7. Drop `discussion_threads.category_id` column + related indexes

### 2.2 Index Strategy

All `uuid[]` fields get GIN indexes for fast overlap (`&&`) and containment (`@>`) queries:

```sql
CREATE INDEX idx_guides_sector_ids ON guides USING GIN (sector_ids);
CREATE INDEX idx_guides_tag_ids ON guides USING GIN (tag_ids);
-- etc.
```

For `sectors.ancestor_ids`:
```sql
CREATE INDEX idx_sectors_ancestor_ids ON sectors USING GIN (ancestor_ids);
```

### 2.3 Entity Changes

#### Guides Module

**Remove:**
- `GuideCategory` entity
- `GuideCategoryCondition` entity
- `GuideCategoryTranslation` entity

**Modify `Guide`:**
```go
type Guide struct {
    model.BaseModel `gorm:"embedded"`
    // Remove: CategoryID, Category
    SectorIDs       []uuid.UUID `gorm:"type:uuid[];index:idx_guides_sector_ids,using:gin"`
    TagIDs          []uuid.UUID `gorm:"type:uuid[];index:idx_guides_tag_ids,using:gin"`
    Slug            string      `gorm:"type:varchar(200);not null;uniqueIndex:idx_guides_slug"`
    // ... rest unchanged
}
```

#### Community Module

**Remove:**
- `CommunityCategory` entity

**Modify `DiscussionThread`:**
```go
type DiscussionThread struct {
    model.BaseModel `gorm:"embedded"`
    // Remove: CategoryID, Category
    SectorIDs       []uuid.UUID `gorm:"type:uuid[];index:idx_threads_sector_ids,using:gin"`
    TagIDs          []uuid.UUID `gorm:"type:uuid[];index:idx_threads_tag_ids,using:gin"`
    // ... rest unchanged
}
```

#### AI Module

**Modify `IngestionDocument`:**
```go
type IngestionDocument struct {
    model.BaseModel `gorm:"embedded"`
    // ... existing fields ...
    SectorIDs []uuid.UUID `gorm:"type:uuid[];index:idx_ingestion_sector_ids,using:gin"`
    TagIDs    []uuid.UUID `gorm:"type:uuid[];index:idx_ingestion_tag_ids,using:gin"`
    Region    *string     `gorm:"type:varchar(50)"`
    Stage     *string     `gorm:"type:varchar(50)"`
}
```

#### Notifications Module

**Modify `NotificationTemplate`:**
```go
type NotificationTemplate struct {
    model.BaseModel `gorm:"embedded"`
    // Remove: Category NotificationCategory
    TemplateGroup   string `gorm:"type:varchar(100);index:idx_notif_templates_group"`
    // ... rest unchanged
}
```

**Modify `NotificationCampaign`:**
```go
type NotificationCampaign struct {
    model.BaseModel `gorm:"embedded"`
    // ... existing fields ...
    SectorIDs []uuid.UUID `gorm:"type:uuid[];index:idx_campaign_sector_ids,using:gin"`
    TagIDs    []uuid.UUID `gorm:"type:uuid[];index:idx_campaign_tag_ids,using:gin"`
    Region    *string     `gorm:"type:varchar(50)"`
    Stage     *string     `gorm:"type:varchar(50)"`
}
```

**Modify `UserNotificationInbox`:**
```go
type UserNotificationInbox struct {
    model.BaseModel `gorm:"embedded"`
    // Remove: Category string
    // ... rest unchanged
}
```

### 2.4 Shared Taxonomy Package

Create `internal/shared/taxonomy/` with:

**Types:**
```go
type TaxonomyFilter struct {
    SectorIDs   []uuid.UUID
    TagIDs      []uuid.UUID
    Region      *string
    Stage       *string
}

type ProfileTaxonomy struct {
    SectorID  *uuid.UUID
    TagIDs    []uuid.UUID
    Region    *string
    Stage     *string
}
```

**Matchers:**
- `MatchSector(filter, profileSectorID, sectorAncestors) bool` — subtree match
- `MatchTags(filterTagIDs, profileTagIDs) bool` — overlap by group rules
- `MatchRegion(filterRegion, profileRegion) bool` — exact or nil
- `MatchStage(filterStage, profileStage) bool` — exact or nil
- `MatchAll(filter, profile) bool` — combines all with AND logic

---

## 3. API Field Naming Conventions

Standardize across all modules:

| Field | Type | Description |
|-------|------|-------------|
| `sector_ids` | `[]string` (UUIDs) | Target sectors |
| `tag_ids` | `[]string` (UUIDs) | Target tags |
| `region` | `string` | Target region |
| `stage` | `string` | Target stage |
| `template_group` | `string` | Template organization |

In responses, include resolved names:
```json
{
  "sector_ids": ["uuid-1"],
  "sectors": [{"id": "uuid-1", "name": "Agriculture", "name_localized": "..."}],
  "tag_ids": ["uuid-2"],
  "tags": [{"id": "uuid-2", "name": "Exporter", "group": "operations"}]
}
```

---

## 4. Migration Strategy

### 4.1 Migration Files

**File 1: `20260507_unified_taxonomy_schema.sql`** (consolidated schema)

Order of operations:
1. Add new columns to existing tables
2. Create new GIN indexes
3. Drop old category columns and indexes
4. Drop old category tables
5. Add `ancestor_ids` to sectors

**File 2: `20260507_unified_taxonomy_seeds.sql`** (seed data)

1. Seed canonical sectors with `ancestor_ids`
2. Seed canonical tags with groups
3. Seed translations for both

### 4.2 Migration Tool

Use existing Atlas/GORM migration pattern. All entity changes must be reflected in:
- Entity structs with GORM tags
- `EntityProvider` implementations
- Migration files (auto-generated by Atlas)

---

## 5. Module-by-Module Implementation Order

| Phase | Module | Changes |
|-------|--------|---------|
| 1 | Shared | Create `internal/shared/taxonomy` package |
| 2 | IAM | Add `ancestor_ids` to Sector entity |
| 3 | Schema | Run consolidated migration |
| 4 | Seeds | Run seed migration |
| 5 | Guides | Remove categories, add sector/tag to Guide, update queries |
| 6 | Community | Remove categories, add sector/tag to DiscussionThread, update queries |
| 7 | AI | Add taxonomy to IngestionDocument, enforce at finalize, propagate to chunks |
| 8 | Notifications | Replace category with template_group, add taxonomy to campaigns, update targeting |
| 9 | Admin | Create sector/tag management endpoints |
| 10 | Tests | Run verification checklist |

---

## 6. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Client breakage from removed category endpoints | High | High | Coordinate with mobile team; category endpoints will be replaced simultaneously |
| Query performance with array filters | Medium | Medium | GIN indexes + query planner verification |
| Tag explosion from admin misuse | Low | Medium | Admin-only tag creation + validation |
| Sector hierarchy corruption | Low | High | Application-level enforcement + validation on writes |
| Missing content after migration (no legacy data) | N/A | N/A | Pre-production, no data to migrate |

---

## 7. Verification Checklist

### Before Implementation
- [ ] All category endpoints identified and listed for replacement
- [ ] Mobile team notified of API contract changes
- [ ] Canonical sector/tag list finalized with stakeholders
- [ ] Seed data prepared and validated

### During Implementation
- [ ] Each migration runs clean on fresh DB
- [ ] All GIN indexes created successfully
- [ ] Entity registration includes new fields
- [ ] No compilation errors after category removal

### After Implementation
- [ ] Admin can create guide with sector/tag targeting
- [ ] User sees only guides matching their profile
- [ ] Admin can upload document with required taxonomy
- [ ] AI retrieval filters by profile taxonomy
- [ ] Notification campaign targets by sector/tag
- [ ] All category endpoints return 404 (or are removed)

---

## 8. Immediate Next Steps (Post Phase 0)

1. **Finalize sector hierarchy** — Get stakeholder sign-off on canonical sector tree
2. **Finalize tag groups and values** — Define all tag groups and initial tags
3. **Coordinate with mobile** — Share API contract changes
4. **Begin Phase 1: Schema Migration** — Write consolidated migration SQL

---

## Appendix A: Complete Category Removal Inventory

### Tables to Drop
- `guide_categories`
- `guide_category_conditions`
- `guide_category_translations`
- `community_categories`
- `library_categories`
- `library_category_translations`

### Columns to Drop
- `guides.category_id`
- `discussion_threads.category_id`
- `notification_templates.category`
- `user_notification_inboxes.category`

### Entities to Remove
- `GuideCategory`
- `GuideCategoryCondition`
- `GuideCategoryTranslation`
- `CommunityCategory`
- `LibraryCategory` (if exists in code)
- `LibraryCategoryTranslation` (if exists in code)
- `NotificationCategory` enum

### Indexes to Drop
- `idx_guides_category`
- `idx_guides_slug_per_category`
- `idx_discussion_threads_category`
- `idx_discussion_threads_slug_per_category`
- `idx_notif_templates_category`
- `idx_notif_inbox_category`
- `idx_guide_categories_parent`
- `idx_guide_categories_slug_per_parent`
- `idx_community_categories_parent`
- `idx_community_categories_slug_per_parent`
- `idx_library_categories_parent`
- `idx_library_categories_slug_per_parent`

### Indexes to Add
- `idx_guides_sector_ids` (GIN)
- `idx_guides_tag_ids` (GIN)
- `idx_discussion_threads_sector_ids` (GIN)
- `idx_discussion_threads_tag_ids` (GIN)
- `idx_ingestion_documents_sector_ids` (GIN)
- `idx_ingestion_documents_tag_ids` (GIN)
- `idx_notification_campaigns_sector_ids` (GIN)
- `idx_notification_campaigns_tag_ids` (GIN)
- `idx_sectors_ancestor_ids` (GIN)
- `idx_notif_templates_group`

---

*Plan completed. Ready for Phase 1 implementation.*
