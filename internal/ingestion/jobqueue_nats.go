package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

type NATSJetStreamQueueConfig struct {
	URL      string
	Stream   string
	Subject  string
	Consumer string
}

type NATSJetStreamJobQueue struct {
	nc      *nats.Conn
	js      nats.JetStreamContext
	subject string
	sub     *nats.Subscription

	mu       sync.Mutex
	inFlight map[string]*nats.Msg
}

func NewNATSJetStreamJobQueue(cfg NATSJetStreamQueueConfig) (*NATSJetStreamJobQueue, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("nats url is required")
	}
	if cfg.Stream == "" {
		return nil, fmt.Errorf("nats stream is required")
	}
	if cfg.Subject == "" {
		return nil, fmt.Errorf("nats subject is required")
	}
	if cfg.Consumer == "" {
		return nil, fmt.Errorf("nats consumer is required")
	}

	nc, err := nats.Connect(cfg.URL)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}

	if _, err := js.StreamInfo(cfg.Stream); err != nil {
		if !errors.Is(err, nats.ErrStreamNotFound) {
			nc.Close()
			return nil, err
		}
		if _, addErr := js.AddStream(&nats.StreamConfig{
			Name:      cfg.Stream,
			Subjects:  []string{cfg.Subject},
			Retention: nats.WorkQueuePolicy,
			Storage:   nats.FileStorage,
		}); addErr != nil {
			nc.Close()
			return nil, addErr
		}
	}

	if _, err := js.ConsumerInfo(cfg.Stream, cfg.Consumer); err != nil {
		if !errors.Is(err, nats.ErrConsumerNotFound) {
			nc.Close()
			return nil, err
		}
		if _, addErr := js.AddConsumer(cfg.Stream, &nats.ConsumerConfig{
			Durable:   cfg.Consumer,
			AckPolicy: nats.AckExplicitPolicy,
			AckWait:   30 * time.Second,
		}); addErr != nil {
			nc.Close()
			return nil, addErr
		}
	}

	sub, err := js.PullSubscribe(cfg.Subject, cfg.Consumer, nats.BindStream(cfg.Stream))
	if err != nil {
		nc.Close()
		return nil, err
	}

	return &NATSJetStreamJobQueue{
		nc:       nc,
		js:       js,
		subject:  cfg.Subject,
		sub:      sub,
		inFlight: map[string]*nats.Msg{},
	}, nil
}

func (q *NATSJetStreamJobQueue) Enqueue(_ context.Context, job JobEnvelope) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}

	_, err = q.js.Publish(q.subject, payload)
	return err
}

func (q *NATSJetStreamJobQueue) Reserve(_ context.Context, wait time.Duration) (JobEnvelope, JobLease, bool, error) {
	if wait <= 0 {
		wait = 5 * time.Second
	}

	msgs, err := q.sub.Fetch(1, nats.MaxWait(wait))
	if err != nil {
		if errors.Is(err, nats.ErrTimeout) {
			return JobEnvelope{}, JobLease{}, false, nil
		}
		return JobEnvelope{}, JobLease{}, false, err
	}
	if len(msgs) == 0 {
		return JobEnvelope{}, JobLease{}, false, nil
	}

	msg := msgs[0]
	var job JobEnvelope
	if err := json.Unmarshal(msg.Data, &job); err != nil {
		_ = msg.Term()
		return JobEnvelope{}, JobLease{}, false, err
	}

	attempt := 1
	if meta, metaErr := msg.Metadata(); metaErr == nil && meta != nil && meta.NumDelivered > 0 {
		attempt = int(meta.NumDelivered)
	}

	receipt := fmt.Sprintf("%s:%d", job.JobID, time.Now().UnixNano())
	q.mu.Lock()
	q.inFlight[receipt] = msg
	q.mu.Unlock()

	lease := JobLease{JobID: job.JobID, Receipt: receipt, Attempt: attempt, LeasedUntil: time.Now().UTC()}
	return job, lease, true, nil
}

func (q *NATSJetStreamJobQueue) Ack(_ context.Context, lease JobLease) error {
	msg, err := q.popInFlight(lease.Receipt)
	if err != nil {
		return err
	}
	return msg.Ack()
}

func (q *NATSJetStreamJobQueue) Retry(_ context.Context, lease JobLease, _ string, nextDelay time.Duration) error {
	msg, err := q.popInFlight(lease.Receipt)
	if err != nil {
		return err
	}

	if nextDelay > 0 {
		return msg.NakWithDelay(nextDelay)
	}
	return msg.Nak()
}

func (q *NATSJetStreamJobQueue) Fail(_ context.Context, lease JobLease, _ string) error {
	msg, err := q.popInFlight(lease.Receipt)
	if err != nil {
		return err
	}
	return msg.Term()
}

func (q *NATSJetStreamJobQueue) Close() {
	if q.nc != nil {
		q.nc.Close()
	}
}

func (q *NATSJetStreamJobQueue) popInFlight(receipt string) (*nats.Msg, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	msg, ok := q.inFlight[receipt]
	if !ok {
		return nil, ErrJobLeaseNotFound
	}
	delete(q.inFlight, receipt)
	return msg, nil
}

var _ JobQueue = (*NATSJetStreamJobQueue)(nil)
