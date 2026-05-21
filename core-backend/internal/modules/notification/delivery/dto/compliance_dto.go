package dto

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	"github.com/google/uuid"
)

// --- List Compliance Entries ---

type ListComplianceEntriesInput struct {
	BusinessProfileID uuid.UUID `query:"businessProfileId"`
}

type ListComplianceEntriesOutput struct {
	Body ListComplianceEntriesResponseBody
}

type ListComplianceEntriesResponseBody struct {
	Data []ComplianceEntryResponse `json:"data"`
}

type ComplianceEntryResponse struct {
	ID                 uuid.UUID                    `json:"id"`
	BusinessProfileID  uuid.UUID                    `json:"businessProfileId"`
	AccountID          uuid.UUID                    `json:"accountId"`
	ComplianceType     entity.ComplianceType        `json:"complianceType"`
	ReferenceNumber    *string                      `json:"referenceNumber,omitempty"`
	IssuedDate         *time.Time                   `json:"issuedDate,omitempty"`
	ExpiryDate         time.Time                    `json:"expiryDate"`
	ReminderDaysBefore int                          `json:"reminderDaysBefore"`
	Status             entity.ComplianceEntryStatus `json:"status"`
	Source             entity.ComplianceSource      `json:"source"`
	LastNotifiedAt     *time.Time                   `json:"lastNotifiedAt,omitempty"`
}

// --- Create Compliance Entry ---

type CreateComplianceEntryRequest struct {
	BusinessProfileID  uuid.UUID             `json:"businessProfileId"`
	ComplianceType     entity.ComplianceType `json:"complianceType"`
	ReferenceNumber    *string               `json:"referenceNumber,omitempty"`
	IssuedDate         *time.Time            `json:"issuedDate,omitempty"`
	ExpiryDate         time.Time             `json:"expiryDate"`
	ReminderDaysBefore int                   `json:"reminderDaysBefore"`
}

type CreateComplianceEntryInput struct {
	Body CreateComplianceEntryRequest
}

type CreateComplianceEntryOutput struct {
	Body CreateComplianceEntryResponseBody
}

type CreateComplianceEntryResponseBody struct {
	ID      uuid.UUID `json:"id"`
	Message string    `json:"message"`
}

// --- Update Compliance Entry ---

type UpdateComplianceEntryRequest struct {
	ReferenceNumber    *string                       `json:"referenceNumber,omitempty"`
	IssuedDate         *time.Time                    `json:"issuedDate,omitempty"`
	ExpiryDate         *time.Time                    `json:"expiryDate,omitempty"`
	ReminderDaysBefore *int                          `json:"reminderDaysBefore,omitempty"`
	Status             *entity.ComplianceEntryStatus `json:"status,omitempty"`
}

type UpdateComplianceEntryInput struct {
	ID   uuid.UUID `path:"id"`
	Body UpdateComplianceEntryRequest
}

type UpdateComplianceEntryOutput struct {
	Body UpdateComplianceEntryResponseBody
}

type UpdateComplianceEntryResponseBody struct {
	Message string `json:"message"`
}

// --- Delete Compliance Entry ---

type DeleteComplianceEntryInput struct {
	ID uuid.UUID `path:"id"`
}

type DeleteComplianceEntryOutput struct {
	Body DeleteComplianceEntryResponseBody
}

type DeleteComplianceEntryResponseBody struct {
	Message string `json:"message"`
}

// --- Compliance Calendar ---

type GetCalendarOutput struct {
	Body GetCalendarResponseBody
}

type GetCalendarResponseBody struct {
	Entries []CalendarEntryResponse `json:"entries"`
}

type CalendarEntryResponse struct {
	ID              uuid.UUID `json:"id"`
	Type            string    `json:"type"`
	Title           string    `json:"title"`
	ReferenceNumber *string   `json:"referenceNumber,omitempty"`
	Date            time.Time `json:"date"`
	DaysRemaining   int       `json:"daysRemaining"`
	Status          string    `json:"status"`
}

// --- Mappers ---

func ToComplianceEntryResponse(entry *entity.ComplianceEntry) ComplianceEntryResponse {
	return ComplianceEntryResponse{
		ID:                 entry.ID,
		BusinessProfileID:  entry.BusinessProfileID,
		AccountID:          entry.AccountID,
		ComplianceType:     entry.ComplianceType,
		ReferenceNumber:    entry.ReferenceNumber,
		IssuedDate:         entry.IssuedDate,
		ExpiryDate:         entry.ExpiryDate,
		ReminderDaysBefore: entry.ReminderDaysBefore,
		Status:             entry.Status,
		Source:             entry.Source,
		LastNotifiedAt:     entry.LastNotifiedAt,
	}
}

func ToComplianceEntryResponses(entries []*entity.ComplianceEntry) []ComplianceEntryResponse {
	if len(entries) == 0 {
		return nil
	}
	resp := make([]ComplianceEntryResponse, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, ToComplianceEntryResponse(e))
	}
	return resp
}

func ToCreateComplianceEntryInput(body CreateComplianceEntryRequest) usecase.CreateComplianceEntryInput {
	return usecase.CreateComplianceEntryInput{
		BusinessProfileID:  body.BusinessProfileID,
		ComplianceType:     body.ComplianceType,
		ReferenceNumber:    body.ReferenceNumber,
		IssuedDate:         body.IssuedDate,
		ExpiryDate:         body.ExpiryDate,
		ReminderDaysBefore: body.ReminderDaysBefore,
	}
}

func ToUpdateComplianceEntryInput(body UpdateComplianceEntryRequest) usecase.UpdateComplianceEntryInput {
	return usecase.UpdateComplianceEntryInput{
		ReferenceNumber:    body.ReferenceNumber,
		IssuedDate:         body.IssuedDate,
		ExpiryDate:         body.ExpiryDate,
		ReminderDaysBefore: body.ReminderDaysBefore,
		Status:             body.Status,
	}
}

func ToCalendarEntryResponse(entry usecase.CalendarEntry) CalendarEntryResponse {
	return CalendarEntryResponse{
		ID:              entry.ID,
		Type:            entry.Type,
		Title:           entry.Title,
		ReferenceNumber: entry.ReferenceNumber,
		Date:            entry.Date,
		DaysRemaining:   entry.DaysRemaining,
		Status:          entry.Status,
	}
}

func ToCalendarEntryResponses(entries []usecase.CalendarEntry) []CalendarEntryResponse {
	if len(entries) == 0 {
		return nil
	}
	resp := make([]CalendarEntryResponse, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, ToCalendarEntryResponse(e))
	}
	return resp
}

// --- Compliance Types ---

type ListComplianceTypesOutput struct {
	Body ListComplianceTypesResponseBody
}

type ListComplianceTypesResponseBody struct {
	Data []ComplianceTypeResponse `json:"data"`
}

type ComplianceTypeResponse struct {
	Slug  string `json:"slug"`
	Label string `json:"label"`
}
