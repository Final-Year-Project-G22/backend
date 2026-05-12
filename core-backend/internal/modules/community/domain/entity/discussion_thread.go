package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/dbtypes"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type DiscussionThread struct {
	model.BaseModel `gorm:"embedded"`
	SectorIDs       dbtypes.UUIDArray    `gorm:"type:uuid[];index:idx_discussion_threads_sector_ids,using:gin"`
	TagIDs          dbtypes.UUIDArray    `gorm:"type:uuid[];index:idx_discussion_threads_tag_ids,using:gin"`
	ParentThreadID  *uuid.UUID           `gorm:"type:uuid;index:idx_threads_parent;uniqueIndex:idx_threads_slug_per_parent,priority:1"`
	ParentThread    *DiscussionThread    `gorm:"foreignKey:ParentThreadID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	SubThreads      []DiscussionThread   `gorm:"foreignKey:ParentThreadID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	AuthorAccountID uuid.UUID            `gorm:"type:uuid;not null;index:idx_discussion_threads_author"`
	Title           string               `gorm:"type:varchar(200);not null"`
	Slug            string               `gorm:"type:varchar(200);not null;uniqueIndex:idx_discussion_threads_slug"`
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
