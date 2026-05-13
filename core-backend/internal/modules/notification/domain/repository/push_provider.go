package repository

import "context"

type PushProvider interface {
	Send(ctx context.Context, deviceToken, title, body string, data map[string]string) error
}
