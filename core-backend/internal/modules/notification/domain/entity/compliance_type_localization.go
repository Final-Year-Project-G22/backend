package entity

import "github.com/Final-Year-Project-G22/backend/core/internal/shared/model"

type ComplianceTypeLocalization struct {
	model.BaseModel `gorm:"embedded"`
	ComplianceType  string `gorm:"type:varchar(64);not null;uniqueIndex:idx_ct_lang,priority:1"`
	Locale          string `gorm:"type:varchar(10);not null;uniqueIndex:idx_ct_lang,priority:2"`
	Label           string `gorm:"type:varchar(255);not null"`
}

func (ComplianceTypeLocalization) TableName() string {
	return "compliance_type_localizations"
}
