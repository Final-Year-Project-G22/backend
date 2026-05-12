package dbtypes

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// UUIDArray adapts PostgreSQL uuid[] values for GORM.
type UUIDArray []uuid.UUID

// Scan implements sql.Scanner.
func (a *UUIDArray) Scan(src any) error {
	if src == nil {
		*a = UUIDArray{}
		return nil
	}

	var raw string
	switch value := src.(type) {
	case string:
		raw = value
	case []byte:
		raw = string(value)
	default:
		return fmt.Errorf("unsupported Scan source for UUIDArray: %T", src)
	}

	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		*a = UUIDArray{}
		return nil
	}

	raw = strings.TrimPrefix(raw, "{")
	raw = strings.TrimSuffix(raw, "}")
	if raw == "" {
		*a = UUIDArray{}
		return nil
	}

	parts := strings.Split(raw, ",")
	result := make(UUIDArray, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(strings.Trim(part, `"`))
		if part == "" {
			continue
		}
		u, err := uuid.Parse(part)
		if err != nil {
			return fmt.Errorf("parse uuid array element: %w", err)
		}
		result = append(result, u)
	}

	*a = result
	return nil
}

// Value implements driver.Valuer.
func (a UUIDArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}

	parts := make([]string, 0, len(a))
	for _, u := range a {
		parts = append(parts, u.String())
	}

	return "{" + strings.Join(parts, ",") + "}", nil
}
