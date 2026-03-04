//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/andreas-lindfalk/videra/internal/ingestion"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	natsPort  = nat.Port("4222/tcp")
	redisPort = nat.Port("6379/tcp")
)

func TestNATSJetStreamJobQueueContractIntegration(t *testing.T) {
	ctx := context.Background()
	natsContainer, endpoint := startNATSJetStreamContainer(t, ctx)
	t.Cleanup(func() {
		_ = natsContainer.Terminate(ctx)
	})

	factory := func(t *testing.T) (ingestion.JobQueue, func()) {
		t.Helper()
		suffix := fmt.Sprintf("%d", time.Now().UnixNano())
		queue, err := ingestion.NewNATSJetStreamJobQueue(ingestion.NATSJetStreamQueueConfig{
			URL:      "nats://" + endpoint,
			Stream:   "videra_jobs_" + suffix,
			Subject:  "videra.jobs." + suffix,
			Consumer: "videra-consumer-" + suffix,
		})
		require.NoError(t, err)
		return queue, func() { queue.Close() }
	}

	runExternalQueueContractSuite(t, factory)
}

func TestRedisStreamsJobQueueContractIntegration(t *testing.T) {
	ctx := context.Background()
	redisContainer, endpoint := startRedisContainer(t, ctx)
	t.Cleanup(func() {
		_ = redisContainer.Terminate(ctx)
	})

	factory := func(t *testing.T) (ingestion.JobQueue, func()) {
		t.Helper()
		suffix := fmt.Sprintf("%d", time.Now().UnixNano())
		queue, err := ingestion.NewRedisStreamsJobQueue(ingestion.RedisStreamsQueueConfig{
			Addr:     endpoint,
			Stream:   "videra:index:jobs:" + suffix,
			Group:    "videra-workers-" + suffix,
			Consumer: "videra-consumer-" + suffix,
		})
		require.NoError(t, err)
		return queue, func() { _ = queue.Close() }
	}

	runExternalQueueContractSuite(t, factory)
}

func runExternalQueueContractSuite(t *testing.T, newQueue func(*testing.T) (ingestion.JobQueue, func())) {
	t.Run("enqueue reserve ack", func(t *testing.T) {
		queue, cleanup := newQueue(t)
		defer cleanup()

		job := ingestion.JobEnvelope{JobID: "job-1", SourcePath: "https://example.com/video-1.mp4", RequestedAt: time.Now().UTC()}
		err := queue.Enqueue(context.Background(), job)
		require.NoError(t, err)

		reserved, lease, ok, err := queue.Reserve(context.Background(), 2*time.Second)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, job.JobID, reserved.JobID)

		err = queue.Ack(context.Background(), lease)
		require.NoError(t, err)
	})

	t.Run("reserve timeout empty", func(t *testing.T) {
		queue, cleanup := newQueue(t)
		defer cleanup()

		_, _, ok, err := queue.Reserve(context.Background(), 50*time.Millisecond)
		require.NoError(t, err)
		require.False(t, ok)
	})

	t.Run("retry requeues job", func(t *testing.T) {
		queue, cleanup := newQueue(t)
		defer cleanup()

		job := ingestion.JobEnvelope{JobID: "job-retry", SourcePath: "https://example.com/retry.mp4", RequestedAt: time.Now().UTC()}
		err := queue.Enqueue(context.Background(), job)
		require.NoError(t, err)

		_, lease, ok, err := queue.Reserve(context.Background(), 2*time.Second)
		require.NoError(t, err)
		require.True(t, ok)

		err = queue.Retry(context.Background(), lease, "transient", 20*time.Millisecond)
		require.NoError(t, err)

		retried, retryLease, ok, err := queue.Reserve(context.Background(), 2*time.Second)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, job.JobID, retried.JobID)

		err = queue.Ack(context.Background(), retryLease)
		require.NoError(t, err)
	})

	t.Run("fail removes in flight lease", func(t *testing.T) {
		queue, cleanup := newQueue(t)
		defer cleanup()

		job := ingestion.JobEnvelope{JobID: "job-fail", SourcePath: "https://example.com/fail.mp4", RequestedAt: time.Now().UTC()}
		err := queue.Enqueue(context.Background(), job)
		require.NoError(t, err)

		_, lease, ok, err := queue.Reserve(context.Background(), 2*time.Second)
		require.NoError(t, err)
		require.True(t, ok)

		err = queue.Fail(context.Background(), lease, "terminal")
		require.NoError(t, err)

		err = queue.Ack(context.Background(), lease)
		require.ErrorIs(t, err, ingestion.ErrJobLeaseNotFound)
	})

	t.Run("duplicate enqueue allowed", func(t *testing.T) {
		queue, cleanup := newQueue(t)
		defer cleanup()

		job := ingestion.JobEnvelope{JobID: "job-dup", SourcePath: "https://example.com/dup.mp4", RequestedAt: time.Now().UTC()}
		err := queue.Enqueue(context.Background(), job)
		require.NoError(t, err)
		err = queue.Enqueue(context.Background(), job)
		require.NoError(t, err)

		first, firstLease, ok, err := queue.Reserve(context.Background(), 2*time.Second)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, job.JobID, first.JobID)

		second, secondLease, ok, err := queue.Reserve(context.Background(), 2*time.Second)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, job.JobID, second.JobID)

		err = queue.Ack(context.Background(), firstLease)
		require.NoError(t, err)
		err = queue.Ack(context.Background(), secondLease)
		require.NoError(t, err)
	})
}

func startNATSJetStreamContainer(t *testing.T, ctx context.Context) (testcontainers.Container, string) {
	t.Helper()
	testcontainers.SkipIfProviderIsNotHealthy(t)

	container, err := testcontainers.Run(ctx,
		"nats:2.10-alpine",
		testcontainers.WithExposedPorts(string(natsPort)),
		testcontainers.WithCmd("-js"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(natsPort)),
	)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	mappedPort, err := container.MappedPort(ctx, natsPort)
	require.NoError(t, err)

	return container, fmt.Sprintf("%s:%s", host, mappedPort.Port())
}

func startRedisContainer(t *testing.T, ctx context.Context) (testcontainers.Container, string) {
	t.Helper()
	testcontainers.SkipIfProviderIsNotHealthy(t)

	container, err := testcontainers.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithExposedPorts(string(redisPort)),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(redisPort)),
	)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	mappedPort, err := container.MappedPort(ctx, redisPort)
	require.NoError(t, err)

	return container, fmt.Sprintf("%s:%s", host, mappedPort.Port())
}
