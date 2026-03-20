package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type UserGuideBookmark struct {
	model.BaseModel `gorm:"embedded"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_bookmark_user_step,priority:1;index:idx_bookmarks_user"`
	StepID          uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_bookmark_user_step,priority:2;index:idx_bookmarks_step"`
	Step            GuideStep `gorm:"foreignKey:StepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Note            *string   `gorm:"type:text"`
}

func (UserGuideBookmark) TableName() string {
	return "user_guide_bookmarks"
}
