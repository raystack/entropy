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

// RunError can be returned by a JobFn on failure to process a job, to provide
// instructions to the JobQueue about the next steps to be taken.
type RunError struct {
	JobID string
	Cause error

	// RetryAfter indicates when to retry this failed job again. Zero-value
	// indicates no-retry.
	RetryAfter time.Duration
}

func (re *RunError) ShouldRetry() bool { return re.RetryAfter > 0 }

func (re *RunError) Error() string {
	return fmt.Sprintf("failed to run job '%s': %v", re.JobID, re.Cause)
}
