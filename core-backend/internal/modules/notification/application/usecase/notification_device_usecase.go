package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/google/uuid"
)

type notificationDeviceUsecase struct {
	deviceRepo repository.UserDeviceRepository
}

func NewNotificationDeviceUsecase(
	deviceRepo repository.UserDeviceRepository,
) usecase.NotificationDeviceUsecase {
	return &notificationDeviceUsecase{
		deviceRepo: deviceRepo,
	}
}

func (uc *notificationDeviceUsecase) RegisterDevice(ctx context.Context, accountID uuid.UUID, input usecase.RegisterDeviceInput) (*entity.UserDevice, error) {
	existing, err := uc.deviceRepo.GetByDeviceToken(ctx, input.DeviceToken)
	if err != nil && err != notiferror.ErrDeviceNotFound {
		return nil, err
	}

	if existing != nil {
		existing.AccountID = accountID
		existing.DeviceType = input.DeviceType
		if input.PushToken != nil {
			existing.PushToken = input.PushToken
		}
		if input.DeviceName != nil {
			existing.DeviceName = input.DeviceName
		}
		if input.DeviceModel != nil {
			existing.DeviceModel = input.DeviceModel
		}
		if input.OSVersion != nil {
			existing.OSVersion = input.OSVersion
		}
		if input.AppVersion != nil {
			existing.AppVersion = input.AppVersion
		}
		existing.IsActive = true
		now := time.Now().UTC()
		existing.LastActiveAt = &now

		if err := uc.deviceRepo.Update(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	now := time.Now().UTC()
	device := &entity.UserDevice{
		AccountID:    accountID,
		DeviceType:   input.DeviceType,
		DeviceToken:  input.DeviceToken,
		PushToken:    input.PushToken,
		DeviceName:   input.DeviceName,
		DeviceModel:  input.DeviceModel,
		OSVersion:    input.OSVersion,
		AppVersion:   input.AppVersion,
		IsActive:     true,
		LastActiveAt: &now,
	}

	if err := uc.deviceRepo.Create(ctx, device); err != nil {
		return nil, err
	}
	return device, nil
}

func (uc *notificationDeviceUsecase) UpdateDevice(ctx context.Context, accountID uuid.UUID, deviceID uuid.UUID, input usecase.UpdateDeviceInput) (*entity.UserDevice, error) {
	device, err := uc.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	if device.AccountID != accountID {
		return nil, notiferror.ErrDeviceNotFound
	}

	if input.PushToken != nil {
		device.PushToken = input.PushToken
	}
	if input.DeviceName != nil {
		device.DeviceName = input.DeviceName
	}
	if input.OSVersion != nil {
		device.OSVersion = input.OSVersion
	}
	if input.AppVersion != nil {
		device.AppVersion = input.AppVersion
	}
	if input.IsActive != nil {
		device.IsActive = *input.IsActive
	}

	now := time.Now().UTC()
	device.LastActiveAt = &now

	if err := uc.deviceRepo.Update(ctx, device); err != nil {
		return nil, err
	}
	return device, nil
}

func (uc *notificationDeviceUsecase) DeactivateDevice(ctx context.Context, accountID uuid.UUID, deviceID uuid.UUID) error {
	device, err := uc.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return err
	}
	if device.AccountID != accountID {
		return notiferror.ErrDeviceNotFound
	}

	return uc.deviceRepo.UpdateByID(ctx, deviceID, map[string]interface{}{
		"is_active": false,
	})
}

func (uc *notificationDeviceUsecase) ListDevices(ctx context.Context, accountID uuid.UUID) ([]*entity.UserDevice, error) {
	return uc.deviceRepo.ListByAccount(ctx, accountID)
}

func (uc *notificationDeviceUsecase) DeactivateAllDevices(ctx context.Context, accountID uuid.UUID) error {
	return uc.deviceRepo.DeactivateByAccount(ctx, accountID)
}
