package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"

	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/worker"
	"github.com/goto/entropy/pkg/worker/pgq"
)

var (
	jID     = flag.String("id", "test", "Job ID")
	kind    = flag.String("kind", "print", "Job kind")
	count   = flag.Int("count", 1, "Number of jobs to create")
	after   = flag.Duration("after", 0, "Enqueue a job after")
	payload = flag.String("payload", "", "Payload for the job")

	runWorker = flag.Bool("worker", false, "Run in worker mode")
	queueName = flag.String("queue", "demo", "Queue name")
	pgConStr  = flag.String("pg", "postgresql://postgres@localhost:5432/postgres?sslmode=disable", "PostgreSQL connection string")
)

func main() {
	flag.Parse()

	lg, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	q, err := pgq.Open(*pgConStr, *queueName)
	if err != nil {
		panic(err)
	}

	opts := []worker.Option{
		worker.WithJobKind("test", testJobFn),
		worker.WithLogger(lg),
	}

	w, err := worker.New(q, opts...)
	if err != nil {
		panic(err)
	}

	if *runWorker {
		if err := w.Run(context.Background()); err != nil {
			panic(err)
		}
	} else {
		for i := 0; i < *count; i++ {
			log.Println(w.Enqueue(context.Background(), worker.Job{
				ID:      fmt.Sprintf("%s_%d", *jID, i),
				Kind:    *kind,
				Payload: []byte(*payload),
				RunAt:   time.Now().Add(*after),
			}))
		}
	}
}

func testJobFn(_ context.Context, job worker.Job) ([]byte, error) {
	const maxAttempts = 3
	const attemptBackoff = 5 * time.Second

	switch string(job.Payload) {
	case "fail_after_3":
		if job.AttemptsDone < maxAttempts {
			return nil, &worker.RetryableError{
				Cause:      errors.New("fake error [retryable]"),
				RetryAfter: attemptBackoff,
			}
		}
		return nil, errors.New("fake error [permanent]")

	case "panic":
		panic("simulated panic")

	case "fail":
		return nil, errors.New("fake error [permanent]")

	default:
		log.Printf("Test Job Says Hello! (attempt=%d)\n", job.AttemptsDone+1)
		return []byte("job is done"), nil
	}
}
