package dto

import (
	"time"

	"github.com/google/uuid"
)

type AskRequest struct {
	Query     string  `json:"query" doc:"User question" minLength:"1" maxLength:"10000"`
	Language  string  `json:"language,omitempty" doc:"Language code (e.g. en, am)"`
	SessionID *string `json:"sessionId,omitempty" doc:"Optional conversation ID to continue"`
	TopK      *int32  `json:"topK,omitempty" doc:"Number of context chunks to retrieve" default:"5"`
}

type AskInput struct {
	Body AskRequest
}

type CitationDTO struct {
	DocumentID uuid.UUID `json:"documentId"`
	ChunkID    uuid.UUID `json:"chunkId"`
	SourceType string    `json:"sourceType"`
	Title      *string   `json:"title,omitempty"`
	Score      float64   `json:"score"`
}

type UsageDTO struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

type AskResponseBody struct {
	RequestID uuid.UUID     `json:"requestId"`
	SessionID uuid.UUID     `json:"sessionId"`
	Answer    string        `json:"answer"`
	Citations []CitationDTO `json:"citations"`
	Usage     UsageDTO      `json:"usage"`
	Model     string        `json:"model"`
	LatencyMS int           `json:"latencyMs"`
}

type AskOutput struct {
	Body AskResponseBody
}

type AskStreamInput struct {
	Body AskRequest
}

type AskStreamChunkEventBody struct {
	Text string `json:"text"`
}

type AskStreamCitationEventBody struct {
	Citations []CitationDTO `json:"citations"`
}

type AskStreamDoneEventBody struct {
	Model     string   `json:"model"`
	LatencyMS int      `json:"latencyMs"`
	Usage     UsageDTO `json:"usage"`
}

type AskStreamErrorEventBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ConversationDTO struct {
	ID        uuid.UUID `json:"id"`
	AccountID uuid.UUID `json:"accountId"`
	Title     string    `json:"title"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ListConversationsInput struct {
	Body  struct{}
	Query struct {
		Limit  *int32 `json:"limit,omitempty" query:"limit" default:"20" minimum:"1" maximum:"100"`
		Offset *int32 `json:"offset,omitempty" query:"offset" default:"0" minimum:"0"`
	}
}

type ListConversationsOutput struct {
	Body struct {
		Sessions []ConversationDTO `json:"sessions"`
		Total    int               `json:"total"`
	}
}

type MessageDTO struct {
	ID        uuid.UUID     `json:"id"`
	Role      string        `json:"role"`
	Content   string        `json:"content"`
	Citations []CitationDTO `json:"citations,omitempty"`
	Usage     *UsageDTO     `json:"usage,omitempty"`
	CreatedAt time.Time     `json:"createdAt"`
}

type GetConversationInput struct {
	Body  struct{}
	Query struct {
		MessageLimit   *int32 `json:"messageLimit,omitempty" query:"messageLimit" default:"50" minimum:"1" maximum:"100"`
		MessageOffset  *int32 `json:"messageOffset,omitempty" query:"messageOffset" default:"0" minimum:"0"`
		IncludeDeleted *bool  `json:"includeDeleted,omitempty" query:"includeDeleted" default:"false"`
	}
}

type GetConversationPathInput struct {
	Body struct{} `path:"sessionId"`
}

type GetConversationOutput struct {
	Body struct {
		Session   ConversationDTO `json:"session"`
		Messages  []MessageDTO    `json:"messages"`
		TotalMsgs int             `json:"totalMsgs"`
	}
}

type ArchiveConversationInput struct {
	Body struct{} `path:"sessionId"`
}

type ArchiveConversationOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}
