package repository

import (
	"context"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/entity"
	iamerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/error"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/domain/repository"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type accountRepository struct {
	sharedrepo.GenericRepository[entity.Account]
	db     *core.Database
	logger core.Logger
}

// NewAccountRepository creates a new AccountRepository implementation.
func NewAccountRepository(db *core.Database, logger core.Logger) repository.AccountRepository {
	base := sharedrepo.NewBaseRepository[entity.Account](db, logger)
	return &accountRepository{
		GenericRepository: base,
		db:                db,
		logger:            logger,
	}
}

// getDB returns the appropriate *gorm.DB for the context (tx-aware).
func (r *accountRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *accountRepository) GetByEmailNormalized(ctx context.Context, email string) (*entity.Account, error) {
	var account entity.Account
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	err := r.getDB(ctx).
		Where("email_normalized = ?", normalizedEmail).
		First(&account).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iamerror.ErrAccountNotFound
		}
		r.logger.Error("Failed to get account by email", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return &account, nil
}

func (r *accountRepository) GetByUsernameNormalized(ctx context.Context, username string) (*entity.Account, error) {
	var account entity.Account
	normalizedUsername := strings.ToLower(strings.TrimSpace(username))

	err := r.getDB(ctx).
		Where("username_normalized = ?", normalizedUsername).
		First(&account).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, iamerror.ErrAccountNotFound
		}
		r.logger.Error("Failed to get account by username", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return &account, nil
}

func (r *accountRepository) GetByEmailOrUsername(ctx context.Context, identifier string) (*entity.Account, error) {
	normalized := strings.ToLower(strings.TrimSpace(identifier))

	account, err := r.GetByEmailNormalized(ctx, normalized)
	if err == nil {
		return account, nil
	}
	if err != iamerror.ErrAccountNotFound {
		return nil, err
	}

	return r.GetByUsernameNormalized(ctx, normalized)
}

func (r *accountRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Account, error) {
	var accounts []*entity.Account

	err := r.getDB(ctx).
		Where("user_id = ?", userID).
		Find(&accounts).Error

	if err != nil {
		r.logger.Error("Failed to list accounts by user ID", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	return accounts, nil
}

func (r *accountRepository) ExistsByEmailNormalized(ctx context.Context, email string) (bool, error) {
	var count int64
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	err := r.getDB(ctx).
		Model(&entity.Account{}).
		Where("email_normalized = ?", normalizedEmail).
		Count(&count).Error

	if err != nil {
		r.logger.Error("Failed to check account existence by email", core.Error(err))
		return false, errors.InternalError("errors.databaseError", err)
	}

	return count > 0, nil
}

func (r *accountRepository) ExistsByUsernameNormalized(ctx context.Context, username string) (bool, error) {
	var count int64
	normalizedUsername := strings.ToLower(strings.TrimSpace(username))

	err := r.getDB(ctx).
		Model(&entity.Account{}).
		Where("username_normalized = ?", normalizedUsername).
		Count(&count).Error

	if err != nil {
		r.logger.Error("Failed to check account existence by username", core.Error(err))
		return false, errors.InternalError("errors.databaseError", err)
	}

	return count > 0, nil
}

func (r *accountRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.AccountStatus) error {
	result := r.getDB(ctx).
		Model(&entity.Account{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		r.logger.Error("Failed to update account status", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}

	if result.RowsAffected == 0 {
		return iamerror.ErrAccountNotFound
	}

	return nil
}

func (r *accountRepository) MarkEmailVerifiedAndActivate(ctx context.Context, id uuid.UUID) error {
	result := r.getDB(ctx).
		Model(&entity.Account{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"email_verified": true,
			"status":         entity.AccountStatusActive,
		})

	if result.Error != nil {
		r.logger.Error("Failed to mark account email verified and active", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}

	if result.RowsAffected == 0 {
		return iamerror.ErrAccountNotFound
	}

	return nil
}

func (r *accountRepository) FindBySegment(ctx context.Context, segment map[string]interface{}) ([]*entity.Account, error) {
	var accounts []*entity.Account
	db := r.getDB(ctx).Select("accounts.id")

	needsBPJoin := false
	for key := range segment {
		switch key {
		case "sector_ids", "tag_ids", "region", "stage":
			needsBPJoin = true
		}
	}
	if needsBPJoin {
		db = db.Joins("JOIN business_profiles bp ON bp.account_id = accounts.id")
	}

	for key, value := range segment {
		switch key {
		case "status":
			db = db.Where("status = ?", value)
		case "created_after":
			if t, ok := value.(string); ok {
				db = db.Where("created_at >= ?", t)
			}
		case "created_before":
			if t, ok := value.(string); ok {
				db = db.Where("created_at <= ?", t)
			}
		case "sector_ids":
			if ids, ok := value.(pq.StringArray); ok && len(ids) > 0 {
				db = db.Where("bp.sector_id IN ?", []string(ids))
			}
		case "tag_ids":
			if ids, ok := value.(pq.StringArray); ok && len(ids) > 0 {
				db = db.Joins("JOIN business_profile_tags bpt ON bpt.business_profile_id = bp.id").
					Where("bpt.tag_id IN ?", []string(ids))
			}
		case "region":
			if s, ok := value.(string); ok && s != "" {
				db = db.Where("bp.region = ?", s)
			}
		case "stage":
			if s, ok := value.(string); ok && s != "" {
				db = db.Where("bp.stage = ?", s)
			}
		}
	}

	if err := db.Find(&accounts).Error; err != nil {
		r.logger.Error("Failed to find accounts by segment", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return accounts, nil
}

func (r *accountRepository) ListAdmins(ctx context.Context, permissionCodes []string, queryOpts map[string]interface{}) ([]*entity.Account, int64, error) {
	var accounts []*entity.Account

	db := r.getDB(ctx).
		Joins("JOIN role_assignments ra ON ra.account_id = accounts.id").
		Joins("JOIN roles ro ON ro.id = ra.role_id").
		Joins("JOIN role_permissions rp ON rp.role_id = ro.id").
		Joins("JOIN permissions p ON p.id = rp.permission_id").
		Where("p.code IN ?", permissionCodes).
		Where("ra.revoked_at IS NULL").
		Group("accounts.id")

	if search, ok := queryOpts["search"].(string); ok && search != "" {
		searchTerm := "%" + search + "%"
		db = db.Where("(accounts.email ILIKE ? OR accounts.username ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ?)",
			searchTerm, searchTerm, searchTerm, searchTerm)
		db = db.Joins("JOIN users ON users.id = accounts.user_id")
	}

	if status, ok := queryOpts["status"].(string); ok && status != "" {
		db = db.Where("accounts.status = ?", status)
	}

	if roleID, ok := queryOpts["roleId"].(string); ok && roleID != "" {
		db = db.Where("ro.id = ?", roleID)
	}

	if page, ok := queryOpts["page"].(int); ok && page > 0 {
		pageSize := 20
		if ps, ok := queryOpts["pageSize"].(int); ok && ps > 0 {
			pageSize = ps
		}
		offset := (page - 1) * pageSize
		db = db.Offset(offset).Limit(pageSize)
	}

	db = db.Order("accounts.created_at DESC")

	if err := db.Find(&accounts).Error; err != nil {
		r.logger.Error("Failed to list admin accounts", core.Error(err))
		return nil, 0, errors.InternalError("errors.databaseError", err)
	}

	var total int64
	countDB := r.getDB(ctx).
		Joins("JOIN role_assignments ra ON ra.account_id = accounts.id").
		Joins("JOIN roles ro ON ro.id = ra.role_id").
		Joins("JOIN role_permissions rp ON rp.role_id = ro.id").
		Joins("JOIN permissions p ON p.id = rp.permission_id").
		Where("p.code IN ?", permissionCodes).
		Where("ra.revoked_at IS NULL")

	if search, ok := queryOpts["search"].(string); ok && search != "" {
		searchTerm := "%" + search + "%"
		countDB = countDB.Where("(accounts.email ILIKE ? OR accounts.username ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ?)",
			searchTerm, searchTerm, searchTerm, searchTerm)
		countDB = countDB.Joins("JOIN users ON users.id = accounts.user_id")
	}

	if status, ok := queryOpts["status"].(string); ok && status != "" {
		countDB = countDB.Where("accounts.status = ?", status)
	}

	if roleID, ok := queryOpts["roleId"].(string); ok && roleID != "" {
		countDB = countDB.Where("ro.id = ?", roleID)
	}

	countDB.Model(&entity.Account{}).Count(&total)

	return accounts, total, nil
}
