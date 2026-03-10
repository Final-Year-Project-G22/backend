package event

import (
	"time"

	"github.com/google/uuid"
)

type SessionRevoked struct {
	SessionID  uuid.UUID
	AccountID  uuid.UUID
	OccurredAt time.Time
}
