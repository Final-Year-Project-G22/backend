package entity

import "github.com/Final-Year-Project-G22/backend/core/internal/shared/model"

type User struct {
	model.BaseModel `gorm:"embedded"`
}
