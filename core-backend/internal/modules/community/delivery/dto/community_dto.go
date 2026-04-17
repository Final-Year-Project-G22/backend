package dto

import (
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type CategoryDTO struct {
	ID               uuid.UUID  `json:"id" doc:"Category ID"`
	Name             string     `json:"name" doc:"Category name"`
	Slug             string     `json:"slug" doc:"Category slug"`
	Description      *string    `json:"description,omitempty" doc:"Category description"`
	ParentCategoryID *uuid.UUID `json:"parentCategoryId,omitempty" doc:"Parent category ID"`
	IsActive         bool       `json:"isActive" doc:"Whether category is active"`
	CreatedAt        *time.Time `json:"createdAt,omitempty" doc:"Created timestamp"`
	UpdatedAt        *time.Time `json:"updatedAt,omitempty" doc:"Updated timestamp"`
}

type ThreadDTO struct {
	ID             uuid.UUID           `json:"id" doc:"Thread ID"`
	Title          string              `json:"title" doc:"Thread title"`
	Slug           string              `json:"slug" doc:"Thread slug"`
	Description    *string             `json:"description,omitempty" doc:"Thread description"`
	CategoryID     uuid.UUID           `json:"categoryId" doc:"Category ID"`
	AuthorID       uuid.UUID           `json:"authorId" doc:"Author account ID"`
	IsPinned       bool                `json:"isPinned" doc:"Pinned flag"`
	Status         entity.ThreadStatus `json:"status" doc:"Thread status"`
	ViewCount      int                 `json:"viewCount" doc:"View count"`
	ShareCount     int                 `json:"shareCount" doc:"Share count"`
	ReplyCount     int                 `json:"replyCount" doc:"Reply count"`
	LastActivityAt *time.Time          `json:"lastActivityAt,omitempty" doc:"Last activity timestamp"`
	CreatedAt      *time.Time          `json:"createdAt,omitempty" doc:"Created timestamp"`
	UpdatedAt      *time.Time          `json:"updatedAt,omitempty" doc:"Updated timestamp"`
}

type PostDTO struct {
	ID             uuid.UUID  `json:"id" doc:"Post ID"`
	ThreadID       uuid.UUID  `json:"threadId" doc:"Thread ID"`
	ParentPostID   *uuid.UUID `json:"parentPostId,omitempty" doc:"Parent post ID"`
	AuthorID       uuid.UUID  `json:"authorId" doc:"Author account ID"`
	Content        string     `json:"content" doc:"Post content"`
	IsSolution     bool       `json:"isSolution" doc:"Solution flag"`
	IsPinned       bool       `json:"isPinned" doc:"Pinned flag"`
	UpvoteCount    int        `json:"upvoteCount" doc:"Upvote count"`
	AttachmentURL  *string    `json:"attachmentUrl,omitempty" doc:"Attachment URL"`
	AttachmentType *string    `json:"attachmentType,omitempty" doc:"Attachment type"`
	EditCount      int        `json:"editCount" doc:"Edit count"`
	EditedAt       *time.Time `json:"editedAt,omitempty" doc:"Edited timestamp"`
	CreatedAt      *time.Time `json:"createdAt,omitempty" doc:"Created timestamp"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty" doc:"Updated timestamp"`
}

type ListCategoriesInput struct {
	Page     int    `query:"page" doc:"Page number"`
	PageSize int    `query:"pageSize" doc:"Page size"`
	Search   string `query:"search" doc:"Search term"`
}

type ListCategoriesOutput struct {
	Body ListCategoriesResponseBody
}

type ListCategoriesResponseBody struct {
	Categories []*CategoryDTO `json:"categories" doc:"Category list"`
}

type GetCategoryInput struct {
	ID uuid.UUID `path:"id" doc:"Category ID"`
}

type GetCategoryOutput struct {
	Body GetCategoryResponseBody
}

type GetCategoryResponseBody struct {
	Category *CategoryDTO `json:"category" doc:"Category details"`
}

type ListThreadsByCategoryInput struct {
	CategoryID uuid.UUID `path:"id" doc:"Category ID"`
	Page       int       `query:"page" doc:"Page number"`
	PageSize   int       `query:"pageSize" doc:"Page size"`
	Search     string    `query:"search" doc:"Search term"`
}

type ListThreadsByCategoryOutput struct {
	Body ListThreadsResponseBody
}

type SearchThreadsInput struct {
	Keyword    string `query:"keyword" doc:"Search keyword"`
	CategoryID string `query:"categoryId" doc:"Category ID"`
	Page       int    `query:"page" doc:"Page number"`
	PageSize   int    `query:"pageSize" doc:"Page size"`
}

type SearchThreadsOutput struct {
	Body ListThreadsResponseBody
}

type ListThreadsResponseBody struct {
	Threads []*ThreadDTO `json:"threads" doc:"Thread list"`
}

type GetThreadInput struct {
	ID uuid.UUID `path:"id" doc:"Thread ID"`
}

type GetThreadOutput struct {
	Body GetThreadResponseBody
}

type GetThreadResponseBody struct {
	Thread *ThreadDTO `json:"thread" doc:"Thread details"`
}

type ListPostsInput struct {
	ThreadID uuid.UUID `path:"id" doc:"Thread ID"`
	Page     int       `query:"page" doc:"Page number"`
	PageSize int       `query:"pageSize" doc:"Page size"`
}

type ListPostsOutput struct {
	Body ListPostsResponseBody
}

type ListPostsResponseBody struct {
	Posts []*PostDTO `json:"posts" doc:"Post list"`
}

type CreateThreadFormData struct {
	CategoryID         string        `form:"categoryId" doc:"Category ID"`
	Title              string        `form:"title" doc:"Thread title"`
	Slug               string        `form:"slug" doc:"Thread slug"`
	Description        string        `form:"description" doc:"Thread description"`
	InitialPostContent string        `form:"initialPostContent" doc:"Initial post content"`
	File               huma.FormFile `form:"file" doc:"Optional attachment file"`
}

type CreateThreadInput struct {
	RawBody huma.MultipartFormFiles[CreateThreadFormData]
}

type CreateThreadOutput struct {
	Body CreateThreadResponseBody
}

type CreateThreadResponseBody struct {
	ThreadID uuid.UUID `json:"threadId" doc:"Created thread ID"`
	PostID   uuid.UUID `json:"postId" doc:"Initial post ID"`
}

type CreatePostFormData struct {
	Content string        `form:"content" doc:"Post content"`
	File    huma.FormFile `form:"file" doc:"Optional attachment file"`
}

type CreatePostInput struct {
	ThreadID uuid.UUID `path:"id" doc:"Thread ID"`
	RawBody  huma.MultipartFormFiles[CreatePostFormData]
}

type CreatePostOutput struct {
	Body CreatePostResponseBody
}

type CreatePostResponseBody struct {
	PostID uuid.UUID `json:"postId" doc:"Created post ID"`
}

type ReplyPostInput struct {
	ThreadID uuid.UUID `path:"id" doc:"Thread ID"`
	PostID   uuid.UUID `path:"postId" doc:"Parent post ID"`
	RawBody  huma.MultipartFormFiles[CreatePostFormData]
}

type ReplyPostOutput struct {
	Body CreatePostResponseBody
}

type UpdatePostFormData struct {
	Content          string        `form:"content" doc:"Post content"`
	RemoveAttachment bool          `form:"removeAttachment" doc:"Remove existing attachment"`
	File             huma.FormFile `form:"file" doc:"Optional replacement attachment file"`
}

type UpdatePostInput struct {
	PostID  uuid.UUID `path:"id" doc:"Post ID"`
	RawBody huma.MultipartFormFiles[UpdatePostFormData]
}

type UpdatePostOutput struct {
	Body UpdatePostResponseBody
}

type UpdatePostResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type DeletePostInput struct {
	PostID uuid.UUID `path:"id" doc:"Post ID"`
}

type DeletePostOutput struct {
	Body DeletePostResponseBody
}

type DeletePostResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type MarkSolutionInput struct {
	ThreadID uuid.UUID `path:"id" doc:"Thread ID"`
	PostID   uuid.UUID `path:"postId" doc:"Post ID"`
}

type MarkSolutionOutput struct {
	Body MarkSolutionResponseBody
}

type MarkSolutionResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type FollowThreadInput struct {
	ThreadID uuid.UUID `path:"id" doc:"Thread ID"`
}

type FollowThreadOutput struct {
	Body FollowResponseBody
}

type UnfollowThreadInput struct {
	ThreadID uuid.UUID `path:"id" doc:"Thread ID"`
}

type UnfollowThreadOutput struct {
	Body FollowResponseBody
}

type FollowCategoryInput struct {
	CategoryID uuid.UUID `path:"id" doc:"Category ID"`
}

type FollowCategoryOutput struct {
	Body FollowResponseBody
}

type UnfollowCategoryInput struct {
	CategoryID uuid.UUID `path:"id" doc:"Category ID"`
}

type UnfollowCategoryOutput struct {
	Body FollowResponseBody
}

type FollowResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type ListFollowedThreadsInput struct {
	Page     int    `query:"page" doc:"Page number"`
	PageSize int    `query:"pageSize" doc:"Page size"`
	Search   string `query:"search" doc:"Search term"`
}

type ListFollowedThreadsOutput struct {
	Body ListFollowedThreadsResponseBody
}

type ListFollowedThreadsResponseBody struct {
	Threads []*ThreadDTO `json:"threads" doc:"Followed threads"`
}

type ListFollowedCategoriesInput struct {
	Page     int    `query:"page" doc:"Page number"`
	PageSize int    `query:"pageSize" doc:"Page size"`
	Search   string `query:"search" doc:"Search term"`
}

type ListFollowedCategoriesOutput struct {
	Body ListFollowedCategoriesResponseBody
}

type ListFollowedCategoriesResponseBody struct {
	Categories []*CategoryDTO `json:"categories" doc:"Followed categories"`
}

type ReportContentRequest struct {
	TargetType entity.TargetType `json:"targetType" doc:"Target type"`
	TargetID   uuid.UUID         `json:"targetId" doc:"Target ID"`
	Reason     string            `json:"reason" doc:"Report reason"`
}

type ReportContentInput struct {
	Body ReportContentRequest
}

type ReportContentOutput struct {
	Body ReportContentResponseBody
}

type ReportContentResponseBody struct {
	ReportID uuid.UUID `json:"reportId" doc:"Report ID"`
}

type BlockUserRequest struct {
	BlockedID uuid.UUID `json:"blockedId" doc:"Blocked account ID"`
	Reason    *string   `json:"reason,omitempty" doc:"Block reason"`
}

type BlockUserInput struct {
	ThreadID uuid.UUID `path:"id" doc:"Thread ID"`
	Body     BlockUserRequest
}

type BlockUserOutput struct {
	Body BlockUserResponseBody
}

type BlockUserResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type UnblockUserInput struct {
	ThreadID  uuid.UUID `path:"id" doc:"Thread ID"`
	AccountID uuid.UUID `path:"accountId" doc:"Blocked account ID"`
}

type UnblockUserOutput struct {
	Body BlockUserResponseBody
}

func ToQueryOptions(page, pageSize int) query.QueryOptions {
	opts := query.QueryOptions{}
	if page > 0 {
		opts.Page = page
	}
	if pageSize > 0 {
		opts.PageSize = pageSize
	}
	return opts
}

func ToCategoryDTO(category *entity.CommunityCategory) *CategoryDTO {
	if category == nil {
		return nil
	}
	return &CategoryDTO{
		ID:               category.ID,
		Name:             category.Name,
		Slug:             category.Slug,
		Description:      category.Description,
		ParentCategoryID: category.ParentCategoryID,
		IsActive:         category.IsActive,
		CreatedAt:        category.CreatedAt,
		UpdatedAt:        category.UpdatedAt,
	}
}

func ToThreadDTO(thread *entity.DiscussionThread) *ThreadDTO {
	if thread == nil {
		return nil
	}
	return &ThreadDTO{
		ID:             thread.ID,
		Title:          thread.Title,
		Slug:           thread.Slug,
		Description:    thread.Description,
		CategoryID:     thread.CategoryID,
		AuthorID:       thread.AuthorAccountID,
		IsPinned:       thread.IsPinned,
		Status:         thread.Status,
		ViewCount:      thread.ViewCount,
		ShareCount:     thread.ShareCount,
		ReplyCount:     thread.ReplyCount,
		LastActivityAt: thread.LastActivityAt,
		CreatedAt:      thread.CreatedAt,
		UpdatedAt:      thread.UpdatedAt,
	}
}

func ToPostDTO(post *entity.DiscussionPost) *PostDTO {
	if post == nil {
		return nil
	}
	return &PostDTO{
		ID:             post.ID,
		ThreadID:       post.ThreadID,
		ParentPostID:   post.ParentPostID,
		AuthorID:       post.AuthorAccountID,
		Content:        post.Content,
		IsSolution:     post.IsSolution,
		IsPinned:       post.IsPinned,
		UpvoteCount:    post.UpvoteCount,
		AttachmentURL:  post.AttachmentURL,
		AttachmentType: post.AttachmentType,
		EditCount:      post.EditCount,
		EditedAt:       post.EditedAt,
		CreatedAt:      post.CreatedAt,
		UpdatedAt:      post.UpdatedAt,
	}
}

func ToCreateThreadInput(categoryID uuid.UUID, title, slug, description, initialPostContent string, attachmentURL, attachmentType *string) usecase.CreateThreadInput {
	var descriptionPtr *string
	if strings.TrimSpace(description) != "" {
		d := strings.TrimSpace(description)
		descriptionPtr = &d
	}

	return usecase.CreateThreadInput{
		CategoryID:         categoryID,
		Title:              title,
		Slug:               slug,
		Description:        descriptionPtr,
		InitialPostContent: initialPostContent,
		AttachmentURL:      attachmentURL,
		AttachmentType:     attachmentType,
	}
}

func ToCreatePostInput(content string, attachmentURL, attachmentType *string) usecase.CreatePostInput {
	return usecase.CreatePostInput{
		Content:        content,
		AttachmentURL:  attachmentURL,
		AttachmentType: attachmentType,
	}
}

func ToUpdatePostInput(content string, removeAttachment bool) usecase.UpdatePostInput {
	var contentPtr *string
	if strings.TrimSpace(content) != "" {
		c := strings.TrimSpace(content)
		contentPtr = &c
	}

	var removeAttachmentPtr *bool
	if removeAttachment {
		r := true
		removeAttachmentPtr = &r
	}

	return usecase.UpdatePostInput{
		Content:          contentPtr,
		AttachmentURL:    nil,
		AttachmentType:   nil,
		RemoveAttachment: removeAttachmentPtr,
	}
}
