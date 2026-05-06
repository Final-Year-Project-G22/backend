package taxonomy

import "github.com/google/uuid"

// MatchSector checks if any of the profile sector ancestors overlap with the
// filter sectors. An empty filter means no restriction (matches all).
func MatchSector(filterSectorIDs []uuid.UUID, profileSectorAncestors []uuid.UUID) bool {
	if len(filterSectorIDs) == 0 {
		return true
	}
	if len(profileSectorAncestors) == 0 {
		return false
	}

	filterSet := make(map[uuid.UUID]struct{}, len(filterSectorIDs))
	for _, id := range filterSectorIDs {
		filterSet[id] = struct{}{}
	}

	for _, id := range profileSectorAncestors {
		if _, ok := filterSet[id]; ok {
			return true
		}
	}
	return false
}

// MatchTags checks if profile tags satisfy filter tags grouped by tag group.
// Rules:
//   - Empty filter means no restriction (matches all).
//   - For each tag group present in the filter, the profile must have at least
//     one tag in that group that also appears in the filter for that group.
func MatchTags(filterTagIDs []uuid.UUID, profileTagIDs []uuid.UUID, tagGroups map[uuid.UUID]string) bool {
	if len(filterTagIDs) == 0 {
		return true
	}
	if len(profileTagIDs) == 0 {
		return false
	}

	// Group filter tags by their group.
	filterByGroup := make(map[string]map[uuid.UUID]struct{})
	for _, id := range filterTagIDs {
		group, ok := tagGroups[id]
		if !ok {
			continue // unknown tag, skip
		}
		if filterByGroup[group] == nil {
			filterByGroup[group] = make(map[uuid.UUID]struct{})
		}
		filterByGroup[group][id] = struct{}{}
	}

	// Group profile tags by their group.
	profileByGroup := make(map[string]map[uuid.UUID]struct{})
	for _, id := range profileTagIDs {
		group, ok := tagGroups[id]
		if !ok {
			continue
		}
		if profileByGroup[group] == nil {
			profileByGroup[group] = make(map[uuid.UUID]struct{})
		}
		profileByGroup[group][id] = struct{}{}
	}

	// Every group present in the filter must be satisfied by the profile.
	for group, filterSet := range filterByGroup {
		profileSet, ok := profileByGroup[group]
		if !ok {
			return false // profile missing this group entirely
		}
		matched := false
		for id := range filterSet {
			if _, ok := profileSet[id]; ok {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// MatchRegion checks exact region match. Nil/empty filter means no restriction.
func MatchRegion(filterRegion, profileRegion *string) bool {
	if filterRegion == nil || *filterRegion == "" {
		return true
	}
	if profileRegion == nil || *profileRegion == "" {
		return false
	}
	return *filterRegion == *profileRegion
}

// MatchStage checks exact stage match. Nil/empty filter means no restriction.
func MatchStage(filterStage, profileStage *string) bool {
	if filterStage == nil || *filterStage == "" {
		return true
	}
	if profileStage == nil || *profileStage == "" {
		return false
	}
	return *filterStage == *profileStage
}

// MatchAll evaluates all taxonomy dimensions. It returns true only when sector,
// tag, region, and stage all match (AND logic).
func MatchAll(filter TaxonomyFilter, profile ProfileTaxonomy, profileSectorAncestors []uuid.UUID, tagGroups map[uuid.UUID]string) bool {
	if !MatchSector(filter.SectorIDs, profileSectorAncestors) {
		return false
	}
	if !MatchTags(filter.TagIDs, profile.TagIDs, tagGroups) {
		return false
	}
	if !MatchRegion(filter.Region, profile.Region) {
		return false
	}
	if !MatchStage(filter.Stage, profile.Stage) {
		return false
	}
	return true
}
