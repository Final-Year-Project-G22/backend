package handler

import (
	"context"
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

	var sessionID *uuid.UUID
	if input.Body.SessionID != nil && *input.Body.SessionID != "" {
		sid, err := uuid.Parse(*input.Body.SessionID)
		if err != nil {
			return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("ask.errors.invalidSessionId"))
		}
		sessionID = &sid
	}

	var topK int32 = 5
	if input.Body.TopK != nil && *input.Body.TopK > 0 {
		topK = *input.Body.TopK
	}

	out, err := h.askService.Ask(ctx, service.AskInput{
		UserID:    userID,
		AccountID: accountID,
		Query:     input.Body.Query,
		Language:  input.Body.Language,
		SessionID: sessionID,
		TopK:      topK,
	})
	if err != nil {
		return nil, mapAskError(err)
	}

	citations := make([]dto.CitationDTO, 0, len(out.Citations))
	for _, c := range out.Citations {
		citations = append(citations, dto.CitationDTO{
			DocumentID: c.DocumentID,
			ChunkID:    c.ChunkID,
			SourceType: c.SourceType,
			Title:      c.Title,
			Score:      c.Score,
		})
	}

	return &dto.AskOutput{
		Body: dto.AskResponseBody{
			RequestID: out.RequestID,
			SessionID: out.SessionID,
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
		return nil, mapConversationError(err)
	}

	sessions := make([]dto.ConversationDTO, 0, len(out.Sessions))
	for _, s := range out.Sessions {
		sessions = append(sessions, dto.ConversationDTO{
			ID:        s.ID,
			AccountID: s.AccountID,
			Title:     s.Title,
			Language:  string(s.Language),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
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
		return nil, mapConversationError(err)
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
			CreatedAt: time.Now(),
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
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
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
		return nil, mapConversationError(err)
	}

	return &dto.ArchiveConversationOutput{
		Body: struct {
			Success bool `json:"success"`
		}{Success: out.Success},
	}, nil
}

func mapAskError(err error) error {
	if err == nil {
		return nil
	}
	return apperrors.UnauthorizedError("ask.errors.unauthorized")
}

func mapConversationError(err error) error {
	if err == nil {
		return nil
	}
	return apperrors.NotFoundError("conversation", "session not found")
}
