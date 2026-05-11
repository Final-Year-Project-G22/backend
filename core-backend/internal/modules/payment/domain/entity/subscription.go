package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

// Subscription represents a subscription record for an account.
type Subscription struct {
	model.BaseModel    `gorm:"embedded"`
	AccountID          uuid.UUID          `gorm:"type:uuid;not null;index:idx_subscriptions_account"`
	PlanName           string             `gorm:"type:varchar(50);not null"`
	PlanPeriod         string             `gorm:"type:varchar(20);not null"`
	Amount             int64              `gorm:"not null"`
	Currency           string             `gorm:"type:varchar(3);not null;default:ETB"`
	Status             SubscriptionStatus `gorm:"type:varchar(20);not null;default:active;index:idx_subscriptions_status"`
	CurrentPeriodStart time.Time          `gorm:"not null"`
	CurrentPeriodEnd   time.Time          `gorm:"not null;index:idx_subscriptions_period_end"`
	CancelledAt        *time.Time
	RenewalCount       int `gorm:"not null;default:0"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}
