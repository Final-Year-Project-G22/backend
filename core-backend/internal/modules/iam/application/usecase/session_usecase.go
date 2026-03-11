package appusecase

import (
	"context"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/google/uuid"
)

type sessionUsecase struct {
	sessionRepo repository.SessionRepository
	logger      core.Logger
}

func NewSessionUsecase(
	sessionRepo repository.SessionRepository,
	logger core.Logger,
) usecase.SessionUsecase {
	return &sessionUsecase{
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

func (u *sessionUsecase) CreateSession(ctx context.Context, accountID uuid.UUID, input usecase.CreateSessionInput) (*entity.Session, error) {
	session := &entity.Session{
		AccountID:        accountID,
		RefreshTokenHash: input.RefreshTokenHash,
		UserAgent:        input.UserAgent,
		IPAddress:        input.IPAddress,
		LastActiveAt:     time.Now(),
		ExpiresAt:        input.ExpiresAt,
	}

	if err := u.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	u.logger.Info("Session created",
		core.String("sessionID", session.ID.String()),
		core.String("accountID", accountID.String()),
	)
	return session, nil
}

func (u *sessionUsecase) GetSessionByRefreshTokenHash(ctx context.Context, hash string) (*entity.Session, error) {
	return u.sessionRepo.GetByRefreshTokenHash(ctx, hash)
}

func (u *sessionUsecase) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	if err := u.sessionRepo.RevokeByID(ctx, sessionID, time.Now()); err != nil {
		return err
	}

	u.logger.Info("Session revoked", core.String("sessionID", sessionID.String()))
	return nil
}

func (u *sessionUsecase) RevokeAllSessions(ctx context.Context, accountID uuid.UUID) error {
	if err := u.sessionRepo.RevokeAllByAccountID(ctx, accountID, time.Now()); err != nil {
		return err
	}

	u.logger.Info("All sessions revoked for account", core.String("accountID", accountID.String()))
	return nil
}

func (u *sessionUsecase) CleanupExpiredSessions(ctx context.Context) error {
	if err := u.sessionRepo.DeleteExpired(ctx, time.Now()); err != nil {
		return err
	}

	u.logger.Info("Expired sessions cleaned up")
	return nil
}
