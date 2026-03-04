package ingestion

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrJobLeaseNotFound = errors.New("job lease not found")

type InProcessJobQueue struct {
	queue chan JobEnvelope

	mu        sync.Mutex
	inFlight  map[string]JobEnvelope
	receiptID uint64
}

func NewInProcessJobQueue(bufferSize int) *InProcessJobQueue {
	if bufferSize <= 0 {
		bufferSize = 64
	}

	return &InProcessJobQueue{
		queue:    make(chan JobEnvelope, bufferSize),
		inFlight: map[string]JobEnvelope{},
	}
}

func (q *InProcessJobQueue) Enqueue(ctx context.Context, job JobEnvelope) error {
	select {
	case q.queue <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *InProcessJobQueue) Reserve(ctx context.Context, wait time.Duration) (JobEnvelope, JobLease, bool, error) {
	reserveCtx := ctx
	var cancel context.CancelFunc
	if wait > 0 {
		reserveCtx, cancel = context.WithTimeout(ctx, wait)
		defer cancel()
	}

	select {
	case job := <-q.queue:
		lease := q.newLease(job.JobID)
		q.mu.Lock()
		q.inFlight[lease.Receipt] = job
		q.mu.Unlock()
		return job, lease, true, nil
	case <-reserveCtx.Done():
		if errors.Is(reserveCtx.Err(), context.DeadlineExceeded) {
			return JobEnvelope{}, JobLease{}, false, nil
		}
		return JobEnvelope{}, JobLease{}, false, reserveCtx.Err()
	}
}

func (q *InProcessJobQueue) Ack(_ context.Context, lease JobLease) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, ok := q.inFlight[lease.Receipt]; !ok {
		return ErrJobLeaseNotFound
	}

	delete(q.inFlight, lease.Receipt)
	return nil
}

func (q *InProcessJobQueue) Retry(ctx context.Context, lease JobLease, _ string, nextDelay time.Duration) error {
	q.mu.Lock()
	job, ok := q.inFlight[lease.Receipt]
	if ok {
		delete(q.inFlight, lease.Receipt)
	}
	q.mu.Unlock()

	if !ok {
		return ErrJobLeaseNotFound
	}

	if nextDelay <= 0 {
		return q.Enqueue(ctx, job)
	}

	go func(enqueued JobEnvelope, delay time.Duration) {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		<-timer.C
		_ = q.Enqueue(context.Background(), enqueued)
	}(job, nextDelay)

	return nil
}

func (q *InProcessJobQueue) Fail(_ context.Context, lease JobLease, _ string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, ok := q.inFlight[lease.Receipt]; !ok {
		return ErrJobLeaseNotFound
	}

	delete(q.inFlight, lease.Receipt)
	return nil
}

func (q *InProcessJobQueue) newLease(jobID string) JobLease {
	q.mu.Lock()
	q.receiptID++
	receipt := fmt.Sprintf("%s:%d", jobID, q.receiptID)
	q.mu.Unlock()

	return JobLease{
		JobID:       jobID,
		Receipt:     receipt,
		LeasedUntil: time.Now().UTC(),
	}
}

var _ JobQueue = (*InProcessJobQueue)(nil)
