package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ProgressStatus string

const (
	ProgressStatusLocked     ProgressStatus = "LOCKED"
	ProgressStatusInProgress ProgressStatus = "IN_PROGRESS"
	ProgressStatusCompleted  ProgressStatus = "COMPLETED"
	ProgressStatusSkipped    ProgressStatus = "SKIPPED"
)

type UserGuideProgress struct {
	model.BaseModel   `gorm:"embedded"`
	UserID            uuid.UUID         `gorm:"type:uuid;not null;index:idx_user_progress_user"`
	StepID            uuid.UUID         `gorm:"type:uuid;not null;index:idx_user_progress_step"`
	Step              GuideStep         `gorm:"foreignKey:StepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Status            ProgressStatus    `gorm:"type:varchar(20);not null;default:'LOCKED';index:idx_user_progress_status"`
	StartedAt         *time.Time        `gorm:"type:timestamptz"`
	CompletedAt       *time.Time        `gorm:"type:timestamptz"`
	TimeSpent         *int              `gorm:"column:time_spent"`
	Notes             *string           `gorm:"type:text"`
	UploadedDocuments datatypes.JSONMap `gorm:"type:jsonb;not null;default:'[]'"`
	LastAccessedAt    *time.Time        `gorm:"type:timestamptz"`
	StepVersion       int               `gorm:"not null;column:version"`
}

func (UserGuideProgress) TableName() string {
	return "user_guide_progresses"
}
