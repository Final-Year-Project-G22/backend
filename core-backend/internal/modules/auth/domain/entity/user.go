package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type Tier string

const (
	TierFree    Tier = "free"
	TierPremium Tier = "premium"
)

type User struct {
	model.BaseModel `gorm:"embedded"`

	AccountId uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Account   Account   `gorm:"foreignKey:AccountId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	FirstName string  `gorm:"type:varchar(255);not null"`
	LastName  string  `gorm:"type:varchar(255);not null"`
	ImageUrl  *string `gorm:"type:varchar(255)"`
	Tier      Tier    `gorm:"type:varchar(255);not null;default:'free'"`
}
