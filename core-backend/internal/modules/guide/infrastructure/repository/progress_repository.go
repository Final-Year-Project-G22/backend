package repository

import (
	"context"
	"fmt"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/entity"
	guideerror "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/error"
	guiderepo "github.com/Final-Year-Project-G22/backend/core/internal/modules/guide/domain/repository"
	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	sharedrepo "github.com/Final-Year-Project-G22/backend/core/internal/shared/repository"
	"github.com/Final-Year-Project-G22/backend/core/pkg/errors"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type progressRepository struct {
	sharedrepo.GenericRepository[entity.UserGuideProgress]
	db     *core.Database
	logger core.Logger
}

func NewProgressRepository(db *core.Database, logger core.Logger) guiderepo.ProgressRepository {
	base := sharedrepo.NewBaseRepository[entity.UserGuideProgress](db, logger)
	return &progressRepository{GenericRepository: base, db: db, logger: logger}
}

func (r *progressRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *progressRepository) GetProgress(ctx context.Context, accountID, userID, stepID uuid.UUID) (*entity.UserGuideProgress, error) {
	var progress entity.UserGuideProgress
	if err := r.getDB(ctx).Where("account_id = ? AND user_id = ? AND step_id = ?", accountID, userID, stepID).First(&progress).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, guideerror.ErrProgressNotFound
		}
		r.logger.Error("Failed to get progress", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &progress, nil
}

func (r *progressRepository) ListProgressByGuide(ctx context.Context, accountID, userID, guideID uuid.UUID, q query.QueryOptions) ([]*entity.UserGuideProgress, error) {
	var progress []*entity.UserGuideProgress
	db := r.getDB(ctx).
		Model(&entity.UserGuideProgress{}).
		Joins("JOIN guide_steps ON guide_steps.id = user_guide_progresses.step_id").
		Where("user_guide_progresses.account_id = ? AND user_guide_progresses.user_id = ? AND guide_steps.guide_id = ?", accountID, userID, guideID).
		Preload("Step")
	if len(q.SortBy) > 0 {
		for i, col := range q.SortBy {
			order := "asc"
			if i < len(q.SortOrder) && q.SortOrder[i] == "desc" {
				order = "desc"
			}
			db = db.Order(fmt.Sprintf("%s %s", col, order))
		}
	} else {
		db = db.Order("guide_steps.sort_order asc, user_guide_progresses.created_at asc")
	}
	if q.Page < 1 {
		q.Page = query.DefaultPage
	}
	if q.PageSize < 1 {
		q.PageSize = query.DefaultPageSize
	}
	if q.PageSize > query.MaxPageSize {
		q.PageSize = query.MaxPageSize
	}
	if err := db.Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&progress).Error; err != nil {
		r.logger.Error("Failed to list progress by guide", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return progress, nil
}

func (r *progressRepository) UpsertProgress(ctx context.Context, progress *entity.UserGuideProgress) error {
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}, {Name: "user_id"}, {Name: "step_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"status":             progress.Status,
			"started_at":         progress.StartedAt,
			"completed_at":       progress.CompletedAt,
			"time_spent":         progress.TimeSpent,
			"notes":              progress.Notes,
			"uploaded_documents": progress.UploadedDocuments,
			"last_accessed_at":   progress.LastAccessedAt,
			"version":            progress.StepVersion,
			"updated_at":         gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(progress).Error; err != nil {
		r.logger.Error("Failed to upsert progress", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *progressRepository) DeleteProgress(ctx context.Context, accountID, userID, stepID uuid.UUID) error {
	result := r.getDB(ctx).Where("account_id = ? AND user_id = ? AND step_id = ?", accountID, userID, stepID).Delete(&entity.UserGuideProgress{})
	if result.Error != nil {
		r.logger.Error("Failed to delete progress", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return guideerror.ErrProgressNotFound
	}
	return nil
}

func (r *progressRepository) GetJourney(ctx context.Context, accountID, userID, guideID uuid.UUID) (*entity.UserGuideJourney, error) {
	var journey entity.UserGuideJourney
	if err := r.getDB(ctx).Where("account_id = ? AND user_id = ? AND guide_id = ?", accountID, userID, guideID).First(&journey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, guideerror.ErrJourneyNotFound
		}
		r.logger.Error("Failed to get journey", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &journey, nil
}

func (r *progressRepository) UpsertJourney(ctx context.Context, journey *entity.UserGuideJourney) error {
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}, {Name: "user_id"}, {Name: "guide_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"journey_hash":         journey.JourneyHash,
			"step_sequence":        journey.StepSequence,
			"total_steps":          journey.TotalSteps,
			"completed_steps":      journey.CompletedSteps,
			"estimated_total_time": journey.EstimatedTotalTime,
			"generated_at":         journey.GeneratedAt,
			"expires_at":           journey.ExpiresAt,
			"updated_at":           gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(journey).Error; err != nil {
		r.logger.Error("Failed to upsert journey", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *progressRepository) DeleteJourney(ctx context.Context, accountID, userID, guideID uuid.UUID) error {
	result := r.getDB(ctx).Where("account_id = ? AND user_id = ? AND guide_id = ?", accountID, userID, guideID).Delete(&entity.UserGuideJourney{})
	if result.Error != nil {
		r.logger.Error("Failed to delete journey", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return guideerror.ErrJourneyNotFound
	}
	return nil
}

func (r *progressRepository) InvalidateJourneysForGuide(ctx context.Context, guideID uuid.UUID) error {
	if err := r.getDB(ctx).Where("guide_id = ?", guideID).Delete(&entity.UserGuideJourney{}).Error; err != nil {
		r.logger.Error("Failed to invalidate journeys for guide", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *progressRepository) InvalidateJourneyForUser(ctx context.Context, accountID, userID, guideID uuid.UUID) error {
	return r.DeleteJourney(ctx, accountID, userID, guideID)
}

func (r *progressRepository) GetBookmark(ctx context.Context, accountID, userID, stepID uuid.UUID) (*entity.UserGuideBookmark, error) {
	var bookmark entity.UserGuideBookmark
	if err := r.getDB(ctx).Preload("Step").Where("account_id = ? AND user_id = ? AND step_id = ?", accountID, userID, stepID).First(&bookmark).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, guideerror.ErrBookmarkNotFound
		}
		r.logger.Error("Failed to get bookmark", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return &bookmark, nil
}

func (r *progressRepository) ListBookmarks(ctx context.Context, accountID, userID uuid.UUID, q query.QueryOptions) ([]*entity.UserGuideBookmark, error) {
	var bookmarks []*entity.UserGuideBookmark
	db := r.getDB(ctx).Where("account_id = ? AND user_id = ?", accountID, userID).Preload("Step")
	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	if len(q.SortBy) > 0 {
		for i, col := range q.SortBy {
			order := "asc"
			if i < len(q.SortOrder) && q.SortOrder[i] == "desc" {
				order = "desc"
			}
			db = db.Order(fmt.Sprintf("%s %s", col, order))
		}
	} else {
		db = db.Order("created_at desc")
	}
	if q.Page < 1 {
		q.Page = query.DefaultPage
	}
	if q.PageSize < 1 {
		q.PageSize = query.DefaultPageSize
	}
	if q.PageSize > query.MaxPageSize {
		q.PageSize = query.MaxPageSize
	}
	if err := db.Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&bookmarks).Error; err != nil {
		r.logger.Error("Failed to list bookmarks", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}
	return bookmarks, nil
}

func (r *progressRepository) UpsertBookmark(ctx context.Context, bookmark *entity.UserGuideBookmark) error {
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}, {Name: "user_id"}, {Name: "step_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"note":       bookmark.Note,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(bookmark).Error; err != nil {
		r.logger.Error("Failed to upsert bookmark", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *progressRepository) RemoveBookmark(ctx context.Context, accountID, userID, stepID uuid.UUID) error {
	result := r.getDB(ctx).Where("account_id = ? AND user_id = ? AND step_id = ?", accountID, userID, stepID).Delete(&entity.UserGuideBookmark{})
	if result.Error != nil {
		r.logger.Error("Failed to remove bookmark", core.Error(result.Error))
		return errors.InternalError("errors.databaseError", result.Error)
	}
	if result.RowsAffected == 0 {
		return guideerror.ErrBookmarkNotFound
	}
	return nil
}

func (r *progressRepository) UpsertRecentView(ctx context.Context, accountID, userID, guideID uuid.UUID) error {
	if err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}, {Name: "user_id"}, {Name: "guide_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_viewed_at": gorm.Expr("CURRENT_TIMESTAMP"),
			"view_count":     gorm.Expr("view_count + 1"),
			"updated_at":     gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(&entity.UserGuideRecentView{
		AccountID: accountID,
		UserID:    userID,
		GuideID:   guideID,
	}).Error; err != nil {
		r.logger.Error("Failed to upsert recent view", core.Error(err))
		return errors.InternalError("errors.databaseError", err)
	}
	return nil
}

func (r *progressRepository) ListRecentlyViewedGuides(ctx context.Context, accountID, userID uuid.UUID, q query.QueryOptions, locale constants.Locale) ([]*entity.Guide, error) {
	var guides []*entity.Guide
	db := r.getDB(ctx).
		Model(&entity.Guide{}).
		Select("guides.*").
		Joins("JOIN user_guide_recent_views rv ON rv.guide_id = guides.id").
		Where("rv.account_id = ? AND rv.user_id = ?", accountID, userID).
		Preload("Translations", "language = ?", locale).
		Order("rv.last_viewed_at desc")
	if q.Page < 1 {
		q.Page = query.DefaultPage
	}
	if q.PageSize < 1 {
		q.PageSize = query.DefaultPageSize
	}
	if q.PageSize > query.MaxPageSize {
		q.PageSize = query.MaxPageSize
	}
	if err := db.Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&guides).Error; err != nil {
		r.logger.Error("Failed to list recently viewed guides", core.Error(err))
		return nil, errors.InternalError("errors.databaseError", err)
	}

	for i := range guides {
		if len(guides[i].Translations) == 0 {
			if err := r.getDB(ctx).Preload("Translations", "language = ?", constants.LocaleEnglish).
				Where("id = ?", guides[i].ID).
				First(guides[i]).Error; err != nil {
				r.logger.Error("Failed to load fallback translation for recently viewed guide", core.Error(err))
			}
		}
	}
	return guides, nil
}
