package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type StepCondition struct {
	model.BaseModel `gorm:"embedded"`
	StepID          uuid.UUID         `gorm:"type:uuid;not null;index:idx_step_conditions_step"`
	Step            GuideStep         `gorm:"foreignKey:StepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ConditionType   string            `gorm:"type:varchar(50);not null"`
	Operator        string            `gorm:"type:varchar(20);not null"`
	ConditionValue  datatypes.JSONMap `gorm:"type:jsonb;not null"`
	IsInverse       bool              `gorm:"not null;default:false"`
}

func (StepCondition) TableName() string {
	return "step_conditions"
}
