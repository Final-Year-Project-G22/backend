package dto

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type CreateCommunityCategoryRequest struct {
	Name             string     `json:"name" doc:"Category name"`
	Slug             string     `json:"slug" doc:"Category slug"`
	Description      *string    `json:"description,omitempty" doc:"Category description"`
	ParentCategoryID *uuid.UUID `json:"parentCategoryId,omitempty" doc:"Parent category ID"`
	IsActive         bool       `json:"isActive" doc:"Active flag"`
}

type CreateCommunityCategoryInput struct {
	Body CreateCommunityCategoryRequest
}

type CreateCommunityCategoryOutput struct {
	Body CreateCommunityCategoryResponseBody
}

type CreateCommunityCategoryResponseBody struct {
	ID uuid.UUID `json:"id" doc:"Created category ID"`
}

type UpdateCommunityCategoryRequest struct {
	Name             *string    `json:"name,omitempty" doc:"Category name"`
	Slug             *string    `json:"slug,omitempty" doc:"Category slug"`
	Description      *string    `json:"description,omitempty" doc:"Category description"`
	ParentCategoryID *uuid.UUID `json:"parentCategoryId,omitempty" doc:"Parent category ID"`
	IsActive         *bool      `json:"isActive,omitempty" doc:"Active flag"`
}

type UpdateCommunityCategoryInput struct {
	ID   uuid.UUID `path:"id" doc:"Category ID"`
	Body UpdateCommunityCategoryRequest
}

type UpdateCommunityCategoryOutput struct {
	Body UpdateCommunityCategoryResponseBody
}

type UpdateCommunityCategoryResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type DeleteCommunityCategoryInput struct {
	ID uuid.UUID `path:"id" doc:"Category ID"`
}

type DeleteCommunityCategoryOutput struct {
	Body DeleteCommunityCategoryResponseBody
}

type DeleteCommunityCategoryResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type AdminListCommunityCategoriesInput struct {
	Page            int  `query:"page" doc:"Page number"`
	PageSize        int  `query:"pageSize" doc:"Page size"`
	IncludeInactive bool `query:"includeInactive" doc:"Include inactive categories"`
}

type AdminListCommunityCategoriesOutput struct {
	Body ListCategoriesResponseBody
}

func ToCreateCategoryInput(input CreateCommunityCategoryRequest) usecase.CreateCategoryInput {
	return usecase.CreateCategoryInput{
		Name:             input.Name,
		Slug:             input.Slug,
		Description:      input.Description,
		ParentCategoryID: input.ParentCategoryID,
		IsActive:         input.IsActive,
	}
}

func ToUpdateCategoryInput(input UpdateCommunityCategoryRequest) usecase.UpdateCategoryInput {
	return usecase.UpdateCategoryInput{
		Name:             input.Name,
		Slug:             input.Slug,
		Description:      input.Description,
		ParentCategoryID: input.ParentCategoryID,
		IsActive:         input.IsActive,
	}
}

func ToAdminQueryOptions(page, pageSize int, search ...string) query.QueryOptions {
	opts := ToQueryOptions(page, pageSize)
	if len(search) > 0 {
		opts.Search = search[0]
	}
	return opts
}

type ListReportsInput struct {
	Status     string `query:"status" doc:"Filter by status: pending, under_review, resolved, dismissed"`
	TargetType string `query:"targetType" doc:"Filter by target type: thread, post, user"`
	TargetID   string `query:"targetId" doc:"Filter by target ID"`
	Page       int    `query:"page" doc:"Page number"`
	PageSize   int    `query:"pageSize" doc:"Page size"`
}

type ListReportsOutput struct {
	Body ListReportsResponseBody
}

type ListReportsResponseBody struct {
	Reports    []*ReportWithContentDTO `json:"reports" doc:"Report list with content"`
	Total      int64                   `json:"total" doc:"Total count"`
	Page       int                     `json:"page" doc:"Current page"`
	PageSize   int                     `json:"pageSize" doc:"Page size"`
	TotalPages int                     `json:"totalPages" doc:"Total pages"`
}

type GetReportInput struct {
	ID uuid.UUID `path:"id" doc:"Report ID"`
}

type GetReportOutput struct {
	Body GetReportResponseBody
}

type GetReportResponseBody struct {
	Report  *ReportDTO  `json:"report" doc:"Report details"`
	Content *ContentDTO `json:"content,omitempty" doc:"Reported content details"`
}

type ReportWithContentDTO struct {
	Report  *ReportDTO  `json:"report" doc:"Report details"`
	Content *ContentDTO `json:"content,omitempty" doc:"Reported content details"`
}

type ContentDTO struct {
	Thread *ThreadContentDTO `json:"thread,omitempty" doc:"Thread content"`
	Post   *PostContentDTO   `json:"post,omitempty" doc:"Post content"`
	User   *UserContentDTO   `json:"user,omitempty" doc:"User content"`
}

type ThreadContentDTO struct {
	ID              uuid.UUID `json:"id" doc:"Thread ID"`
	Title           string    `json:"title" doc:"Thread title"`
	Slug            string    `json:"slug" doc:"Thread slug"`
	Description     *string   `json:"description,omitempty" doc:"Thread description"`
	CategoryID      uuid.UUID `json:"categoryId" doc:"Category ID"`
	CategoryName    string    `json:"categoryName,omitempty" doc:"Category name"`
	AuthorAccountID uuid.UUID `json:"authorId" doc:"Author account ID"`
	AuthorFirstName string    `json:"authorFirstName" doc:"Author first name"`
	AuthorLastName  string    `json:"authorLastName" doc:"Author last name"`
	Status          string    `json:"status" doc:"Thread status"`
	ReplyCount      int       `json:"replyCount" doc:"Reply count"`
	ViewCount       int       `json:"viewCount" doc:"View count"`
	CreatedAt       string    `json:"createdAt" doc:"Created timestamp"`
	UpdatedAt       string    `json:"updatedAt,omitempty" doc:"Updated timestamp"`
}

type PostContentDTO struct {
	ID              uuid.UUID  `json:"id" doc:"Post ID"`
	ThreadID        uuid.UUID  `json:"threadId" doc:"Thread ID"`
	ThreadTitle     string     `json:"threadTitle,omitempty" doc:"Thread title"`
	ThreadSlug      string     `json:"threadSlug,omitempty" doc:"Thread slug"`
	ParentPostID    *uuid.UUID `json:"parentPostId,omitempty" doc:"Parent post ID"`
	AuthorAccountID uuid.UUID  `json:"authorId" doc:"Author account ID"`
	AuthorFirstName string     `json:"authorFirstName" doc:"Author first name"`
	AuthorLastName  string     `json:"authorLastName" doc:"Author last name"`
	Content         string     `json:"content" doc:"Post content"`
	IsSolution      bool       `json:"isSolution" doc:"Is solution"`
	CreatedAt       string     `json:"createdAt" doc:"Created timestamp"`
	UpdatedAt       string     `json:"updatedAt,omitempty" doc:"Updated timestamp"`
}

type UserContentDTO struct {
	ID          uuid.UUID  `json:"id" doc:"User ID"`
	Email       string     `json:"email" doc:"User email"`
	FirstName   string     `json:"firstName" doc:"First name"`
	LastName    string     `json:"lastName" doc:"Last name"`
	ThreadID    *uuid.UUID `json:"threadId,omitempty" doc:"Thread ID"`
	ThreadTitle string     `json:"threadTitle,omitempty" doc:"Thread title"`
}

type ReportDTO struct {
	ID                uuid.UUID  `json:"id" doc:"Report ID"`
	ReporterAccountID uuid.UUID  `json:"reporterId" doc:"Reporter account ID"`
	ReporterFirstName string     `json:"reporterFirstName" doc:"Reporter first name"`
	ReporterLastName  string     `json:"reporterLastName" doc:"Reporter last name"`
	ThreadID          *uuid.UUID `json:"threadId,omitempty" doc:"Thread ID"`
	PostID            *uuid.UUID `json:"postId,omitempty" doc:"Post ID"`
	ReportedAccountID *uuid.UUID `json:"reportedAccountId,omitempty" doc:"Reported account ID"`
	Reason            string     `json:"reason" doc:"Report reason"`
	Status            string     `json:"status" doc:"Report status"`
	AdminNote         *string    `json:"adminNote,omitempty" doc:"Admin note"`
	ResolvedBy        *uuid.UUID `json:"resolvedBy,omitempty" doc:"Resolved by account ID"`
	ResolvedAt        *string    `json:"resolvedAt,omitempty" doc:"Resolved timestamp"`
	CreatedAt         string     `json:"createdAt" doc:"Created timestamp"`
}

type UpdateReportStatusInput struct {
	ID   uuid.UUID `path:"id" doc:"Report ID"`
	Body UpdateReportStatusRequest
}

type UpdateReportStatusRequest struct {
	Status    string  `json:"status" doc:"Status: pending, under_review, resolved, dismissed" validate:"required,oneof=pending under_review resolved dismissed"`
	AdminNote *string `json:"adminNote,omitempty" doc:"Admin note"`
}

type UpdateReportStatusOutput struct {
	Body UpdateReportStatusResponseBody
}

type UpdateReportStatusResponseBody struct {
	Report *ReportDTO `json:"report" doc:"Updated report"`
}

type DeleteReportedThreadInput struct {
	ID uuid.UUID `path:"id" doc:"Report ID"`
}

type DeleteReportedThreadOutput struct {
	Body DeleteReportedContentResponseBody
}

type DeleteReportedPostInput struct {
	ID uuid.UUID `path:"id" doc:"Report ID"`
}

type DeleteReportedPostOutput struct {
	Body DeleteReportedContentResponseBody
}

type DeleteReportedContentResponseBody struct {
	Success bool   `json:"success" doc:"Success"`
	Message string `json:"message" doc:"Message"`
}

type BlockReportedUserInput struct {
	ID   uuid.UUID `path:"id" doc:"Report ID"`
	Body BlockReportedUserRequest
}

type BlockReportedUserRequest struct {
	Reason string `json:"reason,omitempty" doc:"Reason for blocking user"`
}

type BlockReportedUserOutput struct {
	Body BlockReportedUserResponseBody
}

type BlockReportedUserResponseBody struct {
	Success bool   `json:"success" doc:"Success"`
	Message string `json:"message" doc:"Message"`
}

func ToReportDTO(report *entity.ContentReport, reporterFirstName, reporterLastName string) *ReportDTO {
	var resolvedAt *string
	if report.ResolvedAt != nil {
		t := report.ResolvedAt.Format("2006-01-02T15:04:05Z07:00")
		resolvedAt = &t
	}
	createdAt := report.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	return &ReportDTO{
		ID:                report.ID,
		ReporterAccountID: report.ReporterAccountID,
		ReporterFirstName: reporterFirstName,
		ReporterLastName:  reporterLastName,
		ThreadID:          report.ThreadID,
		PostID:            report.PostID,
		ReportedAccountID: report.ReportedAccountID,
		Reason:            report.Reason,
		Status:            string(report.Status),
		AdminNote:         report.AdminNote,
		ResolvedBy:        report.ResolvedByAccountID,
		ResolvedAt:        resolvedAt,
		CreatedAt:         createdAt,
	}
}

type ListBlockedUsersInput struct {
	ID       uuid.UUID `path:"id" doc:"Thread ID"`
	Page     int       `query:"page" doc:"Page number"`
	PageSize int       `query:"pageSize" doc:"Page size"`
}

type ListBlockedUsersOutput struct {
	Body ListBlockedUsersResponseBody
}

type ListBlockedUsersResponseBody struct {
	BlockedUsers []*ThreadBlockedUserDTO `json:"blockedUsers" doc:"Blocked users list"`
	Total        int64                   `json:"total" doc:"Total count"`
	Page         int                     `json:"page" doc:"Current page"`
	PageSize     int                     `json:"pageSize" doc:"Page size"`
	TotalPages   int                     `json:"totalPages" doc:"Total pages"`
}

type ListAllBlockedUsersInput struct {
	Page     int `query:"page" doc:"Page number"`
	PageSize int `query:"pageSize" doc:"Page size"`
}

type ListAllBlockedUsersOutput struct {
	Body ListAllBlockedUsersResponseBody
}

type ListAllBlockedUsersResponseBody struct {
	BlockedUsers []*ThreadBlockedUserDTO `json:"blockedUsers" doc:"Blocked users list"`
	Total        int64                   `json:"total" doc:"Total count"`
	Page         int                     `json:"page" doc:"Current page"`
	PageSize     int                     `json:"pageSize" doc:"Page size"`
	TotalPages   int                     `json:"totalPages" doc:"Total pages"`
}

type ThreadBlockedUserDTO struct {
	ThreadID             uuid.UUID `json:"threadId" doc:"Thread ID"`
	ThreadTitle          string    `json:"threadTitle,omitempty" doc:"Thread title"`
	ThreadSlug           string    `json:"threadSlug,omitempty" doc:"Thread slug"`
	BlockedAccountID     uuid.UUID `json:"blockedUserId" doc:"Blocked user account ID"`
	BlockedUserFirstName string    `json:"blockedUserFirstName,omitempty" doc:"Blocked user first name"`
	BlockedUserLastName  string    `json:"blockedUserLastName,omitempty" doc:"Blocked user last name"`
	BlockedUserEmail     string    `json:"blockedUserEmail,omitempty" doc:"Blocked user email"`
	BlockedBy            uuid.UUID `json:"blockedBy" doc:"Blocked by account ID"`
	BlockedByFirstName   string    `json:"blockedByFirstName,omitempty" doc:"Blocked by first name"`
	BlockedByLastName    string    `json:"blockedByLastName,omitempty" doc:"Blocked by last name"`
	Reason               *string   `json:"reason,omitempty" doc:"Reason for blocking"`
	CreatedAt            string    `json:"createdAt" doc:"Created timestamp"`
}

func ToThreadBlockedUserDTO(block *entity.ThreadBlockedUser) *ThreadBlockedUserDTO {
	var createdAt string
	if block.CreatedAt != nil {
		createdAt = block.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return &ThreadBlockedUserDTO{
		ThreadID:         block.ThreadID,
		BlockedAccountID: block.BlockedAccountID,
		BlockedBy:        block.BlockedByAccountID,
		Reason:           block.Reason,
		CreatedAt:        createdAt,
	}
}

type ListThreadReportsInput struct {
	Status   string `query:"status" doc:"Filter by status: pending, under_review, resolved, dismissed"`
	Search   string `query:"search" doc:"Search in thread title, slug, description"`
	Page     int    `query:"page" doc:"Page number"`
	PageSize int    `query:"pageSize" doc:"Page size"`
}

type ListThreadReportsOutput struct {
	Body ListThreadReportsResponseBody
}

type ListThreadReportsResponseBody struct {
	Reports    []*ReportWithContentDTO `json:"reports" doc:"Thread report list"`
	Total      int64                   `json:"total" doc:"Total count"`
	Page       int                     `json:"page" doc:"Current page"`
	PageSize   int                     `json:"pageSize" doc:"Page size"`
	TotalPages int                     `json:"totalPages" doc:"Total pages"`
}

type GetThreadReportInput struct {
	ID uuid.UUID `path:"id" doc:"Report ID"`
}

type GetThreadReportOutput struct {
	Body GetThreadReportResponseBody
}

type GetThreadReportResponseBody struct {
	Report  *ReportDTO  `json:"report" doc:"Report details"`
	Content *ContentDTO `json:"content,omitempty" doc:"Thread content"`
}

type UpdateThreadReportStatusInput struct {
	ID   uuid.UUID `path:"id" doc:"Report ID"`
	Body UpdateThreadReportStatusRequest
}

type UpdateThreadReportStatusRequest struct {
	Status    string  `json:"status" doc:"Status: pending, under_review, resolved, dismissed" validate:"required,oneof=pending under_review resolved dismissed"`
	AdminNote *string `json:"adminNote,omitempty" doc:"Admin note"`
}

type UpdateThreadReportStatusOutput struct {
	Body UpdateThreadReportStatusResponseBody
}

type UpdateThreadReportStatusResponseBody struct {
	Report *ReportDTO `json:"report" doc:"Updated report"`
}

type ListPostReportsInput struct {
	Status   string `query:"status" doc:"Filter by status: pending, under_review, resolved, dismissed"`
	Search   string `query:"search" doc:"Search in post content"`
	Page     int    `query:"page" doc:"Page number"`
	PageSize int    `query:"pageSize" doc:"Page size"`
}

type ListPostReportsOutput struct {
	Body ListPostReportsResponseBody
}

type ListPostReportsResponseBody struct {
	Reports    []*ReportWithContentDTO `json:"reports" doc:"Post report list"`
	Total      int64                   `json:"total" doc:"Total count"`
	Page       int                     `json:"page" doc:"Current page"`
	PageSize   int                     `json:"pageSize" doc:"Page size"`
	TotalPages int                     `json:"totalPages" doc:"Total pages"`
}

type GetPostReportInput struct {
	ID uuid.UUID `path:"id" doc:"Report ID"`
}

type GetPostReportOutput struct {
	Body GetPostReportResponseBody
}

type GetPostReportResponseBody struct {
	Report  *ReportDTO  `json:"report" doc:"Report details"`
	Content *ContentDTO `json:"content,omitempty" doc:"Post content"`
}

type UpdatePostReportStatusInput struct {
	ID   uuid.UUID `path:"id" doc:"Report ID"`
	Body UpdatePostReportStatusRequest
}

type UpdatePostReportStatusRequest struct {
	Status    string  `json:"status" doc:"Status: pending, under_review, resolved, dismissed" validate:"required,oneof=pending under_review resolved dismissed"`
	AdminNote *string `json:"adminNote,omitempty" doc:"Admin note"`
}

type UpdatePostReportStatusOutput struct {
	Body UpdatePostReportStatusResponseBody
}

type UpdatePostReportStatusResponseBody struct {
	Report *ReportDTO `json:"report" doc:"Updated report"`
}

type ListUserReportsInput struct {
	Status   string `query:"status" doc:"Filter by status: pending, under_review, resolved, dismissed"`
	Search   string `query:"search" doc:"Search in reporter first name, last name, email"`
	Page     int    `query:"page" doc:"Page number"`
	PageSize int    `query:"pageSize" doc:"Page size"`
}

type ListUserReportsOutput struct {
	Body ListUserReportsResponseBody
}

type ListUserReportsResponseBody struct {
	Reports    []*ReportWithContentDTO `json:"reports" doc:"User report list"`
	Total      int64                   `json:"total" doc:"Total count"`
	Page       int                     `json:"page" doc:"Current page"`
	PageSize   int                     `json:"pageSize" doc:"Page size"`
	TotalPages int                     `json:"totalPages" doc:"Total pages"`
}

type GetUserReportInput struct {
	ID uuid.UUID `path:"id" doc:"Report ID"`
}

type GetUserReportOutput struct {
	Body GetUserReportResponseBody
}

type GetUserReportResponseBody struct {
	Report  *ReportDTO  `json:"report" doc:"Report details"`
	Content *ContentDTO `json:"content,omitempty" doc:"User content"`
}

type UpdateUserReportStatusInput struct {
	ID   uuid.UUID `path:"id" doc:"Report ID"`
	Body UpdateUserReportStatusRequest
}

type UpdateUserReportStatusRequest struct {
	Status    string  `json:"status" doc:"Status: pending, under_review, resolved, dismissed" validate:"required,oneof=pending under_review resolved dismissed"`
	AdminNote *string `json:"adminNote,omitempty" doc:"Admin note"`
}

type UpdateUserReportStatusOutput struct {
	Body UpdateUserReportStatusResponseBody
}

type UpdateUserReportStatusResponseBody struct {
	Report *ReportDTO `json:"report" doc:"Updated report"`
}
