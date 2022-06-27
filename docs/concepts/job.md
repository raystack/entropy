# Job

A job is an action that needs to be performed by Entropy asynchronously. It has a kind which maps it to a JobFn and a payload that is passed to the JobFn. Jobs are picked by worker treads for execution.

The Job struct looks like this:

```
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
```

## Sanitise

Sanitise the job fields.

```
func (j *Job) Sanitise() error
```

## Attempt

Attempt attempts to safely invoke `fn` for this job. Handles success, failure and panic scenarios and updates the job with result in-place.

```
func (j *Job) Attempt(ctx context.Context, now time.Time, fn JobFn)
```