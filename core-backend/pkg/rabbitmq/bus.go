package rabbitmq

import (
	"context"
	"errors"
)

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

type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string {
	return e.Err.Error()
}

func (e *PermanentError) Unwrap() error {
	return e.Err
}

func NewPermanentError(err error) error {
	return &PermanentError{Err: err}
}

func IsPermanent(err error) bool {
	var permanent *PermanentError
	return errors.As(err, &permanent)
}
