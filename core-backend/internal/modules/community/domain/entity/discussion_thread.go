package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type DiscussionThread struct {
	model.BaseModel `gorm:"embedded"`
	CategoryID      uuid.UUID            `gorm:"type:uuid;not null;uniqueIndex:idx_discussion_threads_slug_per_category,priority:1;index:idx_discussion_threads_category"`
	Category        CommunityCategory    `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AuthorAccountID uuid.UUID            `gorm:"type:uuid;not null;index:idx_discussion_threads_author"`
	Title           string               `gorm:"type:varchar(200);not null"`
	Slug            string               `gorm:"type:varchar(200);not null;uniqueIndex:idx_discussion_threads_slug_per_category,priority:2"`
	Description     *string              `gorm:"type:text"`
	IsPinned        bool                 `gorm:"not null;default:false"`
	Status          ThreadStatus         `gorm:"type:varchar(20);not null;default:'active';index:idx_discussion_threads_status"`
	ViewCount       int                  `gorm:"not null;default:0"`
	ShareCount      int                  `gorm:"not null;default:0"`
	ReplyCount      int                  `gorm:"not null;default:0"`
	LastActivityAt  *time.Time           `gorm:"type:timestamptz;index:idx_discussion_threads_last_activity"`
	Posts           []DiscussionPost     `gorm:"foreignKey:ThreadID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Followers       []UserThreadSettings `gorm:"foreignKey:ThreadID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Blocks          []ThreadBlockedUser  `gorm:"foreignKey:ThreadID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (DiscussionThread) TableName() string {
	return "discussion_threads"
}
