package repository

import "context"

// Transactor enables cross-entity atomic operations without exposing *gorm.DB.
// Use this when an operation needs to span multiple repositories atomically.
// The transaction is carried via context - repositories automatically detect
// and use it via core.TxFromContext.
type Transactor interface {
	// WithinTransaction executes fn within a database transaction.
	// The transaction is passed via context - repositories automatically
	// detect and use it via core.TxFromContext.
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
