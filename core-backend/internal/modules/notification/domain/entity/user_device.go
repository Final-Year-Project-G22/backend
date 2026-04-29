package entity

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/google/uuid"
)

type UserDevice struct {
	model.BaseModel `gorm:"embedded"`
	AccountID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_user_devices_account"`
	DeviceType      DeviceType `gorm:"type:varchar(20);not null"`
	DeviceToken     string     `gorm:"type:varchar(512);not null;uniqueIndex:idx_user_devices_token"`
	PushToken       *string    `gorm:"type:text"`
	DeviceName      *string    `gorm:"type:varchar(200)"`
	DeviceModel     *string    `gorm:"type:varchar(200)"`
	OSVersion       *string    `gorm:"type:varchar(50)"`
	AppVersion      *string    `gorm:"type:varchar(50)"`
	IsActive        bool       `gorm:"not null;default:true"`
	LastActiveAt    *time.Time `gorm:"type:timestamptz"`
}

func (UserDevice) TableName() string {
	return "user_devices"
}
