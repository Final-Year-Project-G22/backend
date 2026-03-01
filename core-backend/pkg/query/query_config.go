// Package query provides a robust query builder system for GORM-based repositories.
// It supports pagination, sorting, filtering, search, and relation preloading.
//
// # Usage
//
// 1. Register entity configurations in your model files:
//
//	query.RegisterConfig("user", query.EntityConfig{
//	    SearchableColumns: []string{"name", "email"},
//	    SortableColumns:   []string{"name", "email", "created_at"},
//	    DefaultSort:       []string{"created_at"},
//	    DefaultIncludes:   []string{"Role"},
//	})
//
// 2. Use in your HTTP handlers:
//
//	func GetUsers(c *gin.Context) {
//	    opts := query.ParseFromRequest(c)
//	    result := userRepo.FindAll(c.Request.Context(), opts)
//	    c.JSON(200, result)
//	}
//
// # Query String Format
//
//	?page=1&pageSize=20&sortBy=name,created_at&sortOrder=asc,desc&search=john&searchColumns=name,email&preload=Role
package query

import "slices"

// EntityConfig defines the queryable configuration for an entity.
// It specifies which columns can be searched, sorted, and preloaded.
type EntityConfig struct {
	// SearchableColumns defines which columns can be used in text search.
	// These columns will be used when a search query is provided.
	SearchableColumns []string

	// SortableColumns defines which columns can be used for ordering results.
	// Column validation prevents SQL injection attacks.
	SortableColumns []string

	// DefaultSort defines the default sorting when no sortBy is specified.
	// Example: []string{"created_at"}
	DefaultSort []string

	// DefaultIncludes defines which relations to preload by default.
	// Example: []string{"Role", "Tenant"}
	DefaultIncludes []string
}

var configRegistry = make(map[string]*EntityConfig)

// RegisterConfig registers an entity's query configuration.
//
// Example:
//
//	query.RegisterConfig("user", query.EntityConfig{
//	    SearchableColumns: []string{"name", "email"},
//	    SortableColumns:   []string{"name", "email", "created_at"},
//	    DefaultSort:       []string{"created_at"},
//	    DefaultIncludes:   []string{"Role"},
//	})
func RegisterConfig(entityType string, config EntityConfig) {
	configRegistry[entityType] = &config
}

// GetConfig retrieves the query configuration for an entity type.
// Returns nil if no configuration is registered.
func GetConfig(entityType string) *EntityConfig {
	if config, ok := configRegistry[entityType]; ok {
		return config
	}
	return nil
}

// IsValidSortColumn checks if a column is allowed for sorting.
// Returns false if the entity has no configuration or column is not whitelisted.
//
// Example:
//
//	if query.IsValidSortColumn("user", "name") {
//	    // column is safe to use in ORDER BY
//	}
func IsValidSortColumn(entityType, column string) bool {
	config := GetConfig(entityType)
	if config == nil {
		return false
	}
	return slices.Contains(config.SortableColumns, column)
}

// IsValidSearchColumn checks if a column is allowed for searching.
// Returns false if the entity has no configuration or column is not whitelisted.
//
// Example:
//
//	if query.IsValidSearchColumn("user", "email") {
//	    // column is safe to use in WHERE LIKE
//	}
func IsValidSearchColumn(entityType, column string) bool {
	config := GetConfig(entityType)
	if config == nil {
		return false
	}
	return slices.Contains(config.SearchableColumns, column)
}
