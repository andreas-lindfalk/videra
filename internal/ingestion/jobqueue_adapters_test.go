package ingestion

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewNATSJetStreamJobQueueValidatesConfig(t *testing.T) {
	_, err := NewNATSJetStreamJobQueue(NATSJetStreamQueueConfig{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "nats url is required")

	_, err = NewNATSJetStreamJobQueue(NATSJetStreamQueueConfig{URL: "nats://127.0.0.1:4222"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "nats stream is required")

	_, err = NewNATSJetStreamJobQueue(NATSJetStreamQueueConfig{URL: "nats://127.0.0.1:4222", Stream: "jobs"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "nats subject is required")

	_, err = NewNATSJetStreamJobQueue(NATSJetStreamQueueConfig{URL: "nats://127.0.0.1:4222", Stream: "jobs", Subject: "videra.jobs"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "nats consumer is required")
}

func TestNewRedisStreamsJobQueueValidatesConfig(t *testing.T) {
	_, err := NewRedisStreamsJobQueue(RedisStreamsQueueConfig{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "redis addr is required")

	_, err = NewRedisStreamsJobQueue(RedisStreamsQueueConfig{Addr: "127.0.0.1:6379"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "redis stream is required")

	_, err = NewRedisStreamsJobQueue(RedisStreamsQueueConfig{Addr: "127.0.0.1:6379", Stream: "videra:jobs"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "redis group is required")

	_, err = NewRedisStreamsJobQueue(RedisStreamsQueueConfig{Addr: "127.0.0.1:6379", Stream: "videra:jobs", Group: "workers"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "redis consumer is required")
}

func TestDecodeRedisJobValue(t *testing.T) {
	job := JobEnvelope{JobID: "job-1", SourcePath: "https://example.com/video.mp4", RequestedAt: time.Now().UTC(), Attempt: 1, MaxAttempts: 3}
	payload, err := json.Marshal(job)
	require.NoError(t, err)

	decodedFromString, err := decodeRedisJobValue(string(payload))
	require.NoError(t, err)
	require.Equal(t, job.JobID, decodedFromString.JobID)
	require.Equal(t, job.SourcePath, decodedFromString.SourcePath)
	require.Equal(t, job.Attempt, decodedFromString.Attempt)
	require.Equal(t, job.MaxAttempts, decodedFromString.MaxAttempts)

	decodedFromBytes, err := decodeRedisJobValue(payload)
	require.NoError(t, err)
	require.Equal(t, job.JobID, decodedFromBytes.JobID)

	_, err = decodeRedisJobValue(123)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported type")
}
