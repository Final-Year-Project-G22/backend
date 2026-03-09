package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type CommunityPreference struct {
	model.BaseModel `gorm:"embedded"`

	AccountID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Account       Account   `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AllowMentions bool      `gorm:"not null;default:true"`
	AllowReplies  bool      `gorm:"not null;default:true"`
	DigestEnabled bool      `gorm:"not null;default:true"`
}

func (CommunityPreference) TableName() string {
	return "community_preferences"
}
