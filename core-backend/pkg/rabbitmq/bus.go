package rabbitmq

import "context"

type Bus interface {
	Publish(ctx context.Context, event string, payload any) error
	Subscribe(event string, handler func(context.Context, []byte) error) error
	Close() error
}

func New(cfg Config) (Bus, error) {
	if !cfg.Enabled {
		return NoOp(), nil
	}
	return newClient(cfg)
}

func NoOp() Bus {
	return &noopBus{}
}

type noopBus struct{}

func (n *noopBus) Publish(ctx context.Context, event string, payload any) error {
	return nil
}

func (n *noopBus) Subscribe(event string, handler func(context.Context, []byte) error) error {
	return nil
}

func (n *noopBus) Close() error {
	return nil
}
