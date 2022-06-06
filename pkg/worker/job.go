package worker

//go:generate mockery --name=JobQueue -r --case underscore --with-expecter --structname JobQueue --filename=job_queue.go --output=./mocks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const minRetryBackoff = 5 * time.Second

const (
	// StatusDone indicates the Job is successfully finished.
	StatusDone = "DONE"

	// StatusPanic indicates there was a panic during job-execution.
	// This is a terminal status and will NOT be retried.
	StatusPanic = "PANIC"

	// StatusFailed indicates job failed to succeed even after retries.
	// This is a terminal status and will NOT be retried.
	StatusFailed = "FAILED"

	// StatusPending indicates at-least 1 attempt is still pending.
	StatusPending = "PENDING"
)

var (
	ErrInvalidJob  = errors.New("job is not valid")
	ErrKindExists  = errors.New("handler for given kind exists")
	ErrUnknownKind = errors.New("job kind is invalid")
)

// Job represents the specification for async processing and also maintains
// the progress so far.
type Job struct {
	// Specification of the job.
	ID      string    `json:"id"`
	Kind    string    `json:"kind"`
	RunAt   time.Time `json:"run_at"`
	Payload []byte    `json:"payload"`

	// Internal metadata.
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Execution information.
	Result        []byte    `json:"result,omitempty"`
	AttemptsDone  int64     `json:"attempts_done"`
	LastAttemptAt time.Time `json:"last_attempt_at,omitempty"`
	LastError     string    `json:"last_error,omitempty"`
}

// JobQueue represents a special queue that holds jobs and releases them via
// Dequeue() only after their RunAt time.
type JobQueue interface {
	// Enqueue all jobs. Enqueue must ensure all-or-nothing behaviour.
	// Jobs with zero-value or historical value for ReadyAt must be
	// executed immediately.
	Enqueue(ctx context.Context, jobs ...Job) error

	// Dequeue one job having one of the given kinds and invoke `fn`.
	// The job should be 'locked' until `fn` returns. Refer DequeueFn.
	Dequeue(ctx context.Context, kinds []string, fn DequeueFn) error
}

// DequeueFn is invoked by the JobQueue for ready jobs. It is responsible for
// handling a ready job and returning the updated version after the attempt.
type DequeueFn func(ctx context.Context, j Job) (*Job, error)

// RetryableError can be returned by a JobFn to instruct the worker to attempt
// retry after time specified by the RetryAfter field. RetryAfter can have min
// of 5 seconds.
type RetryableError struct {
	Cause      error
	RetryAfter time.Duration
}

func (j *Job) Sanitise() error {
	now := time.Now()

	j.ID = strings.TrimSpace(j.ID)
	j.Kind = strings.TrimSpace(strings.ToLower(j.Kind))

	if j.ID == "" {
		return fmt.Errorf("%w: job id must be set", ErrInvalidJob)
	}

	if j.Kind == "" {
		return fmt.Errorf("%w: job kind must be set", ErrInvalidJob)
	}

	j.Status = StatusPending
	j.CreatedAt = now
	j.UpdatedAt = now

	if j.RunAt.IsZero() {
		j.RunAt = now
	}

	j.AttemptsDone = 0
	j.LastAttemptAt = time.Time{}
	j.LastError = ""
	return nil
}

// Attempt attempts to safely invoke `fn` for this job. Handles success, failure
// and panic scenarios and updates the job with result in-place.
func (j *Job) Attempt(ctx context.Context, now time.Time, fn JobFn) {
	defer func() {
		if v := recover(); v != nil {
			j.LastError = fmt.Sprintf("panic: %v", v)
			j.Status = StatusPanic
		}

		j.AttemptsDone++
		j.LastAttemptAt = now
		j.UpdatedAt = now
	}()

	select {
	case <-ctx.Done():
		j.Status = StatusPending
		j.RunAt = now.Add(minRetryBackoff)
		j.LastError = fmt.Sprintf("cancelled: %v", ctx.Err())

	default:
		res, err := fn(ctx, *j)
		if err != nil {
			re := &RetryableError{}
			if errors.As(err, &re) {
				j.RunAt = now.Add(re.backoff())
				j.LastError = re.Error()
				j.Status = StatusPending
			} else {
				j.LastError = err.Error()
				j.Status = StatusFailed
			}
		} else {
			j.Result = res
			j.Status = StatusDone
		}
	}
}

func (re *RetryableError) Error() string {
	return fmt.Sprintf("retryable-error: %v", re.Cause)
}

func (re RetryableError) backoff() time.Duration {
	if re.RetryAfter <= minRetryBackoff {
		return minRetryBackoff
	}
	return re.RetryAfter
}
