package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type UserGuideJourney struct {
	model.BaseModel    `gorm:"embedded"`
	AccountID          uuid.UUID      `gorm:"type:uuid;not null;index:idx_journey_account;uniqueIndex:idx_journey_account_user_guide,priority:1"`
	UserID             uuid.UUID      `gorm:"type:uuid;not null;index:idx_journey_user;uniqueIndex:idx_journey_account_user_guide,priority:2"`
	GuideID            uuid.UUID      `gorm:"type:uuid;not null;index:idx_journey_guide;uniqueIndex:idx_journey_account_user_guide,priority:3"`
	Guide              Guide          `gorm:"foreignKey:GuideID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	JourneyHash        *string        `gorm:"type:text"`
	StepSequence       datatypes.JSON `gorm:"type:jsonb;not null"`
	TotalSteps         int            `gorm:"not null"`
	CompletedSteps     int            `gorm:"not null;default:0"`
	EstimatedTotalTime *int           `gorm:"column:estimated_total_time"`
	GeneratedAt        time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
	ExpiresAt          *time.Time     `gorm:"type:timestamptz;index:idx_journey_expires"`
}

func (UserGuideJourney) TableName() string {
	return "user_guide_journeys"
}
