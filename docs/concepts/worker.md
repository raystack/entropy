# Worker

Worker provides asynchronous job processing using a job-queue.

The worker struct looks like this:

```
type Worker struct {
	workers int
	pollInt time.Duration

	queue  JobQueue
	logger *zap.Logger

	mu       sync.RWMutex
	handlers map[string]JobFn
}
```

- ***Workers*** defines the number of worker threads that will run parallelly.
- ***JobQueue*** is a queue supporting Job Enqueues and Dequeues.
- ***JobFn*** are functions invoked by the Worker for ready jobs.
- ***Handlers*** is a map of JobFn which registers differet kind of jobs with their respective JobFn.

The JobQueue interface looks like:

```
type JobQueue interface {
	// Enqueue all jobs. Enqueue must ensure all-or-nothing behaviour.
	// Jobs with zero-value or historical value for ReadyAt must be
	// executed immediately.
	Enqueue(ctx context.Context, jobs ...Job) error

	// Dequeue one job having one of the given kinds and invoke `fn`.
	// The job should be 'locked' until `fn` returns. Refer DequeueFn.
	Dequeue(ctx context.Context, kinds []string, fn DequeueFn) error
}
```

## Register

Register registers a job-kind and the function that should be invoked for handling it. Returns ErrKindExists if the kind already registered.

```
func (w *Worker) Register(kind string, h JobFn) error
```

## Enqueue

Enqueue enqueues all jobs for processing. It internally sanitize a job and validate the job's kind, after will it enqueues to job into it's JobQueue.

```
func (w *Worker) Enqueue(ctx context.Context, jobs ...Job) error
```

## Run (Dequeue)

Run starts the worker threads that dequeue and process ready jobs. Run blocks until all workers exit or context is cancelled. Context cancellation will do graceful shutdown of the worker threads.

```
func (w *Worker) Run(baseCtx context.Context) error
```

A running worker dequeues a job by passing a dequeueFn which makes an attempt to run the job. It triggers the Attempt function of a job, along with the JobFn based on it's kind.