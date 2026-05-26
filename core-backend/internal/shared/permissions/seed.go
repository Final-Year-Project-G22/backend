package permissions

import "strings"

type SeedPermission struct {
	Code        string
	Name        string
	Description *string
	Module      string
}

type SeedRole struct {
	Code            string
	Name            string
	Description     *string
	PermissionCodes []string
}

func StringPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
