package entity

import "github.com/Final-Year-Project-G22/backend/core/internal/shared/model"

// Plan represents a subscription plan seeded in the database.
type Plan struct {
	model.BaseModel `gorm:"embedded"`
	Name            string `gorm:"type:varchar(50);not null;uniqueIndex:idx_plans_name_period"`
	Period          string `gorm:"type:varchar(20);not null;uniqueIndex:idx_plans_name_period"`
	Amount          int64  `gorm:"not null"`
	Currency        string `gorm:"type:varchar(3);not null;default:ETB"`
	IsActive        bool   `gorm:"not null;default:true"`
}

func (Plan) TableName() string {
	return "plans"
}
