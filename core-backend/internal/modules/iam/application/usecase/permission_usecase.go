package appusecase

import (
	"context"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
)

type permissionUsecase struct {
	permissionRepo repository.PermissionRepository
	logger         core.Logger
}

func NewPermissionUsecase(
	permissionRepo repository.PermissionRepository,
	logger core.Logger,
) usecase.PermissionUsecase {
	return &permissionUsecase{
		permissionRepo: permissionRepo,
		logger:         logger,
	}
}

func (u *permissionUsecase) GetPermissionByCode(ctx context.Context, code string) (*entity.Permission, error) {
	permission, err := u.permissionRepo.GetByCode(ctx, code)
	if err == iamerror.ErrPermissionNotFound {
		return nil, errors.NotFoundErrorWithKey("iam.errors.notFound")
	}
	if err != nil {
		return nil, err
	}
	return permission, nil
}

func (u *permissionUsecase) ListPermissionsByRole(ctx context.Context, roleID uuid.UUID) ([]*entity.Permission, error) {
	return u.permissionRepo.ListByRoleID(ctx, roleID)
}
