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
	ParentThreadID     *uuid.UUID
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

type ReportThreadInput struct {
	ThreadID uuid.UUID
	Reason   string
}

type ReportPostInput struct {
	ThreadID uuid.UUID
	PostID   uuid.UUID
	Reason   string
}

type ReportUserInput struct {
	ThreadID          uuid.UUID
	ReportedAccountID uuid.UUID
	Reason            string
}

type UpdateReportStatusInput struct {
	Status    entity.ReportStatus
	AdminNote *string
}

type DeleteReportedContentInput struct {
	ReportID uuid.UUID
	AdminID  uuid.UUID
}

type BlockUserInput struct {
	ActorID   uuid.UUID
	ThreadID  uuid.UUID
	BlockedID uuid.UUID
	Reason    *string
	IsAdmin   bool
}
