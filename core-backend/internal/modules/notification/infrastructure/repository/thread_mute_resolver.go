package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	notifrepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type threadMuteResolver struct {
	db     *core.Database
	logger core.Logger
}

func NewThreadMuteResolver(db *core.Database, logger core.Logger) notifrepo.MuteResolver {
	return &threadMuteResolver{db: db, logger: logger}
}

func (r *threadMuteResolver) IsMuted(ctx context.Context, accountID uuid.UUID, itemType string, itemID uuid.UUID) (bool, error) {
	fmt.Printf("[MUTE_RESOLVER] IsMuted called: accountID=%s, itemType=%s, itemID=%s\n", accountID.String(), itemType, itemID.String())
	if itemType != "thread" {
		fmt.Printf("[MUTE_RESOLVER] not a thread item, skipping\n")
		return false, nil
	}

	var isMuted bool
	err := r.getDB(ctx).
		Table("user_thread_settings").
		Select("is_muted").
		Where("account_id = ? AND thread_id = ?", accountID, itemID).
		Limit(1).
		Pluck("is_muted", &isMuted).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Printf("[MUTE_RESOLVER] no row found, returning false\n")
		return false, nil
	}
	if err != nil {
		r.logger.Error("Failed to resolve thread mute status",
			core.String("accountID", accountID.String()),
			core.String("threadID", itemID.String()),
			core.Error(err),
		)
		return false, err
	}
	fmt.Printf("[MUTE_RESOLVER] isMuted=%v\n", isMuted)
	return isMuted, nil
}

func (r *threadMuteResolver) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}
