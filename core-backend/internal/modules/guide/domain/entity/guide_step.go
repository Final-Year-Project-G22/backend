package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type StepType string

const (
	StepTypeInformational      StepType = "INFORMATIONAL"
	StepTypeActionRequired     StepType = "ACTION_REQUIRED"
	StepTypeDocumentSubmission StepType = "DOCUMENT_SUBMISSION"
	StepTypeVerification       StepType = "VERIFICATION"
)

type GuideStep struct {
	model.BaseModel     `gorm:"embedded"`
	GuideID             uuid.UUID              `gorm:"type:uuid;not null;uniqueIndex:idx_guide_steps_slug_per_guide,priority:1;uniqueIndex:idx_guide_steps_sort_per_guide,priority:1;index:idx_guide_steps_guide"`
	Guide               Guide                  `gorm:"foreignKey:GuideID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Slug                string                 `gorm:"type:varchar(200);not null;uniqueIndex:idx_guide_steps_slug_per_guide,priority:2"`
	StepType            StepType               `gorm:"type:varchar(64);not null"`
	SortOrder           int                    `gorm:"not null;default:0;uniqueIndex:idx_guide_steps_sort_per_guide,priority:2"`
	EstimatedTime       *int                   `gorm:"column:estimated_time"`
	DifficultyLevel     *int                   `gorm:"check:difficulty_level_check,difficulty_level BETWEEN 1 AND 5"`
	IsOptional          bool                   `gorm:"not null;default:false"`
	ExternalLinks       *datatypes.JSON        `gorm:"type:jsonb;default:'[]'"`
	FeeEstimate         *int                   `gorm:"column:fee_estimate"`
	Version             int                    `gorm:"not null;default:1"`
	EffectiveDate       time.Time              `gorm:"type:date;not null;default:CURRENT_DATE"`
	ExpiryDate          *time.Time             `gorm:"type:date"`
	Conditions          []StepCondition        `gorm:"foreignKey:StepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Dependencies        []StepDependency       `gorm:"foreignKey:StepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ReverseDependencies []StepDependency       `gorm:"foreignKey:RequiredStepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Progress            []UserGuideProgress    `gorm:"foreignKey:StepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Bookmarks           []UserGuideBookmark    `gorm:"foreignKey:StepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Versions            []GuideStepVersion     `gorm:"foreignKey:StepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Translations        []GuideStepTranslation `gorm:"foreignKey:GuideStepID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (GuideStep) TableName() string {
	return "guide_steps"
}
