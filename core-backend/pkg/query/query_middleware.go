package query

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ParseFromRequest parses HTTP query parameters into QueryOptions.
// This is the main entry point for handling paginated, filterable API endpoints.
//
// # Supported Query Parameters
//
//	page          - Page number (1-indexed), default: 1
//	pageSize      - Items per page, default: 20, max: 100
//	sortBy        - Comma-separated columns to sort by
//	sortOrder     - Comma-separated order (asc/desc), defaults to asc
//	search        - Text search term
//	searchColumns - Comma-separated columns to search in
//	preload       - Comma-separated relations to eager load
//	includeArchived - Include soft-deleted records (true/1)
//	Any other query parameter is treated as an exact-match filter
//
// # Example Request
//
//	GET /users?page=2&pageSize=50&search=john&searchColumns=name,email&sortBy=name,created_at&sortOrder=asc,desc&preload=Role&status=active
//
// # Example Handler
//
//	func GetUsers(c *gin.Context) {
//	    opts := query.ParseFromRequest(c)
//	    result := userRepo.FindAll(c.Request.Context(), opts)
//	    c.JSON(200, result)
//	}
func ParseFromRequest(c *gin.Context) QueryOptions {
	opts := DefaultQueryOptions()

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			opts.Page = page
		}
	}

	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			opts.PageSize = pageSize
		}
	}

	if sortBy := c.Query("sortBy"); sortBy != "" {
		opts.SortBy = strings.Split(sortBy, ",")
	}

	if sortOrder := c.Query("sortOrder"); sortOrder != "" {
		opts.SortOrder = strings.Split(sortOrder, ",")
	}

	if search := c.Query("search"); search != "" {
		opts.Search = search
	}

	if searchColumns := c.Query("searchColumns"); searchColumns != "" {
		opts.SearchColumns = strings.Split(searchColumns, ",")
	}

	if preload := c.Query("preload"); preload != "" {
		opts.Preload = strings.Split(preload, ",")
	}

	if includeArchived := c.Query("includeArchived"); includeArchived != "" {
		opts.IncludeArchived = includeArchived == "true" || includeArchived == "1"
	}

	parseFilters(c, &opts)

	return opts
}

// parseFilters extracts custom filters from query string.
// Reserved parameter names are skipped (page, pageSize, sortBy, etc.)
func parseFilters(c *gin.Context, opts *QueryOptions) {
	filters := make(map[string]interface{})

	for key, values := range c.Request.URL.Query() {
		if key == "page" || key == "pageSize" || key == "sortBy" ||
			key == "sortOrder" || key == "search" || key == "searchColumns" ||
			key == "preload" || key == "includeArchived" {
			continue
		}

		if len(values) > 0 {
			filters[key] = values[0]
		}
	}

	if len(filters) > 0 {
		opts.Filters = filters
	}
}

// PaginatedResponse creates a standardized paginated response.
// Use this to format your response data with pagination metadata.
//
// # Example
//
//	response := query.PaginatedResponse(users, total, page, pageSize)
//	c.JSON(200, response)
func PaginatedResponse[T any](data []*T, total int64, page, pageSize int) gin.H {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return gin.H{
		"data":        data,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	}
}

// HandleError sends a standardized error response.
//
// # Example
//
//	query.HandleError(c, http.StatusBadRequest, "Invalid page parameter")
func HandleError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// ParseEntityType extracts the entity type name from an interface{}.
// This is useful for automatic entity type detection.
func ParseEntityType(entity interface{}) string {
	parts := strings.Split(string(rune('t')), ".")
	if len(parts) > 0 {
		return strings.ToLower(parts[len(parts)-1])
	}
	return ""
}

// Paginate validates and normalizes pagination parameters.
// Ensures page >= 1 and pageSize is within limits.
//
// # Example
//
//	page, pageSize := query.Paginate(page, pageSize)
func Paginate(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}

// GetPageFromRequest extracts and validates page from query string.
// Returns DefaultPage if invalid or not provided.
//
// # Example
//
//	page := query.GetPageFromRequest(c)
func GetPageFromRequest(c *gin.Context) int {
	pageStr := c.Query("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return DefaultPage
	}
	return page
}

// GetPageSizeFromRequest extracts and validates pageSize from query string.
// Returns DefaultPageSize if invalid, MaxPageSize if too large.
//
// # Example
//
//	pageSize := query.GetPageSizeFromRequest(c)
func GetPageSizeFromRequest(c *gin.Context) int {
	pageSizeStr := c.Query("pageSize")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		return DefaultPageSize
	}
	if pageSize > MaxPageSize {
		return MaxPageSize
	}
	return pageSize
}

// RespondWithPagination sends a standardized paginated JSON response.
//
// # Example
//
//	query.RespondWithPagination(c, 200, total, page, pageSize)
func RespondWithPagination(c *gin.Context, status int, total int64, page, pageSize int) gin.H {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return gin.H{
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	}
}
