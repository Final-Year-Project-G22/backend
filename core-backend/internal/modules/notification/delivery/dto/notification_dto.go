package dto

import (
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/google/uuid"
)

// --- Inbox List ---

type ListInboxInput struct {
	Category *entity.NotificationCategory `query:"category" doc:"Filter by notification category"`
	Page     int                          `query:"page" doc:"Page number"`
	PageSize int                          `query:"pageSize" doc:"Items per page"`
}

type ListInboxOutput struct {
	Body ListInboxResponseBody
}

type ListInboxResponseBody struct {
	Data       []InboxEntryResponse `json:"data" doc:"Inbox entries"`
	Total      int64                `json:"total" doc:"Total count"`
	Page       int                  `json:"page" doc:"Current page"`
	PageSize   int                  `json:"pageSize" doc:"Items per page"`
	TotalPages int                  `json:"totalPages" doc:"Total pages"`
}

type InboxEntryResponse struct {
	ID           uuid.UUID                   `json:"id" doc:"Inbox entry ID"`
	Category     entity.NotificationCategory `json:"category" doc:"Notification category"`
	ActionUrl    *string                     `json:"actionUrl,omitempty" doc:"Action URL"`
	IsRead       bool                        `json:"isRead" doc:"Whether the notification has been read"`
	IsArchived   bool                        `json:"isArchived" doc:"Whether the notification has been archived"`
	ExpiresAt    *time.Time                  `json:"expiresAt,omitempty" doc:"Expiration time"`
	Notification NotificationSummaryResponse `json:"notification" doc:"Notification details"`
}

type NotificationSummaryResponse struct {
	Title            string                  `json:"title" doc:"Notification title"`
	Content          string                  `json:"content" doc:"Notification content"`
	NotificationType entity.NotificationType `json:"notificationType" doc:"Notification type"`
	Channel          entity.Channel          `json:"channel" doc:"Delivery channel"`
	SentAt           time.Time               `json:"sentAt" doc:"Time the notification was sent"`
	DeliveredAt      *time.Time              `json:"deliveredAt,omitempty" doc:"Time the notification was delivered"`
	ReadAt           *time.Time              `json:"readAt,omitempty" doc:"Time the notification was read"`
}

// --- Unread Count ---

type UnreadCountOutput struct {
	Body UnreadCountResponseBody
}

type UnreadCountResponseBody struct {
	Count int64 `json:"count" doc:"Number of unread notifications"`
}

// --- Mark As Read ---

type MarkAsReadInput struct {
	ID uuid.UUID `path:"id" doc:"Inbox entry ID"`
}

type MarkAsReadOutput struct {
	Body MarkAsReadResponseBody
}

type MarkAsReadResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Mark All As Read ---

type MarkAllAsReadOutput struct {
	Body MarkAllAsReadResponseBody
}

type MarkAllAsReadResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Mark Category As Read ---

type MarkCategoryAsReadInput struct {
	Category entity.NotificationCategory `path:"category" doc:"Notification category"`
}

type MarkCategoryAsReadOutput struct {
	Body MarkCategoryAsReadResponseBody
}

type MarkCategoryAsReadResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Archive ---

type ArchiveNotificationInput struct {
	ID uuid.UUID `path:"id" doc:"Inbox entry ID"`
}

type ArchiveNotificationOutput struct {
	Body ArchiveNotificationResponseBody
}

type ArchiveNotificationResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- Delete ---

type DeleteNotificationInput struct {
	ID uuid.UUID `path:"id" doc:"Inbox entry ID"`
}

type DeleteNotificationOutput struct {
	Body DeleteNotificationResponseBody
}

type DeleteNotificationResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

// --- History List ---

type ListHistoryInput struct {
	Page     int `query:"page" doc:"Page number"`
	PageSize int `query:"pageSize" doc:"Items per page"`
}

type ListHistoryOutput struct {
	Body ListHistoryResponseBody
}

type ListHistoryResponseBody struct {
	Data       []HistoryEntryResponse `json:"data" doc:"History entries"`
	Total      int64                  `json:"total" doc:"Total count"`
	Page       int                    `json:"page" doc:"Current page"`
	PageSize   int                    `json:"pageSize" doc:"Items per page"`
	TotalPages int                    `json:"totalPages" doc:"Total pages"`
}

type HistoryEntryResponse struct {
	ID               uuid.UUID               `json:"id" doc:"History entry ID"`
	NotificationType entity.NotificationType `json:"notificationType" doc:"Notification type"`
	Channel          entity.Channel          `json:"channel" doc:"Delivery channel"`
	Title            string                  `json:"title" doc:"Notification title"`
	Content          string                  `json:"content" doc:"Notification content"`
	ActionUrl        *string                 `json:"actionUrl,omitempty" doc:"Action URL"`
	SentAt           time.Time               `json:"sentAt" doc:"Time the notification was sent"`
	DeliveredAt      *time.Time              `json:"deliveredAt,omitempty" doc:"Time the notification was delivered"`
	ReadAt           *time.Time              `json:"readAt,omitempty" doc:"Time the notification was read"`
	ClickedAt        *time.Time              `json:"clickedAt,omitempty" doc:"Time the notification was clicked"`
	DeliveryStatus   entity.DeliveryStatus   `json:"deliveryStatus" doc:"Current delivery status"`
	FailureReason    *string                 `json:"failureReason,omitempty" doc:"Failure reason if delivery failed"`
}

// --- History Detail ---

type GetHistoryInput struct {
	ID uuid.UUID `path:"id" doc:"History entry ID"`
}

type GetHistoryOutput struct {
	Body HistoryEntryResponse
}

// --- Mappers ---

func ToInboxEntryResponse(inbox *entity.UserNotificationInbox) InboxEntryResponse {
	resp := InboxEntryResponse{
		ID:         inbox.ID,
		Category:   inbox.Category,
		ActionUrl:  inbox.ActionUrl,
		IsRead:     inbox.IsRead,
		IsArchived: inbox.IsArchived,
		ExpiresAt:  inbox.ExpiresAt,
	}
	if inbox.NotificationHistory.ID != uuid.Nil {
		resp.Notification = NotificationSummaryResponse{
			Title:            inbox.NotificationHistory.Title,
			Content:          inbox.NotificationHistory.Content,
			NotificationType: inbox.NotificationHistory.NotificationType,
			Channel:          inbox.NotificationHistory.Channel,
			SentAt:           inbox.NotificationHistory.SentAt,
			DeliveredAt:      inbox.NotificationHistory.DeliveredAt,
			ReadAt:           inbox.NotificationHistory.ReadAt,
		}
	}
	return resp
}

func ToInboxEntryResponses(inboxes []*entity.UserNotificationInbox) []InboxEntryResponse {
	if len(inboxes) == 0 {
		return nil
	}
	resp := make([]InboxEntryResponse, 0, len(inboxes))
	for _, inbox := range inboxes {
		resp = append(resp, ToInboxEntryResponse(inbox))
	}
	return resp
}

func ToHistoryEntryResponse(history *entity.NotificationHistory) HistoryEntryResponse {
	return HistoryEntryResponse{
		ID:               history.ID,
		NotificationType: history.NotificationType,
		Channel:          history.Channel,
		Title:            history.Title,
		Content:          history.Content,
		ActionUrl:        history.ActionUrl,
		SentAt:           history.SentAt,
		DeliveredAt:      history.DeliveredAt,
		ReadAt:           history.ReadAt,
		ClickedAt:        history.ClickedAt,
		DeliveryStatus:   history.DeliveryStatus,
		FailureReason:    history.FailureReason,
	}
}

func ToHistoryEntryResponses(histories []*entity.NotificationHistory) []HistoryEntryResponse {
	if len(histories) == 0 {
		return nil
	}
	resp := make([]HistoryEntryResponse, 0, len(histories))
	for _, h := range histories {
		resp = append(resp, ToHistoryEntryResponse(h))
	}
	return resp
}
