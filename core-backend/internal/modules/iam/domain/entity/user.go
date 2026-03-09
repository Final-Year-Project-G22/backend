package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
)

type User struct {
	model.BaseModel `gorm:"embedded"`

	FirstName string    `gorm:"type:varchar(100);not null"`
	LastName  string    `gorm:"type:varchar(100);not null"`
	ImageURL  *string   `gorm:"type:varchar(512)"`
	Bio       *string   `gorm:"type:text"`
	Accounts  []Account `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (User) TableName() string {
	return "users"
}
