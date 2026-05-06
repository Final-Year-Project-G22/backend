package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	iamusecase "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type CommunityAdminHandler struct {
	categoryUsecase  usecase.CommunityCategoryUsecase
	communityService service.CommunityService
	reportUsecase    usecase.ContentReportUsecase
	threadUsecase    usecase.DiscussionThreadUsecase
	postUsecase      usecase.DiscussionPostUsecase
	blockUsecase     usecase.ThreadBlockUsecase
	accountUsecase   iamusecase.AccountUsecase
	userUsecase      iamusecase.UserUsecase
	logger           core.Logger
}

func NewCommunityAdminHandler(
	categoryUsecase usecase.CommunityCategoryUsecase,
	communityService service.CommunityService,
	reportUsecase usecase.ContentReportUsecase,
	threadUsecase usecase.DiscussionThreadUsecase,
	postUsecase usecase.DiscussionPostUsecase,
	blockUsecase usecase.ThreadBlockUsecase,
	accountUsecase iamusecase.AccountUsecase,
	userUsecase iamusecase.UserUsecase,
	logger core.Logger,
) *CommunityAdminHandler {
	return &CommunityAdminHandler{
		categoryUsecase:  categoryUsecase,
		communityService: communityService,
		reportUsecase:    reportUsecase,
		threadUsecase:    threadUsecase,
		postUsecase:      postUsecase,
		blockUsecase:     blockUsecase,
		accountUsecase:   accountUsecase,
		userUsecase:      userUsecase,
		logger:           logger,
	}
}

func (h *CommunityAdminHandler) HandleCreateCategory(ctx context.Context, input *dto.CreateCommunityCategoryInput) (*dto.CreateCommunityCategoryOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	category, err := h.categoryUsecase.CreateCategory(ctx, accountID, dto.ToCreateCategoryInput(input.Body))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.CreateCommunityCategoryOutput{Body: dto.CreateCommunityCategoryResponseBody{ID: category.ID}}, nil
}

func (h *CommunityAdminHandler) HandleUpdateCategory(ctx context.Context, input *dto.UpdateCommunityCategoryInput) (*dto.UpdateCommunityCategoryOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	_, err := h.categoryUsecase.UpdateCategory(ctx, accountID, input.ID, dto.ToUpdateCategoryInput(input.Body))
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UpdateCommunityCategoryOutput{Body: dto.UpdateCommunityCategoryResponseBody{Message: "Category updated"}}, nil
}

func (h *CommunityAdminHandler) HandleDeleteCategory(ctx context.Context, input *dto.DeleteCommunityCategoryInput) (*dto.DeleteCommunityCategoryOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.categoryUsecase.DeleteCategory(ctx, accountID, input.ID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.DeleteCommunityCategoryOutput{Body: dto.DeleteCommunityCategoryResponseBody{Message: "Category deleted"}}, nil
}

func (h *CommunityAdminHandler) HandleListCategories(ctx context.Context, input *dto.AdminListCommunityCategoriesInput) (*dto.AdminListCommunityCategoriesOutput, error) {
	opts := dto.ToAdminQueryOptions(input.Page, input.PageSize)
	categories, err := h.categoryUsecase.ListCategories(ctx, input.IncludeInactive, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	items := make([]*dto.CategoryDTO, 0, len(categories))
	for _, category := range categories {
		items = append(items, dto.ToCategoryDTO(category))
	}
	return &dto.AdminListCommunityCategoriesOutput{Body: dto.ListCategoriesResponseBody{Categories: items}}, nil
}

func (h *CommunityAdminHandler) HandleBlockUser(ctx context.Context, input *dto.BlockUserInput) (*dto.BlockUserOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.communityService.BlockUserInThread(ctx, usecase.BlockUserInput{
		ActorID:   accountID,
		ThreadID:  input.ThreadID,
		BlockedID: input.Body.BlockedID,
		Reason:    input.Body.Reason,
		IsAdmin:   true,
	}); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	if err := h.reportUsecase.ResolveUserReport(ctx, input.ThreadID, input.Body.BlockedID, accountID); err != nil {
		h.logger.Error("Failed to resolve user report", core.Error(err))
	}

	return &dto.BlockUserOutput{Body: dto.BlockUserResponseBody{Message: "User blocked"}}, nil
}

func (h *CommunityAdminHandler) HandleUnblockUser(ctx context.Context, input *dto.UnblockUserInput) (*dto.UnblockUserOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))
	if err := h.communityService.UnblockUserInThread(ctx, usecase.BlockUserInput{
		ActorID:   accountID,
		ThreadID:  input.ThreadID,
		BlockedID: input.AccountID,
		IsAdmin:   true,
	}); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}
	return &dto.UnblockUserOutput{Body: dto.BlockUserResponseBody{Message: "User unblocked"}}, nil
}

func (h *CommunityAdminHandler) HandleListBlockedUsers(ctx context.Context, input *dto.ListBlockedUsersInput) (*dto.ListBlockedUsersOutput, error) {
	page := input.Page
	if page == 0 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	opts := dto.ToAdminQueryOptions(page, pageSize)

	blockedUsers, total, err := h.blockUsecase.ListBlockedUsers(ctx, input.ID, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	items := make([]*dto.ThreadBlockedUserDTO, 0, len(blockedUsers))
	for _, b := range blockedUsers {
		items = append(items, dto.ToThreadBlockedUserDTO(b))
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.ListBlockedUsersOutput{Body: dto.ListBlockedUsersResponseBody{
		BlockedUsers: items,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   totalPages,
	}}, nil
}

func (h *CommunityAdminHandler) HandleListAllBlockedUsers(ctx context.Context, input *dto.ListAllBlockedUsersInput) (*dto.ListAllBlockedUsersOutput, error) {
	page := input.Page
	if page == 0 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	opts := dto.ToAdminQueryOptions(page, pageSize)

	blockedUsers, total, err := h.blockUsecase.ListAllBlockedUsers(ctx, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	items := make([]*dto.ThreadBlockedUserDTO, 0, len(blockedUsers))
	for _, b := range blockedUsers {
		blockDTO := dto.ToThreadBlockedUserDTO(b)

		if thread, err := h.threadUsecase.GetThread(ctx, b.ThreadID); err == nil {
			blockDTO.ThreadTitle = thread.Title
			blockDTO.ThreadSlug = thread.Slug
		}

		if blockedAccount, err := h.accountUsecase.GetAccount(ctx, b.BlockedAccountID); err == nil {
			if blockedUser, err := h.userUsecase.GetUser(ctx, blockedAccount.UserID); err == nil {
				blockDTO.BlockedUserFirstName = blockedUser.FirstName
				blockDTO.BlockedUserLastName = blockedUser.LastName
			}
			blockDTO.BlockedUserEmail = blockedAccount.Email
		}

		if blockedByAccount, err := h.accountUsecase.GetAccount(ctx, b.BlockedByAccountID); err == nil {
			if blockedByUser, err := h.userUsecase.GetUser(ctx, blockedByAccount.UserID); err == nil {
				blockDTO.BlockedByFirstName = blockedByUser.FirstName
				blockDTO.BlockedByLastName = blockedByUser.LastName
			}
		}

		items = append(items, blockDTO)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.ListAllBlockedUsersOutput{Body: dto.ListAllBlockedUsersResponseBody{
		BlockedUsers: items,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
		TotalPages:   totalPages,
	}}, nil
}

func (h *CommunityAdminHandler) HandleListThreadReports(ctx context.Context, input *dto.ListThreadReportsInput) (*dto.ListThreadReportsOutput, error) {
	page := input.Page
	if page == 0 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	opts := dto.ToAdminQueryOptions(page, pageSize, input.Search)

	var status *entity.ReportStatus
	if input.Status != "" {
		s := entity.ReportStatus(input.Status)
		status = &s
	}

	reports, total, err := h.reportUsecase.ListThreadReports(ctx, status, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	items := make([]*dto.ReportWithContentDTO, 0, len(reports))
	for _, r := range reports {
		reporterFirstName, reporterLastName := "", ""
		account, err := h.accountUsecase.GetAccount(ctx, r.ReporterAccountID)
		if err == nil && account != nil {
			user, err := h.userUsecase.GetUser(ctx, account.UserID)
			if err == nil && user != nil {
				reporterFirstName = user.FirstName
				reporterLastName = user.LastName
			}
		}

		reportDTO := dto.ToReportDTO(r, reporterFirstName, reporterLastName)
		content := &dto.ContentDTO{}

		if r.ThreadID != nil {
			if thread, err := h.threadUsecase.GetThread(ctx, *r.ThreadID); err == nil {
				authorFirstName, authorLastName := "", ""
				if threadAccount, err := h.accountUsecase.GetAccount(ctx, thread.AuthorAccountID); err == nil && threadAccount != nil {
					if threadUser, err := h.userUsecase.GetUser(ctx, threadAccount.UserID); err == nil && threadUser != nil {
						authorFirstName = threadUser.FirstName
						authorLastName = threadUser.LastName
					}
				}
				var createdAt, updatedAt string
				if thread.CreatedAt != nil {
					createdAt = thread.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
				}
				if thread.UpdatedAt != nil {
					updatedAt = thread.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
				}
				content.Thread = &dto.ThreadContentDTO{
					ID:              thread.ID,
					Title:           thread.Title,
					Slug:            thread.Slug,
					Description:     thread.Description,
					SectorIDs:       thread.SectorIDs,
					TagIDs:          thread.TagIDs,
					AuthorAccountID: thread.AuthorAccountID,
					AuthorFirstName: authorFirstName,
					AuthorLastName:  authorLastName,
					Status:          string(thread.Status),
					ReplyCount:      thread.ReplyCount,
					ViewCount:       thread.ViewCount,
					CreatedAt:       createdAt,
					UpdatedAt:       updatedAt,
				}
			}
		}

		items = append(items, &dto.ReportWithContentDTO{
			Report:  reportDTO,
			Content: content,
		})
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.ListThreadReportsOutput{Body: dto.ListThreadReportsResponseBody{
		Reports:    items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}}, nil
}

func (h *CommunityAdminHandler) HandleListPostReports(ctx context.Context, input *dto.ListPostReportsInput) (*dto.ListPostReportsOutput, error) {
	page := input.Page
	if page == 0 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	opts := dto.ToAdminQueryOptions(page, pageSize, input.Search)

	var status *entity.ReportStatus
	if input.Status != "" {
		s := entity.ReportStatus(input.Status)
		status = &s
	}

	reports, total, err := h.reportUsecase.ListPostReports(ctx, status, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	items := make([]*dto.ReportWithContentDTO, 0, len(reports))
	for _, r := range reports {
		reporterFirstName, reporterLastName := "", ""
		account, err := h.accountUsecase.GetAccount(ctx, r.ReporterAccountID)
		if err == nil && account != nil {
			user, err := h.userUsecase.GetUser(ctx, account.UserID)
			if err == nil && user != nil {
				reporterFirstName = user.FirstName
				reporterLastName = user.LastName
			}
		}

		reportDTO := dto.ToReportDTO(r, reporterFirstName, reporterLastName)
		content := &dto.ContentDTO{}

		if r.PostID != nil {
			if post, err := h.postUsecase.GetPost(ctx, *r.PostID); err == nil {
				authorFirstName, authorLastName := "", ""
				if postAccount, err := h.accountUsecase.GetAccount(ctx, post.AuthorAccountID); err == nil && postAccount != nil {
					if postUser, err := h.userUsecase.GetUser(ctx, postAccount.UserID); err == nil && postUser != nil {
						authorFirstName = postUser.FirstName
						authorLastName = postUser.LastName
					}
				}
				threadTitle, threadSlug := "", ""
				if thread, err := h.threadUsecase.GetThread(ctx, post.ThreadID); err == nil {
					threadTitle = thread.Title
					threadSlug = thread.Slug
				}
				var createdAt, updatedAt string
				if post.CreatedAt != nil {
					createdAt = post.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
				}
				if post.UpdatedAt != nil {
					updatedAt = post.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
				}
				content.Post = &dto.PostContentDTO{
					ID:              post.ID,
					ThreadID:        post.ThreadID,
					ThreadTitle:     threadTitle,
					ThreadSlug:      threadSlug,
					ParentPostID:    post.ParentPostID,
					AuthorAccountID: post.AuthorAccountID,
					AuthorFirstName: authorFirstName,
					AuthorLastName:  authorLastName,
					Content:         post.Content,
					IsSolution:      post.IsSolution,
					CreatedAt:       createdAt,
					UpdatedAt:       updatedAt,
				}
			}
		}

		items = append(items, &dto.ReportWithContentDTO{
			Report:  reportDTO,
			Content: content,
		})
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.ListPostReportsOutput{Body: dto.ListPostReportsResponseBody{
		Reports:    items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}}, nil
}

func (h *CommunityAdminHandler) HandleListUserReports(ctx context.Context, input *dto.ListUserReportsInput) (*dto.ListUserReportsOutput, error) {
	page := input.Page
	if page == 0 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	opts := dto.ToAdminQueryOptions(page, pageSize, input.Search)

	var status *entity.ReportStatus
	if input.Status != "" {
		s := entity.ReportStatus(input.Status)
		status = &s
	}

	reports, total, err := h.reportUsecase.ListUserReports(ctx, status, opts)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	items := make([]*dto.ReportWithContentDTO, 0, len(reports))
	for _, r := range reports {
		reporterFirstName, reporterLastName := "", ""
		account, err := h.accountUsecase.GetAccount(ctx, r.ReporterAccountID)
		if err == nil && account != nil {
			user, err := h.userUsecase.GetUser(ctx, account.UserID)
			if err == nil && user != nil {
				reporterFirstName = user.FirstName
				reporterLastName = user.LastName
			}
		}

		reportDTO := dto.ToReportDTO(r, reporterFirstName, reporterLastName)
		content := &dto.ContentDTO{}

		if r.ReportedAccountID != nil {
			if targetAccount, err := h.accountUsecase.GetAccount(ctx, *r.ReportedAccountID); err == nil && targetAccount != nil {
				targetUser, err := h.userUsecase.GetUser(ctx, targetAccount.UserID)
				if err == nil && targetUser != nil {
					threadTitle := ""
					if r.ThreadID != nil {
						if thread, err := h.threadUsecase.GetThread(ctx, *r.ThreadID); err == nil {
							threadTitle = thread.Title
						}
					}
					content.User = &dto.UserContentDTO{
						ID:          targetUser.ID,
						FirstName:   targetUser.FirstName,
						LastName:    targetUser.LastName,
						Email:       targetAccount.Email,
						ThreadID:    r.ThreadID,
						ThreadTitle: threadTitle,
					}
				}
			}
		}

		items = append(items, &dto.ReportWithContentDTO{
			Report:  reportDTO,
			Content: content,
		})
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.ListUserReportsOutput{Body: dto.ListUserReportsResponseBody{
		Reports:    items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}}, nil
}

func (h *CommunityAdminHandler) HandleGetThreadReport(ctx context.Context, input *dto.GetThreadReportInput) (*dto.GetThreadReportOutput, error) {
	report, err := h.reportUsecase.GetThreadReport(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	reporterFirstName, reporterLastName := "", ""
	account, err := h.accountUsecase.GetAccount(ctx, report.ReporterAccountID)
	if err == nil && account != nil {
		user, err := h.userUsecase.GetUser(ctx, account.UserID)
		if err == nil && user != nil {
			reporterFirstName = user.FirstName
			reporterLastName = user.LastName
		}
	}

	reportDTO := dto.ToReportDTO(report, reporterFirstName, reporterLastName)
	content := &dto.ContentDTO{}

	if report.ThreadID != nil {
		if thread, err := h.threadUsecase.GetThread(ctx, *report.ThreadID); err == nil {
			authorFirstName, authorLastName := "", ""
			if threadAccount, err := h.accountUsecase.GetAccount(ctx, thread.AuthorAccountID); err == nil && threadAccount != nil {
				if threadUser, err := h.userUsecase.GetUser(ctx, threadAccount.UserID); err == nil && threadUser != nil {
					authorFirstName = threadUser.FirstName
					authorLastName = threadUser.LastName
				}
			}
			var createdAt, updatedAt string
			if thread.CreatedAt != nil {
				createdAt = thread.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
			}
			if thread.UpdatedAt != nil {
				updatedAt = thread.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
			}
			content.Thread = &dto.ThreadContentDTO{
				ID:              thread.ID,
				Title:           thread.Title,
				Slug:            thread.Slug,
				Description:     thread.Description,
				SectorIDs:       thread.SectorIDs,
				TagIDs:          thread.TagIDs,
				AuthorAccountID: thread.AuthorAccountID,
				AuthorFirstName: authorFirstName,
				AuthorLastName:  authorLastName,
				Status:          string(thread.Status),
				ReplyCount:      thread.ReplyCount,
				ViewCount:       thread.ViewCount,
				CreatedAt:       createdAt,
				UpdatedAt:       updatedAt,
			}
		}
	}

	return &dto.GetThreadReportOutput{Body: dto.GetThreadReportResponseBody{
		Report:  reportDTO,
		Content: content,
	}}, nil
}

func (h *CommunityAdminHandler) HandleGetPostReport(ctx context.Context, input *dto.GetPostReportInput) (*dto.GetPostReportOutput, error) {
	report, err := h.reportUsecase.GetPostReport(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	reporterFirstName, reporterLastName := "", ""
	account, err := h.accountUsecase.GetAccount(ctx, report.ReporterAccountID)
	if err == nil && account != nil {
		user, err := h.userUsecase.GetUser(ctx, account.UserID)
		if err == nil && user != nil {
			reporterFirstName = user.FirstName
			reporterLastName = user.LastName
		}
	}

	reportDTO := dto.ToReportDTO(report, reporterFirstName, reporterLastName)
	content := &dto.ContentDTO{}

	if report.PostID != nil {
		if post, err := h.postUsecase.GetPost(ctx, *report.PostID); err == nil {
			authorFirstName, authorLastName := "", ""
			if postAccount, err := h.accountUsecase.GetAccount(ctx, post.AuthorAccountID); err == nil && postAccount != nil {
				if postUser, err := h.userUsecase.GetUser(ctx, postAccount.UserID); err == nil && postUser != nil {
					authorFirstName = postUser.FirstName
					authorLastName = postUser.LastName
				}
			}
			threadTitle, threadSlug := "", ""
			if thread, err := h.threadUsecase.GetThread(ctx, post.ThreadID); err == nil {
				threadTitle = thread.Title
				threadSlug = thread.Slug
			}
			var createdAt, updatedAt string
			if post.CreatedAt != nil {
				createdAt = post.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
			}
			if post.UpdatedAt != nil {
				updatedAt = post.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
			}
			content.Post = &dto.PostContentDTO{
				ID:              post.ID,
				ThreadID:        post.ThreadID,
				ThreadTitle:     threadTitle,
				ThreadSlug:      threadSlug,
				ParentPostID:    post.ParentPostID,
				AuthorAccountID: post.AuthorAccountID,
				AuthorFirstName: authorFirstName,
				AuthorLastName:  authorLastName,
				Content:         post.Content,
				IsSolution:      post.IsSolution,
				CreatedAt:       createdAt,
				UpdatedAt:       updatedAt,
			}
		}
	}

	return &dto.GetPostReportOutput{Body: dto.GetPostReportResponseBody{
		Report:  reportDTO,
		Content: content,
	}}, nil
}

func (h *CommunityAdminHandler) HandleGetUserReport(ctx context.Context, input *dto.GetUserReportInput) (*dto.GetUserReportOutput, error) {
	report, err := h.reportUsecase.GetUserReport(ctx, input.ID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	reporterFirstName, reporterLastName := "", ""
	account, err := h.accountUsecase.GetAccount(ctx, report.ReporterAccountID)
	if err == nil && account != nil {
		user, err := h.userUsecase.GetUser(ctx, account.UserID)
		if err == nil && user != nil {
			reporterFirstName = user.FirstName
			reporterLastName = user.LastName
		}
	}

	reportDTO := dto.ToReportDTO(report, reporterFirstName, reporterLastName)
	content := &dto.ContentDTO{}

	if report.ReportedAccountID != nil {
		if targetAccount, err := h.accountUsecase.GetAccount(ctx, *report.ReportedAccountID); err == nil && targetAccount != nil {
			targetUser, err := h.userUsecase.GetUser(ctx, targetAccount.UserID)
			if err == nil && targetUser != nil {
				threadTitle := ""
				if report.ThreadID != nil {
					if thread, err := h.threadUsecase.GetThread(ctx, *report.ThreadID); err == nil {
						threadTitle = thread.Title
					}
				}
				content.User = &dto.UserContentDTO{
					ID:          targetUser.ID,
					FirstName:   targetUser.FirstName,
					LastName:    targetUser.LastName,
					Email:       targetAccount.Email,
					ThreadID:    report.ThreadID,
					ThreadTitle: threadTitle,
				}
			}
		}
	}

	return &dto.GetUserReportOutput{Body: dto.GetUserReportResponseBody{
		Report:  reportDTO,
		Content: content,
	}}, nil
}

func (h *CommunityAdminHandler) HandleUpdateThreadReportStatus(ctx context.Context, input *dto.UpdateThreadReportStatusInput) (*dto.UpdateThreadReportStatusOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))

	inputStatus := entity.ReportStatus(input.Body.Status)
	updated, err := h.reportUsecase.UpdateReportStatus(ctx, input.ID, usecase.UpdateReportStatusInput{
		Status:    inputStatus,
		AdminNote: input.Body.AdminNote,
	}, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.UpdateThreadReportStatusOutput{Body: dto.UpdateThreadReportStatusResponseBody{Report: dto.ToReportDTO(updated, "", "")}}, nil
}

func (h *CommunityAdminHandler) HandleUpdatePostReportStatus(ctx context.Context, input *dto.UpdatePostReportStatusInput) (*dto.UpdatePostReportStatusOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))

	inputStatus := entity.ReportStatus(input.Body.Status)
	updated, err := h.reportUsecase.UpdateReportStatus(ctx, input.ID, usecase.UpdateReportStatusInput{
		Status:    inputStatus,
		AdminNote: input.Body.AdminNote,
	}, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.UpdatePostReportStatusOutput{Body: dto.UpdatePostReportStatusResponseBody{Report: dto.ToReportDTO(updated, "", "")}}, nil
}

func (h *CommunityAdminHandler) HandleUpdateUserReportStatus(ctx context.Context, input *dto.UpdateUserReportStatusInput) (*dto.UpdateUserReportStatusOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))

	inputStatus := entity.ReportStatus(input.Body.Status)
	updated, err := h.reportUsecase.UpdateReportStatus(ctx, input.ID, usecase.UpdateReportStatusInput{
		Status:    inputStatus,
		AdminNote: input.Body.AdminNote,
	}, accountID)
	if err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.UpdateUserReportStatusOutput{Body: dto.UpdateUserReportStatusResponseBody{Report: dto.ToReportDTO(updated, "", "")}}, nil
}

func (h *CommunityAdminHandler) HandleDeleteReportedThread(ctx context.Context, input *dto.DeleteReportedThreadInput) (*dto.DeleteReportedThreadOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))

	if err := h.reportUsecase.DeleteReportedThread(ctx, input.ID, accountID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.DeleteReportedThreadOutput{Body: dto.DeleteReportedContentResponseBody{
		Success: true,
		Message: "Thread deleted and report resolved",
	}}, nil
}

func (h *CommunityAdminHandler) HandleDeleteReportedPost(ctx context.Context, input *dto.DeleteReportedPostInput) (*dto.DeleteReportedPostOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))

	if err := h.reportUsecase.DeleteReportedPost(ctx, input.ID, accountID); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.DeleteReportedPostOutput{Body: dto.DeleteReportedContentResponseBody{
		Success: true,
		Message: "Post deleted and report resolved",
	}}, nil
}

func (h *CommunityAdminHandler) HandleBlockReportedUser(ctx context.Context, input *dto.BlockReportedUserInput) (*dto.BlockReportedUserOutput, error) {
	accountID := contextkeys.GetAccountID(ctx.Value(contextkeys.AccountID))

	var reason *string
	if input.Body.Reason != "" {
		r := input.Body.Reason
		reason = &r
	}

	if err := h.reportUsecase.BlockReportedUser(ctx, input.ID, accountID, reason); err != nil {
		return nil, apperrors.ToHumaError(ctx, err)
	}

	return &dto.BlockReportedUserOutput{Body: dto.BlockReportedUserResponseBody{
		Success: true,
		Message: "User blocked and report resolved",
	}}, nil
}
