package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type BusinessProfile struct {
	model.BaseModel `gorm:"embedded"`

	AccountID               uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex"`
	Account                 Account           `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CompanyName             string            `gorm:"type:varchar(255);not null"`
	CompanyEmail            string            `gorm:"type:varchar(255);not null"`
	CompanyPhoneNumber      string            `gorm:"type:varchar(50);not null"`
	BusinessType            *string           `gorm:"type:varchar(100)"`
	BusinessSector          *string           `gorm:"type:varchar(100)"`
	RegistrationNumber      *string           `gorm:"type:varchar(100)"`
	RegistrationDate        *time.Time        `gorm:"type:date"`
	TaxIdentificationNumber *string           `gorm:"type:varchar(100)"`
	TradeLicenseNumber      *string           `gorm:"type:varchar(100)"`
	Location                *string           `gorm:"type:varchar(255)"`
	Description             *string           `gorm:"type:text"`
	LogoURL                 *string           `gorm:"type:varchar(512)"`
	BannerURL               *string           `gorm:"type:varchar(512)"`
	SocialLinks             datatypes.JSONMap `gorm:"type:jsonb;not null;default:'{}'"`
}

func (BusinessProfile) TableName() string {
	return "business_profiles"
}
