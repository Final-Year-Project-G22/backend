package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type NotificationCampaign struct {
	model.BaseModel `gorm:"embedded"`
	Name            string             `gorm:"type:varchar(200);not null"`
	Description     *string            `gorm:"type:text"`
	CampaignType    CampaignType       `gorm:"type:varchar(20);not null"`
	TargetSegment   *datatypes.JSONMap `gorm:"type:jsonb"`
	TemplateID      uuid.UUID          `gorm:"type:uuid;not null"`
	CustomSubject   *string            `gorm:"type:varchar(500)"`
	CustomContent   *datatypes.JSONMap `gorm:"type:jsonb"`
	ScheduledFor    *time.Time         `gorm:"type:timestamptz"`
	SentAt          *time.Time         `gorm:"type:timestamptz"`
	Status          CampaignStatus     `gorm:"type:varchar(20);not null;default:'draft'"`
	CreatedBy       uuid.UUID          `gorm:"type:uuid;not null;index:idx_notif_campaigns_creator"`
}

func (NotificationCampaign) TableName() string {
	return "notification_campaigns"
}
