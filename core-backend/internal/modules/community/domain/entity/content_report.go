package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type ContentReport struct {
	model.BaseModel     `gorm:"embedded"`
	ReporterAccountID   uuid.UUID    `gorm:"type:uuid;not null;index:idx_content_reports_reporter"`
	TargetType          TargetType   `gorm:"type:varchar(20);not null;index:idx_content_reports_target"`
	TargetID            uuid.UUID    `gorm:"type:uuid;not null;index:idx_content_reports_target"`
	Reason              string       `gorm:"type:text;not null"`
	Status              ReportStatus `gorm:"type:varchar(20);not null;default:'pending';index:idx_content_reports_status"`
	AdminNote           *string      `gorm:"type:text"`
	ResolvedByAccountID *uuid.UUID   `gorm:"type:uuid;index:idx_content_reports_resolved_by"`
	ResolvedAt          *time.Time   `gorm:"type:timestamptz"`
}

func (ContentReport) TableName() string {
	return "content_reports"
}
