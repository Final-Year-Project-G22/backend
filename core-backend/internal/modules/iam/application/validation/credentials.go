package validation

import (
	"net/mail"
	"regexp"
	"strings"
)

type IdentifierKind string

const (
	IdentifierKindEmail    IdentifierKind = "email"
	IdentifierKindUsername IdentifierKind = "username"
)

var usernameRegex = regexp.MustCompile(`^[a-z0-9_]{3,32}$`)

func NormalizeEmail(email string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return "", false
	}
	if _, err := mail.ParseAddress(normalized); err != nil {
		return "", false
	}
	return normalized, true
}

func NormalizeUsername(username string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(username))
	if !usernameRegex.MatchString(normalized) {
		return "", false
	}
	return normalized, true
}

func NormalizeIdentifier(identifier string) (IdentifierKind, string, bool) {
	trimmed := strings.TrimSpace(identifier)
	if strings.Contains(trimmed, "@") {
		email, ok := NormalizeEmail(trimmed)
		if !ok {
			return "", "", false
		}
		return IdentifierKindEmail, email, true
	}

	username, ok := NormalizeUsername(trimmed)
	if !ok {
		return "", "", false
	}
	return IdentifierKindUsername, username, true
}
