package routes

import (
	"context"
	"testing"

	"github.com/Final-Year-Project-G22/backend/core/internal/modules/iam/delivery/contextkeys"
	"github.com/google/uuid"
)

func TestGetSessionIDFromContext(t *testing.T) {
	tests := []struct {
		name    string
		ctxVal  interface{}
		wantNil bool
	}{
		{
			name:    "valid UUID",
			ctxVal:  uuid.NewString(),
			wantNil: false,
		},
		{
			name:    "invalid UUID string",
			ctxVal:  "not-a-uuid",
			wantNil: true,
		},
		{
			name:    "nil value",
			ctxVal:  nil,
			wantNil: true,
		},
		{
			name:    "non-string value",
			ctxVal:  123,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context
			if tt.ctxVal != nil {
				ctx = context.WithValue(context.Background(), contextkeys.SessionID, tt.ctxVal)
			} else {
				ctx = context.Background()
			}
			got := getSessionIDFromContext(ctx)
			if tt.wantNil {
				if got != uuid.Nil {
					t.Errorf("getSessionIDFromContext() = %v, want nil UUID", got)
				}
			} else {
				if got == uuid.Nil {
					t.Errorf("getSessionIDFromContext() = nil, want a valid UUID")
				}
				// Note: we don't check the exact value because we don't know what the input UUID is
				// but we can check that it's not nil
			}
		})
	}
}
