package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type ComplianceEntry struct {
	model.BaseModel    `gorm:"embedded"`
	BusinessProfileID  uuid.UUID             `gorm:"type:uuid;not null;index:idx_compliance_profile"`
	AccountID          uuid.UUID             `gorm:"type:uuid;not null;index:idx_compliance_account"`
	ComplianceType     ComplianceType        `gorm:"type:varchar(64);not null"`
	ReferenceNumber    *string               `gorm:"type:varchar(255)"`
	IssuedDate         *time.Time            `gorm:"type:date"`
	ExpiryDate         time.Time             `gorm:"type:timestamptz;not null;index:idx_compliance_expiry"`
	ReminderDaysBefore int                   `gorm:"not null;default:30"`
	Source             ComplianceSource      `gorm:"type:varchar(20);not null;default:'manual';index:idx_compliance_source"`
	Status             ComplianceEntryStatus `gorm:"type:varchar(20);not null;default:'active';index:idx_compliance_status"`
	LastNotifiedAt     *time.Time            `gorm:"type:timestamptz"`
}

func (ComplianceEntry) TableName() string {
	return "compliance_entries"
}
