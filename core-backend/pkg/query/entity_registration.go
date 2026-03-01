package query

// RegisterDefaultConfigs registers base configurations for common fields.
// This allows basic sorting by created_at, updated_at, and id.
func RegisterDefaultConfigs() {
	RegisterConfig("base_model", EntityConfig{
		SearchableColumns: []string{"id", "created_at", "updated_at"},
		SortableColumns:   []string{"id", "created_at", "updated_at"},
		DefaultSort:       []string{"created_at"},
		DefaultIncludes:   []string{},
	})
}

// ExampleRegistration demonstrates how to register entity configurations.
// Call this function (e.g., in your model's init() or main()) to register your entities.
//
// # Best Practice
//
// Register entity configs in your model files' init() function:
//
//	import "github.com/Final-Year-Project-G22/backend/core/pkg/query"
//
//	func init() {
//	    query.RegisterConfig("user", query.EntityConfig{
//	        SearchableColumns: []string{"name", "email", "phone"},
//	        SortableColumns:   []string{"name", "email", "created_at", "updated_at"},
//	        DefaultSort:       []string{"created_at"},
//	        DefaultIncludes:   []string{"Role", "Tenant"},
//	    })
//	}
func ExampleRegistration() {
	RegisterConfig("user", EntityConfig{
		SearchableColumns: []string{"name", "email", "phone"},
		SortableColumns:   []string{"name", "email", "created_at", "updated_at"},
		DefaultSort:       []string{"created_at"},
		DefaultIncludes:   []string{"Role"},
	})

	RegisterConfig("product", EntityConfig{
		SearchableColumns: []string{"name", "description", "sku"},
		SortableColumns:   []string{"name", "price", "created_at", "updated_at"},
		DefaultSort:       []string{"name"},
		DefaultIncludes:   []string{"Category", "Tags"},
	})
}
