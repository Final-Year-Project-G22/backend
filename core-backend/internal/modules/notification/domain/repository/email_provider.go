package repository

import "context"

type EmailProvider interface {
	Send(ctx context.Context, to, subject, body string, metadata map[string]string) (providerMessageID string, err error)
	VerifyWebhookSignature(ctx context.Context, payload []byte, svixTimestamp, svixSignature string) (bool, error)
}
