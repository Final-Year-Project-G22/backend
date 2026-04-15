package repository

import (
	"fmt"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
	"gorm.io/gorm"
)

func applyPaginationAndSorting(db *gorm.DB, q query.QueryOptions, defaultOrder string) *gorm.DB {
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
			if i < len(q.SortOrder) && strings.EqualFold(q.SortOrder[i], "desc") {
				order = "desc"
			}
			db = db.Order(fmt.Sprintf("%s %s", col, order))
		}
	} else if defaultOrder != "" {
		db = db.Order(defaultOrder)
	}

	offset := (q.Page - 1) * q.PageSize
	return db.Offset(offset).Limit(q.PageSize)
}
