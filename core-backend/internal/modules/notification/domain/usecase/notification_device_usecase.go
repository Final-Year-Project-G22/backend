package usecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/google/uuid"
)

type NotificationDeviceUsecase interface {
	RegisterDevice(ctx context.Context, accountID uuid.UUID, input RegisterDeviceInput) (*entity.UserDevice, error)
	UpdateDevice(ctx context.Context, accountID uuid.UUID, deviceID uuid.UUID, input UpdateDeviceInput) (*entity.UserDevice, error)
	DeactivateDevice(ctx context.Context, accountID uuid.UUID, deviceID uuid.UUID) error
	ListDevices(ctx context.Context, accountID uuid.UUID) ([]*entity.UserDevice, error)
	DeactivateAllDevices(ctx context.Context, accountID uuid.UUID) error
}
