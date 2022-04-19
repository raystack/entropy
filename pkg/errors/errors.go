package errors

import (
	"fmt"
	"strings"
)

// Common timer domain errors. Use `ErrX.WithCausef()` to clone and add context.
var (
	ErrInvalid     = Error{Code: "bad_request", Message: "Request is not valid"}
	ErrNotFound    = Error{Code: "not_found", Message: "Requested resource not found"}
	ErrConflict    = Error{Code: "conflict", Message: "A resource with conflicting identifier exists"}
	ErrInternal    = Error{Code: "internal_error", Message: "Some unexpected error occurred"}
	ErrUnsupported = Error{Code: "unsupported", Message: "Requested feature is not supported"}
)

// Error represents any error returned by the Timer components along with any
// relevant context.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   string `json:"cause,omitempty"`
}

// WithCausef returns clone of err with the cause added.
func (err Error) WithCausef(format string, args ...interface{}) Error {
	cloned := err
	cloned.Cause = fmt.Sprintf(format, args...)
	return cloned
}

// WithMsgf returns a clone of the error with message set.
func (err Error) WithMsgf(format string, args ...interface{}) Error {
	cloned := err
	cloned.Message = fmt.Sprintf(format, args...)
	return cloned
}

// Is checks if 'other' is of type Error and has the same code.
// See https://blog.golang.org/go1.13-errors.
func (err Error) Is(other error) bool {
	oe, ok := other.(Error)
	equivalent := ok && oe.Code == err.Code
	return equivalent
}

func (err Error) Error() string {
	if err.Cause == "" {
		return fmt.Sprintf("%s: %s", err.Code, strings.ToLower(err.Message))
	}
	return fmt.Sprintf("%s: %s", err.Code, err.Cause)
}

// Errorf returns a formatted error similar to `fmt.Errorf` but uses the
// Error type defined in this package. returned value is equivalent to
// ErrInternal (i.e., errors.Is(retVal, ErrInternal) = true).
func Errorf(format string, args ...interface{}) error {
	return ErrInternal.WithMsgf(format, args...)
}
