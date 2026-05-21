package service

import "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"

type ComplianceDefault struct {
	ReminderDays       int
	ExpiryDurationDays int
}

var ComplianceDefaults = map[entity.ComplianceType]ComplianceDefault{
	entity.ComplianceTypeTaxRegistration:      {ReminderDays: 30, ExpiryDurationDays: 365},
	entity.ComplianceTypeTradeLicense:         {ReminderDays: 30, ExpiryDurationDays: 365},
	entity.ComplianceTypeBusinessRegistration: {ReminderDays: 30, ExpiryDurationDays: 365},
}
