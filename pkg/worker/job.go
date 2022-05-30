package worker

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	StatusDone    = "DONE"
	StatusFailed  = "FAILED"
	StatusPending = "PENDING"
)

// Job represents the specification for async processing and also maintains
// the progress so far.
type Job struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	RunAt         time.Time `json:"run_at"`
	Payload       []byte    `json:"payload"`
	AttemptsDone  int64     `json:"attempts_done"`
	LastAttemptAt time.Time `json:"last_attempt_at"`
	LastError     string    `json:"last_error"`
}

// JobQueue represents a special queue that holds jobs and releases them via
// Dequeue() only after their ReadyAt time.
type JobQueue interface {
	// Enqueue all jobs. Enqueue must ensure all-or-nothing behaviour.
	// Jobs with zero-value or historical value for ReadyAt must be
	// executed immediately.
	Enqueue(ctx context.Context, jobs ...Job) error

	// Dequeue one job having one of the given kinds and invoke `fn`.
	// The job should be 'locked' until `fn` returns. Refer JobFn.
	Dequeue(ctx context.Context, kinds []string, fn JobFn) error
}

// JobFn is invoked by the JobQueue for ready jobs. If it returns no error,
// job will be marked with StatusDone by the JobQueue. If it returns error
// job should be retried or marked with StatusFailed accordingly.
// Refer RunError for expected behaviour on error.
type JobFn func(ctx context.Context, job Job) error

func (j *Job) sanitise() error {
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
