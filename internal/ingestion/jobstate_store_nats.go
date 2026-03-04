package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
)

type NATSIndexJobStateStoreConfig struct {
	URL    string
	Bucket string
}

type NATSIndexJobStateStore struct {
	nc *nats.Conn
	kv nats.KeyValue
}

func NewNATSIndexJobStateStore(cfg NATSIndexJobStateStoreConfig) (*NATSIndexJobStateStore, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, fmt.Errorf("nats url is required")
	}
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, fmt.Errorf("nats key-value bucket is required")
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

	kv, err := js.KeyValue(cfg.Bucket)
	if err != nil {
		if !errors.Is(err, nats.ErrBucketNotFound) {
			nc.Close()
			return nil, err
		}
		kv, err = js.CreateKeyValue(&nats.KeyValueConfig{
			Bucket:      cfg.Bucket,
			Description: "Videra index job lifecycle state",
			History:     1,
			Storage:     nats.FileStorage,
		})
		if err != nil {
			nc.Close()
			return nil, err
		}
	}

	return &NATSIndexJobStateStore{nc: nc, kv: kv}, nil
}

func (s *NATSIndexJobStateStore) Set(_ context.Context, result IndexJobResult) error {
	jobID := strings.TrimSpace(result.JobID)
	if jobID == "" {
		return nil
	}

	payload, err := json.Marshal(cloneJobResult(result))
	if err != nil {
		return err
	}

	_, err = s.kv.Put(jobID, payload)
	return err
}

func (s *NATSIndexJobStateStore) Get(_ context.Context, jobID string) (IndexJobResult, bool, error) {
	trimmedJobID := strings.TrimSpace(jobID)
	if trimmedJobID == "" {
		return IndexJobResult{}, false, nil
	}

	entry, err := s.kv.Get(trimmedJobID)
	if err != nil {
		if errors.Is(err, nats.ErrKeyNotFound) {
			return IndexJobResult{}, false, nil
		}
		return IndexJobResult{}, false, err
	}

	var result IndexJobResult
	if err := json.Unmarshal(entry.Value(), &result); err != nil {
		return IndexJobResult{}, false, err
	}

	return cloneJobResult(result), true, nil
}

func (s *NATSIndexJobStateStore) Close() {
	if s.nc != nil {
		s.nc.Close()
	}
}

var _ IndexJobStateStore = (*NATSIndexJobStateStore)(nil)
