package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type RoleHandler struct {
	roleUsecase       usecase.RoleUsecase
	permissionUsecase usecase.PermissionUsecase
}

func NewRoleHandler(roleUsecase usecase.RoleUsecase, permissionUsecase usecase.PermissionUsecase) *RoleHandler {
	return &RoleHandler{
		roleUsecase:       roleUsecase,
		permissionUsecase: permissionUsecase,
	}
}

func (h *RoleHandler) HandleListRoles(ctx context.Context, _ *dto.ListRolesInput) (*dto.ListRolesOutput, error) {
	roles, err := h.roleUsecase.ListRoles(ctx)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	response := make([]dto.RoleDTO, 0, len(roles))
	for _, role := range roles {
		response = append(response, dto.RoleDTO{
			ID:          role.ID,
			Code:        role.Code,
			Name:        role.Name,
			Description: role.Description,
			Type:        string(role.Type),
			IsSystem:    role.IsSystem,
			IsMutable:   role.IsMutable,
		})
	}

	return &dto.ListRolesOutput{
		Body: dto.ListRolesResponseBody{Roles: response},
	}, nil
}

func (h *RoleHandler) HandleGetRole(ctx context.Context, input *dto.GetRoleInput) (*dto.GetRoleOutput, error) {
	role, err := h.roleUsecase.GetRole(ctx, input.RoleID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	permissions, err := h.permissionUsecase.ListPermissionsByRole(ctx, input.RoleID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	permissionDTOs := make([]dto.PermissionDTO, 0, len(permissions))
	for _, permission := range permissions {
		permissionDTOs = append(permissionDTOs, dto.PermissionDTO{
			ID:          permission.ID,
			Code:        permission.Code,
			Name:        permission.Name,
			Description: permission.Description,
			Module:      permission.Module,
		})
	}

	return &dto.GetRoleOutput{
		Body: dto.GetRoleResponseBody{
			Role: dto.RoleDTO{
				ID:          role.ID,
				Code:        role.Code,
				Name:        role.Name,
				Description: role.Description,
				Type:        string(role.Type),
				IsSystem:    role.IsSystem,
				IsMutable:   role.IsMutable,
			},
			Permissions: permissionDTOs,
		},
	}, nil
}

func (h *RoleHandler) HandleCreateRole(ctx context.Context, input *dto.CreateRoleInput) (*dto.CreateRoleOutput, error) {
	role, err := h.roleUsecase.CreateRole(ctx, usecase.CreateRoleInput{
		Code:          input.Body.Code,
		Name:          input.Body.Name,
		Description:   input.Body.Description,
		PermissionIDs: input.Body.PermissionIDs,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.CreateRoleOutput{
		Body: dto.CreateRoleResponseBody{
			Role: dto.RoleDTO{
				ID:          role.ID,
				Code:        role.Code,
				Name:        role.Name,
				Description: role.Description,
				Type:        string(role.Type),
				IsSystem:    role.IsSystem,
				IsMutable:   role.IsMutable,
			},
		},
	}, nil
}

func (h *RoleHandler) HandleUpdateRole(ctx context.Context, input *dto.UpdateRoleInput) (*dto.UpdateRoleOutput, error) {
	permissionIDs := input.Body.PermissionIDs
	role, err := h.roleUsecase.UpdateRole(ctx, input.RoleID, usecase.UpdateRoleInput{
		Name:          input.Body.Name,
		Description:   input.Body.Description,
		PermissionIDs: &permissionIDs,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.UpdateRoleOutput{
		Body: dto.UpdateRoleResponseBody{
			Role: dto.RoleDTO{
				ID:          role.ID,
				Code:        role.Code,
				Name:        role.Name,
				Description: role.Description,
				Type:        string(role.Type),
				IsSystem:    role.IsSystem,
				IsMutable:   role.IsMutable,
			},
		},
	}, nil
}

func (h *RoleHandler) HandleDeleteRole(ctx context.Context, input *dto.DeleteRoleInput) (*dto.DeleteRoleOutput, error) {
	if err := h.roleUsecase.DeleteRole(ctx, input.RoleID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.DeleteRoleOutput{
		Body: struct {
			Message string `json:"message" doc:"Status message"`
		}{
			Message: "Role deleted",
		},
	}, nil
}
