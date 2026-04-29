package repository

import (
	"context"

	"github.com/google/uuid"
)

type MuteResolver interface {
	IsMuted(ctx context.Context, accountID uuid.UUID, itemType string, itemID uuid.UUID) (bool, error)
}
