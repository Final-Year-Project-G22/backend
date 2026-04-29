package usecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	notiferror "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type notificationMuteUsecase struct {
	mutedRepo repository.MutedAccountRepository
}

func NewNotificationMuteUsecase(
	mutedRepo repository.MutedAccountRepository,
) usecase.NotificationMuteUsecase {
	return &notificationMuteUsecase{
		mutedRepo: mutedRepo,
	}
}

func (uc *notificationMuteUsecase) MuteAccount(ctx context.Context, accountID uuid.UUID, input usecase.MuteAccountInput) error {
	existing, err := uc.findMute(ctx, accountID, input.MutedAccountID)
	if err != nil {
		return err
	}
	if existing != nil {
		if existing.MuteUntil == nil || existing.MuteUntil.After(time.Now().UTC()) {
			return notiferror.ErrAlreadyMuted
		}
		existing.MuteUntil = input.MuteUntil
		existing.Reason = input.Reason
		return uc.mutedRepo.Update(ctx, existing)
	}

	mute := &entity.MutedAccount{
		AccountID:      accountID,
		MutedAccountID: input.MutedAccountID,
		MuteUntil:      input.MuteUntil,
		Reason:         input.Reason,
	}
	return uc.mutedRepo.Create(ctx, mute)
}

func (uc *notificationMuteUsecase) UnmuteAccount(ctx context.Context, accountID uuid.UUID, mutedAccountID uuid.UUID) error {
	existing, err := uc.findMute(ctx, accountID, mutedAccountID)
	if err != nil {
		return err
	}
	if existing == nil {
		return notiferror.ErrMuteNotFound
	}
	return uc.mutedRepo.DeleteByAccountPair(ctx, accountID, mutedAccountID)
}

func (uc *notificationMuteUsecase) IsMuted(ctx context.Context, accountID uuid.UUID, mutedAccountID uuid.UUID) (bool, error) {
	existing, err := uc.findMute(ctx, accountID, mutedAccountID)
	if err != nil {
		return false, err
	}
	if existing == nil {
		return false, nil
	}
	if existing.MuteUntil != nil && existing.MuteUntil.Before(time.Now().UTC()) {
		return false, nil
	}
	return true, nil
}

func (uc *notificationMuteUsecase) ListMutedAccounts(ctx context.Context, accountID uuid.UUID, q query.QueryOptions) ([]*entity.MutedAccount, error) {
	return uc.mutedRepo.ListByAccount(ctx, accountID, q)
}

func (uc *notificationMuteUsecase) findMute(ctx context.Context, accountID uuid.UUID, mutedAccountID uuid.UUID) (*entity.MutedAccount, error) {
	opts := query.DefaultQueryOptions()
	opts.Filters = map[string]interface{}{
		"account_id":       accountID,
		"muted_account_id": mutedAccountID,
	}
	mutes, err := uc.mutedRepo.Find(ctx, opts)
	if err != nil {
		return nil, err
	}
	if len(mutes) == 0 {
		return nil, nil
	}
	return mutes[0], nil
}
