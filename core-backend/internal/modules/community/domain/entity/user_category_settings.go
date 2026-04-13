package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type UserCategorySettings struct {
	model.BaseModel `gorm:"embedded"`
	AccountID       uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex:idx_user_category_settings_account_category,priority:1;index:idx_user_category_settings_account"`
	CategoryID      uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex:idx_user_category_settings_account_category,priority:2;index:idx_user_category_settings_category"`
	Category        CommunityCategory `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	IsFollowing     bool              `gorm:"not null;default:true"`
	IsMuted         bool              `gorm:"not null;default:false"`
	LastReadAt      *time.Time        `gorm:"type:timestamptz"`
}

func (UserCategorySettings) TableName() string {
	return "user_category_settings"
}
