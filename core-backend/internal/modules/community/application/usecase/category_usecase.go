package usecase

import (
	"context"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/entity"
	communityerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/community/domain/usecase"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	apperrors "github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
)

type communityCategoryUsecase struct {
	catRepo    repository.CommunityCategoryRepository
	transactor sharedrepo.Transactor
}

func NewCommunityCategoryUsecase(
	catRepo repository.CommunityCategoryRepository,
	transactor sharedrepo.Transactor,
) usecase.CommunityCategoryUsecase {
	return &communityCategoryUsecase{
		catRepo:    catRepo,
		transactor: transactor,
	}
}

func (u *communityCategoryUsecase) CreateCategory(ctx context.Context, actorID uuid.UUID, input usecase.CreateCategoryInput) (*entity.CommunityCategory, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.TrimSpace(input.Slug)
	if name == "" {
		return nil, apperrors.RequiredFieldError("name")
	}
	if slug == "" {
		return nil, apperrors.RequiredFieldError("slug")
	}

	if input.ParentCategoryID != nil {
		if _, err := u.catRepo.GetByID(ctx, *input.ParentCategoryID); err != nil {
			return nil, err
		}
	}

	exists, err := u.catRepo.ExistsBySlug(ctx, input.ParentCategoryID, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.AlreadyExistsError("category", "slug", slug)
	}

	category := &entity.CommunityCategory{
		Name:             name,
		Slug:             slug,
		Description:      input.Description,
		ParentCategoryID: input.ParentCategoryID,
		IsActive:         input.IsActive,
	}
	if err := u.catRepo.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (u *communityCategoryUsecase) UpdateCategory(ctx context.Context, actorID, id uuid.UUID, input usecase.UpdateCategoryInput) (*entity.CommunityCategory, error) {
	var updated *entity.CommunityCategory
	if err := u.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		category, err := u.catRepo.GetByID(txCtx, id)
		if err != nil {
			return err
		}
		if input.Name != nil {
			name := strings.TrimSpace(*input.Name)
			if name == "" {
				return apperrors.RequiredFieldError("name")
			}
			category.Name = name
		}
		if input.Slug != nil {
			slug := strings.TrimSpace(*input.Slug)
			if slug == "" {
				return apperrors.RequiredFieldError("slug")
			}
			parentID := category.ParentCategoryID
			if input.ParentCategoryID != nil {
				parentID = input.ParentCategoryID
			}
			existing, err := u.catRepo.GetBySlug(txCtx, parentID, slug)
			if err != nil && err != communityerror.ErrCategoryNotFound {
				return err
			}
			if existing != nil && existing.ID != id {
				return apperrors.AlreadyExistsError("category", "slug", slug)
			}
			category.Slug = slug
		}
		if input.Description != nil {
			category.Description = input.Description
		}
		if input.ParentCategoryID != nil {
			if _, err := u.catRepo.GetByID(txCtx, *input.ParentCategoryID); err != nil {
				return err
			}
			category.ParentCategoryID = input.ParentCategoryID
		}
		if input.IsActive != nil {
			category.IsActive = *input.IsActive
		}
		if err := u.catRepo.Update(txCtx, category); err != nil {
			return err
		}
		updated = category
		return nil
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

func (u *communityCategoryUsecase) DeleteCategory(ctx context.Context, actorID, id uuid.UUID) error {
	return u.catRepo.Delete(ctx, id)
}

func (u *communityCategoryUsecase) ListCategories(ctx context.Context, includeInactive bool, q query.QueryOptions) ([]*entity.CommunityCategory, error) {
	if includeInactive {
		return u.catRepo.ListTree(ctx, true)
	}
	return u.catRepo.ListActive(ctx, q)
}

func (u *communityCategoryUsecase) GetCategory(ctx context.Context, id uuid.UUID) (*entity.CommunityCategory, error) {
	return u.catRepo.GetByID(ctx, id)
}
