package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Payment represents a single payment transaction attempt.
type Payment struct {
	model.BaseModel `gorm:"embedded"`
	AccountID       uuid.UUID     `gorm:"type:uuid;not null;index:idx_payments_account_id;index:idx_payments_account_status,priority:1"`
	SubscriptionID  *uuid.UUID    `gorm:"type:uuid;index"`
	TxRef           string        `gorm:"type:varchar(255);not null;uniqueIndex:idx_payments_tx_ref"`
	ChapaRef        *string       `gorm:"type:varchar(255);index:idx_payments_chapa_ref"`
	Amount          int64         `gorm:"not null"`
	Currency        string        `gorm:"type:varchar(3);not null;default:ETB"`
	PlanName        string        `gorm:"type:varchar(50);not null"`
	PlanPeriod      string        `gorm:"type:varchar(20);not null"`
	Status          PaymentStatus `gorm:"type:varchar(20);not null;default:pending;index:idx_payments_status;index:idx_payments_account_status,priority:2"`
	PaymentMethod   *string       `gorm:"type:varchar(50)"`
	CheckedOutAt    *time.Time
	VerifiedAt      *time.Time
	FailedAt        *time.Time
	Metadata        datatypes.JSONMap `gorm:"type:jsonb;default:'{}'"`
}

func (Payment) TableName() string {
	return "payments"
}
