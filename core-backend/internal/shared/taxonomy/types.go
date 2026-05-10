package taxonomy

import "github.com/google/uuid"

// TaxonomyFilter represents targeting criteria on content.
type TaxonomyFilter struct {
	SectorIDs []uuid.UUID
	TagIDs    []uuid.UUID
	Region    *string
	Stage     *string
}

// ProfileTaxonomy represents the taxonomy dimensions of a user profile.
type ProfileTaxonomy struct {
	SectorID *uuid.UUID
	TagIDs   []uuid.UUID
	Region   *string
	Stage    *string
}
