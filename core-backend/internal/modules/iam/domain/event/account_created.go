package event

import (
	"time"

	"github.com/google/uuid"
)

type AccountCreated struct {
	AccountID  uuid.UUID
	UserID     uuid.UUID
	OccurredAt time.Time
}
