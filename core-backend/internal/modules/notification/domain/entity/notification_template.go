package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"gorm.io/datatypes"
)

type NotificationTemplate struct {
	model.BaseModel  `gorm:"embedded"`
	Name             string                            `gorm:"type:varchar(200);not null;uniqueIndex:idx_notif_templates_name"`
	Description      *string                           `gorm:"type:text"`
	NotificationType NotificationType                  `gorm:"type:varchar(64);not null;uniqueIndex:idx_notif_templates_type"`
	TemplateGroup    string                            `gorm:"type:varchar(100);index:idx_notif_templates_group"`
	Priority         NotificationPriority              `gorm:"type:smallint;not null;default:1"`
	IsSystemManaged  bool                              `gorm:"not null;default:false"`
	DefaultContent   datatypes.JSONMap                 `gorm:"type:jsonb;not null"`
	VariablesSchema  *datatypes.JSONMap                `gorm:"type:jsonb"`
	DefaultTTL       *int                              `gorm:"type:integer"`
	EnablePushMirror bool                              `gorm:"not null;default:false"`
	Translations     []NotificationTemplateTranslation `gorm:"foreignKey:TemplateID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (NotificationTemplate) TableName() string {
	return "notification_templates"
}
