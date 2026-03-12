package event

import (
	"time"

	"github.com/google/uuid"
)

type RoleAssigned struct {
	RoleAssignmentID uuid.UUID
	AccountID        uuid.UUID
	RoleID           uuid.UUID
	AssignedBy       uuid.UUID
	OccurredAt       time.Time
}
