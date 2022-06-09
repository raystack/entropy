package worker

import (
	"time"

	"go.uber.org/zap"
)

func WithJobKind(kind string, fn JobFn) Option {
	return func(w *Worker) error {
		return w.Register(kind, fn)
	}
}

func WithLogger(l *zap.Logger) Option {
	return func(w *Worker) error {
		if l == nil {
			l = zap.NewNop()
		}
		w.logger = l
		return nil
	}
}

func WithRunConfig(workers int, pollInterval time.Duration) Option {
	return func(w *Worker) error {
		if workers == 0 {
			workers = 1
		}

		const minPollInterval = 100 * time.Millisecond
		if pollInterval < minPollInterval {
			pollInterval = minPollInterval
		}

		w.pollInt = pollInterval
		w.workers = workers
		return nil
	}
}

func withDefaults(opts []Option) []Option {
	return append([]Option{
		WithLogger(nil),
		WithRunConfig(1, 1*time.Second),
	}, opts...)
}
