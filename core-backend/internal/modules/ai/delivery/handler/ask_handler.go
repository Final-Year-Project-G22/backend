package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
)

type AskHandler struct {
	askService *service.AskService
}

func NewAskHandler(askService *service.AskService) *AskHandler {
	return &AskHandler{askService: askService}
}

func (h *AskHandler) HandleAsk(ctx context.Context, input *dto.AskInput) (*dto.AskOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	if accountID == contextkeys.NilUUID || userID == contextkeys.NilUUID {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("ask.errors.unauthorized"))
	}

	askIn, err := h.parseAskInput(ctx, input.Body, accountID, userID)
	if err != nil {
		return nil, err
	}

	out, err := h.askService.Ask(ctx, askIn)
	if err != nil {
		return nil, mapAskError(ctx, err)
	}

	citations := make([]dto.CitationDTO, 0, len(out.Citations))
	for _, c := range out.Citations {
		citations = append(citations, dto.CitationDTO{
			DocumentID: c.DocumentID,
			ChunkID:    c.ChunkID,
			SourceType: c.SourceType,
			Title:      c.Title,
			Score:      c.Score,
			Excerpt:    c.Excerpt,
		})
	}

	return &dto.AskOutput{
		Body: dto.AskResponseBody{
			RequestID: out.RequestID,
			SessionID: out.SessionID,
			CreatedAt: parseTimeFromRFC3339(out.CreatedAt),
			UpdatedAt: parseTimeFromRFC3339(out.UpdatedAt),
			Answer:    out.Answer,
			Citations: citations,
			Usage: dto.UsageDTO{
				PromptTokens:     out.Usage.PromptTokens,
				CompletionTokens: out.Usage.CompletionTokens,
				TotalTokens:      out.Usage.TotalTokens,
			},
			Model:     out.Model,
			LatencyMS: out.LatencyMS,
		},
	}, nil
}

func (h *AskHandler) HandleAskStream(ctx context.Context, input *dto.AskStreamInput) (<-chan service.AskStreamChunk, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	userID := contextkeys.GetUserID(ctx.Value(contextkeys.UserID))

	if accountID == contextkeys.NilUUID || userID == contextkeys.NilUUID {
		return nil, apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("ask.errors.unauthorized"))
	}

	askIn, err := h.parseAskInput(ctx, input.Body, accountID, userID)
	if err != nil {
		return nil, err
	}

	streamIn := service.AskStreamInput(askIn)

	ch, streamErr := h.askService.AskStream(ctx, streamIn)
	if streamErr != nil {
		return nil, mapAskError(ctx, streamErr)
	}

	return ch, nil
}

func (h *AskHandler) HandleListConversations(ctx context.Context, input *dto.ListConversationsInput, accountID, userID uuid.UUID) (*dto.ListConversationsOutput, error) {
	limit := int32(20)
	if input.Query.Limit != nil && *input.Query.Limit > 0 {
		limit = *input.Query.Limit
	}
	offset := int32(0)
	if input.Query.Offset != nil && *input.Query.Offset >= 0 {
		offset = *input.Query.Offset
	}

	out, err := h.askService.ListConversations(ctx, service.ListConversationsInput{
		UserID:    userID,
		AccountID: accountID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, mapConversationError(ctx, err)
	}

	sessions := make([]dto.ConversationDTO, 0, len(out.Sessions))
	for _, s := range out.Sessions {
		sessions = append(sessions, dto.ConversationDTO{
			ID:        s.ID,
			AccountID: s.AccountID,
			Title:     s.Title,
			Language:  string(s.Language),
			CreatedAt: parseTimeFromRFC3339(s.CreatedAt),
			UpdatedAt: parseTimeFromRFC3339(s.UpdatedAt),
		})
	}

	return &dto.ListConversationsOutput{
		Body: struct {
			Sessions []dto.ConversationDTO `json:"sessions"`
			Total    int                   `json:"total"`
		}{
			Sessions: sessions,
			Total:    int(out.Total),
		},
	}, nil
}

func (h *AskHandler) HandleGetConversation(ctx context.Context, input *dto.GetConversationInput, sessionID, accountID uuid.UUID) (*dto.GetConversationOutput, error) {
	msgLimit := int32(50)
	if input.Query.MessageLimit != nil && *input.Query.MessageLimit > 0 {
		msgLimit = *input.Query.MessageLimit
	}
	msgOffset := int32(0)
	if input.Query.MessageOffset != nil && *input.Query.MessageOffset >= 0 {
		msgOffset = *input.Query.MessageOffset
	}
	includeDeleted := false
	if input.Query.IncludeDeleted != nil {
		includeDeleted = *input.Query.IncludeDeleted
	}

	out, err := h.askService.GetConversation(ctx, service.GetConversationInput{
		SessionID:      sessionID,
		AccountID:      accountID,
		MessageLimit:   msgLimit,
		MessageOffset:  msgOffset,
		IncludeDeleted: includeDeleted,
	})
	if err != nil {
		return nil, mapConversationError(ctx, err)
	}

	messages := make([]dto.MessageDTO, 0, len(out.Messages))
	for _, m := range out.Messages {
		citations := make([]dto.CitationDTO, 0, len(m.Citations))
		for _, c := range m.Citations {
			citations = append(citations, dto.CitationDTO{
				DocumentID: c.DocumentID,
				ChunkID:    c.ChunkID,
				SourceType: c.SourceType,
				Title:      c.Title,
				Score:      c.Score,
				Excerpt:    c.Excerpt,
			})
		}
		var usage *dto.UsageDTO
		if m.TokenUsage.TotalTokens > 0 {
			usage = &dto.UsageDTO{
				PromptTokens:     m.TokenUsage.PromptTokens,
				CompletionTokens: m.TokenUsage.CompletionTokens,
				TotalTokens:      m.TokenUsage.TotalTokens,
			}
		}
		messages = append(messages, dto.MessageDTO{
			ID:        m.ID,
			Role:      m.Role,
			Content:   m.Content,
			Citations: citations,
			Usage:     usage,
			CreatedAt: parseTimeFromRFC3339(m.CreatedAt),
		})
	}

	return &dto.GetConversationOutput{
		Body: struct {
			Session   dto.ConversationDTO `json:"session"`
			Messages  []dto.MessageDTO    `json:"messages"`
			TotalMsgs int                 `json:"totalMsgs"`
		}{
			Session: dto.ConversationDTO{
				ID:        out.Session.ID,
				AccountID: out.Session.AccountID,
				Title:     out.Session.Title,
				Language:  string(out.Session.Language),
				CreatedAt: parseTimeFromRFC3339(out.Session.CreatedAt),
				UpdatedAt: parseTimeFromRFC3339(out.Session.UpdatedAt),
			},
			Messages:  messages,
			TotalMsgs: int(out.TotalMsgs),
		},
	}, nil
}

func (h *AskHandler) HandleArchiveConversation(ctx context.Context, input *dto.ArchiveConversationInput, sessionID, accountID uuid.UUID) (*dto.ArchiveConversationOutput, error) {
	out, err := h.askService.ArchiveConversation(ctx, service.ArchiveConversationInput{
		SessionID: sessionID,
		AccountID: accountID,
	})
	if err != nil {
		return nil, mapConversationError(ctx, err)
	}

	return &dto.ArchiveConversationOutput{
		Body: struct {
			Success   bool      `json:"success"`
			UpdatedAt time.Time `json:"updatedAt"`
		}{Success: out.Success, UpdatedAt: parseTimeFromRFC3339(out.UpdatedAt)},
	}, nil
}

func (h *AskHandler) parseAskInput(
	ctx context.Context,
	body dto.AskRequest,
	accountID uuid.UUID,
	userID uuid.UUID,
) (service.AskInput, error) {
	query := strings.TrimSpace(body.Query)
	if query == "" {
		return service.AskInput{}, apperrors.ToHumaError(ctx, apperrors.BadRequestError("ask.errors.queryRequired"))
	}

	var sessionID *uuid.UUID
	if body.SessionID != nil && strings.TrimSpace(*body.SessionID) != "" {
		sid, err := uuid.Parse(strings.TrimSpace(*body.SessionID))
		if err != nil {
			return service.AskInput{}, apperrors.ToHumaError(ctx, apperrors.BadRequestError("ask.errors.invalidSessionId"))
		}
		sessionID = &sid
	}

	var title *string
	if body.Title != nil && strings.TrimSpace(*body.Title) != "" {
		t := strings.TrimSpace(*body.Title)
		title = &t
	}

	topK := int32(5)
	if body.TopK != nil {
		if *body.TopK < 1 || *body.TopK > 20 {
			return service.AskInput{}, apperrors.ToHumaError(ctx, apperrors.BadRequestError("ask.errors.invalidTopK"))
		}
		topK = *body.TopK
	}

	return service.AskInput{
		UserID:    userID,
		AccountID: accountID,
		Query:     query,
		Language:  strings.TrimSpace(body.Language),
		SessionID: sessionID,
		Title:     title,
		TopK:      topK,
	}, nil
}

func mapAskError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	switch {
	case strings.Contains(msg, "invalid argument"):
		return apperrors.ToHumaError(ctx, apperrors.BadRequestError("ask.errors.invalidInput"))
	case strings.Contains(msg, "quota exceeded"):
		return apperrors.ToHumaError(ctx, apperrors.NewError(apperrors.ErrCodeBadRequest, "ask.errors.quotaExceeded", 429))
	case strings.Contains(msg, "authentication failed"), strings.Contains(msg, "unauthenticated"):
		return apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("ask.errors.unauthorized"))
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline exceeded"):
		return apperrors.ToHumaError(ctx, apperrors.NewError(apperrors.ErrCodeInternal, "ask.errors.timeout", 504))
	default:
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return apperrors.ToHumaError(ctx, appErr)
		}
		return apperrors.ToHumaError(ctx, apperrors.InternalError("ask.errors.failed", fmt.Errorf("%w", err)))
	}
}

func mapConversationError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	switch {
	case strings.Contains(msg, "not found"):
		return apperrors.ToHumaError(ctx, apperrors.NotFoundErrorWithKey("conversation.errors.notFound"))
	case strings.Contains(msg, "invalid argument"):
		return apperrors.ToHumaError(ctx, apperrors.BadRequestError("conversation.errors.invalidInput"))
	case strings.Contains(msg, "permission denied"), strings.Contains(msg, "forbidden"):
		return apperrors.ToHumaError(ctx, apperrors.ForbiddenError("conversation.errors.permissionDenied"))
	case strings.Contains(msg, "unauthenticated"):
		return apperrors.ToHumaError(ctx, apperrors.UnauthorizedError("conversation.errors.unauthorized"))
	default:
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return apperrors.ToHumaError(ctx, appErr)
		}
		return apperrors.ToHumaError(ctx, apperrors.InternalError("conversation.errors.failed", fmt.Errorf("%w", err)))
	}
}

func parseTimeFromRFC3339(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return t
	}
	t, err = time.Parse(time.RFC3339, value)
	if err == nil {
		return t
	}
	return time.Time{}
}
