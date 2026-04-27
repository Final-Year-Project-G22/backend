package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type NotificationHandler struct {
	inboxUC   usecase.NotificationInboxUsecase
	historyUC usecase.NotificationHistoryUsecase
}

func NewNotificationHandler(
	inboxUC usecase.NotificationInboxUsecase,
	historyUC usecase.NotificationHistoryUsecase,
) *NotificationHandler {
	return &NotificationHandler{
		inboxUC:   inboxUC,
		historyUC: historyUC,
	}
}

// --- Inbox ---

func (h *NotificationHandler) HandleListInbox(ctx context.Context, input *dto.ListInboxInput) (*dto.ListInboxOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	q := dto.ToQueryOptions(input.Page, input.PageSize)
	inboxes, err := h.inboxUC.ListInbox(ctx, accountID, input.Category, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	data := dto.ToInboxEntryResponses(inboxes)
	return &dto.ListInboxOutput{Body: dto.ListInboxResponseBody{
		Data:       data,
		Total:      int64(len(data)),
		Page:       q.Page,
		PageSize:   q.PageSize,
		TotalPages: calcTotalPages(len(data), q.PageSize),
	}}, nil
}

func (h *NotificationHandler) HandleGetUnreadCount(ctx context.Context, input *struct{}) (*dto.UnreadCountOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	count, err := h.inboxUC.GetUnreadCount(ctx, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UnreadCountOutput{Body: dto.UnreadCountResponseBody{Count: count}}, nil
}

func (h *NotificationHandler) HandleMarkAsRead(ctx context.Context, input *dto.MarkAsReadInput) (*dto.MarkAsReadOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.inboxUC.MarkAsRead(ctx, accountID, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.MarkAsReadOutput{Body: dto.MarkAsReadResponseBody{Message: "Marked as read"}}, nil
}

func (h *NotificationHandler) HandleMarkAllAsRead(ctx context.Context, input *struct{}) (*dto.MarkAllAsReadOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.inboxUC.MarkAllAsRead(ctx, accountID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.MarkAllAsReadOutput{Body: dto.MarkAllAsReadResponseBody{Message: "All marked as read"}}, nil
}

func (h *NotificationHandler) HandleMarkCategoryAsRead(ctx context.Context, input *dto.MarkCategoryAsReadInput) (*dto.MarkCategoryAsReadOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.inboxUC.MarkCategoryAsRead(ctx, accountID, input.Category); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.MarkCategoryAsReadOutput{Body: dto.MarkCategoryAsReadResponseBody{Message: "Category marked as read"}}, nil
}

func (h *NotificationHandler) HandleArchiveNotification(ctx context.Context, input *dto.ArchiveNotificationInput) (*dto.ArchiveNotificationOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.inboxUC.ArchiveNotification(ctx, accountID, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.ArchiveNotificationOutput{Body: dto.ArchiveNotificationResponseBody{Message: "Notification archived"}}, nil
}

func (h *NotificationHandler) HandleDeleteNotification(ctx context.Context, input *dto.DeleteNotificationInput) (*dto.DeleteNotificationOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.inboxUC.DeleteNotification(ctx, accountID, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeleteNotificationOutput{Body: dto.DeleteNotificationResponseBody{Message: "Notification deleted"}}, nil
}

// --- History ---

func (h *NotificationHandler) HandleListHistory(ctx context.Context, input *dto.ListHistoryInput) (*dto.ListHistoryOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	q := dto.ToQueryOptions(input.Page, input.PageSize)
	histories, err := h.historyUC.ListByAccount(ctx, accountID, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	data := dto.ToHistoryEntryResponses(histories)
	return &dto.ListHistoryOutput{Body: dto.ListHistoryResponseBody{
		Data:       data,
		Total:      int64(len(data)),
		Page:       q.Page,
		PageSize:   q.PageSize,
		TotalPages: calcTotalPages(len(data), q.PageSize),
	}}, nil
}

func (h *NotificationHandler) HandleGetHistoryDetail(ctx context.Context, input *dto.GetHistoryInput) (*dto.GetHistoryOutput, error) {
	history, err := h.historyUC.GetByID(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.GetHistoryOutput{Body: dto.ToHistoryEntryResponse(history)}, nil
}

// --- Helpers ---

func calcTotalPages(total, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	return (total + pageSize - 1) / pageSize
}
