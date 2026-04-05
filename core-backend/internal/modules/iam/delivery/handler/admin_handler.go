package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type AdminHandler struct {
	adminService service.AdminService
}

func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

func (h *AdminHandler) HandleRegisterAdmin(ctx context.Context, input *dto.AdminRegisterInput) (*dto.AdminRegisterOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, apperrors.UnauthorizedError("iam.errors.unauthorized")
	}

	result, err := h.adminService.RegisterAdmin(ctx, service.RegisterAdminInput{
		Email:      input.Body.Email,
		Username:   input.Body.Username,
		FirstName:  input.Body.FirstName,
		LastName:   input.Body.LastName,
		RoleIDs:    input.Body.RoleIDs,
		AssignedBy: accountID,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.AdminRegisterOutput{
		Body: dto.AdminRegisterResponseBody{
			AccountID: result.AccountID,
			Message:   "Admin account created",
		},
	}, nil
}

func (h *AdminHandler) HandleUpdateAdminRoles(ctx context.Context, input *dto.AdminUpdateRolesInput) (*dto.AdminUpdateRolesOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, apperrors.UnauthorizedError("iam.errors.unauthorized")
	}

	if err := h.adminService.UpdateAdminRoles(ctx, service.UpdateAdminRolesInput{
		AccountID: input.AccountID,
		RoleIDs:   input.Body.RoleIDs,
		UpdatedBy: accountID,
	}); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.AdminUpdateRolesOutput{
		Body: struct {
			Message string `json:"message" doc:"Status message"`
		}{
			Message: "Admin roles updated",
		},
	}, nil
}
