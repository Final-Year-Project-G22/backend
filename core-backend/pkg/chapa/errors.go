package chapa

import "fmt"

// Error represents an error returned by the Chapa API or client.
type Error struct {
	HTTPStatus  int    // HTTP status code
	Message     string // Human-readable message
	RawResponse []byte // Raw response body for debugging
}

func (e *Error) Error() string {
	return fmt.Sprintf("chapa error: status=%d, message=%s", e.HTTPStatus, e.Message)
}

// IsError checks if an error is a *chapa.Error.
func IsError(err error) bool {
	_, ok := err.(*Error)
	return ok
}

// ErrorStatus extracts the HTTP status from a chapa.Error, or returns 0.
func ErrorStatus(err error) int {
	if e, ok := err.(*Error); ok {
		return e.HTTPStatus
	}
	return 0
}
