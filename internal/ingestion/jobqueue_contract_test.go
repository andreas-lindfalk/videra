package ingestion

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInProcessJobQueueContract(t *testing.T) {
	runJobQueueContractSuite(t, func() JobQueue {
		return NewInProcessJobQueue(16)
	})
}

func runJobQueueContractSuite(t *testing.T, newQueue func() JobQueue) {
	t.Run("enqueue reserve ack", func(t *testing.T) {
		queue := newQueue()
		job := JobEnvelope{JobID: "job-1", SourcePath: "https://example.com/video-1.mp4", RequestedAt: time.Now().UTC()}

		err := queue.Enqueue(context.Background(), job)
		require.NoError(t, err)

		reserved, lease, ok, err := queue.Reserve(context.Background(), 100*time.Millisecond)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, job.JobID, reserved.JobID)
		require.Equal(t, job.SourcePath, reserved.SourcePath)
		require.NotEmpty(t, lease.Receipt)

		err = queue.Ack(context.Background(), lease)
		require.NoError(t, err)
	})

	t.Run("reserve timeout empty", func(t *testing.T) {
		queue := newQueue()

		_, _, ok, err := queue.Reserve(context.Background(), 20*time.Millisecond)
		require.NoError(t, err)
		require.False(t, ok)
	})

	t.Run("retry requeues job", func(t *testing.T) {
		queue := newQueue()
		job := JobEnvelope{JobID: "job-retry", SourcePath: "https://example.com/retry.mp4", RequestedAt: time.Now().UTC()}

		err := queue.Enqueue(context.Background(), job)
		require.NoError(t, err)

		firstJob, firstLease, ok, err := queue.Reserve(context.Background(), 100*time.Millisecond)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, job.JobID, firstJob.JobID)

		err = queue.Retry(context.Background(), firstLease, "transient", 10*time.Millisecond)
		require.NoError(t, err)

		retried, retryLease, ok, err := queue.Reserve(context.Background(), 200*time.Millisecond)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, job.JobID, retried.JobID)

		err = queue.Ack(context.Background(), retryLease)
		require.NoError(t, err)
	})

	t.Run("fail removes in flight lease", func(t *testing.T) {
		queue := newQueue()
		job := JobEnvelope{JobID: "job-fail", SourcePath: "https://example.com/fail.mp4", RequestedAt: time.Now().UTC()}

		err := queue.Enqueue(context.Background(), job)
		require.NoError(t, err)

		_, lease, ok, err := queue.Reserve(context.Background(), 100*time.Millisecond)
		require.NoError(t, err)
		require.True(t, ok)

		err = queue.Fail(context.Background(), lease, "terminal")
		require.NoError(t, err)

		err = queue.Ack(context.Background(), lease)
		require.ErrorIs(t, err, ErrJobLeaseNotFound)
	})

	t.Run("duplicate enqueue allowed", func(t *testing.T) {
		queue := newQueue()
		job := JobEnvelope{JobID: "job-dup", SourcePath: "https://example.com/dup.mp4", RequestedAt: time.Now().UTC()}

		err := queue.Enqueue(context.Background(), job)
		require.NoError(t, err)
		err = queue.Enqueue(context.Background(), job)
		require.NoError(t, err)

		first, firstLease, ok, err := queue.Reserve(context.Background(), 100*time.Millisecond)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, job.JobID, first.JobID)

		second, secondLease, ok, err := queue.Reserve(context.Background(), 100*time.Millisecond)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, job.JobID, second.JobID)

		err = queue.Ack(context.Background(), firstLease)
		require.NoError(t, err)
		err = queue.Ack(context.Background(), secondLease)
		require.NoError(t, err)
	})
}
