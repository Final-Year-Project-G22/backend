package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type DiscussionPost struct {
	model.BaseModel `gorm:"embedded"`
	ThreadID        uuid.UUID        `gorm:"type:uuid;not null;index:idx_discussion_posts_thread"`
	Thread          DiscussionThread `gorm:"foreignKey:ThreadID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ParentPostID    *uuid.UUID       `gorm:"type:uuid;index:idx_discussion_posts_parent"`
	ParentPost      *DiscussionPost  `gorm:"foreignKey:ParentPostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Replies         []DiscussionPost `gorm:"foreignKey:ParentPostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	AuthorAccountID uuid.UUID        `gorm:"type:uuid;not null;index:idx_discussion_posts_author"`
	Content         string           `gorm:"type:text;not null"`
	IsSolution      bool             `gorm:"not null;default:false;index:idx_discussion_posts_solution"`
	IsPinned        bool             `gorm:"not null;default:false"`
	UpvoteCount     int              `gorm:"not null;default:0"`
	EditCount       int              `gorm:"not null;default:0"`
	EditedAt        *time.Time       `gorm:"type:timestamptz"`
	Attachments     []Attachment     `gorm:"foreignKey:PostID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (DiscussionPost) TableName() string {
	return "discussion_posts"
}
