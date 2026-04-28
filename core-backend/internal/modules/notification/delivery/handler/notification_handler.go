package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/notification/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type NotificationHandler struct {
	inboxUC   usecase.NotificationInboxUsecase
	historyUC usecase.NotificationHistoryUsecase
	prefUC    usecase.NotificationPreferenceUsecase
	muteUC    usecase.NotificationMuteUsecase
	deviceUC  usecase.NotificationDeviceUsecase
}

func NewNotificationHandler(
	inboxUC usecase.NotificationInboxUsecase,
	historyUC usecase.NotificationHistoryUsecase,
	prefUC usecase.NotificationPreferenceUsecase,
	muteUC usecase.NotificationMuteUsecase,
	deviceUC usecase.NotificationDeviceUsecase,
) *NotificationHandler {
	return &NotificationHandler{
		inboxUC:   inboxUC,
		historyUC: historyUC,
		prefUC:    prefUC,
		muteUC:    muteUC,
		deviceUC:  deviceUC,
	}
}

// --- Inbox ---

func (h *NotificationHandler) HandleListInbox(ctx context.Context, input *dto.ListInboxInput) (*dto.ListInboxOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	q := dto.ToQueryOptions(input.Page, input.PageSize)
	var category *entity.NotificationCategory
	if input.Category != "" {
		cat := entity.NotificationCategory(input.Category)
		category = &cat
	}
	inboxes, err := h.inboxUC.ListInbox(ctx, accountID, category, q)
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

// --- Preferences ---

func (h *NotificationHandler) HandleListPreferences(ctx context.Context, input *struct{}) (*dto.ListPreferencesOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	prefs, err := h.prefUC.GetPreferences(ctx, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.ListPreferencesOutput{Body: dto.ToPreferenceResponses(prefs)}, nil
}

func (h *NotificationHandler) HandleSetPreference(ctx context.Context, input *dto.SetPreferenceInput) (*dto.SetPreferenceOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.prefUC.SetPreference(ctx, accountID, dto.ToSetPreferenceInput(input.Body)); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.SetPreferenceOutput{Body: dto.SetPreferenceResponseBody{Message: "Preference set"}}, nil
}

func (h *NotificationHandler) HandleDeletePreference(ctx context.Context, input *dto.DeletePreferenceInput) (*dto.DeletePreferenceOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.prefUC.DeletePreference(ctx, accountID, input.NotificationType, input.Channel); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeletePreferenceOutput{Body: dto.DeletePreferenceResponseBody{Message: "Preference deleted"}}, nil
}

// --- Mutes ---

func (h *NotificationHandler) HandleListMutes(ctx context.Context, input *dto.ListMutesInput) (*dto.ListMutesOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	q := dto.ToQueryOptions(input.Page, input.PageSize)
	mutes, err := h.muteUC.ListMutedAccounts(ctx, accountID, q)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	data := dto.ToMuteEntryResponses(mutes)
	return &dto.ListMutesOutput{Body: dto.ListMutesResponseBody{
		Data:       data,
		Total:      int64(len(data)),
		Page:       q.Page,
		PageSize:   q.PageSize,
		TotalPages: calcTotalPages(len(data), q.PageSize),
	}}, nil
}

func (h *NotificationHandler) HandleMuteAccount(ctx context.Context, input *dto.MuteAccountInput) (*dto.MuteAccountOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.muteUC.MuteAccount(ctx, accountID, dto.ToMuteAccountInput(input.Body)); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.MuteAccountOutput{Body: dto.MuteAccountResponseBody{Message: "Account muted"}}, nil
}

func (h *NotificationHandler) HandleUnmuteAccount(ctx context.Context, input *dto.UnmuteAccountInput) (*dto.UnmuteAccountOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.muteUC.UnmuteAccount(ctx, accountID, input.MutedAccountID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UnmuteAccountOutput{Body: dto.UnmuteAccountResponseBody{Message: "Account unmuted"}}, nil
}

// --- Devices ---

func (h *NotificationHandler) HandleListDevices(ctx context.Context, input *struct{}) (*dto.ListDevicesOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	devices, err := h.deviceUC.ListDevices(ctx, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.ListDevicesOutput{Body: dto.ToDeviceResponses(devices)}, nil
}

func (h *NotificationHandler) HandleRegisterDevice(ctx context.Context, input *dto.RegisterDeviceInput) (*dto.RegisterDeviceOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	device, err := h.deviceUC.RegisterDevice(ctx, accountID, dto.ToRegisterDeviceInput(input.Body))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.RegisterDeviceOutput{Body: dto.RegisterDeviceResponseBody{ID: device.ID, Message: "Device registered"}}, nil
}

func (h *NotificationHandler) HandleUpdateDevice(ctx context.Context, input *dto.UpdateDeviceInput) (*dto.UpdateDeviceOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	device, err := h.deviceUC.UpdateDevice(ctx, accountID, input.ID, dto.ToUpdateDeviceInput(input.Body))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateDeviceOutput{Body: dto.ToDeviceResponse(device)}, nil
}

func (h *NotificationHandler) HandleDeactivateDevice(ctx context.Context, input *dto.DeactivateDeviceInput) (*dto.DeactivateDeviceOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.deviceUC.DeactivateDevice(ctx, accountID, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeactivateDeviceOutput{Body: dto.DeactivateDeviceResponseBody{Message: "Device deactivated"}}, nil
}

// --- Helpers ---

func calcTotalPages(total, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	return (total + pageSize - 1) / pageSize
}
