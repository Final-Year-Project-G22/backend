package service

import (
	"context"
	"testing"
	"time"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/usecase"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/model"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/permissions"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type listAdminsAccountUsecaseStub struct {
	usecase.AccountUsecase
	gotCodes []string
	accounts []*entity.Account
}

func (s *listAdminsAccountUsecaseStub) ListAdmins(_ context.Context, permissionCodes []string, _ map[string]interface{}) ([]*entity.Account, int64, error) {
	s.gotCodes = permissionCodes
	return s.accounts, int64(len(s.accounts)), nil
}

type listAdminsUserUsecaseStub struct {
	usecase.UserUsecase
	user *entity.User
}

func (s *listAdminsUserUsecaseStub) GetUser(_ context.Context, _ uuid.UUID) (*entity.User, error) {
	return s.user, nil
}

type listAdminsRoleAssignmentRepoStub struct {
	repository.RoleAssignmentRepository
}

func (s *listAdminsRoleAssignmentRepoStub) ListByAccountID(_ context.Context, _ uuid.UUID) ([]*entity.RoleAssignment, error) {
	return nil, nil
}

type listAdminsRoleRepoStub struct {
	repository.RoleRepository
}

func (s *listAdminsRoleRepoStub) GetByID(_ context.Context, _ uuid.UUID) (*entity.Role, error) {
	return nil, nil
}

func TestListAdmins_PermissionGateIncludesNonListIAMAdminPermissions(t *testing.T) {
	now := time.Now()
	account := &entity.Account{
		BaseModel: model.BaseModel{
			ID:        uuid.New(),
			CreatedAt: &now,
		},
		UserID: uuid.New(),
		Email:  "yohannes.solomon.210@gmail.com",
		Status: entity.AccountStatusActive,
	}

	accountUC := &listAdminsAccountUsecaseStub{accounts: []*entity.Account{account}}
	userUC := &listAdminsUserUsecaseStub{user: &entity.User{FirstName: "Yohannes", LastName: "Solomon"}}

	svc := NewAdminService(
		nil,
		userUC,
		accountUC,
		&listAdminsRoleRepoStub{},
		&listAdminsRoleAssignmentRepoStub{},
		nil, nil, nil, nil, nil,
	)

	out, err := svc.ListAdmins(context.Background(), ListAdminsInput{})
	assert.NoError(t, err)
	if assert.NotNil(t, out) && assert.Len(t, out.Admins, 1) {
		assert.Equal(t, account.Email, out.Admins[0].Email)
		assert.Equal(t, "Yohannes", out.Admins[0].FirstName)
	}

	// The permission gate must include every iam.admin.* capability so that
	// admins registered via the hub with roles such as ai_content_manager
	// (which carries iam.admin.read) are listed — not just accounts whose
	// roles happen to carry iam.admin.list.
	for _, code := range []string{
		permissions.AdminList,
		permissions.AdminRead,
		permissions.AdminCreate,
		permissions.AdminRolesUpdate,
		permissions.AdminResetPassword,
		permissions.AdminStatusUpdate,
		permissions.RoleRead,
	} {
		assert.Contains(t, accountUC.gotCodes, code, "missing permission code %s", code)
	}
}
