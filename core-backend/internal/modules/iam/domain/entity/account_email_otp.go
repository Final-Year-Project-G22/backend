package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type AccountEmailOTP struct {
	model.BaseModel `gorm:"embedded"`

	AccountID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_account_email_otps_account_id"`
	Account      Account    `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CodeHash     string     `gorm:"type:varchar(255);not null"`
	ExpiresAt    time.Time  `gorm:"type:timestamptz;not null;index:idx_account_email_otps_expires_at"`
	ConsumedAt   *time.Time `gorm:"type:timestamptz"`
	AttemptCount int        `gorm:"not null;default:0"`
	ResendCount  int        `gorm:"not null;default:0"`
	LastSentAt   time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
	Purpose      string     `gorm:"type:varchar(32);not null;default:'email_verification'"`
}

type EmailOTPPurpose string

const (
	EmailOTPPurposeVerification  EmailOTPPurpose = "email_verification"
	EmailOTPPurposePasswordReset EmailOTPPurpose = "password_reset"
)

func (AccountEmailOTP) TableName() string {
	return "account_email_otps"
}
