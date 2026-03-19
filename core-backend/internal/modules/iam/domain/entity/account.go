package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type AccountStatus string

const (
	AccountStatusPendingVerification AccountStatus = "pending_verification"
	AccountStatusActive              AccountStatus = "active"
	AccountStatusLocked              AccountStatus = "locked"
	AccountStatusSuspended           AccountStatus = "suspended"
	AccountStatusDisabled            AccountStatus = "disabled"
)

type Account struct {
	model.BaseModel `gorm:"embedded"`

	UserID              uuid.UUID     `gorm:"type:uuid;not null;index"`
	User                User          `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Email               string        `gorm:"type:varchar(255);not null"`
	EmailNormalized     string        `gorm:"type:varchar(255);not null;uniqueIndex:idx_accounts_email_normalized"`
	PasswordHash        *string       `gorm:"type:varchar(255)"`
	PhoneNumber         *string       `gorm:"type:varchar(50)"`
	EmailVerified       bool          `gorm:"not null;default:false"`
	PhoneVerified       bool          `gorm:"not null;default:false"`
	Status              AccountStatus `gorm:"type:varchar(64);not null;default:'pending_verification'"`
	FailedLoginAttempts int           `gorm:"not null;default:0"`
	LockedUntil         *time.Time    `gorm:"type:timestamptz"`
	LastLoginAt         *time.Time    `gorm:"type:timestamptz"`
	BusinessProfile     *BusinessProfile
	OAuthIdentities     []OAuthIdentity         `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Sessions            []Session               `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RoleAssignments     []RoleAssignment        `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AccountPreference   *AccountPreference      `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	NotificationPref    *NotificationPreference `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CommunityPreference *CommunityPreference    `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AIPreference        *AIPreference           `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TemplatePreference  *TemplatePreference     `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Account) TableName() string {
	return "accounts"
}
