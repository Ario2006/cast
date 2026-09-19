// Package apperror defines errors that can be presented clearly at the CLI boundary.
package apperror

import "fmt"

// Error describes a user-actionable application failure.
type Error struct {
	Message string
	Hint    string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *Error) Unwrap() error { return e.Cause }

// New creates an error with a concise next step for the user.
func New(message, hint string) *Error {
	return &Error{Message: message, Hint: hint}
}

// Wrap retains a cause for diagnostics while keeping the CLI message actionable.
func Wrap(message, hint string, cause error) *Error {
	return &Error{Message: message, Hint: hint, Cause: cause}
}
