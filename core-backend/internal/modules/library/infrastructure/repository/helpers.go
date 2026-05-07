package repository

import (
	"context"
	"fmt"

	"github.com/Final-Year-Project-G22/backend/core/internal/core"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"gorm.io/gorm"
)

func getDB(ctx context.Context, db *core.Database) *gorm.DB {
	if tx, ok := core.TxFromContext(ctx); ok {
		return tx
	}
	return db.WithContext(ctx)
}

func applyPaginationAndSorting(q query.QueryOptions, defaultSort string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if q.Page < 1 {
			q.Page = query.DefaultPage
		}
		if q.PageSize < 1 {
			q.PageSize = query.DefaultPageSize
		}
		if q.PageSize > query.MaxPageSize {
			q.PageSize = query.MaxPageSize
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
			db = db.Order(defaultSort)
		}

		for _, preload := range q.Preload {
			db = db.Preload(preload)
		}

		return db.Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize)
	}
}
