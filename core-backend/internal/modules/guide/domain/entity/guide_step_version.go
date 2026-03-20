package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type GuideStepVersion struct {
	model.BaseModel `gorm:"embedded"`
	StepID          uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex:idx_version_step_version,priority:1;index:idx_versions_step"`
	Step            GuideStep         `gorm:"foreignKey:StepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Version         int               `gorm:"not null;uniqueIndex:idx_version_step_version,priority:2"`
	Title           string            `gorm:"type:varchar(200);not null"`
	Content         datatypes.JSONMap `gorm:"type:jsonb;not null"`
	EffectiveDate   time.Time         `gorm:"type:date;not null"`
}

func (GuideStepVersion) TableName() string {
	return "guide_step_versions"
}
