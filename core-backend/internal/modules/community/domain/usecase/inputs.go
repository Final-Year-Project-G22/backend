package usecase

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/google/uuid"
)

type CreateCategoryInput struct {
	Name             string
	Slug             string
	Description      *string
	ParentCategoryID *uuid.UUID
	IsActive         bool
}

type UpdateCategoryInput struct {
	Name             *string
	Slug             *string
	Description      *string
	ParentCategoryID *uuid.UUID
	IsActive         *bool
}

type CreateThreadInput struct {
	CategoryID         uuid.UUID
	Title              string
	Slug               string
	Description        *string
	InitialPostContent string
	AttachmentURL      *string
	AttachmentType     *string
}

type UpdateThreadInput struct {
	Title       *string
	Description *string
	IsPinned    *bool
	Status      *entity.ThreadStatus
}

type CreatePostInput struct {
	Content        string
	AttachmentURL  *string
	AttachmentType *string
}

type UpdatePostInput struct {
	Content          *string
	AttachmentURL    *string
	AttachmentType   *string
	RemoveAttachment *bool
}

type ReportInput struct {
	TargetType entity.TargetType
	TargetID   uuid.UUID
	Reason     string
}

type BlockUserInput struct {
	ActorID   uuid.UUID
	ThreadID  uuid.UUID
	BlockedID uuid.UUID
	Reason    *string
	IsAdmin   bool
}
