package payment

import (
	"github.com/Final-Year-Project-G22/backend/core/internal/modules/payment/domain/entity"
	"github.com/Final-Year-Project-G22/backend/core/pkg/query"
)

func init() {
	query.RegisterConfig("Plan", query.EntityConfig{
		SearchableColumns: []string{"name", "period"},
		SortableColumns:   []string{"name", "period", "amount", "created_at"},
		DefaultSort:       []string{"name", "period"},
	})

	query.RegisterConfig("Payment", query.EntityConfig{
		SearchableColumns: []string{"tx_ref", "chapa_ref"},
		SortableColumns:   []string{"status", "created_at", "verified_at"},
		DefaultSort:       []string{"created_at"},
	})

	query.RegisterConfig("Subscription", query.EntityConfig{
		SearchableColumns: []string{"plan_name", "plan_period"},
		SortableColumns:   []string{"status", "current_period_end", "created_at"},
		DefaultSort:       []string{"created_at"},
	})
}

// EntityProvider registers payment module entities for schema generation.
type EntityProvider struct{}

func NewEntityProvider() *EntityProvider {
	return &EntityProvider{}
}

func (e *EntityProvider) Entities() []any {
	return []any{
		&entity.Plan{},
		&entity.Payment{},
		&entity.Subscription{},
	}
}

func (e *EntityProvider) ModuleName() string {
	return "payment"
}
