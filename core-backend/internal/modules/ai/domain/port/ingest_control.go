package port

import "context"

type IngestControl interface {
	IsEnabled(ctx context.Context) bool
	SetEnabled(ctx context.Context, enabled bool) error
	GetToggleState(ctx context.Context) (bool, bool, error)
}
