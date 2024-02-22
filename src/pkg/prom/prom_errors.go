package prom

import "fmt"

// NewError returns a new error with a status code and message.
func NewError(statusCode int, msg string) error {
	return Error{
		StatusCode: statusCode,
		Message:    msg,
	}
}

// NewErrorf returns a new error with a status code and formatted message.
func NewErrorf(statusCode int, msg string, args ...any) error {
	return Error{
		StatusCode: statusCode,
		Message:    fmt.Sprintf(msg, args...),
	}
}

// An Error is a Prometheus error.
type Error struct {
	StatusCode int
	Message    string
}

// Error returns the error message.
func (err Error) Error() string {
	return fmt.Sprintf("status_code: %d, msg=%s", err.StatusCode, err.Message)
}
