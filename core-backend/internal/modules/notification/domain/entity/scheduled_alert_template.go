package entity

import "github.com/Final-Year-Project-G22/backend/core/internal/shared/model"

type ScheduledAlertTemplate struct {
	model.BaseModel `gorm:"embedded"`
	Slug            string   `gorm:"type:varchar(64);uniqueIndex;not null"`
	Name            string   `gorm:"type:varchar(255);not null"`
	DefaultTitle    string   `gorm:"type:varchar(255);not null"`
	DefaultBody     string   `gorm:"type:text;not null"`
	DefaultChannel  *Channel `gorm:"type:varchar(20)"`
	IsActive        bool     `gorm:"not null;default:true"`
}

func (ScheduledAlertTemplate) TableName() string {
	return "scheduled_alert_templates"
}
