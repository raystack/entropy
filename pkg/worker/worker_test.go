package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/worker"
	"github.com/goto/entropy/pkg/worker/mocks"
)

func Test_New(t *testing.T) {
	t.Parallel()

	q := &mocks.JobQueue{}

	t.Run("DuplicateKind", func(t *testing.T) {
		w, err := worker.New(q,
			worker.WithJobKind("test", nil),
			worker.WithJobKind("test", nil),
		)
		assert.Error(t, err)
		assert.EqualError(t, err, "handler for given kind exists: kind 'test'")
		assert.Nil(t, w)
	})

	t.Run("Success", func(t *testing.T) {
		w, err := worker.New(q,
			worker.WithJobKind("test", nil),
			worker.WithRunConfig(0, 0),
		)
		assert.NoError(t, err)
		assert.NotNil(t, w)
	})
}

func TestWorker_Enqueue(t *testing.T) {
	t.Parallel()

	table := []struct {
		title   string
		queue   func(t *testing.T) worker.JobQueue
		opts    []worker.Option
		jobs    []worker.Job
		wantErr error
	}{
		{
			title: "InvalidJobID",
			queue: func(t *testing.T) worker.JobQueue {
				t.Helper()
				return &mocks.JobQueue{}
			},
			opts: []worker.Option{
				worker.WithJobKind("test", func(ctx context.Context, job worker.Job) ([]byte, error) {
					return nil, nil
				}),
			},
			jobs: []worker.Job{
				{ID: "", Kind: "test"},
			},
			wantErr: worker.ErrInvalidJob,
		},
		{
			title: "UnknownJobKind",
			queue: func(t *testing.T) worker.JobQueue {
				t.Helper()
				return &mocks.JobQueue{}
			},
			opts: []worker.Option{
				worker.WithJobKind("test", func(ctx context.Context, job worker.Job) ([]byte, error) {
					return nil, nil
				}),
			},
			jobs: []worker.Job{
				{ID: "foo1", Kind: "test"},
				{ID: "foo2", Kind: "unknown_kind"},
			},
			wantErr: worker.ErrUnknownKind,
		},
		{
			title: "Success",
			queue: func(t *testing.T) worker.JobQueue {
				t.Helper()

				q := &mocks.JobQueue{}
				q.EXPECT().
					Enqueue(mock.Anything, mock.Anything, mock.Anything).
					Run(func(ctx context.Context, jobs ...worker.Job) {
						require.Len(t, jobs, 2)
						assert.Equal(t, "foo1", jobs[0].ID)
						assert.Equal(t, "foo2", jobs[1].ID)
					}).
					Return(nil).
					Once()
				return q
			},
			opts: []worker.Option{
				worker.WithJobKind("test", func(ctx context.Context, job worker.Job) ([]byte, error) {
					return nil, nil
				}),
			},
			jobs: []worker.Job{
				{ID: "foo1", Kind: "test"},
				{ID: "foo2", Kind: "test"},
			},
			wantErr: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			w, err := worker.New(tt.queue(t), tt.opts...)
			require.NoError(t, err)
			require.NotNil(t, w)

			got := w.Enqueue(context.Background(), tt.jobs...)
			if tt.wantErr != nil {
				assert.Error(t, got)
				assert.True(t, errors.Is(got, tt.wantErr))
			} else {
				assert.NoError(t, got)
			}
		})
	}
}

func TestWorker_Run(t *testing.T) {
	t.Parallel()

	opts := []worker.Option{
		worker.WithJobKind("test", func(ctx context.Context, job worker.Job) ([]byte, error) {
			return []byte("test_result"), nil
		}),
		worker.WithRunConfig(1, 10*time.Millisecond),
	}

	t.Run("ContextCancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // immediately cancel the context.

		q := &mocks.JobQueue{}

		w, err := worker.New(q, opts...)
		require.NoError(t, err)
		require.NotNil(t, w)

		got := w.Run(ctx)
		assert.NoError(t, got)
	})

	t.Run("ContextDeadline", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
		defer cancel()

		q := &mocks.JobQueue{}

		w, err := worker.New(q, opts...)
		require.NoError(t, err)
		require.NotNil(t, w)

		got := w.Run(ctx)
		assert.NoError(t, got)
	})

	t.Run("Dequeue_ReturnsUnknownKind", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dequeued := 0
		sampleJob := worker.Job{
			ID:   "test1",
			Kind: "unknown_kind",
		}

		q := &mocks.JobQueue{}
		q.EXPECT().
			Dequeue(mock.Anything, []string{"test"}, mock.Anything).
			Run(func(ctx context.Context, kinds []string, fn worker.DequeueFn) {
				_, err := fn(ctx, sampleJob)
				assert.Error(t, err)
				assert.EqualError(t, err, "retryable-error: job kind is invalid")

				dequeued++
				cancel() // cancel context to stop the worker.
			}).
			Return(nil)

		w, err := worker.New(q, opts...)
		require.NoError(t, err)
		require.NotNil(t, w)

		got := w.Run(ctx)
		assert.NoError(t, got)
		assert.Equal(t, 1, dequeued)
	})

	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dequeued := 0
		sampleJob := worker.Job{
			ID:   "test1",
			Kind: "test",
		}

		q := &mocks.JobQueue{}
		q.EXPECT().
			Dequeue(mock.Anything, []string{"test"}, mock.Anything).
			Run(func(ctx context.Context, kinds []string, fn worker.DequeueFn) {
				res, err := fn(ctx, sampleJob)
				assert.NoError(t, err)
				assert.Equal(t, []byte("test_result"), res.Result)

				dequeued++
				cancel() // cancel context to stop the worker.
			}).
			Return(nil)

		w, err := worker.New(q, opts...)
		require.NoError(t, err)
		require.NotNil(t, w)

		got := w.Run(ctx)
		assert.NoError(t, got)
		assert.Equal(t, 1, dequeued)
	})
}
