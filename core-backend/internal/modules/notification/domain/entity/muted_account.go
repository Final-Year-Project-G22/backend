package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type MutedAccount struct {
	model.BaseModel `gorm:"embedded"`
	AccountID       uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_muted_accounts_account_pair,priority:1;index:idx_muted_accounts_account"`
	MutedAccountID  uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_muted_accounts_account_pair,priority:2;index:idx_muted_accounts_muted"`
	MuteUntil       *time.Time `gorm:"type:timestamptz"`
	Reason          *string    `gorm:"type:text"`
}

func (MutedAccount) TableName() string {
	return "muted_accounts"
}
