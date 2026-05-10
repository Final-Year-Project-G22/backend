package iam

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
)

// EntityProvider implements core.EntityProvider for iam module
type EntityProvider struct{}

// NewEntityProvider creates the entity provider
func NewEntityProvider() *EntityProvider {
	return &EntityProvider{}
}

// Entities returns all domain entities for auth module
func (p *EntityProvider) Entities() []interface{} {
	return []interface{}{
		&entity.User{},
		&entity.Account{},
		&entity.AccountEmailOTP{},
		&entity.Role{},
		&entity.Permission{},
		&entity.RolePermission{},
		&entity.RoleAssignment{},
		&entity.Session{},
		&entity.OAuthIdentity{},
		&entity.AIPreference{},
		&entity.TemplatePreference{},
		&entity.AccountPreference{},
		&entity.NotificationPreference{},
		&entity.CommunityPreference{},
		&entity.BusinessProfile{},
		&entity.Sector{},
		&entity.SectorTranslation{},
		&entity.Tag{},
		&entity.TagTranslation{},
	}
}

// ModuleName returns the module identifier
func (p *EntityProvider) ModuleName() string {
	return "iam"
}
