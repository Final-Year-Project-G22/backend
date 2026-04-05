package middleware

import (
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/danielgtaylor/huma/v2"
)

func PermissionMiddleware(api huma.API, roleAssignmentUsecase usecase.RoleAssignmentUsecase, permissionName string, allowedRoles []string) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		accountID := contextkeys.GetAccountID(ctx.Context().Value(contextkeys.AccountID))
		if accountID == contextkeys.NilUUID {
			_ = huma.WriteErr(api, ctx, 401, "missing account context")
			return
		}

		if len(allowedRoles) > 0 {
			roles, err := roleAssignmentUsecase.ListAccountRoles(ctx.Context(), accountID)
			if err != nil {
				_ = huma.WriteErr(api, ctx, 500, "failed to load roles")
				return
			}
			if hasRole(roles, allowedRoles) {
				next(ctx)
				return
			}
		}

		hasPermission, err := roleAssignmentUsecase.HasPermission(ctx.Context(), accountID, permissionName)
		if err != nil {
			if appErr, ok := err.(*apperrors.AppError); ok {
				_ = huma.WriteErr(api, ctx, appErr.GetStatus(), appErr.GetMessage("en"))
				return
			}
			_ = huma.WriteErr(api, ctx, 500, "failed to validate permissions")
			return
		}
		if !hasPermission {
			_ = huma.WriteErr(api, ctx, 403, "permission denied")
			return
		}

		next(ctx)
	}
}

func hasRole(roles []*entity.Role, allowedRoles []string) bool {
	for _, role := range roles {
		for _, allowed := range allowedRoles {
			if strings.EqualFold(role.Code, allowed) {
				return true
			}
		}
	}
	return false
}
