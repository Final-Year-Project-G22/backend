package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type UserGuideRecentView struct {
	model.BaseModel `gorm:"embedded"`
	AccountID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_recent_view_account_user_guide,priority:1;index:idx_recent_view_account"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_recent_view_account_user_guide,priority:2;index:idx_recent_view_user"`
	GuideID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_recent_view_account_user_guide,priority:3;index:idx_recent_view_guide"`
	Guide           Guide     `gorm:"foreignKey:GuideID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	LastViewedAt    time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP;index:idx_recent_view_last_viewed"`
	ViewCount       int       `gorm:"not null;default:1"`
}

func (UserGuideRecentView) TableName() string {
	return "user_guide_recent_views"
}
