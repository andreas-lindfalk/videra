package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStreamsQueueConfig struct {
	Addr     string
	Password string
	DB       int
	Stream   string
	Group    string
	Consumer string
}

type redisInFlightJob struct {
	messageID string
	job       JobEnvelope
}

type RedisStreamsJobQueue struct {
	client   *redis.Client
	stream   string
	group    string
	consumer string

	mu       sync.Mutex
	inFlight map[string]redisInFlightJob
}

func NewRedisStreamsJobQueue(cfg RedisStreamsQueueConfig) (*RedisStreamsJobQueue, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("redis addr is required")
	}
	if cfg.Stream == "" {
		return nil, fmt.Errorf("redis stream is required")
	}
	if cfg.Group == "" {
		return nil, fmt.Errorf("redis group is required")
	}
	if cfg.Consumer == "" {
		return nil, fmt.Errorf("redis consumer is required")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	if err := client.XGroupCreateMkStream(ctx, cfg.Stream, cfg.Group, "$").Err(); err != nil {
		if !strings.Contains(err.Error(), "BUSYGROUP") {
			_ = client.Close()
			return nil, err
		}
	}

	return &RedisStreamsJobQueue{
		client:   client,
		stream:   cfg.Stream,
		group:    cfg.Group,
		consumer: cfg.Consumer,
		inFlight: map[string]redisInFlightJob{},
	}, nil
}

func (q *RedisStreamsJobQueue) Enqueue(ctx context.Context, job JobEnvelope) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: q.stream,
		Values: map[string]any{"job": payload},
	}).Err()
}

func (q *RedisStreamsJobQueue) Reserve(ctx context.Context, wait time.Duration) (JobEnvelope, JobLease, bool, error) {
	if wait <= 0 {
		wait = 5 * time.Second
	}

	result, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    q.group,
		Consumer: q.consumer,
		Streams:  []string{q.stream, ">"},
		Count:    1,
		Block:    wait,
	}).Result()
	if err == redis.Nil {
		return JobEnvelope{}, JobLease{}, false, nil
	}
	if err != nil {
		return JobEnvelope{}, JobLease{}, false, err
	}
	if len(result) == 0 || len(result[0].Messages) == 0 {
		return JobEnvelope{}, JobLease{}, false, nil
	}

	message := result[0].Messages[0]
	job, err := decodeRedisJobValue(message.Values["job"])
	if err != nil {
		_ = q.client.XAck(ctx, q.stream, q.group, message.ID).Err()
		return JobEnvelope{}, JobLease{}, false, err
	}

	receipt := message.ID
	q.mu.Lock()
	q.inFlight[receipt] = redisInFlightJob{messageID: message.ID, job: job}
	q.mu.Unlock()

	lease := JobLease{JobID: job.JobID, Receipt: receipt, Attempt: job.Attempt + 1, LeasedUntil: time.Now().UTC()}
	return job, lease, true, nil
}

func (q *RedisStreamsJobQueue) Ack(ctx context.Context, lease JobLease) error {
	inFlight, err := q.popInFlight(lease.Receipt)
	if err != nil {
		return err
	}
	return q.client.XAck(ctx, q.stream, q.group, inFlight.messageID).Err()
}

func (q *RedisStreamsJobQueue) Retry(ctx context.Context, lease JobLease, _ string, nextDelay time.Duration) error {
	inFlight, err := q.popInFlight(lease.Receipt)
	if err != nil {
		return err
	}

	if err := q.client.XAck(ctx, q.stream, q.group, inFlight.messageID).Err(); err != nil {
		return err
	}

	inFlight.job.Attempt++

	if nextDelay <= 0 {
		return q.Enqueue(ctx, inFlight.job)
	}

	go func(job JobEnvelope, delay time.Duration) {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		<-timer.C
		_ = q.Enqueue(context.Background(), job)
	}(inFlight.job, nextDelay)

	return nil
}

func (q *RedisStreamsJobQueue) Fail(ctx context.Context, lease JobLease, _ string) error {
	inFlight, err := q.popInFlight(lease.Receipt)
	if err != nil {
		return err
	}
	return q.client.XAck(ctx, q.stream, q.group, inFlight.messageID).Err()
}

func (q *RedisStreamsJobQueue) Close() error {
	if q.client == nil {
		return nil
	}
	return q.client.Close()
}

func (q *RedisStreamsJobQueue) popInFlight(receipt string) (redisInFlightJob, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	job, ok := q.inFlight[receipt]
	if !ok {
		return redisInFlightJob{}, ErrJobLeaseNotFound
	}
	delete(q.inFlight, receipt)
	return job, nil
}

func decodeRedisJobValue(value any) (JobEnvelope, error) {
	var payload []byte
	switch typed := value.(type) {
	case string:
		payload = []byte(typed)
	case []byte:
		payload = typed
	default:
		return JobEnvelope{}, fmt.Errorf("redis job payload has unsupported type: %T", value)
	}

	var job JobEnvelope
	if err := json.Unmarshal(payload, &job); err != nil {
		return JobEnvelope{}, err
	}

	return job, nil
}

var _ JobQueue = (*RedisStreamsJobQueue)(nil)
