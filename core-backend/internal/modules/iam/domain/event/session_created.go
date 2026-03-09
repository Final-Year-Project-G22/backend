package event

import (
	"time"

	"github.com/google/uuid"
)

type SessionCreated struct {
	SessionID  uuid.UUID
	AccountID  uuid.UUID
	OccurredAt time.Time
	ExpiresAt  time.Time
}
