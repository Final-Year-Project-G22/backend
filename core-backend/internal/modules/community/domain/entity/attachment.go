package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type Attachment struct {
	model.BaseModel `gorm:"embedded"`
	StorageKey      string           `gorm:"type:text;not null"`
	FileURL         string           `gorm:"type:text;not null"`
	FileType        string           `gorm:"type:varchar(50);not null"`
	FileName        string           `gorm:"type:varchar(255);not null"`
	FileSize        *int64           `gorm:"type:bigint"`
	PostID          *uuid.UUID       `gorm:"type:uuid;index:idx_attachments_post"`
	Post            *DiscussionPost  `gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	UploadedBy      uuid.UUID        `gorm:"type:uuid;not null;index:idx_attachments_uploaded_by"`
	Status          AttachmentStatus `gorm:"type:varchar(20);not null;default:'pending'"`
}

func (Attachment) TableName() string {
	return "attachments"
}
