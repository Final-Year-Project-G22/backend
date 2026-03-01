package query

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

const (
	// DefaultPage is the default page number (1-indexed).
	DefaultPage = 1

	// DefaultPageSize is the default number of items per page.
	DefaultPageSize = 20

	// MaxPageSize is the maximum allowed page size to prevent abuse.
	MaxPageSize = 100
)

// QueryOptions contains all options for building a query.
// Use DefaultQueryOptions() for sensible defaults.
type QueryOptions struct {
	// Page is the page number (1-indexed). Defaults to 1.
	Page int

	// PageSize is the number of items per page. Defaults to 20, max 100.
	PageSize int

	// SortBy defines the columns to sort by. Multiple columns supported.
	// Example: []string{"name", "created_at"}
	SortBy []string

	// SortOrder defines the sort order for each SortBy column.
	// Use "asc" or "desc". Defaults to "asc" for each.
	// Example: []string{"asc", "desc"}
	SortOrder []string

	// Search is the text search term. Uses ILIKE for case-insensitive matching.
	Search string

	// SearchColumns specifies which columns to search in.
	// If empty, uses entity's SearchableColumns from config.
	SearchColumns []string

	// Filters provides exact-match filters as key-value pairs.
	// Example: map[string]interface{}{"status": "active", "role_id": uuid}
	Filters map[string]interface{}

	// Preload specifies which relations to eager load.
	// Example: []string{"Role", "Tenant"}
	Preload []string

	// IncludeArchived when true includes soft-deleted records.
	IncludeArchived bool
}

// QueryBuilder transforms QueryOptions into GORM query clauses.
// It validates columns against entity configuration to prevent SQL injection.
type QueryBuilder struct {
	entityType string
	config     *EntityConfig
}

// NewQueryBuilder creates a new QueryBuilder for the given entity type.
//
// Example:
//
//	builder := query.NewQueryBuilder("user")
//	db := builder.Build(db, opts)
func NewQueryBuilder(entityType string) *QueryBuilder {
	return &QueryBuilder{
		entityType: entityType,
		config:     GetConfig(entityType),
	}
}

// Build applies QueryOptions to the GORM db for non-archived (active) records.
// It excludes soft-deleted records (where deleted_at IS NULL).
//
// Example:
//
//	opts := query.QueryOptions{Page: 1, PageSize: 10, Search: "john"}
//	builder := query.NewQueryBuilder("user")
//	db := builder.Build(db.Session(&gorm.Session{}), opts)
//	var users []User
//	db.Find(&users)
func (q *QueryBuilder) Build(db *gorm.DB, opts QueryOptions) *gorm.DB {
	return q.applyQuery(db, opts, false)
}

// BuildArchived applies QueryOptions to include archived (soft-deleted) records.
// It includes only soft-deleted records (where deleted_at IS NOT NULL).
//
// Example:
//
//	opts := query.QueryOptions{Page: 1, PageSize: 10}
//	builder := query.NewQueryBuilder("user")
//	db := builder.BuildArchived(db.Session(&gorm.Session{}), opts)
//	var archivedUsers []User
//	db.Find(&archivedUsers)
func (q *QueryBuilder) BuildArchived(db *gorm.DB, opts QueryOptions) *gorm.DB {
	return q.applyQuery(db, opts, true)
}

func (q *QueryBuilder) applyQuery(db *gorm.DB, opts QueryOptions, archived bool) *gorm.DB {
	db = q.applyPagination(db, opts)
	db = q.applySorting(db, opts)
	db = q.applyFilters(db, opts)
	db = q.applySearch(db, opts)
	db = q.applyPreload(db, opts)

	if archived {
		db = db.Unscoped().Where("deleted_at IS NOT NULL")
	} else {
		db = db.Where("deleted_at IS NULL")
	}

	return db
}

// applyPagination applies pagination (OFFSET/LIMIT) to the query.
// Page is 1-indexed. Page size is capped at MaxPageSize.
func (q *QueryBuilder) applyPagination(db *gorm.DB, opts QueryOptions) *gorm.DB {
	if opts.Page < 1 {
		opts.Page = DefaultPage
	}
	if opts.PageSize < 1 {
		opts.PageSize = DefaultPageSize
	}
	if opts.PageSize > MaxPageSize {
		opts.PageSize = MaxPageSize
	}

	offset := (opts.Page - 1) * opts.PageSize
	return db.Offset(offset).Limit(opts.PageSize)
}

// applySorting applies ORDER BY clause.
// Validates columns against entity config to prevent SQL injection.
// Defaults to created_at desc if no SortBy provided and no config.
func (q *QueryBuilder) applySorting(db *gorm.DB, opts QueryOptions) *gorm.DB {
	if len(opts.SortBy) == 0 {
		if q.config != nil && len(q.config.DefaultSort) > 0 {
			opts.SortBy = q.config.DefaultSort
			opts.SortOrder = make([]string, len(opts.SortBy))
			for i := range opts.SortOrder {
				opts.SortOrder[i] = "asc"
			}
		} else {
			return db.Order("created_at desc")
		}
	}

	var orderClauses []string
	for i, col := range opts.SortBy {
		if q.entityType != "" && !IsValidSortColumn(q.entityType, col) {
			continue
		}
		order := "asc"
		if i < len(opts.SortOrder) && strings.ToLower(opts.SortOrder[i]) == "desc" {
			order = "desc"
		}
		orderClauses = append(orderClauses, fmt.Sprintf("%s %s", col, order))
	}

	if len(orderClauses) > 0 {
		return db.Order(strings.Join(orderClauses, ", "))
	}
	return db
}

// applyFilters applies WHERE conditions for exact-match filters.
func (q *QueryBuilder) applyFilters(db *gorm.DB, opts QueryOptions) *gorm.DB {
	if len(opts.Filters) > 0 {
		return db.Where(opts.Filters)
	}
	return db
}

// applySearch applies ILIKE search across specified columns.
// Combines columns with OR logic. Case-insensitive.
func (q *QueryBuilder) applySearch(db *gorm.DB, opts QueryOptions) *gorm.DB {
	if opts.Search == "" {
		return db
	}

	searchColumns := opts.SearchColumns
	if len(searchColumns) == 0 && q.config != nil {
		searchColumns = q.config.SearchableColumns
	}

	if len(searchColumns) == 0 {
		return db
	}

	var conditions []string
	var args []interface{}
	searchTerm := "%" + opts.Search + "%"

	for _, col := range searchColumns {
		if q.entityType != "" && !IsValidSearchColumn(q.entityType, col) {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s ILIKE ?", col))
		args = append(args, searchTerm)
	}

	if len(conditions) > 0 {
		return db.Where("("+strings.Join(conditions, " OR ")+")", args...)
	}
	return db
}

// applyPreload eager loads specified relations.
func (q *QueryBuilder) applyPreload(db *gorm.DB, opts QueryOptions) *gorm.DB {
	for _, preload := range opts.Preload {
		db = db.Preload(preload)
	}
	return db
}

// Count returns the total number of records matching the query options.
// Uses a separate query session to avoid affecting the main query.
func (q *QueryBuilder) Count(db *gorm.DB, opts QueryOptions, archived bool) int64 {
	var count int64
	query := q.applyQuery(db.Session(&gorm.Session{}), opts, archived)
	query.Count(&count)
	return count
}

// DefaultQueryOptions returns QueryOptions with sensible defaults.
// Use this as a starting point and modify as needed.
//
// Example:
//
//	opts := query.DefaultQueryOptions()
//	opts.Page = 1
//	opts.PageSize = 50
//	opts.Search = "john"
func DefaultQueryOptions() QueryOptions {
	return QueryOptions{
		Page:            DefaultPage,
		PageSize:        DefaultPageSize,
		SortBy:          []string{"created_at"},
		SortOrder:       []string{"desc"},
		Search:          "",
		SearchColumns:   nil,
		Filters:         nil,
		Preload:         nil,
		IncludeArchived: false,
	}
}
