package dto

import (
	"strconv"
	"strings"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
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
	ParentThreadID *uuid.UUID          `json:"parentThreadId,omitempty" doc:"Parent thread ID for sub-threads"`
	Title          string              `json:"title" doc:"Thread title"`
	Slug           string              `json:"slug" doc:"Thread slug"`
	Description    *string             `json:"description,omitempty" doc:"Thread description"`
	SectorIDs      []uuid.UUID         `json:"sectorIds" doc:"Sector IDs"`
	TagIDs         []uuid.UUID         `json:"tagIds" doc:"Tag IDs"`
	AuthorID       uuid.UUID           `json:"authorId" doc:"Author account ID"`
	AuthorUsername string              `json:"authorUsername,omitempty" doc:"Author username"`
	AuthorDisplay  string              `json:"authorDisplayName" doc:"Author display name"`
	AuthorAvatar   *string             `json:"authorAvatarUrl,omitempty" doc:"Author avatar URL"`
	IsPinned       bool                `json:"isPinned" doc:"Pinned flag"`
	IsFollowed     bool                `json:"isFollowed" doc:"Whether current user follows this thread"`
	IsMuted        bool                `json:"isMuted" doc:"Whether current user has muted this thread"`
	UnreadCount    int                 `json:"unreadCount" doc:"Number of unread posts/replies"`
	HasSolution    bool                `json:"hasSolution" doc:"Whether thread has an accepted solution"`
	Status         entity.ThreadStatus `json:"status" doc:"Thread status"`
	ViewCount      int                 `json:"viewCount" doc:"View count"`
	ShareCount     int                 `json:"shareCount" doc:"Share count"`
	ReplyCount     int                 `json:"replyCount" doc:"Reply count"`
	LastActivityAt *time.Time          `json:"lastActivityAt,omitempty" doc:"Last activity timestamp"`
	CreatedAt      *time.Time          `json:"createdAt,omitempty" doc:"Created timestamp"`
	UpdatedAt      *time.Time          `json:"updatedAt,omitempty" doc:"Updated timestamp"`
}

type AttachmentDTO struct {
	ID       uuid.UUID `json:"id"`
	FileURL  string    `json:"fileUrl"`
	FileType string    `json:"fileType"`
	FileName string    `json:"fileName"`
	FileSize *int64    `json:"fileSize,omitempty"`
}

type PostDTO struct {
	ID             uuid.UUID       `json:"id" doc:"Post ID"`
	ThreadID       uuid.UUID       `json:"threadId" doc:"Thread ID"`
	ParentPostID   *uuid.UUID      `json:"parentPostId,omitempty" doc:"Parent post ID"`
	AuthorID       uuid.UUID       `json:"authorId" doc:"Author account ID"`
	AuthorUsername string          `json:"authorUsername,omitempty" doc:"Author username"`
	AuthorDisplay  string          `json:"authorDisplayName" doc:"Author display name"`
	AuthorAvatar   *string         `json:"authorAvatarUrl,omitempty" doc:"Author avatar URL"`
	Content        string          `json:"content" doc:"Post content"`
	IsSolution     bool            `json:"isSolution" doc:"Solution flag"`
	IsPinned       bool            `json:"isPinned" doc:"Pinned flag"`
	UpvoteCount    int             `json:"upvoteCount" doc:"Upvote count"`
	Attachments    []AttachmentDTO `json:"attachments,omitempty" doc:"Post attachments"`
	EditCount      int             `json:"editCount" doc:"Edit count"`
	EditedAt       *time.Time      `json:"editedAt,omitempty" doc:"Edited timestamp"`
	CreatedAt      *time.Time      `json:"createdAt,omitempty" doc:"Created timestamp"`
	UpdatedAt      *time.Time      `json:"updatedAt,omitempty" doc:"Updated timestamp"`
}

type AuthorMeta struct {
	Username    string
	DisplayName string
	AvatarURL   *string
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

type ListThreadsInput struct {
	Search   string `query:"search" doc:"Search term"`
	Page     int    `query:"page" doc:"Page number"`
	PageSize int    `query:"pageSize" doc:"Page size"`
}

type SearchThreadsInput struct {
	Keyword  string `query:"keyword" doc:"Search keyword"`
	Page     int    `query:"page" doc:"Page number"`
	PageSize int    `query:"pageSize" doc:"Page size"`
}

type SearchThreadsOutput struct {
	Body ListThreadsResponseBody
}

type ListThreadsOutput struct {
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
	SectorIDs          string `form:"sectorIds" doc:"Sector IDs (comma-separated)" required:"false"`
	TagIDs             string `form:"tagIds" doc:"Tag IDs (comma-separated)" required:"false"`
	ParentThreadID     string `form:"parentThreadId" doc:"Parent thread ID for sub-threads" required:"false"`
	Title              string `form:"title" doc:"Thread title"`
	Slug               string `form:"slug" doc:"Thread slug"`
	Description        string `form:"description" doc:"Thread description" required:"false"`
	InitialPostContent string `form:"initialPostContent" doc:"Initial post content"`
	AttachmentIds      string `form:"attachmentIds" doc:"Pre-uploaded attachment IDs (comma-separated)" required:"false"`
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

type UpdateThreadFormData struct {
	SectorIDs   string `form:"sectorIds" doc:"Sector IDs (comma-separated)" required:"false"`
	TagIDs      string `form:"tagIds" doc:"Tag IDs (comma-separated)" required:"false"`
	Title       string `form:"title" doc:"Thread title" required:"false"`
	Description string `form:"description" doc:"Thread description" required:"false"`
	IsPinned    string `form:"isPinned" doc:"Pinned flag (true/false)" required:"false"`
	Status      string `form:"status" doc:"Thread status" required:"false"`
}

type UpdateThreadInput struct {
	ID      uuid.UUID `path:"id" doc:"Thread ID"`
	RawBody huma.MultipartFormFiles[UpdateThreadFormData]
}

type UpdateThreadOutput struct {
	Body UpdateThreadResponseBody
}

type UpdateThreadResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type DeleteThreadInput struct {
	ID uuid.UUID `path:"id" doc:"Thread ID"`
}

type DeleteThreadOutput struct {
	Body DeleteThreadResponseBody
}

type DeleteThreadResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type CreatePostFormData struct {
	Content       string `form:"content" doc:"Post content"`
	AttachmentIds string `form:"attachmentIds" doc:"Pre-uploaded attachment IDs (comma-separated)" required:"false"`
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
	Content              string `form:"content" doc:"Post content"`
	AttachmentIds        string `form:"attachmentIds" doc:"New attachment IDs to link (comma-separated)" required:"false" explode:"true"`
	RemoveAttachmentIds  string `form:"removeAttachmentIds" doc:"Attachment IDs to unlink (comma-separated)" required:"false"`
	RemoveAllAttachments bool   `form:"removeAllAttachments" doc:"Remove all attachments"`
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

type MarkThreadReadInput struct {
	ThreadID uuid.UUID `path:"id" doc:"Thread ID"`
}

type MarkThreadReadOutput struct {
	Body MarkThreadReadResponseBody
}

type MarkThreadReadResponseBody struct {
	Message string `json:"message" doc:"Success message"`
}

type MuteThreadInput struct {
	ThreadID uuid.UUID `path:"id" doc:"Thread ID"`
}

type MuteThreadOutput struct {
	Body FollowResponseBody
}

type UnmuteThreadInput struct {
	ThreadID uuid.UUID `path:"id" doc:"Thread ID"`
}

type UnmuteThreadOutput struct {
	Body FollowResponseBody
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

type ReportThreadRequest struct {
	Reason string `json:"reason" doc:"Report reason" validate:"required"`
}

type ReportThreadInput struct {
	ID   uuid.UUID `path:"id" doc:"Thread ID"`
	Body ReportThreadRequest
}

type ReportThreadOutput struct {
	Body ReportThreadResponseBody
}

type ReportThreadResponseBody struct {
	ReportID uuid.UUID `json:"reportId" doc:"Report ID"`
}

type ReportPostRequest struct {
	Reason string `json:"reason" doc:"Report reason" validate:"required"`
}

type ReportPostInput struct {
	ID     uuid.UUID `path:"id" doc:"Thread ID"`
	PostID uuid.UUID `path:"postId" doc:"Post ID"`
	Body   ReportPostRequest
}

type ReportPostOutput struct {
	Body ReportPostResponseBody
}

type ReportPostResponseBody struct {
	ReportID uuid.UUID `json:"reportId" doc:"Report ID"`
}

type ReportUserRequest struct {
	ReportedAccountID uuid.UUID `json:"reportedAccountId" doc:"Reported account ID" validate:"required"`
	Reason            string    `json:"reason" doc:"Report reason" validate:"required"`
}

type ReportUserInput struct {
	ID   uuid.UUID `path:"id" doc:"Thread ID"`
	Body ReportUserRequest
}

type ReportUserOutput struct {
	Body ReportUserResponseBody
}

type ReportUserResponseBody struct {
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

func ToThreadDTO(thread *entity.DiscussionThread, authorMeta *AuthorMeta, isFollowed bool, isMuted bool, unreadCount int, hasSolution bool) *ThreadDTO {
	if thread == nil {
		return nil
	}
	displayName := "User " + thread.AuthorAccountID.String()[:6]
	authorUsername := ""
	var authorAvatarURL *string
	if authorMeta != nil {
		if authorMeta.DisplayName != "" {
			displayName = authorMeta.DisplayName
		}
		authorUsername = authorMeta.Username
		authorAvatarURL = authorMeta.AvatarURL
	}

	return &ThreadDTO{
		ID:             thread.ID,
		ParentThreadID: thread.ParentThreadID,
		Title:          thread.Title,
		Slug:           thread.Slug,
		Description:    thread.Description,
		SectorIDs:      thread.SectorIDs,
		TagIDs:         thread.TagIDs,
		AuthorID:       thread.AuthorAccountID,
		AuthorUsername: authorUsername,
		AuthorDisplay:  displayName,
		AuthorAvatar:   authorAvatarURL,
		IsPinned:       thread.IsPinned,
		IsFollowed:     isFollowed,
		IsMuted:        isMuted,
		UnreadCount:    unreadCount,
		HasSolution:    hasSolution,
		Status:         thread.Status,
		ViewCount:      thread.ViewCount,
		ShareCount:     thread.ShareCount,
		ReplyCount:     thread.ReplyCount,
		LastActivityAt: thread.LastActivityAt,
		CreatedAt:      thread.CreatedAt,
		UpdatedAt:      thread.UpdatedAt,
	}
}

func ToPostDTO(post *entity.DiscussionPost, authorMeta *AuthorMeta) *PostDTO {
	if post == nil {
		return nil
	}
	displayName := "User " + post.AuthorAccountID.String()[:6]
	authorUsername := ""
	var authorAvatarURL *string
	if authorMeta != nil {
		if authorMeta.DisplayName != "" {
			displayName = authorMeta.DisplayName
		}
		authorUsername = authorMeta.Username
		authorAvatarURL = authorMeta.AvatarURL
	}

	attachments := make([]AttachmentDTO, 0, len(post.Attachments))
	for _, att := range post.Attachments {
		attDTO := AttachmentDTO{
			ID:       att.ID,
			FileURL:  att.FileURL,
			FileType: att.FileType,
			FileName: att.FileName,
			FileSize: att.FileSize,
		}
		attachments = append(attachments, attDTO)
	}

	return &PostDTO{
		ID:             post.ID,
		ThreadID:       post.ThreadID,
		ParentPostID:   post.ParentPostID,
		AuthorID:       post.AuthorAccountID,
		AuthorUsername: authorUsername,
		AuthorDisplay:  displayName,
		AuthorAvatar:   authorAvatarURL,
		Content:        post.Content,
		IsSolution:     post.IsSolution,
		IsPinned:       post.IsPinned,
		UpvoteCount:    post.UpvoteCount,
		Attachments:    attachments,
		EditCount:      post.EditCount,
		EditedAt:       post.EditedAt,
		CreatedAt:      post.CreatedAt,
		UpdatedAt:      post.UpdatedAt,
	}
}

// parseStringSliceToUUID converts a slice of strings to a slice of UUIDs.
// Invalid UUID strings are silently skipped.
func parseStringSliceToUUID(strs []string) []uuid.UUID {
	if len(strs) == 0 {
		return nil
	}
	result := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		if parsed, err := uuid.Parse(s); err == nil {
			result = append(result, parsed)
		}
	}
	return result
}

// parseCSVToStringSlice converts a comma-separated string to a slice of strings.
func parseCSVToStringSlice(csv string) []string {
	if strings.TrimSpace(csv) == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func ToCreateThreadInput(sectorIDs, tagIDs string, parentThreadID *uuid.UUID, title, slug, description, initialPostContent string, attachmentIds string) usecase.CreateThreadInput {
	var descriptionPtr *string
	if strings.TrimSpace(description) != "" {
		d := strings.TrimSpace(description)
		descriptionPtr = &d
	}

	return usecase.CreateThreadInput{
		SectorIDs:          parseStringSliceToUUID(parseCSVToStringSlice(sectorIDs)),
		TagIDs:             parseStringSliceToUUID(parseCSVToStringSlice(tagIDs)),
		ParentThreadID:     parentThreadID,
		Title:              title,
		Slug:               slug,
		Description:        descriptionPtr,
		InitialPostContent: initialPostContent,
		AttachmentIds:      parseStringSliceToUUID(parseCSVToStringSlice(attachmentIds)),
	}
}

func ToCreatePostInput(content string, attachmentIds string) usecase.CreatePostInput {
	return usecase.CreatePostInput{
		Content:       content,
		AttachmentIds: parseStringSliceToUUID(parseCSVToStringSlice(attachmentIds)),
	}
}

func ToUpdatePostInput(content string, attachmentIds, removeAttachmentIds string, removeAllAttachments bool) usecase.UpdatePostInput {
	var contentPtr *string
	if strings.TrimSpace(content) != "" {
		c := strings.TrimSpace(content)
		contentPtr = &c
	}

	return usecase.UpdatePostInput{
		Content:              contentPtr,
		AttachmentIds:        parseStringSliceToUUID(parseCSVToStringSlice(attachmentIds)),
		RemoveAttachmentIds:  parseStringSliceToUUID(parseCSVToStringSlice(removeAttachmentIds)),
		RemoveAllAttachments: removeAllAttachments,
	}
}

func ToUpdateThreadInput(formData *UpdateThreadFormData) (usecase.UpdateThreadInput, error) {
	input := usecase.UpdateThreadInput{}

	if formData.SectorIDs != "" {
		ids := parseStringSliceToUUID(parseCSVToStringSlice(formData.SectorIDs))
		input.SectorIDs = &ids
	}
	if formData.TagIDs != "" {
		ids := parseStringSliceToUUID(parseCSVToStringSlice(formData.TagIDs))
		input.TagIDs = &ids
	}
	if strings.TrimSpace(formData.Title) != "" {
		t := strings.TrimSpace(formData.Title)
		input.Title = &t
	}
	if strings.TrimSpace(formData.Description) != "" {
		d := strings.TrimSpace(formData.Description)
		input.Description = &d
	}
	if strings.TrimSpace(formData.IsPinned) != "" {
		b, err := strconv.ParseBool(strings.TrimSpace(formData.IsPinned))
		if err != nil {
			return input, apperrors.InvalidInputError("isPinned", "must be true or false")
		}
		input.IsPinned = &b
	}
	if strings.TrimSpace(formData.Status) != "" {
		s := entity.ThreadStatus(strings.TrimSpace(formData.Status))
		input.Status = &s
	}

	return input, nil
}
