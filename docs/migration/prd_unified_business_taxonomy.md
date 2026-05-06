## Problem Statement

The platform currently relies on scattered, inconsistent category fields across modules (guides, community, AI ingestion, notifications), which forces hard-coded routing and generic content delivery. This prevents personalized guidance, makes AI retrieval noisy, and leads to irrelevant notifications. The system needs a single, structured taxonomy (sectors, tags, region, stage) that can be applied consistently across modules to power precise filtering, targeting, and personalization.

## Solution

Adopt a unified business taxonomy as the shared targeting system across modules. Sectors are hierarchical and single-select on the business profile, while tags are multi-select attributes. Modules store sector and tag targeting metadata on content (guides, threads, documents, campaigns), and user-facing queries are automatically filtered based on the user profile. The AI ingestion pipeline stores taxonomy at the ingestion-document level and propagates it into chunks. Notifications use taxonomy-based targeting in campaigns; templates are grouped separately via a lightweight admin-only field. Category-based models and endpoints are removed entirely, since this is pre-production.

## User Stories

1. As a business owner, I want to select my sector during onboarding, so that the platform can personalize my experience.
2. As a business owner, I want to select multiple operational tags, so that the platform understands my specific business activities.
3. As a business owner, I want to see only guides relevant to my sector and tags, so that I do not waste time on irrelevant steps.
4. As a business owner, I want guides that apply to my region to appear automatically, so that I follow the correct local rules.
5. As a business owner, I want content for my stage to appear automatically, so that I receive guidance appropriate to my maturity level.
6. As a business owner, I want community threads tailored to my sector, so that I can learn from peers in my industry.
7. As a business owner, I want to filter community threads by tags, so that I can find discussions relevant to my operations.
8. As a business owner, I want AI answers filtered to my sector and region, so that I receive accurate regulatory guidance.
9. As a business owner, I want AI answers to respect my tags, so that I receive advice specific to my tax and operational profile.
10. As a business owner, I want notifications targeted to my business profile, so that I only receive relevant alerts.
11. As an admin, I want to create guides with sector and tag targeting, so that only relevant users see them.
12. As an admin, I want to tag community threads at creation time, so that they are discoverable by the right audience.
13. As an admin, I want to upload regulatory documents with taxonomy metadata, so that AI retrieval is accurate.
14. As an admin, I want to target notification campaigns by sector and tags, so that announcements reach the correct audience.
15. As an admin, I want to target notification campaigns by region and stage, so that deadlines and policy updates are accurate.
16. As an admin, I want to use a shared taxonomy list for sectors and tags, so that targeting is consistent across modules.
17. As a translator, I want sector and tag labels localized, so that onboarding is clear for all languages.
18. As a user, I want content to be shown based on my current profile, so that changes in my business update my experience.
19. As a user, I want general content to remain visible even if I have no tags yet, so that I still see helpful information.
20. As a user, I want region-specific content hidden if my region is missing, so that I do not get incorrect guidance.
21. As a system admin, I want to enforce unique tag and sector slugs, so that filters remain predictable.
22. As an admin, I want to update tags when policies change, so that targeting stays aligned with regulations.
23. As an admin, I want sector hierarchies to support parent-child matching, so that top-level sector content reaches relevant sub-sectors.
24. As an admin, I want AI ingestion metadata to be required for regulatory documents, so that the corpus stays clean.
25. As a developer, I want shared matching rules for taxonomy, so that every module filters content the same way.
26. As a developer, I want to remove category models entirely, so that there is a single source of truth for classification.
27. As a developer, I want consistent API field names for sector and tags, so that clients can integrate easily.
28. As a developer, I want strict validation for taxonomy IDs, so that data integrity is protected.
29. As a mobile client, I want to fetch sector and tag options from the backend, so that I can render a guided onboarding flow.
30. As a community moderator, I want to edit thread targeting metadata, so that misclassified threads are corrected.
31. As an admin, I want notification templates grouped for organization, so that templates are easy to manage.
32. As a system operator, I want fast filters for sector and tag queries, so that listing endpoints stay performant.
33. As a user, I want AI answers to be filtered by my sector subtree, so that parent-sector policies apply where relevant.
34. As a user, I want my selected tags validated by group rules, so that I do not pick incompatible options.
35. As a platform owner, I want the taxonomy to power dynamic routing across modules, so that the app feels context-aware.

## Implementation Decisions

- The taxonomy is a unified model consisting of sectors, tags, region, and stage.
- Sectors are hierarchical and stored with ancestor relationships to support subtree matching.
- Business profiles remain single-sector and multi-tag.
- Tags are grouped with multi-select rules enforced at write time.
- Targeting metadata is stored on guides, community threads, AI ingestion documents, AI chunks, and notification campaigns.
- Array-based storage for sector and tag targeting is used with overlap and containment queries as needed.
- Region and stage targeting use exact-match rules only.
- Admin ingestion of documents requires explicit taxonomy selection; no profile-derived defaults.
- Notifications target users by profile at dispatch time, not at campaign creation time.
- Notification templates are grouped using a lightweight internal grouping field separate from targeting.
- All category entities, fields, and endpoints are removed, and replaced by sector/tag management endpoints.
- A shared taxonomy helper module provides matching, validation, and tag-group rules without owning persistence.
- API contracts standardize naming for sector/tag fields across modules.
- SQL migrations are consolidated into one schema migration followed by one seed migration.
- Seed data is delivered via SQL migrations for sectors, tags, and translations.

## Testing Decisions

- Good tests focus on external behavior: given profile metadata and content targeting, the system returns the correct results.
- Targeting logic is tested through module usecases and query-level behavior rather than internal helper details.
- Shared taxonomy matching and validation is tested with table-driven cases for subtree, tag group rules, and region/stage matching.
- Guide listing tests cover profile-based filtering and client-side narrowing.
- Community listing tests cover thread targeting and filter combinations.
- AI ingestion tests cover required taxonomy enforcement and propagation into chunks.
- Notification campaign tests cover audience filtering and dispatch-time matching.
- Prior art: existing usecase tests in IAM and community modules using table-driven cases and mocked repositories.

## Out of Scope

- Client UI implementation details beyond consistent API fields and translated option lists.
- Real-time content delivery mechanisms such as push subscriptions or websocket updates.
- Analytics dashboards built specifically around taxonomy metrics.
- End-user creation of sectors or tags.

## Further Notes

- Sector subtree matching should be fast and index-backed to support heavy filtering.
- Empty sector/tag arrays should represent no targeting restrictions for general content.
- Region and stage filters should be treated as strict constraints to avoid incorrect guidance.
