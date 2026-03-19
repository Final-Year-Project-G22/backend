package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type UserHandler struct {
	authService service.AuthService
}

func NewUserHandler(authService service.AuthService) *UserHandler {
	return &UserHandler{
		authService: authService,
	}
}

func (h *UserHandler) HandleUserUpdate(ctx context.Context, input *dto.UpdateUserProfileInput) (*dto.UpdateUserProfileOutput, error) {
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))
	if userID == contextkeys.NilUUID {
		return nil, apperrors.UnauthorizedError("iam.errors.unauthorized")
	}

	result, err := h.authService.UpdateUserProfile(ctx, userID, service.UpdateUserProfileInput{
		FirstName: input.Body.FirstName,
		LastName:  input.Body.LastName,
		Bio:       input.Body.Bio,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.UpdateUserProfileOutput{
		Body: dto.UpdateUserProfileResponseBody{
			FirstName: result.FirstName,
			LastName:  result.LastName,
			Bio:       result.Bio,
		},
	}, nil
}
