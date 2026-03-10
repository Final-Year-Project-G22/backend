package event

import (
	"time"

	"github.com/google/uuid"
)

type OAuthIdentityLinked struct {
	OAuthIdentityID uuid.UUID
	AccountID       uuid.UUID
	Provider        string
	OccurredAt      time.Time
}
