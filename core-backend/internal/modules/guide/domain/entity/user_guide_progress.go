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
	AccountID         uuid.UUID      `gorm:"type:uuid;not null;index:idx_user_progress_account;uniqueIndex:idx_user_progress_account_user_step,priority:1"`
	UserID            uuid.UUID      `gorm:"type:uuid;not null;index:idx_user_progress_user;uniqueIndex:idx_user_progress_account_user_step,priority:2"`
	StepID            uuid.UUID      `gorm:"type:uuid;not null;index:idx_user_progress_step;uniqueIndex:idx_user_progress_account_user_step,priority:3"`
	Step              GuideStep      `gorm:"foreignKey:StepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Status            ProgressStatus `gorm:"type:varchar(20);not null;default:'LOCKED';index:idx_user_progress_status"`
	StartedAt         *time.Time     `gorm:"type:timestamptz"`
	CompletedAt       *time.Time     `gorm:"type:timestamptz"`
	TimeSpent         *int           `gorm:"column:time_spent"`
	Notes             *string        `gorm:"type:text"`
	UploadedDocuments datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'"`
	LastAccessedAt    *time.Time     `gorm:"type:timestamptz"`
	StepVersion       int            `gorm:"not null;column:version"`
}

func (UserGuideProgress) TableName() string {
	return "user_guide_progresses"
}
