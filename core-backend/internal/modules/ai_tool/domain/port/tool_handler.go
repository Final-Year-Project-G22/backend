package port

import (
	"context"

	"github.com/google/uuid"
)

type ToolHandler interface {
	Name() string
	Description() string
	ParameterSchema() string
	Execute(ctx context.Context, argsJSON string, accountID, userID uuid.UUID) (string, error)
}
