package worker

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidJob  = errors.New("job is not valid")
	ErrKindExists  = errors.New("handler for given kind exists")
	ErrUnknownKind = errors.New("job kind is invalid")
)

// RetryableError can be returned by a JobFn to instruct the worker to attempt
// retry after time specified by the RetryAfter field.
type RetryableError struct {
	Cause      error
	RetryAfter time.Duration
}

func (re *RetryableError) Error() string {
	return fmt.Sprintf("retryable-error: %v", re.Cause)
}
