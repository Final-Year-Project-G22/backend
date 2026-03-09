package entity

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type NotificationPreference struct {
	model.BaseModel `gorm:"embedded"`

	AccountID               uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Account                 Account   `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	EnableEmailNotification bool      `gorm:"not null;default:true"`
	EnableSMSNotification   bool      `gorm:"not null;default:false"`
	EnablePushNotification  bool      `gorm:"not null;default:false"`
	CampaignDigestEnabled   bool      `gorm:"not null;default:true"`
}

func (NotificationPreference) TableName() string {
	return "notification_preferences"
}
