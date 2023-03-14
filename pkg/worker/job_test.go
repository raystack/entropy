package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/worker"
)

func TestJob_Attempt(t *testing.T) {
	cancelledCtx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()

	createdAt := time.Unix(1654081526, 0)
	frozenTime := time.Unix(1654082526, 0)

	table := []struct {
		title string
		ctx   context.Context
		job   worker.Job
		fn    worker.JobFn
		want  worker.Job
	}{
		{
			title: "ContextCancelled",
			ctx:   cancelledCtx,
			job: worker.Job{
				UpdatedAt:     createdAt,
				AttemptsDone:  0,
				LastAttemptAt: frozenTime,
				Result:        nil,
				LastError:     "",
			},
			fn: func(ctx context.Context, job worker.Job) ([]byte, error) {
				return nil, nil
			},
			want: worker.Job{
				Status:        worker.StatusPending,
				RunAt:         frozenTime.Add(5 * time.Second),
				UpdatedAt:     frozenTime,
				AttemptsDone:  1,
				LastAttemptAt: frozenTime,
				Result:        nil,
				LastError:     "cancelled: context deadline exceeded",
			},
		},
		{
			title: "Panic",
			job: worker.Job{
				UpdatedAt:     createdAt,
				AttemptsDone:  0,
				LastAttemptAt: frozenTime,
				Result:        nil,
				LastError:     "",
			},
			fn: func(ctx context.Context, job worker.Job) ([]byte, error) {
				panic("blown up")
			},
			want: worker.Job{
				Status:        worker.StatusPanic,
				UpdatedAt:     frozenTime,
				AttemptsDone:  1,
				LastAttemptAt: frozenTime,
				Result:        nil,
				LastError:     "panic: blown up",
			},
		},
		{
			title: "NonRetryableError",
			job: worker.Job{
				UpdatedAt:     createdAt,
				AttemptsDone:  0,
				LastAttemptAt: frozenTime,
				Result:        nil,
				LastError:     "",
			},
			fn: func(ctx context.Context, job worker.Job) ([]byte, error) {
				return nil, errors.New("a non-retryable error occurred")
			},
			want: worker.Job{
				Status:        worker.StatusFailed,
				UpdatedAt:     frozenTime,
				AttemptsDone:  1,
				LastAttemptAt: frozenTime,
				Result:        nil,
				LastError:     "a non-retryable error occurred",
			},
		},
		{
			title: "RetryableError",
			job: worker.Job{
				UpdatedAt:     createdAt,
				AttemptsDone:  0,
				LastAttemptAt: frozenTime,
				Result:        nil,
				LastError:     "",
			},
			fn: func(ctx context.Context, job worker.Job) ([]byte, error) {
				return nil, &worker.RetryableError{
					Cause:      errors.New("some retryable error occurred"),
					RetryAfter: 10 * time.Second,
				}
			},
			want: worker.Job{
				Status:        worker.StatusPending,
				RunAt:         frozenTime.Add(10 * time.Second),
				UpdatedAt:     frozenTime,
				AttemptsDone:  1,
				LastAttemptAt: frozenTime,
				Result:        nil,
				LastError:     "retryable-error: some retryable error occurred",
			},
		},
		{
			title: "Successful_FirstAttempt",
			job: worker.Job{
				UpdatedAt:     createdAt,
				AttemptsDone:  0,
				LastAttemptAt: frozenTime,
				Result:        nil,
				LastError:     "",
			},
			fn: func(ctx context.Context, job worker.Job) ([]byte, error) {
				return []byte("The answer to life is 42"), nil
			},
			want: worker.Job{
				Status:        worker.StatusDone,
				UpdatedAt:     frozenTime,
				AttemptsDone:  1,
				LastAttemptAt: frozenTime,
				Result:        []byte("The answer to life is 42"),
				LastError:     "",
			},
		},
		{
			title: "Successful_SecondAttempt",
			job: worker.Job{
				UpdatedAt:     createdAt,
				AttemptsDone:  1,
				LastAttemptAt: frozenTime,
				Result:        nil,
				LastError:     "attempt 1 failed with some retryable error",
			},
			fn: func(ctx context.Context, job worker.Job) ([]byte, error) {
				return []byte("The answer to life is 42"), nil
			},
			want: worker.Job{
				Status:        worker.StatusDone,
				UpdatedAt:     frozenTime,
				AttemptsDone:  2,
				LastAttemptAt: frozenTime,
				Result:        []byte("The answer to life is 42"),
				LastError:     "attempt 1 failed with some retryable error",
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			if tt.ctx == nil {
				tt.ctx = context.Background()
			}

			tt.job.Attempt(tt.ctx, frozenTime, tt.fn)

			assert.Truef(t, cmp.Equal(tt.want, tt.job), cmp.Diff(tt.want, tt.job))
		})
	}
}
