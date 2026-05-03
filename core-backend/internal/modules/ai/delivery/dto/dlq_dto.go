package dto

import (
	"time"

	"github.com/google/uuid"
)

type ListDeadEventsRequest struct {
	Limit  *int `json:"limit,omitempty" query:"limit" default:"20" minimum:"1" maximum:"100"`
	Offset *int `json:"offset,omitempty" query:"offset" default:"0" minimum:"0"`
}

type ListDeadEventsInput struct {
	Query ListDeadEventsRequest
}

type DeadEventDTO struct {
	EventID      uuid.UUID `json:"eventId"`
	EventType    string    `json:"eventType"`
	Payload      string    `json:"payload"`
	Status       string    `json:"status"`
	ErrorMessage *string   `json:"errorMessage,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	ReplayCount  int32     `json:"replayCount"`
}

type ListDeadEventsOutput struct {
	Body struct {
		Events []DeadEventDTO `json:"events"`
		Total  int            `json:"total"`
	}
}

type GetDeadEventInput struct {
	EventID uuid.UUID `path:"eventId" doc:"Event ID"`
}

type GetDeadEventOutput struct {
	Body DeadEventDTO
}

type RedriveEventInput struct {
	EventID uuid.UUID `path:"eventId" doc:"Event ID"`
	Body    struct{}
}

type RedriveEventOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

type RedriveBatchInput struct {
	Body struct {
		EventIDs []uuid.UUID `json:"eventIds"`
	}
}

type RedriveBatchOutput struct {
	Body struct {
		SuccessCount int `json:"successCount"`
	}
}

type IngestToggleStateResponse struct {
	Enabled bool `json:"enabled"`
}

type GetIngestToggleInput struct{}

type GetIngestToggleOutput struct {
	Body IngestToggleStateResponse
}

type SetIngestToggleInput struct {
	Body struct {
		Enabled bool `json:"enabled"`
	}
}

type SetIngestToggleOutput struct {
	Body IngestToggleStateResponse
}
