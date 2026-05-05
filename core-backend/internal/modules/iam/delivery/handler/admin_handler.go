package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
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

func (h *AdminHandler) HandleListAdmins(ctx context.Context, input *dto.ListAdminsInput) (*dto.ListAdminsOutput, error) {
	result, err := h.adminService.ListAdmins(ctx, service.ListAdminsInput{
		Search:   input.Search,
		Status:   input.Status,
		RoleID:   input.RoleID,
		Page:     input.Page,
		PageSize: input.PageSize,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	adminDTOs := make([]dto.AdminAccountDTO, 0, len(result.Admins))
	for _, a := range result.Admins {
		roleDTOs := make([]dto.RoleDTO, 0, len(a.Roles))
		for _, r := range a.Roles {
			roleDTOs = append(roleDTOs, dto.RoleDTO{
				ID:   r.ID,
				Code: r.Code,
				Name: r.Name,
			})
		}
		adminDTOs = append(adminDTOs, dto.AdminAccountDTO{
			ID:          a.ID,
			Email:       a.Email,
			Username:    a.Username,
			Status:      a.Status,
			FirstName:   a.FirstName,
			LastName:    a.LastName,
			Roles:       roleDTOs,
			CreatedAt:   a.CreatedAt,
			LastLoginAt: a.LastLogin,
		})
	}

	totalPages := result.TotalPages
	if totalPages == 0 && result.Total > 0 {
		totalPages = 1
	}

	return &dto.ListAdminsOutput{
		Body: dto.AdminListResponseBody{
			Admins:     adminDTOs,
			Total:      result.Total,
			Page:       result.Page,
			PageSize:   result.PageSize,
			TotalPages: totalPages,
		},
	}, nil
}

func (h *AdminHandler) HandleUpdateAdminStatus(ctx context.Context, input *dto.UpdateAdminStatusInput) (*dto.UpdateAdminStatusOutput, error) {
	triggeredBy := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if triggeredBy == contextkeys.NilUUID {
		return nil, apperrors.UnauthorizedError("iam.errors.unauthorized")
	}

	status, err := entity.ParseAccountStatus(input.Body.Status)
	if err != nil {
		return nil, apperrors.BadRequestError("iam.errors.invalidInput")
	}

	if err := h.adminService.UpdateAdminStatus(ctx, service.UpdateAdminStatusInput{
		AccountID: input.AccountID,
		Status:    status,
	}); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.UpdateAdminStatusOutput{
		Body: struct {
			Message string `json:"message" doc:"Status message"`
		}{
			Message: "Admin status updated",
		},
	}, nil
}

func (h *AdminHandler) HandleResetAdminPassword(ctx context.Context, input *dto.ResetAdminPasswordInput) (*dto.ResetAdminPasswordOutput, error) {
	triggeredBy := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if triggeredBy == contextkeys.NilUUID {
		return nil, apperrors.UnauthorizedError("iam.errors.unauthorized")
	}

	if err := h.adminService.ResetAdminPassword(ctx, service.ResetAdminPasswordInput{
		AccountID:   input.AccountID,
		TriggeredBy: triggeredBy,
	}); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.ResetAdminPasswordOutput{
		Body: struct {
			Message string `json:"message" doc:"Status message"`
		}{
			Message: "Password reset link sent",
		},
	}, nil
}

func (h *AdminHandler) HandleCompleteAdminPasswordReset(ctx context.Context, input *dto.CompleteAdminPasswordResetInput) (*dto.CompleteAdminPasswordResetOutput, error) {
	if err := h.adminService.CompletePasswordReset(ctx, input.Body.Token, input.Body.NewPassword); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.CompleteAdminPasswordResetOutput{
		Body: struct {
			Message string `json:"message" doc:"Status message"`
		}{
			Message: "Password reset successfully",
		},
	}, nil
}

func (h *AdminHandler) HandleListAdminsOld(ctx context.Context, input *dto.ListAdminsInput) (*dto.ListAdminsOutput, error) {
	page := input.Page
	if page <= 0 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	result, err := h.adminService.ListAdmins(ctx, service.ListAdminsInput{
		Search:   input.Search,
		Status:   input.Status,
		RoleID:   input.RoleID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	admins := make([]dto.AdminAccountDTO, 0, len(result.Admins))
	for _, a := range result.Admins {
		roles := make([]dto.RoleDTO, 0, len(a.Roles))
		for _, r := range a.Roles {
			roles = append(roles, dto.RoleDTO{ID: r.ID, Code: r.Code, Name: r.Name})
		}
		admins = append(admins, dto.AdminAccountDTO{
			ID:          a.ID,
			Email:       a.Email,
			Username:    a.Username,
			Status:      a.Status,
			FirstName:   a.FirstName,
			LastName:    a.LastName,
			Roles:       roles,
			CreatedAt:   a.CreatedAt,
			LastLoginAt: a.LastLogin,
		})
	}
	totalPages := result.TotalPages
	if totalPages == 0 && result.Total > 0 {
		totalPages = 1
	}
	return &dto.ListAdminsOutput{
		Body: dto.AdminListResponseBody{
			Admins:     admins,
			Total:      result.Total,
			Page:       result.Page,
			PageSize:   result.PageSize,
			TotalPages: totalPages,
		},
	}, nil
}
