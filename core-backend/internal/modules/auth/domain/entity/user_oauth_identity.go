package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type UserOauthIdentity struct {
	model.BaseModel `gorm:"embedded"`

	AccountId      uuid.UUID `gorm:"type:uuid;not null;index"`
	Provider       string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_provider_provider_user_id"`
	ProviderUserId string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_provider_provider_user_id"`
	AccessToken    string    `gorm:"type:varchar(255);not null"`
	RefreshToken   *string   `gorm:"type:varchar(255)"`
}

func (UserOauthIdentity) TableName() string {
	return "user_oauth_identities"
}
