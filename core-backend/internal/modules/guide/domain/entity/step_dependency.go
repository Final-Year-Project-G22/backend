package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type DependencyType string

const (
	DependencyTypeMandatory DependencyType = "MANDATORY"
	DependencyTypeOptional  DependencyType = "OPTIONAL"
)

type StepDependency struct {
	model.BaseModel `gorm:"embedded"`
	StepID          uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_step_dependencies_unique,priority:1;index:idx_step_dependencies_step"`
	Step            GuideStep      `gorm:"foreignKey:StepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RequiredStepID  uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_step_dependencies_unique,priority:2;index:idx_step_dependencies_required"`
	RequiredStep    GuideStep      `gorm:"foreignKey:RequiredStepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	DependencyType  DependencyType `gorm:"type:varchar(20);not null;default:'MANDATORY'"`
}

func (StepDependency) TableName() string {
	return "step_dependencies"
}
