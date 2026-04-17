package handler

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/application/service"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/delivery/dto"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
)

type CommunityAdminHandler struct {
	categoryUsecase  usecase.CommunityCategoryUsecase
	communityService service.CommunityService
}

func NewCommunityAdminHandler(categoryUsecase usecase.CommunityCategoryUsecase, communityService service.CommunityService) *CommunityAdminHandler {
	return &CommunityAdminHandler{categoryUsecase: categoryUsecase, communityService: communityService}
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
