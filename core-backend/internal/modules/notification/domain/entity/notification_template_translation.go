package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type NotificationTemplateTranslation struct {
	ID         uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TemplateID uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex:idx_notif_template_trans,priority:1"`
	Language   string            `gorm:"type:varchar(10);not null;uniqueIndex:idx_notif_template_trans,priority:2"`
	Subject    string            `gorm:"type:varchar(500);not null"`
	Content    datatypes.JSONMap `gorm:"type:jsonb;not null"`
	CreatedAt  *time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt  *time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (t *NotificationTemplateTranslation) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

func (NotificationTemplateTranslation) TableName() string {
	return "notification_template_translations"
}
