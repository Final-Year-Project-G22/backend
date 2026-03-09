package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type OAuthIdentity struct {
	model.BaseModel `gorm:"embedded"`

	AccountID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	Account         Account    `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Provider        string     `gorm:"type:varchar(100);not null;uniqueIndex:idx_oauth_provider_subject,priority:1"`
	ProviderSubject string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_oauth_provider_subject,priority:2"`
	ProviderEmail   *string    `gorm:"type:varchar(255)"`
	AccessToken     *string    `gorm:"type:text"`
	RefreshToken    *string    `gorm:"type:text"`
	TokenExpiresAt  *time.Time `gorm:"type:timestamptz"`
	LastUsedAt      *time.Time `gorm:"type:timestamptz"`
}

func (OAuthIdentity) TableName() string {
	return "oauth_identities"
}
