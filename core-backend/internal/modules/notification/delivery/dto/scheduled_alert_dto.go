package dto

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/google/uuid"
)

// --- List Scheduled Alerts ---

type ListScheduledAlertsOutput struct {
	Body ListScheduledAlertsResponseBody
}

type ListScheduledAlertsResponseBody struct {
	Data []ScheduledAlertResponse `json:"data"`
}

type ScheduledAlertResponse struct {
	ID              uuid.UUID             `json:"id"`
	TemplateSlug    *string               `json:"templateSlug,omitempty"`
	Title           string                `json:"title"`
	Body            string                `json:"body"`
	Channels        []entity.Channel      `json:"channels"`
	ScheduledFor    time.Time             `json:"scheduledFor"`
	Status          entity.ScheduleStatus `json:"status"`
	RescheduledFrom *time.Time            `json:"rescheduledFrom,omitempty"`
	SentAt          *time.Time            `json:"sentAt,omitempty"`
	CancelledAt     *time.Time            `json:"cancelledAt,omitempty"`
	CreatedAt       *time.Time            `json:"createdAt"`
}

// --- Create Scheduled Alert ---

type CreateScheduledAlertRequest struct {
	TemplateSlug *string          `json:"templateSlug,omitempty"`
	Title        string           `json:"title"`
	Body         string           `json:"body"`
	Channels     []entity.Channel `json:"channels"`
	ScheduledFor time.Time        `json:"scheduledFor"`
}

type CreateScheduledAlertInput struct {
	Body CreateScheduledAlertRequest
}

type CreateScheduledAlertOutput struct {
	Body CreateScheduledAlertResponseBody
}

type CreateScheduledAlertResponseBody struct {
	ID      uuid.UUID `json:"id"`
	Message string    `json:"message"`
}

// --- Cancel Scheduled Alert ---

type CancelScheduledAlertInput struct {
	ID uuid.UUID `path:"id"`
}

type CancelScheduledAlertOutput struct {
	Body CancelScheduledAlertResponseBody
}

type CancelScheduledAlertResponseBody struct {
	Message string `json:"message"`
}

// --- Reschedule Scheduled Alert ---

type RescheduleScheduledAlertRequest struct {
	ScheduledFor time.Time `json:"scheduledFor"`
}

type RescheduleScheduledAlertInput struct {
	ID   uuid.UUID `path:"id"`
	Body RescheduleScheduledAlertRequest
}

type RescheduleScheduledAlertOutput struct {
	Body RescheduleScheduledAlertResponseBody
}

type RescheduleScheduledAlertResponseBody struct {
	Message              string    `json:"message"`
	PreviousScheduledFor time.Time `json:"previousScheduledFor"`
}

// --- List Scheduled Alert Templates ---

type ListScheduledTemplatesOutput struct {
	Body ListScheduledTemplatesResponseBody
}

type ListScheduledTemplatesResponseBody struct {
	Data []ScheduledTemplateResponse `json:"data"`
}

type ScheduledTemplateResponse struct {
	Slug           string          `json:"slug"`
	Name           string          `json:"name"`
	DefaultTitle   string          `json:"defaultTitle"`
	DefaultBody    string          `json:"defaultBody"`
	DefaultChannel *entity.Channel `json:"defaultChannel,omitempty"`
}

// --- Mappers ---

func ToScheduledAlertResponse(notif *entity.UserScheduledNotification) ScheduledAlertResponse {
	channels := make([]entity.Channel, len(notif.Channels))
	for i, ch := range notif.Channels {
		channels[i] = entity.Channel(ch)
	}
	return ScheduledAlertResponse{
		ID:              notif.ID,
		TemplateSlug:    notif.TemplateSlug,
		Title:           notif.Title,
		Body:            notif.Body,
		Channels:        channels,
		ScheduledFor:    notif.ScheduledFor,
		Status:          notif.Status,
		RescheduledFrom: notif.RescheduledFrom,
		SentAt:          notif.SentAt,
		CancelledAt:     notif.CancelledAt,
		CreatedAt:       notif.CreatedAt,
	}
}

func ToScheduledAlertResponses(notifs []*entity.UserScheduledNotification) []ScheduledAlertResponse {
	if len(notifs) == 0 {
		return nil
	}
	resp := make([]ScheduledAlertResponse, 0, len(notifs))
	for _, n := range notifs {
		resp = append(resp, ToScheduledAlertResponse(n))
	}
	return resp
}

func ToCreateScheduledAlertInput(body CreateScheduledAlertRequest) usecase.ScheduleUserNotificationInput {
	return usecase.ScheduleUserNotificationInput{
		TemplateSlug: body.TemplateSlug,
		Title:        body.Title,
		Body:         body.Body,
		Channels:     body.Channels,
		ScheduledFor: body.ScheduledFor,
	}
}

func ToScheduledTemplateResponse(tmpl *entity.ScheduledAlertTemplate) ScheduledTemplateResponse {
	return ScheduledTemplateResponse{
		Slug:           tmpl.Slug,
		Name:           tmpl.Name,
		DefaultTitle:   tmpl.DefaultTitle,
		DefaultBody:    tmpl.DefaultBody,
		DefaultChannel: tmpl.DefaultChannel,
	}
}

func ToScheduledTemplateResponses(templates []*entity.ScheduledAlertTemplate) []ScheduledTemplateResponse {
	if len(templates) == 0 {
		return nil
	}
	resp := make([]ScheduledTemplateResponse, 0, len(templates))
	for _, t := range templates {
		resp = append(resp, ToScheduledTemplateResponse(t))
	}
	return resp
}
