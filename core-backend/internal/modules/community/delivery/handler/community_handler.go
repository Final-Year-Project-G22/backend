package handler

import (
	"context"
	"io"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	iamusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type CommunityHandler struct {
	communityService  service.CommunityService
	attachmentUsecase usecase.AttachmentUsecase
	categoryUsecase   usecase.CommunityCategoryUsecase
	threadUsecase     usecase.DiscussionThreadUsecase
	postUsecase       usecase.DiscussionPostUsecase
	accountUsecase    iamusecase.AccountUsecase
	userUsecase       iamusecase.UserUsecase
}

func NewCommunityHandler(
	communityService service.CommunityService,
	attachmentUsecase usecase.AttachmentUsecase,
	categoryUsecase usecase.CommunityCategoryUsecase,
	threadUsecase usecase.DiscussionThreadUsecase,
	postUsecase usecase.DiscussionPostUsecase,
	accountUsecase iamusecase.AccountUsecase,
	userUsecase iamusecase.UserUsecase,
) *CommunityHandler {
	return &CommunityHandler{
		communityService:  communityService,
		attachmentUsecase: attachmentUsecase,
		categoryUsecase:   categoryUsecase,
		threadUsecase:     threadUsecase,
		postUsecase:       postUsecase,
		accountUsecase:    accountUsecase,
		userUsecase:       userUsecase,
	}
}

func (h *CommunityHandler) HandleListCategories(ctx context.Context, input *dto.ListCategoriesInput) (*dto.ListCategoriesOutput, error) {
	opts := dto.ToQueryOptions(input.Page, input.PageSize)
	opts.Search = input.Search
	categories, err := h.categoryUsecase.ListCategories(ctx, false, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	items := make([]*dto.CategoryDTO, 0, len(categories))
	for _, category := range categories {
		items = append(items, dto.ToCategoryDTO(category))
	}
	return &dto.ListCategoriesOutput{Body: dto.ListCategoriesResponseBody{Categories: items}}, nil
}

func (h *CommunityHandler) HandleGetCategory(ctx context.Context, input *dto.GetCategoryInput) (*dto.GetCategoryOutput, error) {
	category, err := h.categoryUsecase.GetCategory(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.GetCategoryOutput{Body: dto.GetCategoryResponseBody{Category: dto.ToCategoryDTO(category)}}, nil
}

func (h *CommunityHandler) HandleListThreads(ctx context.Context, input *dto.ListThreadsInput) (*dto.ListThreadsOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	opts := dto.ToQueryOptions(input.Page, input.PageSize)
	opts.Search = input.Search

	threads, err := h.threadUsecase.ListThreads(ctx, accountID, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	threadIDs := make([]uuid.UUID, 0, len(threads))
	for _, thread := range threads {
		threadIDs = append(threadIDs, thread.ID)
	}
	followStatus, err := h.communityService.ListThreadFollowStatus(ctx, accountID, threadIDs)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	unreadCounts, err := h.communityService.ListThreadUnreadCounts(ctx, accountID, threadIDs)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	authorCache := make(map[uuid.UUID]*dto.AuthorMeta)
	items := make([]*dto.ThreadDTO, 0, len(threads))
	for _, thread := range threads {
		authorMeta := h.resolveAuthorMeta(ctx, thread.AuthorAccountID, authorCache)
		items = append(items, dto.ToThreadDTO(thread, authorMeta, followStatus[thread.ID], unreadCounts[thread.ID]))
	}

	return &dto.ListThreadsOutput{Body: dto.ListThreadsResponseBody{Threads: items}}, nil
}

func (h *CommunityHandler) HandleListAllThreads(ctx context.Context, input *dto.ListThreadsInput) (*dto.ListThreadsOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	opts := dto.ToQueryOptions(input.Page, input.PageSize)

	threads, err := h.threadUsecase.ListAllThreads(ctx, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	threadIDs := make([]uuid.UUID, 0, len(threads))
	for _, thread := range threads {
		threadIDs = append(threadIDs, thread.ID)
	}
	followStatus, err := h.communityService.ListThreadFollowStatus(ctx, accountID, threadIDs)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	unreadCounts, err := h.communityService.ListThreadUnreadCounts(ctx, accountID, threadIDs)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	authorCache := make(map[uuid.UUID]*dto.AuthorMeta)
	items := make([]*dto.ThreadDTO, 0, len(threads))
	for _, thread := range threads {
		authorMeta := h.resolveAuthorMeta(ctx, thread.AuthorAccountID, authorCache)
		items = append(items, dto.ToThreadDTO(thread, authorMeta, followStatus[thread.ID], unreadCounts[thread.ID]))
	}

	return &dto.ListThreadsOutput{Body: dto.ListThreadsResponseBody{Threads: items}}, nil
}

func (h *CommunityHandler) HandleSearchThreads(ctx context.Context, input *dto.SearchThreadsInput) (*dto.SearchThreadsOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	opts := dto.ToQueryOptions(input.Page, input.PageSize)

	threads, err := h.threadUsecase.SearchThreads(ctx, input.Keyword, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	threadIDs := make([]uuid.UUID, 0, len(threads))
	for _, thread := range threads {
		threadIDs = append(threadIDs, thread.ID)
	}
	followStatus, err := h.communityService.ListThreadFollowStatus(ctx, accountID, threadIDs)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	unreadCounts, err := h.communityService.ListThreadUnreadCounts(ctx, accountID, threadIDs)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	authorCache := make(map[uuid.UUID]*dto.AuthorMeta)
	items := make([]*dto.ThreadDTO, 0, len(threads))
	for _, thread := range threads {
		authorMeta := h.resolveAuthorMeta(ctx, thread.AuthorAccountID, authorCache)
		items = append(items, dto.ToThreadDTO(thread, authorMeta, followStatus[thread.ID], unreadCounts[thread.ID]))
	}
	return &dto.SearchThreadsOutput{Body: dto.ListThreadsResponseBody{Threads: items}}, nil
}

func (h *CommunityHandler) HandleGetThread(ctx context.Context, input *dto.GetThreadInput) (*dto.GetThreadOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	thread, err := h.threadUsecase.GetThread(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	_ = h.communityService.RecordThreadView(ctx, accountID, thread.ID)
	followStatus, err := h.communityService.ListThreadFollowStatus(ctx, accountID, []uuid.UUID{thread.ID})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	unreadCounts, err := h.communityService.ListThreadUnreadCounts(ctx, accountID, []uuid.UUID{thread.ID})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	authorMeta := h.resolveAuthorMeta(ctx, thread.AuthorAccountID, make(map[uuid.UUID]*dto.AuthorMeta))
	return &dto.GetThreadOutput{Body: dto.GetThreadResponseBody{Thread: dto.ToThreadDTO(thread, authorMeta, followStatus[thread.ID], unreadCounts[thread.ID])}}, nil
}

func (h *CommunityHandler) HandleListPosts(ctx context.Context, input *dto.ListPostsInput) (*dto.ListPostsOutput, error) {
	opts := dto.ToQueryOptions(input.Page, input.PageSize)
	opts.Preload = []string{"Attachments"}
	posts, err := h.postUsecase.ListPostsByThread(ctx, input.ThreadID, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	authorCache := make(map[uuid.UUID]*dto.AuthorMeta)
	items := make([]*dto.PostDTO, 0, len(posts))
	for _, post := range posts {
		authorMeta := h.resolveAuthorMeta(ctx, post.AuthorAccountID, authorCache)
		items = append(items, dto.ToPostDTO(post, authorMeta))
	}
	return &dto.ListPostsOutput{Body: dto.ListPostsResponseBody{Posts: items}}, nil
}

func (h *CommunityHandler) HandleCreateThread(ctx context.Context, input *dto.CreateThreadInput) (*dto.CreateThreadOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	formData := input.RawBody.Data()
	if formData == nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("community.errors.invalidInput"))
	}

	var parentThreadID *uuid.UUID
	if formData.ParentThreadID != "" {
		parsed, err := uuid.Parse(formData.ParentThreadID)
		if err != nil {
			return nil, apperrors.ToHumaError(ctx, apperrors.InvalidInputError("parentThreadId", "community.errors.invalidInput"))
		}
		parentThreadID = &parsed
	}

	thread, post, err := h.communityService.CreateThreadWithPost(ctx, accountID, dto.ToCreateThreadInput(
		formData.SectorIDs,
		formData.TagIDs,
		parentThreadID,
		formData.Title,
		formData.Slug,
		formData.Description,
		formData.InitialPostContent,
		formData.AttachmentIds,
	))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateThreadOutput{Body: dto.CreateThreadResponseBody{ThreadID: thread.ID, PostID: post.ID}}, nil
}

func (h *CommunityHandler) HandleUpdateThread(ctx context.Context, input *dto.UpdateThreadInput) (*dto.UpdateThreadOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	formData := input.RawBody.Data()
	if formData == nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("community.errors.invalidInput"))
	}

	updateInput, err := dto.ToUpdateThreadInput(formData)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	_, err = h.threadUsecase.UpdateThread(ctx, accountID, input.ID, updateInput)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.UpdateThreadOutput{Body: dto.UpdateThreadResponseBody{Message: "Thread updated"}}, nil
}

func (h *CommunityHandler) HandleDeleteThread(ctx context.Context, input *dto.DeleteThreadInput) (*dto.DeleteThreadOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))

	if err := h.threadUsecase.DeleteThread(ctx, accountID, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.DeleteThreadOutput{Body: dto.DeleteThreadResponseBody{Message: "Thread deleted"}}, nil
}

func (h *CommunityHandler) HandleCreatePost(ctx context.Context, input *dto.CreatePostInput) (*dto.CreatePostOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	formData := input.RawBody.Data()
	if formData == nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("community.errors.invalidInput"))
	}

	post, err := h.communityService.ReplyToThread(ctx, accountID, input.ThreadID, nil, dto.ToCreatePostInput(
		formData.Content,
		formData.AttachmentIds,
	))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreatePostOutput{Body: dto.CreatePostResponseBody{PostID: post.ID}}, nil
}

func (h *CommunityHandler) HandleReplyPost(ctx context.Context, input *dto.ReplyPostInput) (*dto.ReplyPostOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	formData := input.RawBody.Data()
	if formData == nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("community.errors.invalidInput"))
	}

	post, err := h.communityService.ReplyToThread(ctx, accountID, input.ThreadID, &input.PostID, dto.ToCreatePostInput(
		formData.Content,
		formData.AttachmentIds,
	))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.ReplyPostOutput{Body: dto.CreatePostResponseBody{PostID: post.ID}}, nil
}

func (h *CommunityHandler) HandleUpdatePost(ctx context.Context, input *dto.UpdatePostInput) (*dto.UpdatePostOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	formData := input.RawBody.Data()
	if formData == nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("community.errors.invalidInput"))
	}

	_, err := h.communityService.UpdatePost(ctx, accountID, input.PostID, dto.ToUpdatePostInput(
		formData.Content,
		formData.AttachmentIds,
		formData.RemoveAttachmentIds,
		formData.RemoveAllAttachments,
	))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdatePostOutput{Body: dto.UpdatePostResponseBody{Message: "Post updated"}}, nil
}

func (h *CommunityHandler) HandleDeletePost(ctx context.Context, input *dto.DeletePostInput) (*dto.DeletePostOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.postUsecase.DeletePost(ctx, accountID, input.PostID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeletePostOutput{Body: dto.DeletePostResponseBody{Message: "Post deleted"}}, nil
}

func (h *CommunityHandler) HandleMarkSolution(ctx context.Context, input *dto.MarkSolutionInput) (*dto.MarkSolutionOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.communityService.MarkSolution(ctx, accountID, input.ThreadID, input.PostID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.MarkSolutionOutput{Body: dto.MarkSolutionResponseBody{Message: "Solution marked"}}, nil
}

func (h *CommunityHandler) HandleFollowThread(ctx context.Context, input *dto.FollowThreadInput) (*dto.FollowThreadOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.communityService.FollowThread(ctx, accountID, input.ThreadID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.FollowThreadOutput{Body: dto.FollowResponseBody{Message: "Thread followed"}}, nil
}

func (h *CommunityHandler) HandleUnfollowThread(ctx context.Context, input *dto.UnfollowThreadInput) (*dto.UnfollowThreadOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.communityService.UnfollowThread(ctx, accountID, input.ThreadID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UnfollowThreadOutput{Body: dto.FollowResponseBody{Message: "Thread unfollowed"}}, nil
}

func (h *CommunityHandler) HandleMarkThreadRead(ctx context.Context, input *dto.MarkThreadReadInput) (*dto.MarkThreadReadOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.communityService.MarkThreadRead(ctx, accountID, input.ThreadID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.MarkThreadReadOutput{Body: dto.MarkThreadReadResponseBody{Message: "Thread marked as read"}}, nil
}

func (h *CommunityHandler) HandleFollowCategory(ctx context.Context, input *dto.FollowCategoryInput) (*dto.FollowCategoryOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.communityService.FollowCategory(ctx, accountID, input.CategoryID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.FollowCategoryOutput{Body: dto.FollowResponseBody{Message: "Category followed"}}, nil
}

func (h *CommunityHandler) HandleUnfollowCategory(ctx context.Context, input *dto.UnfollowCategoryInput) (*dto.UnfollowCategoryOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.communityService.UnfollowCategory(ctx, accountID, input.CategoryID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UnfollowCategoryOutput{Body: dto.FollowResponseBody{Message: "Category unfollowed"}}, nil
}

func (h *CommunityHandler) HandleListFollowedThreads(ctx context.Context, input *dto.ListFollowedThreadsInput) (*dto.ListFollowedThreadsOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	opts := dto.ToQueryOptions(input.Page, input.PageSize)
	opts.Search = input.Search
	settings, err := h.communityService.ListFollowedThreads(ctx, accountID, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	threadIDs := make([]uuid.UUID, 0, len(settings))
	for _, setting := range settings {
		threadIDs = append(threadIDs, setting.Thread.ID)
	}
	unreadCounts, err := h.communityService.ListThreadUnreadCounts(ctx, accountID, threadIDs)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	items := make([]*dto.ThreadDTO, 0, len(settings))
	authorCache := make(map[uuid.UUID]*dto.AuthorMeta)
	for _, setting := range settings {
		authorMeta := h.resolveAuthorMeta(ctx, setting.Thread.AuthorAccountID, authorCache)
		items = append(items, dto.ToThreadDTO(&setting.Thread, authorMeta, true, unreadCounts[setting.Thread.ID]))
	}
	return &dto.ListFollowedThreadsOutput{Body: dto.ListFollowedThreadsResponseBody{Threads: items}}, nil
}

func (h *CommunityHandler) resolveAuthorMeta(ctx context.Context, accountID uuid.UUID, cache map[uuid.UUID]*dto.AuthorMeta) *dto.AuthorMeta {
	if cached, ok := cache[accountID]; ok {
		return cached
	}

	fallback := &dto.AuthorMeta{
		DisplayName: "User " + accountID.String()[:6],
	}

	account, err := h.accountUsecase.GetAccount(ctx, accountID)
	if err != nil || account == nil {
		cache[accountID] = fallback
		return fallback
	}

	username := ""
	if account.Username != nil {
		username = strings.TrimSpace(*account.Username)
	}

	displayName := fallback.DisplayName
	if username != "" {
		displayName = username
	}

	meta := &dto.AuthorMeta{
		Username:    username,
		DisplayName: displayName,
	}

	user, userErr := h.userUsecase.GetUser(ctx, account.UserID)
	if userErr == nil && user != nil {
		fullName := strings.TrimSpace(strings.TrimSpace(user.FirstName) + " " + strings.TrimSpace(user.LastName))
		if meta.DisplayName == fallback.DisplayName && fullName != "" {
			meta.DisplayName = fullName
		}
		meta.AvatarURL = user.ImageURL
	}

	cache[accountID] = meta
	return meta
}

func (h *CommunityHandler) HandleListFollowedCategories(ctx context.Context, input *dto.ListFollowedCategoriesInput) (*dto.ListFollowedCategoriesOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	opts := dto.ToQueryOptions(input.Page, input.PageSize)
	opts.Search = input.Search
	settings, err := h.communityService.ListFollowedCategories(ctx, accountID, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	items := make([]*dto.CategoryDTO, 0, len(settings))
	for _, setting := range settings {
		items = append(items, dto.ToCategoryDTO(&setting.Category))
	}
	return &dto.ListFollowedCategoriesOutput{Body: dto.ListFollowedCategoriesResponseBody{Categories: items}}, nil
}

func (h *CommunityHandler) HandleReportThread(ctx context.Context, input *dto.ReportThreadInput) (*dto.ReportThreadOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	report, err := h.communityService.ReportThread(ctx, accountID, usecase.ReportThreadInput{
		ThreadID: input.ID,
		Reason:   input.Body.Reason,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.ReportThreadOutput{Body: dto.ReportThreadResponseBody{ReportID: report.ID}}, nil
}

func (h *CommunityHandler) HandleReportPost(ctx context.Context, input *dto.ReportPostInput) (*dto.ReportPostOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	report, err := h.communityService.ReportPost(ctx, accountID, usecase.ReportPostInput{
		ThreadID: input.ID,
		PostID:   input.PostID,
		Reason:   input.Body.Reason,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.ReportPostOutput{Body: dto.ReportPostResponseBody{ReportID: report.ID}}, nil
}

func (h *CommunityHandler) HandleReportUser(ctx context.Context, input *dto.ReportUserInput) (*dto.ReportUserOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	report, err := h.communityService.ReportUser(ctx, accountID, usecase.ReportUserInput{
		ThreadID:          input.ID,
		ReportedAccountID: input.Body.ReportedAccountID,
		Reason:            input.Body.Reason,
	})
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.ReportUserOutput{Body: dto.ReportUserResponseBody{ReportID: report.ID}}, nil
}

func (h *CommunityHandler) HandleBlockUser(ctx context.Context, input *dto.BlockUserInput) (*dto.BlockUserOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.communityService.BlockUserInThread(ctx, usecase.BlockUserInput{
		ActorID:   accountID,
		ThreadID:  input.ThreadID,
		BlockedID: input.Body.BlockedID,
		Reason:    input.Body.Reason,
		IsAdmin:   false,
	}); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.BlockUserOutput{Body: dto.BlockUserResponseBody{Message: "User blocked"}}, nil
}

func (h *CommunityHandler) HandleUnblockUser(ctx context.Context, input *dto.UnblockUserInput) (*dto.UnblockUserOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.communityService.UnblockUserInThread(ctx, usecase.BlockUserInput{
		ActorID:   accountID,
		ThreadID:  input.ThreadID,
		BlockedID: input.AccountID,
		IsAdmin:   false,
	}); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UnblockUserOutput{Body: dto.BlockUserResponseBody{Message: "User unblocked"}}, nil
}

type UploadAttachmentsFormData struct {
	Files []huma.FormFile `form:"files" doc:"Attachment files" validate:"required,min=1,max=10"`
}

type UploadAttachmentsInput struct {
	RawBody huma.MultipartFormFiles[UploadAttachmentsFormData]
}

type UploadAttachmentsOutput struct {
	Body UploadAttachmentsResponseBody
}

type UploadAttachmentsResponseBody struct {
	Attachments []*dto.AttachmentDTO `json:"attachments"`
}

func (h *CommunityHandler) HandleUploadAttachments(ctx context.Context, input *UploadAttachmentsInput) (*UploadAttachmentsOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	formData := input.RawBody.Data()
	if formData == nil || len(formData.Files) == 0 {
		return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("community.errors.invalidInput"))
	}

	if len(formData.Files) > 10 {
		return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("community.errors.tooManyAttachments"))
	}

	inputs := make([]usecase.AttachmentUploadInput, 0, len(formData.Files))
	for _, f := range formData.Files {
		if !f.IsSet || f.File == nil {
			continue
		}

		limitedReader := io.LimitReader(f.File, int64(service.MaxCommunityAttachmentSize)+1)
		fileBytes, err := io.ReadAll(limitedReader)
		_ = f.Close()
		if err != nil {
			return nil, apperrors.ToHumaError(ctx, apperrors.InternalError("community.errors.readFileFailed", err))
		}

		if len(fileBytes) > service.MaxCommunityAttachmentSize {
			return nil, apperrors.ToHumaError(ctx, apperrors.PayloadTooLargeError("community.errors.fileTooLarge"))
		}

		inputs = append(inputs, usecase.AttachmentUploadInput{
			FileBytes: fileBytes,
			Filename:  f.Filename,
		})
	}

	attachments, err := h.attachmentUsecase.Upload(ctx, accountID, inputs)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	items := make([]*dto.AttachmentDTO, 0, len(attachments))
	for _, att := range attachments {
		items = append(items, &dto.AttachmentDTO{
			ID:       att.ID,
			FileURL:  att.FileURL,
			FileType: att.FileType,
			FileName: att.FileName,
			FileSize: att.FileSize,
		})
	}

	return &UploadAttachmentsOutput{Body: UploadAttachmentsResponseBody{Attachments: items}}, nil
}

type DeleteOrphanAttachmentInput struct {
	ID uuid.UUID `path:"id" doc:"Attachment ID"`
}

func (h *CommunityHandler) HandleDeleteOrphanAttachment(ctx context.Context, input *DeleteOrphanAttachmentInput) (*dto.DeletePostOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.attachmentUsecase.DeleteOrphan(ctx, input.ID, accountID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeletePostOutput{Body: dto.DeletePostResponseBody{Message: "Attachment deleted"}}, nil
}
