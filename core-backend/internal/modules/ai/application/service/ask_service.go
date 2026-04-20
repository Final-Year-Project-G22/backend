package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/ai/domain/port"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	"github.com/google/uuid"
)

type AskService struct {
	inferencePort port.AIInferencePort
}

func NewAskService(inferencePort port.AIInferencePort) *AskService {
	return &AskService{inferencePort: inferencePort}
}

type AskInput struct {
	UserID    uuid.UUID
	AccountID uuid.UUID
	Query     string
	Language  string
	SessionID *uuid.UUID
	TopK      int32
}

type AskOutput struct {
	RequestID uuid.UUID
	SessionID uuid.UUID
	Answer    string
	Citations []port.Citation
	Usage     port.Usage
	Model     string
	LatencyMS int
}

func (s *AskService) Ask(ctx context.Context, in AskInput) (AskOutput, error) {
	if in.TopK < 1 {
		in.TopK = 5
	}
	if in.TopK > 20 {
		in.TopK = 20
	}

	language := constants.LocaleEnglish
	if in.Language != "" {
		language = constants.Locale(in.Language)
	}

	req := port.AskRequest{
		RequestID: uuid.New(),
		UserID:    in.UserID,
		AccountID: in.AccountID,
		Query:     in.Query,
		Language:  language,
		SessionID: in.SessionID,
		TopK:      in.TopK,
	}

	start := time.Now()
	resp, err := s.inferencePort.Ask(ctx, req)
	latency := time.Since(start)

	if err != nil {
		return AskOutput{}, fmt.Errorf("ask failed: %w", err)
	}

	return AskOutput{
		RequestID: resp.RequestID,
		SessionID: resp.SessionID,
		Answer:    resp.Answer,
		Citations: resp.Citations,
		Usage:     resp.Usage,
		Model:     resp.Model,
		LatencyMS: int(latency.Milliseconds()),
	}, nil
}

type AskStreamInput struct {
	UserID    uuid.UUID
	AccountID uuid.UUID
	Query     string
	Language  string
	SessionID *uuid.UUID
	TopK      int32
}

func (s *AskService) AskStream(ctx context.Context, in AskStreamInput) (<-chan port.AskStreamChunk, error) {
	if in.TopK < 1 {
		in.TopK = 5
	}
	if in.TopK > 20 {
		in.TopK = 20
	}

	language := constants.LocaleEnglish
	if in.Language != "" {
		language = constants.Locale(in.Language)
	}

	req := port.AskRequest{
		RequestID: uuid.New(),
		UserID:    in.UserID,
		AccountID: in.AccountID,
		Query:     in.Query,
		Language:  language,
		SessionID: in.SessionID,
		TopK:      in.TopK,
	}

	return s.inferencePort.AskStream(ctx, req)
}

type ListConversationsInput struct {
	UserID    uuid.UUID
	AccountID uuid.UUID
	Limit     int32
	Offset    int32
}

type ListConversationsOutput struct {
	Sessions []port.Conversation
	Total    int32
}

func (s *AskService) ListConversations(ctx context.Context, in ListConversationsInput) (ListConversationsOutput, error) {
	if in.Limit < 1 || in.Limit > 100 {
		in.Limit = 20
	}
	if in.Offset < 0 {
		in.Offset = 0
	}

	req := port.ListConversationsRequest{
		UserID:    in.UserID,
		AccountID: in.AccountID,
		Limit:     in.Limit,
		Offset:    in.Offset,
	}

	resp, err := s.inferencePort.ListConversations(ctx, req)
	if err != nil {
		return ListConversationsOutput{}, fmt.Errorf("list conversations failed: %w", err)
	}

	return ListConversationsOutput{
		Sessions: resp.Sessions,
		Total:    resp.Total,
	}, nil
}

type GetConversationInput struct {
	SessionID      uuid.UUID
	AccountID      uuid.UUID
	MessageLimit   int32
	MessageOffset  int32
	IncludeDeleted bool
}

type GetConversationOutput struct {
	Session   port.Conversation
	Messages  []port.Message
	TotalMsgs int32
}

func (s *AskService) GetConversation(ctx context.Context, in GetConversationInput) (GetConversationOutput, error) {
	if in.MessageLimit < 1 || in.MessageLimit > 100 {
		in.MessageLimit = 50
	}
	if in.MessageOffset < 0 {
		in.MessageOffset = 0
	}

	req := port.GetConversationRequest{
		SessionID:      in.SessionID,
		AccountID:      in.AccountID,
		MessageLimit:   in.MessageLimit,
		MessageOffset:  in.MessageOffset,
		IncludeDeleted: in.IncludeDeleted,
	}

	resp, err := s.inferencePort.GetConversation(ctx, req)
	if err != nil {
		return GetConversationOutput{}, fmt.Errorf("get conversation failed: %w", err)
	}

	return GetConversationOutput{
		Session:   resp.Session,
		Messages:  resp.Messages,
		TotalMsgs: resp.TotalMsgs,
	}, nil
}

type ArchiveConversationInput struct {
	SessionID uuid.UUID
	AccountID uuid.UUID
}

type ArchiveConversationOutput struct {
	Success bool
}

func (s *AskService) ArchiveConversation(ctx context.Context, in ArchiveConversationInput) (ArchiveConversationOutput, error) {
	req := port.ArchiveConversationRequest{
		SessionID: in.SessionID,
		AccountID: in.AccountID,
	}

	_, err := s.inferencePort.ArchiveConversation(ctx, req)
	if err != nil {
		return ArchiveConversationOutput{}, fmt.Errorf("archive conversation failed: %w", err)
	}

	return ArchiveConversationOutput{Success: true}, nil
}
