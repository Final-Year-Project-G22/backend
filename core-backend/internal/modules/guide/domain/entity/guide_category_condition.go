package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type GuideCategoryCondition struct {
	model.BaseModel `gorm:"embedded"`
	CategoryID      uuid.UUID         `gorm:"type:uuid;not null;index:idx_guide_category_conditions_category"`
	Category        GuideCategory     `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ConditionType   string            `gorm:"type:varchar(50);not null"`
	Operator        string            `gorm:"type:varchar(20);not null"`
	ConditionValue  datatypes.JSONMap `gorm:"type:jsonb;not null"`
	IsInverse       bool              `gorm:"not null;default:false"`
}

func (GuideCategoryCondition) TableName() string {
	return "guide_category_conditions"
}
