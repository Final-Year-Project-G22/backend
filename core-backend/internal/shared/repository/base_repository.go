package repository

import (
	"context"
	"reflect"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseRepository[T any] struct {
	db         *core.Database
	logger     core.Logger
	entityType string
}

func NewBaseRepository[T any](db *core.Database, logger core.Logger) GenericRepository[T] {
	var entity T
	entityType := getEntityType(entity)
	return &BaseRepository[T]{
		db:         db,
		logger:     logger,
		entityType: entityType,
	}
}

func getEntityType(entity any) string {
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		r.logger.Error("Failed to create entity", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *BaseRepository[T]) BulkCreate(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).CreateInBatches(entities, 100).Error; err != nil {
		r.logger.Error("Failed to bulk create entities", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *BaseRepository[T]) GetByID(ctx context.Context, id uuid.UUID) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFoundError(r.entityType, id)
		}
		r.logger.Error("Failed to get entity by ID", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &entity, nil
}

func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		r.logger.Error("Failed to update entity", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *BaseRepository[T]) UpdateByID(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(updates).Error; err != nil {
		r.logger.Error("Failed to update entity by ID", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(new(T), "id = ?", id).Error; err != nil {
		r.logger.Error("Failed to delete entity", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *BaseRepository[T]) HardDelete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Unscoped().Delete(new(T), "id = ?", id).Error; err != nil {
		r.logger.Error("Failed to hard delete entity", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *BaseRepository[T]) FindAll(ctx context.Context, opts query.QueryOptions) PaginatedResult[T] {
	return r.findAll(ctx, opts, false)
}

func (r *BaseRepository[T]) FindAllArchived(ctx context.Context, opts query.QueryOptions) PaginatedResult[T] {
	return r.findAll(ctx, opts, true)
}

func (r *BaseRepository[T]) First(ctx context.Context, opts query.QueryOptions) (*T, error) {
	builder := query.NewQueryBuilder(r.entityType)

	db := builder.Build(r.db.WithContext(ctx), opts)

	var entity T
	if err := db.First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFoundError(r.entityType, nil)
		}
		r.logger.Error("Failed to get first entity", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &entity, nil
}

func (r *BaseRepository[T]) Find(ctx context.Context, opts query.QueryOptions) ([]*T, error) {
	builder := query.NewQueryBuilder(r.entityType)

	db := builder.Build(r.db.WithContext(ctx), opts)

	var entities []*T
	if err := db.Find(&entities).Error; err != nil {
		r.logger.Error("Failed to find entities", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return entities, nil
}

func (r *BaseRepository[T]) findAll(ctx context.Context, opts query.QueryOptions, archived bool) PaginatedResult[T] {
	var entities []*T

	builder := query.NewQueryBuilder(r.entityType)

	// Apply pagination defaults
	if opts.Page < 1 {
		opts.Page = query.DefaultPage
	}
	if opts.PageSize < 1 {
		opts.PageSize = query.DefaultPageSize
	}
	if opts.PageSize > query.MaxPageSize {
		opts.PageSize = query.MaxPageSize
	}

	// Count total records
	total := builder.Count(r.db.WithContext(ctx).Session(&gorm.Session{}), opts, archived)

	// Build and execute query
	var db *gorm.DB
	if archived {
		db = builder.BuildArchived(r.db.WithContext(ctx), opts)
	} else {
		db = builder.Build(r.db.WithContext(ctx), opts)
	}

	if err := db.Find(&entities).Error; err != nil {
		r.logger.Error("Failed to find entities", core.Error(err))
		return PaginatedResult[T]{
			Data:       nil,
			Total:      0,
			Page:       opts.Page,
			PageSize:   opts.PageSize,
			TotalPages: 0,
		}
	}

	// Calculate total pages
	totalPages := int(total) / opts.PageSize
	if int(total)%opts.PageSize > 0 {
		totalPages++
	}

	return PaginatedResult[T]{
		Data:       entities,
		Total:      total,
		Page:       opts.Page,
		PageSize:   opts.PageSize,
		TotalPages: totalPages,
	}
}

func (r *BaseRepository[T]) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*T, error) {
	if len(ids) == 0 {
		return []*T{}, nil
	}
	var entities []*T
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&entities).Error; err != nil {
		r.logger.Error("Failed to find entities by IDs", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return entities, nil
}

func (r *BaseRepository[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(new(T)).Count(&count).Error; err != nil {
		r.logger.Error("Failed to count entities", core.Error(err))
		return 0, errors.InternalError("errors.databaseError", err)
	}
	return count, nil
}

func (r *BaseRepository[T]) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Count(&count).Error; err != nil {
		r.logger.Error("Failed to check entity existence", core.Error(err))
		return false, errors.InternalError("errors.databaseError", err)
	}
	return count > 0, nil
}

func (r *BaseRepository[T]) Transaction(ctx context.Context, fn func(repo GenericRepository[T]) error) error {
	return r.db.Transaction(ctx, func(tx *gorm.DB) error {
		txRepo := &BaseRepository[T]{
			db:         &core.Database{DB: tx},
			logger:     r.logger,
			entityType: r.entityType,
		}
		return fn(txRepo)
	})
}

func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db.DB
}
