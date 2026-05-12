package entity

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type UUIDArray []uuid.UUID

func (ua *UUIDArray) Scan(src interface{}) error {
	if src == nil {
		*ua = nil
		return nil
	}

	var s string
	switch v := src.(type) {
	case []byte:
		s = string(v)
	case string:
		s = v
	default:
		return fmt.Errorf("UUIDArray.Scan: unexpected type %T", src)
	}

	s = strings.Trim(s, "{}")
	if s == "" {
		*ua = UUIDArray{}
		return nil
	}

	parts := strings.Split(s, ",")
	result := make(UUIDArray, len(parts))
	for i, p := range parts {
		uid, err := uuid.Parse(strings.TrimSpace(p))
		if err != nil {
			return fmt.Errorf("UUIDArray.Scan: invalid UUID %q: %w", p, err)
		}
		result[i] = uid
	}
	*ua = result
	return nil
}

func (ua UUIDArray) Value() (driver.Value, error) {
	if ua == nil {
		return nil, nil
	}
	parts := make([]string, len(ua))
	for i, uid := range ua {
		parts[i] = uid.String()
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}
