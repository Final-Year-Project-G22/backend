package event

import (
	"time"

	"github.com/google/uuid"
)

type BusinessProfileCreated struct {
	BusinessProfileID uuid.UUID
	AccountID         uuid.UUID
	OccurredAt        time.Time
}
