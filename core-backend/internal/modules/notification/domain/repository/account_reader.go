package repository

import (
	"context"

	"github.com/google/uuid"
)

type AccountInfo struct {
	Email  string
	Locale string
	Name   string
}

type AccountReader interface {
	FindAll(ctx context.Context) ([]uuid.UUID, error)
	FindBySegment(ctx context.Context, segment map[string]interface{}) ([]uuid.UUID, error)
	GetAccountInfo(ctx context.Context, accountID uuid.UUID) (*AccountInfo, error)
}
