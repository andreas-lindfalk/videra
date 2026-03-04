package ingestion

import (
	"context"
	"time"
)

type JobEnvelope struct {
	JobID       string
	SourcePath  string
	RequestedAt time.Time
	Attempt     int
	MaxAttempts int
}

type JobLease struct {
	JobID       string
	Receipt     string
	Attempt     int
	LeasedUntil time.Time
}

type JobQueue interface {
	Enqueue(ctx context.Context, job JobEnvelope) error
	Reserve(ctx context.Context, wait time.Duration) (JobEnvelope, JobLease, bool, error)
	Ack(ctx context.Context, lease JobLease) error
	Retry(ctx context.Context, lease JobLease, cause string, nextDelay time.Duration) error
	Fail(ctx context.Context, lease JobLease, cause string) error
}
