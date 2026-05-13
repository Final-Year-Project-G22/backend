package push

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
)

type NoopProvider struct {
	logger core.Logger
}

func NewNoopProvider(logger core.Logger) *NoopProvider {
	return &NoopProvider{logger: logger}
}

func (p *NoopProvider) Send(ctx context.Context, deviceToken, title, body string, data map[string]string) error {
	p.logger.Info("Push notification (noop)",
		core.String("deviceToken", deviceToken),
		core.String("title", title),
	)
	return nil
}
