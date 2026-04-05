package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type UserGuideBookmark struct {
	model.BaseModel `gorm:"embedded"`
	AccountID       uuid.UUID `gorm:"type:uuid;not null;index:idx_bookmarks_account;uniqueIndex:idx_bookmark_account_user_step,priority:1"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;index:idx_bookmarks_user;uniqueIndex:idx_bookmark_account_user_step,priority:2"`
	StepID          uuid.UUID `gorm:"type:uuid;not null;index:idx_bookmarks_step;uniqueIndex:idx_bookmark_account_user_step,priority:3"`
	Step            GuideStep `gorm:"foreignKey:StepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Note            *string   `gorm:"type:text"`
}

func (UserGuideBookmark) TableName() string {
	return "user_guide_bookmarks"
}
