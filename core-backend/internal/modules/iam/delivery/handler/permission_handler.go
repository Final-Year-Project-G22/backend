package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type PermissionHandler struct {
	permissionUsecase usecase.PermissionUsecase
}

func NewPermissionHandler(permissionUsecase usecase.PermissionUsecase) *PermissionHandler {
	return &PermissionHandler{
		permissionUsecase: permissionUsecase,
	}
}

func (h *PermissionHandler) HandleListPermissions(ctx context.Context, input *dto.ListPermissionsInput) (*dto.ListPermissionsOutput, error) {
	permissions, err := h.permissionUsecase.ListPermissions(ctx, usecase.ListPermissionsInput{
		Codes:  input.Codes,
		Module: input.Module,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	response := make([]dto.PermissionDTO, 0, len(permissions))
	for _, permission := range permissions {
		response = append(response, dto.PermissionDTO{
			ID:          permission.ID,
			Code:        permission.Code,
			Name:        permission.Name,
			Description: permission.Description,
			Module:      permission.Module,
		})
	}

	return &dto.ListPermissionsOutput{
		Body: dto.ListPermissionsResponseBody{
			Permissions: response,
		},
	}, nil
}
