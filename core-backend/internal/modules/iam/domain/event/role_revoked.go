package event

import (
	"time"

	"github.com/google/uuid"
)

type RoleRevoked struct {
	RoleAssignmentID uuid.UUID
	AccountID        uuid.UUID
	RoleID           uuid.UUID
	OccurredAt       time.Time
	Reason           *string
}
