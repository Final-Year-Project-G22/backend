package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
)

type AccountStatus string

const (
	AccountStatusActive              AccountStatus = "active"
	AccountStatusSuspended           AccountStatus = "suspended"
	AccountStatusDeleted             AccountStatus = "deleted"
	AccountStatusPendingVerification AccountStatus = "pending_verification"
	AccountStatusDisabled            AccountStatus = "disabled"
)

type Account struct {
	model.BaseModel `gorm:"embedded"`

	Email                      string        `gorm:"type:varchar(255);not null;uniqueIndex"`
	PasswordHash               string        `gorm:"type:varchar(255);not null"`
	PhoneNumber                *string       `gorm:"type:varchar(255)"`
	EmailVerified              bool          `gorm:"not null;default:false"`
	PhoneNumberVerified        bool          `gorm:"not null;default:false"`
	AccountStatus              AccountStatus `gorm:"type:varchar(255);not null;default:'pending_verification'"`
	FailedLoginAttempts        int           `gorm:"not null;default:0"`
	MFAEnabled                 bool          `gorm:"not null;default:false"`
	MFASecret                  *string       `gorm:"type:varchar(255)"`
	LastLoginAt                time.Time     `gorm:"type:timestamp"`
	EmailVerificationToken     *string       `gorm:"type:varchar(255);uniqueIndex:idx_email_verification_token"`
	EmailVerificationExpiresAt *time.Time    `gorm:"type:timestamp"`
	PasswordResetToken         *string       `gorm:"type:varchar(255);uniqueIndex:idx_password_reset_token"`
	PasswordResetExpiresAt     *time.Time    `gorm:"type:timestamp"`
}

func (Account) TableName() string {
	return "accounts"
}
