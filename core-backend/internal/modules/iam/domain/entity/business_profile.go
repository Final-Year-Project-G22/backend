package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type BusinessProfile struct {
	model.BaseModel `gorm:"embedded"`

	AccountID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Account   Account   `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// --- Standard Profile Info ---
	CompanyName        string            `gorm:"type:varchar(255);not null"`
	CompanyEmail       string            `gorm:"type:varchar(255);not null"`
	CompanyPhoneNumber string            `gorm:"type:varchar(50);not null"`
	PhysicalAddress    *string           `gorm:"type:varchar(255)"` // Specific street/Kebele address
	Description        *string           `gorm:"type:text"`
	LogoURL            *string           `gorm:"type:varchar(512)"`
	BannerURL          *string           `gorm:"type:varchar(512)"`
	SocialLinks        datatypes.JSONMap `gorm:"type:jsonb;not null;default:'{}'"`

	// --- Official Registrations ---
	RegistrationNumber      *string    `gorm:"type:varchar(100)"`
	RegistrationDate        *time.Time `gorm:"type:date"`
	TaxIdentificationNumber *string    `gorm:"type:varchar(100)"`
	TradeLicenseNumber      *string    `gorm:"type:varchar(100)"`

	// ==========================================
	// TAXONOMY ENGINE (Routing & Filtering)
	// ==========================================

	// --- Geography and Lifecycle ---
	Region *Region        `gorm:"type:varchar(50);index"`
	Stage  *BusinessStage `gorm:"type:varchar(50);index"`

	// --- Business Sector ---
	SectorID *uuid.UUID `gorm:"type:uuid;index"`
	Sector   *Sector    `gorm:"foreignKey:SectorID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	// --- Tags ---
	Tags []Tag `gorm:"many2many:business_profile_tags;"`
}

func (BusinessProfile) TableName() string {
	return "business_profiles"
}
