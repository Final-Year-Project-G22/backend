package handler

import (
	"context"
	"io"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type CommunityHandler struct {
	communityService service.CommunityService
	categoryUsecase  usecase.CommunityCategoryUsecase
	threadUsecase    usecase.DiscussionThreadUsecase
	postUsecase      usecase.DiscussionPostUsecase
}

func NewCommunityHandler(
	communityService service.CommunityService,
	categoryUsecase usecase.CommunityCategoryUsecase,
	threadUsecase usecase.DiscussionThreadUsecase,
	postUsecase usecase.DiscussionPostUsecase,
) *CommunityHandler {
	return &CommunityHandler{
		communityService: communityService,
		categoryUsecase:  categoryUsecase,
		threadUsecase:    threadUsecase,
		postUsecase:      postUsecase,
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

func (h *CommunityHandler) HandleListThreadsByCategory(ctx context.Context, input *dto.ListThreadsByCategoryInput) (*dto.ListThreadsByCategoryOutput, error) {
	opts := dto.ToQueryOptions(input.Page, input.PageSize)
	opts.Search = input.Search
	threads, err := h.threadUsecase.ListThreadsByCategory(ctx, input.CategoryID, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	items := make([]*dto.ThreadDTO, 0, len(threads))
	for _, thread := range threads {
		items = append(items, dto.ToThreadDTO(thread))
	}
	return &dto.ListThreadsByCategoryOutput{Body: dto.ListThreadsResponseBody{Threads: items}}, nil
}

func (h *CommunityHandler) HandleSearchThreads(ctx context.Context, input *dto.SearchThreadsInput) (*dto.SearchThreadsOutput, error) {
	opts := dto.ToQueryOptions(input.Page, input.PageSize)
	var categoryID *uuid.UUID
	if input.CategoryID != "" {
		parsed, err := uuid.Parse(input.CategoryID)
		if err != nil {
			return nil, apperrors.ToHumaError(ctx, apperrors.InvalidInputError("categoryId", "community.errors.invalidInput"))
		}
		categoryID = &parsed
	}
	threads, err := h.threadUsecase.SearchThreads(ctx, input.Keyword, categoryID, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	items := make([]*dto.ThreadDTO, 0, len(threads))
	for _, thread := range threads {
		items = append(items, dto.ToThreadDTO(thread))
	}
	return &dto.SearchThreadsOutput{Body: dto.ListThreadsResponseBody{Threads: items}}, nil
}

func (h *CommunityHandler) HandleGetThread(ctx context.Context, input *dto.GetThreadInput) (*dto.GetThreadOutput, error) {
	thread, err := h.threadUsecase.GetThread(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.GetThreadOutput{Body: dto.GetThreadResponseBody{Thread: dto.ToThreadDTO(thread)}}, nil
}

func (h *CommunityHandler) HandleListPosts(ctx context.Context, input *dto.ListPostsInput) (*dto.ListPostsOutput, error) {
	opts := dto.ToQueryOptions(input.Page, input.PageSize)
	posts, err := h.postUsecase.ListPostsByThread(ctx, input.ThreadID, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	items := make([]*dto.PostDTO, 0, len(posts))
	for _, post := range posts {
		items = append(items, dto.ToPostDTO(post))
	}
	return &dto.ListPostsOutput{Body: dto.ListPostsResponseBody{Posts: items}}, nil
}

func (h *CommunityHandler) HandleCreateThread(ctx context.Context, input *dto.CreateThreadInput) (*dto.CreateThreadOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	formData := input.RawBody.Data()
	if formData == nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("community.errors.invalidInput"))
	}

	categoryID, err := uuid.Parse(formData.CategoryID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.InvalidInputError("categoryId", "community.errors.invalidInput"))
	}

	var parentThreadID *uuid.UUID
	if formData.ParentThreadID != "" {
		parsed, err := uuid.Parse(formData.ParentThreadID)
		if err != nil {
			return nil, apperrors.ToHumaError(ctx, apperrors.InvalidInputError("parentThreadId", "community.errors.invalidInput"))
		}
		parentThreadID = &parsed
	}

	attachment, err := readOptionalAttachment(formData.File)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	thread, post, err := h.communityService.CreateThreadWithPost(ctx, accountID, dto.ToCreateThreadInput(
		categoryID,
		parentThreadID,
		formData.Title,
		formData.Slug,
		formData.Description,
		formData.InitialPostContent,
		nil,
		nil,
	), attachment)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateThreadOutput{Body: dto.CreateThreadResponseBody{ThreadID: thread.ID, PostID: post.ID}}, nil
}

func (h *CommunityHandler) HandleCreatePost(ctx context.Context, input *dto.CreatePostInput) (*dto.CreatePostOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	formData := input.RawBody.Data()
	if formData == nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("community.errors.invalidInput"))
	}

	attachment, err := readOptionalAttachment(formData.File)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	post, err := h.communityService.ReplyToThread(ctx, accountID, input.ThreadID, nil, dto.ToCreatePostInput(formData.Content, nil, nil), attachment)
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

	attachment, err := readOptionalAttachment(formData.File)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	post, err := h.communityService.ReplyToThread(ctx, accountID, input.ThreadID, &input.PostID, dto.ToCreatePostInput(formData.Content, nil, nil), attachment)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.ReplyPostOutput{Body: dto.CreatePostResponseBody{PostID: post.ID}}, nil
}

func readOptionalAttachment(file huma.FormFile) (*service.AttachmentUploadInput, error) {
	if !file.IsSet {
		return nil, nil
	}

	f := file.File
	if f == nil {
		return nil, apperrors.BadRequestError("community.errors.invalidFile")
	}
	defer func() { _ = f.Close() }()

	limitedReader := io.LimitReader(f, int64(service.MaxCommunityAttachmentSize)+1)
	fileBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, apperrors.InternalError("community.errors.readFileFailed", err)
	}

	if len(fileBytes) > service.MaxCommunityAttachmentSize {
		return nil, apperrors.PayloadTooLargeError("community.errors.fileTooLarge")
	}

	return &service.AttachmentUploadInput{FileBytes: fileBytes, Filename: file.Filename}, nil
}

func (h *CommunityHandler) HandleUpdatePost(ctx context.Context, input *dto.UpdatePostInput) (*dto.UpdatePostOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	formData := input.RawBody.Data()
	if formData == nil {
		return nil, apperrors.ToHumaError(ctx, apperrors.BadRequestError("community.errors.invalidInput"))
	}

	attachment, err := readOptionalAttachment(formData.File)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	_, err = h.communityService.UpdatePost(ctx, accountID, input.PostID, dto.ToUpdatePostInput(formData.Content, formData.RemoveAttachment), attachment)
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
	items := make([]*dto.ThreadDTO, 0, len(settings))
	for _, setting := range settings {
		items = append(items, dto.ToThreadDTO(&setting.Thread))
	}
	return &dto.ListFollowedThreadsOutput{Body: dto.ListFollowedThreadsResponseBody{Threads: items}}, nil
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
