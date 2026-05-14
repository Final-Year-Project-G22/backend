package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type NotificationCampaign struct {
	model.BaseModel    `gorm:"embedded"`
	Name               string             `gorm:"type:varchar(200);not null"`
	Description        *string            `gorm:"type:text"`
	CampaignType       CampaignType       `gorm:"type:varchar(20);not null"`
	TargetSegment      *datatypes.JSONMap `gorm:"type:jsonb"`
	CampaignTemplateID uuid.UUID          `gorm:"type:uuid;not null"`
	ScheduledFor       *time.Time         `gorm:"type:timestamptz"`
	SentAt             *time.Time         `gorm:"type:timestamptz"`
	Status             CampaignStatus     `gorm:"type:varchar(20);not null;default:'draft'"`
	CreatedBy          uuid.UUID          `gorm:"type:uuid;not null;index:idx_notif_campaigns_creator"`
	SectorIDs          pq.StringArray     `gorm:"type:uuid[];index:idx_notif_campaigns_sector_ids,using:gin"`
	TagIDs             pq.StringArray     `gorm:"type:uuid[];index:idx_notif_campaigns_tag_ids,using:gin"`
	Region             *string            `gorm:"type:varchar(50)"`
	Stage              *string            `gorm:"type:varchar(50)"`
}

func (NotificationCampaign) TableName() string {
	return "notification_campaigns"
}
