package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/dto"
	iamentity "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
)

type AccountPreferenceHandler struct {
	repo iamrepo.AccountPreferenceRepository
}

func NewAccountPreferenceHandler(repo iamrepo.AccountPreferenceRepository) *AccountPreferenceHandler {
	return &AccountPreferenceHandler{repo: repo}
}

func (h *AccountPreferenceHandler) HandleGetAccountPreferences(ctx context.Context, input *struct{}) (*dto.GetAccountPreferenceOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, apperrors.UnauthorizedError("iam.errors.unauthorized")
	}

	pref, err := h.repo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	if pref == nil {
		return &dto.GetAccountPreferenceOutput{
			Body: dto.AccountPreferenceResponse{
				Language: "en",
				Timezone: "UTC",
			},
		}, nil
	}
	return &dto.GetAccountPreferenceOutput{
		Body: dto.AccountPreferenceResponse{
			Language: pref.Language,
			Timezone: pref.Timezone,
		},
	}, nil
}

func (h *AccountPreferenceHandler) HandleUpdateAccountPreferences(ctx context.Context, input *dto.UpdateAccountPreferenceInput) (*dto.UpdateAccountPreferenceOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if accountID == contextkeys.NilUUID {
		return nil, apperrors.UnauthorizedError("iam.errors.unauthorized")
	}

	pref, err := h.repo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	if pref == nil {
		pref = &iamentity.AccountPreference{
			AccountID: accountID,
			Language:  "en",
			Timezone:  "UTC",
		}
	}

	if input.Body.Language != nil {
		pref.Language = *input.Body.Language
	}

	if pref.ID == uuid.Nil {
		if err := h.repo.Create(ctx, pref); err != nil {
			return nil, apperrors.ToHumaError(ctx, err)
		}
	} else {
		if err := h.repo.Update(ctx, pref); err != nil {
			return nil, apperrors.ToHumaError(ctx, err)
		}
	}

	return &dto.UpdateAccountPreferenceOutput{
		Body: dto.AccountPreferenceResponse{
			Language: pref.Language,
			Timezone: pref.Timezone,
		},
	}, nil
}
