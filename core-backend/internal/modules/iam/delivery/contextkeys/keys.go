package contextkeys

import "github.com/google/uuid"

// NilUUID is a convenience constant for checking invalid UUIDs.
var NilUUID = uuid.Nil

// contextKey is a private type to prevent key collisions.
type contextKey struct {
	name string
}

// String returns the context key's name for debugging.
func (c contextKey) String() string {
	return "iam context key: " + c.name
}

// Context keys for authenticated request values.1
var (
	SessionID = contextKey{"session_id"}

	Email = contextKey{"email"}

	AccountID = contextKey{"account_id"}

	UserID = contextKey{"user_id"}

	DocumentID = contextKey{"document_id"}
)

// GetSessionID extracts the session ID from context values.
func GetSessionID(val any) uuid.UUID {
	if id, ok := val.(uuid.UUID); ok {
		return id
	}
	return NilUUID
}

// GetEmail extracts the email from context values.
func GetEmail(val any) string {
	if email, ok := val.(string); ok {
		return email
	}
	return ""
}

func GetAccountID(val any) uuid.UUID {
	if id, ok := val.(uuid.UUID); ok {
		return id
	}

	return NilUUID
}

func GetUserID(val any) uuid.UUID {
	if id, ok := val.(uuid.UUID); ok {
		return id
	}

	return NilUUID
}
