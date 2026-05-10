package taxonomy

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMatchSector(t *testing.T) {
	sectorA := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sectorB := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	sectorC := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	tests := []struct {
		name      string
		filter    []uuid.UUID
		ancestors []uuid.UUID
		wantMatch bool
	}{
		{
			name:      "empty filter matches all",
			filter:    []uuid.UUID{},
			ancestors: []uuid.UUID{sectorA},
			wantMatch: true,
		},
		{
			name:      "empty ancestors with filter returns false",
			filter:    []uuid.UUID{sectorA},
			ancestors: []uuid.UUID{},
			wantMatch: false,
		},
		{
			name:      "direct match",
			filter:    []uuid.UUID{sectorA},
			ancestors: []uuid.UUID{sectorA},
			wantMatch: true,
		},
		{
			name:      "ancestor match",
			filter:    []uuid.UUID{sectorA},
			ancestors: []uuid.UUID{sectorB, sectorA},
			wantMatch: true,
		},
		{
			name:      "no match",
			filter:    []uuid.UUID{sectorA},
			ancestors: []uuid.UUID{sectorC},
			wantMatch: false,
		},
		{
			name:      "multiple filters match one",
			filter:    []uuid.UUID{sectorA, sectorC},
			ancestors: []uuid.UUID{sectorB, sectorA},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchSector(tt.filter, tt.ancestors)
			assert.Equal(t, tt.wantMatch, got)
		})
	}
}

func TestMatchTags(t *testing.T) {
	opExporter := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	opImporter := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	taxVAT := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	taxTOT := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	empFull := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")

	tagGroups := map[uuid.UUID]string{
		opExporter: "GENERAL_OPERATIONS",
		opImporter: "GENERAL_OPERATIONS",
		taxVAT:     "TAX_STATUS",
		taxTOT:     "TAX_STATUS",
		empFull:    "EMPLOYMENT",
	}

	tests := []struct {
		name      string
		filter    []uuid.UUID
		profile   []uuid.UUID
		wantMatch bool
	}{
		{
			name:      "empty filter matches all",
			filter:    []uuid.UUID{},
			profile:   []uuid.UUID{opExporter},
			wantMatch: true,
		},
		{
			name:      "empty profile with filter returns false",
			filter:    []uuid.UUID{opExporter},
			profile:   []uuid.UUID{},
			wantMatch: false,
		},
		{
			name:      "single tag match",
			filter:    []uuid.UUID{opExporter},
			profile:   []uuid.UUID{opExporter},
			wantMatch: true,
		},
		{
			name:      "any-of within group",
			filter:    []uuid.UUID{opExporter, opImporter},
			profile:   []uuid.UUID{opExporter},
			wantMatch: true,
		},
		{
			name:      "all groups must match - both match",
			filter:    []uuid.UUID{opExporter, taxVAT},
			profile:   []uuid.UUID{opExporter, taxVAT},
			wantMatch: true,
		},
		{
			name:      "all groups must match - missing group",
			filter:    []uuid.UUID{opExporter, taxVAT},
			profile:   []uuid.UUID{opExporter},
			wantMatch: false,
		},
		{
			name:      "all groups must match - wrong tag in group",
			filter:    []uuid.UUID{opExporter},
			profile:   []uuid.UUID{opImporter},
			wantMatch: false,
		},
		{
			name:      "profile has extra tags",
			filter:    []uuid.UUID{opExporter},
			profile:   []uuid.UUID{opExporter, taxVAT, empFull},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchTags(tt.filter, tt.profile, tagGroups)
			assert.Equal(t, tt.wantMatch, got)
		})
	}
}

func TestMatchRegion(t *testing.T) {
	regionA := "ADDIS_ABABA"
	regionB := "OROMIA"

	tests := []struct {
		name      string
		filter    *string
		profile   *string
		wantMatch bool
	}{
		{
			name:      "nil filter matches all",
			filter:    nil,
			profile:   &regionA,
			wantMatch: true,
		},
		{
			name:      "empty filter matches all",
			filter:    strPtr(""),
			profile:   &regionA,
			wantMatch: true,
		},
		{
			name:      "nil profile with filter returns false",
			filter:    &regionA,
			profile:   nil,
			wantMatch: false,
		},
		{
			name:      "exact match",
			filter:    &regionA,
			profile:   &regionA,
			wantMatch: true,
		},
		{
			name:      "no match",
			filter:    &regionA,
			profile:   &regionB,
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchRegion(tt.filter, tt.profile)
			assert.Equal(t, tt.wantMatch, got)
		})
	}
}

func TestMatchStage(t *testing.T) {
	stageA := "IDEA"
	stageB := "OPERATIONAL"

	tests := []struct {
		name      string
		filter    *string
		profile   *string
		wantMatch bool
	}{
		{
			name:      "nil filter matches all",
			filter:    nil,
			profile:   &stageA,
			wantMatch: true,
		},
		{
			name:      "exact match",
			filter:    &stageA,
			profile:   &stageA,
			wantMatch: true,
		},
		{
			name:      "no match",
			filter:    &stageA,
			profile:   &stageB,
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchStage(tt.filter, tt.profile)
			assert.Equal(t, tt.wantMatch, got)
		})
	}
}

func TestMatchAll(t *testing.T) {
	sectorA := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tagA := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	regionA := "ADDIS_ABABA"
	stageA := "IDEA"

	tagGroups := map[uuid.UUID]string{
		tagA: "GENERAL_OPERATIONS",
	}

	tests := []struct {
		name      string
		filter    TaxonomyFilter
		profile   ProfileTaxonomy
		ancestors []uuid.UUID
		wantMatch bool
	}{
		{
			name: "all match",
			filter: TaxonomyFilter{
				SectorIDs: []uuid.UUID{sectorA},
				TagIDs:    []uuid.UUID{tagA},
				Region:    &regionA,
				Stage:     &stageA,
			},
			profile: ProfileTaxonomy{
				SectorID: &sectorA,
				TagIDs:   []uuid.UUID{tagA},
				Region:   &regionA,
				Stage:    &stageA,
			},
			ancestors: []uuid.UUID{sectorA},
			wantMatch: true,
		},
		{
			name: "sector mismatch",
			filter: TaxonomyFilter{
				SectorIDs: []uuid.UUID{sectorA},
			},
			profile: ProfileTaxonomy{
				SectorID: nil,
			},
			ancestors: []uuid.UUID{},
			wantMatch: false,
		},
		{
			name:   "empty filter matches all",
			filter: TaxonomyFilter{},
			profile: ProfileTaxonomy{
				SectorID: &sectorA,
				TagIDs:   []uuid.UUID{tagA},
				Region:   &regionA,
				Stage:    &stageA,
			},
			ancestors: []uuid.UUID{sectorA},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchAll(tt.filter, tt.profile, tt.ancestors, tagGroups)
			assert.Equal(t, tt.wantMatch, got)
		})
	}
}

func strPtr(s string) *string {
	return &s
}
