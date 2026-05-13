package port

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	"github.com/google/uuid"
)

type AskRequest struct {
	RequestID uuid.UUID
	UserID    uuid.UUID
	AccountID uuid.UUID
	Query     string
	Language  constants.Locale
	SessionID *uuid.UUID
	Title     *string
	TopK      int32
	// Taxonomy filters for retrieval narrowing
	SectorIDs []uuid.UUID
	TagIDs    []uuid.UUID
	Region    *string
	Stage     *string
}

type Citation struct {
	DocumentID uuid.UUID
	ChunkID    uuid.UUID
	SourceType string
	Title      *string
	Score      float64
	Excerpt    *string
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type AskResponse struct {
	RequestID uuid.UUID
	SessionID uuid.UUID
	CreatedAt string
	UpdatedAt string
	Answer    string
	Citations []Citation
	Usage     Usage
	Model     string
	LatencyMS int
}

type Conversation struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	Title     string
	Language  constants.Locale
	CreatedAt string
	UpdatedAt string
}

type Message struct {
	ID         uuid.UUID
	Role       string
	Content    string
	Citations  []Citation
	TokenUsage Usage
	CreatedAt  string
}

type ListConversationsRequest struct {
	UserID    uuid.UUID
	AccountID uuid.UUID
	Limit     int32
	Offset    int32
}

type ListConversationsResponse struct {
	Sessions []Conversation
	Total    int32
}

type GetConversationRequest struct {
	SessionID      uuid.UUID
	AccountID      uuid.UUID
	MessageLimit   int32
	MessageOffset  int32
	IncludeDeleted bool
}

type GetConversationResponse struct {
	Session   Conversation
	Messages  []Message
	TotalMsgs int32
}

type ArchiveConversationRequest struct {
	SessionID uuid.UUID
	AccountID uuid.UUID
}

type ArchiveConversationResponse struct {
	Success   bool
	UpdatedAt string
}

type AskStreamChunk struct {
	Text      *string
	Citations []Citation
	Done      *DoneInfo
	Error     *ErrorInfo
}

type DoneInfo struct {
	Model     string
	Usage     Usage
	LatencyMs int
	SessionID uuid.UUID
	CreatedAt string
	UpdatedAt string
}

type ErrorInfo struct {
	Code    string
	Message string
}

type AIInferencePort interface {
	Ask(ctx context.Context, req AskRequest) (AskResponse, error)
	AskStream(ctx context.Context, req AskRequest) (<-chan AskStreamChunk, error)
	ListConversations(ctx context.Context, req ListConversationsRequest) (ListConversationsResponse, error)
	GetConversation(ctx context.Context, req GetConversationRequest) (GetConversationResponse, error)
	ArchiveConversation(ctx context.Context, req ArchiveConversationRequest) (ArchiveConversationResponse, error)
}
