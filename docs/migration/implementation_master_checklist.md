# Unified Business Taxonomy — Implementation Master Checklist

> **Rule:** Do NOT proceed to the next phase until every item in the current phase is checked.

---

## Phase 0 — Preparation & Planning

- [x] Complete Phase 0 implementation plan document
- [x] Inventory all category entities, tables, and columns to remove
- [x] Inventory all taxonomy additions (sector_ids, tag_ids, region, stage, template_group)
- [x] Define canonical API field naming conventions
- [x] Define shared taxonomy package interface
- [x] Confirm sector hierarchy and tag groups with stakeholders
- [x] Confirm seed data for sectors and tags
- [x] Coordinate API contract changes with mobile team
- [x] Review and approve Phase 0 plan

**Phase 0 Status:** ✅ Complete

**Decision Gate:** Proceed to Phase 1 only after Phase 0 is fully checked and approved.

---

## Phase 1 — Shared Taxonomy Package

- [x] Create `internal/shared/taxonomy` package directory
- [x] Define `TaxonomyFilter` struct
- [x] Define `ProfileTaxonomy` struct
- [x] Implement `MatchSector()` — subtree matching with ancestor_ids
- [x] Implement `MatchTags()` — overlap with group rules
- [x] Implement `MatchRegion()` — exact match or nil handling
- [x] Implement `MatchStage()` — exact match or nil handling
- [x] Implement `MatchAll()` — combines all matchers with AND logic
- [x] Write unit tests for all matchers
- [x] Verify all tests pass

**Phase 1 Status:** ✅ Complete

**Decision Gate:** All matcher tests must pass before proceeding.

---

## Phase 2 — Schema Migration (Consolidated)

### 2.1 Create Migration Files

- [ ] Create `20260507_unified_taxonomy_schema.sql`
- [ ] Add `ancestor_ids uuid[]` to `sectors` table
- [ ] Add GIN index on `sectors.ancestor_ids`
- [ ] Add `sector_ids uuid[]` to `guides` table
- [ ] Add `tag_ids uuid[]` to `guides` table
- [ ] Add GIN indexes on `guides.sector_ids` and `guides.tag_ids`
- [ ] Add `sector_ids uuid[]` to `discussion_threads` table
- [ ] Add `tag_ids uuid[]` to `discussion_threads` table
- [ ] Add GIN indexes on `discussion_threads.sector_ids` and `discussion_threads.tag_ids`
- [ ] Add `sector_ids uuid[]` to `ingestion_documents` table
- [ ] Add `tag_ids uuid[]` to `ingestion_documents` table
- [ ] Add `region varchar(50)` to `ingestion_documents` table
- [ ] Add `stage varchar(50)` to `ingestion_documents` table
- [ ] Add GIN indexes on `ingestion_documents.sector_ids` and `ingestion_documents.tag_ids`
- [ ] Add `sector_ids uuid[]` to `notification_campaigns` table
- [ ] Add `tag_ids uuid[]` to `notification_campaigns` table
- [ ] Add `region varchar(50)` to `notification_campaigns` table
- [ ] Add `stage varchar(50)` to `notification_campaigns` table
- [ ] Add GIN indexes on `notification_campaigns.sector_ids` and `notification_campaigns.tag_ids`
- [ ] Add `template_group varchar(100)` to `notification_templates` table
- [ ] Add index on `notification_templates.template_group`

### 2.2 Remove Category Schema

- [ ] Drop `guides.category_id` column
- [ ] Drop `idx_guides_category` index
- [ ] Drop `idx_guides_slug_per_category` index
- [ ] Drop `discussion_threads.category_id` column
- [ ] Drop `idx_discussion_threads_category` index
- [ ] Drop `idx_discussion_threads_slug_per_category` index
- [ ] Drop `notification_templates.category` column
- [ ] Drop `idx_notif_templates_category` index
- [ ] Drop `user_notification_inboxes.category` column
- [ ] Drop `idx_notif_inbox_category` index
- [ ] Drop `guide_categories` table
- [ ] Drop `guide_category_conditions` table
- [ ] Drop `guide_category_translations` table
- [ ] Drop `community_categories` table
- [ ] Drop `library_categories` table
- [ ] Drop `library_category_translations` table

### 2.3 Run & Verify

- [ ] Run migration on fresh database
- [ ] Verify all new columns exist
- [ ] Verify all GIN indexes created
- [ ] Verify all old columns/tables removed
- [ ] Verify no compilation errors

**Phase 2 Status:** ⏳ Pending

**Decision Gate:** Migration runs clean with zero errors before proceeding.

---

## Phase 3 — Seed Data Migration

- [ ] Create `20260507_unified_taxonomy_seeds.sql`
- [ ] Seed canonical sectors with parent relationships
- [ ] Seed `ancestor_ids` for all sectors
- [ ] Seed canonical tags with groups
- [ ] Seed sector translations
- [ ] Seed tag translations
- [ ] Run seed migration
- [ ] Verify all sectors have correct ancestor_ids
- [ ] Verify all tags have correct groups
- [ ] Verify translations exist

**Phase 3 Status:** ⏳ Pending

**Decision Gate:** Seed data verified correct before proceeding.

---

## Phase 4 — IAM Updates

- [ ] Update `Sector` entity with `AncestorIDs []uuid.UUID` field
- [ ] Update sector repository to handle ancestor_ids
- [ ] Add validation: sector slug uniqueness
- [ ] Add validation: tag group rules
- [ ] Update BusinessProfile usecase: enforce tag group constraints
- [ ] Update BusinessProfile DTOs: include resolved sector/tag names
- [ ] Verify BusinessProfile CRUD still works
- [ ] Write tests for tag group validation
- [ ] Verify all IAM tests pass

**Phase 4 Status:** ⏳ Pending

**Decision Gate:** IAM tests pass before proceeding.

---

## Phase 5 — Guides Module Update

- [ ] Remove `GuideCategory` entity
- [ ] Remove `GuideCategoryCondition` entity
- [ ] Remove `GuideCategoryTranslation` entity
- [ ] Update `Guide` entity: remove CategoryID, add SectorIDs and TagIDs
- [ ] Update guide repository queries: filter by sector/tag arrays
- [ ] Update guide usecase: apply profile-based targeting
- [ ] Update guide DTOs: standardize field names
- [ ] Replace category admin endpoints with sector/tag endpoints
- [ ] Update guide creation flow: select sectors/tags instead of category
- [ ] Write tests for guide filtering
- [ ] Verify all guide tests pass

**Phase 5 Status:** ⏳ Pending

**Decision Gate:** Guide tests pass before proceeding.

---

## Phase 6 — Community Module Update

- [ ] Remove `CommunityCategory` entity
- [ ] Update `DiscussionThread` entity: remove CategoryID, add SectorIDs and TagIDs
- [ ] Update thread repository queries: filter by sector/tag arrays
- [ ] Update thread usecase: apply profile-based targeting
- [ ] Update thread DTOs: standardize field names
- [ ] Replace category endpoints with sector/tag endpoints
- [ ] Update thread creation flow: select sectors/tags instead of category
- [ ] Write tests for thread filtering
- [ ] Verify all community tests pass

**Phase 6 Status:** ⏳ Pending

**Decision Gate:** Community tests pass before proceeding.

---

## Phase 7 — AI Module Update

- [ ] Update `IngestionDocument` entity: add SectorIDs, TagIDs, Region, Stage
- [ ] Add validation: require sector and region at finalize
- [ ] Update ingestion usecase: enforce taxonomy rules
- [ ] Update ingestion DTOs: include taxonomy fields
- [ ] Propagate taxonomy from ingestion document to chunks
- [ ] Update AI retrieval: filter by profile taxonomy
- [ ] Allow user overrides in AI queries (narrow only)
- [ ] Write tests for ingestion validation
- [ ] Write tests for AI retrieval filtering
- [ ] Verify all AI tests pass

**Phase 7 Status:** ⏳ Pending

**Decision Gate:** AI tests pass before proceeding.

---

## Phase 8 — Notifications Module Update

- [ ] Update `NotificationTemplate` entity: remove Category, add TemplateGroup
- [ ] Update `NotificationCampaign` entity: add SectorIDs, TagIDs, Region, Stage
- [ ] Update `UserNotificationInbox` entity: remove Category
- [ ] Update template usecase: template_group for organization
- [ ] Update campaign usecase: targeting by taxonomy
- [ ] Update delivery usecase: evaluate targeting at dispatch time
- [ ] Update notification DTOs: standardize field names
- [ ] Replace category endpoints with sector/tag endpoints
- [ ] Write tests for campaign targeting
- [ ] Write tests for dispatch-time matching
- [ ] Verify all notification tests pass

**Phase 8 Status:** ⏳ Pending

**Decision Gate:** Notification tests pass before proceeding.

---

## Phase 9 — Admin Endpoints & API Standardization

- [ ] Create sector management endpoints (CRUD)
- [ ] Create tag management endpoints (CRUD)
- [ ] Standardize request/response field names across all modules
- [ ] Include resolved sector/tag names in responses
- [ ] Add strict validation for unknown sector/tag IDs
- [ ] Verify OpenAPI generation works with new routes
- [ ] Document API changes for mobile team
- [ ] Write integration tests for admin flows
- [ ] Verify all integration tests pass

**Phase 9 Status:** ⏳ Pending

**Decision Gate:** Integration tests pass before proceeding.

---

## Phase 10 — Final Verification & Cleanup

### 10.1 End-to-End Verification

- [ ] Admin creates sector → appears in options
- [ ] Admin creates tag → appears in options
- [ ] Admin creates guide with sector/tag targeting
- [ ] User sees only guides matching their profile
- [ ] Admin uploads document with taxonomy → enforced at finalize
- [ ] AI retrieval filters by user profile
- [ ] Admin creates notification campaign targeting specific sector/tag
- [ ] Campaign delivers only to matching profiles
- [ ] All old category endpoints return 404 or are removed

### 10.2 Code Cleanup

- [ ] Remove all category entity files
- [ ] Remove all category handler files
- [ ] Remove all category route registrations
- [ ] Remove all category DTOs
- [ ] Remove `NotificationCategory` enum
- [ ] Verify no dead code references to categories
- [ ] Run linter, fix any issues

### 10.3 Performance Verification

- [ ] Verify GIN indexes are used in query plans
- [ ] Test guide listing with targeting (acceptable response time)
- [ ] Test community listing with targeting (acceptable response time)
- [ ] Test AI retrieval with taxonomy filters (acceptable response time)

### 10.4 Documentation

- [ ] Update API documentation
- [ ] Update admin documentation
- [ ] Mark implementation as complete

**Phase 10 Status:** ⏳ Pending

**Decision Gate:** All verification checks pass before declaring project complete.

---

## Overall Project Status

| Phase | Status | Date Completed |
|-------|--------|---------------|
| Phase 0 | 🟡 In Progress | - |
| Phase 1 | ⏳ Pending | - |
| Phase 2 | ⏳ Pending | - |
| Phase 3 | ⏳ Pending | - |
| Phase 4 | ⏳ Pending | - |
| Phase 5 | ⏳ Pending | - |
| Phase 6 | ⏳ Pending | - |
| Phase 7 | ⏳ Pending | - |
| Phase 8 | ⏳ Pending | - |
| Phase 9 | ⏳ Pending | - |
| Phase 10 | ⏳ Pending | - |

**Last Updated:** 2025-05-06

**Next Action:** Approve Phase 0 and proceed to Phase 1
