package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type User struct {
	model.BaseModel `gorm:"embedded"`

	AccountId uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Account   Account   `gorm:"foreignKey:AccountId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
