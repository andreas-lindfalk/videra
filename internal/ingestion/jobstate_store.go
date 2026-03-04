package ingestion

import (
	"context"
	"strings"
	"sync"
)

type IndexJobStateStore interface {
	Set(ctx context.Context, result IndexJobResult) error
	Get(ctx context.Context, jobID string) (IndexJobResult, bool, error)
}

type InMemoryIndexJobStateStore struct {
	mu   sync.RWMutex
	jobs map[string]IndexJobResult
}

func NewInMemoryIndexJobStateStore() *InMemoryIndexJobStateStore {
	return &InMemoryIndexJobStateStore{jobs: map[string]IndexJobResult{}}
}

func (s *InMemoryIndexJobStateStore) Set(_ context.Context, result IndexJobResult) error {
	jobID := strings.TrimSpace(result.JobID)
	if jobID == "" {
		return nil
	}

	s.mu.Lock()
	if s.jobs == nil {
		s.jobs = map[string]IndexJobResult{}
	}
	s.jobs[jobID] = cloneJobResult(result)
	s.mu.Unlock()

	return nil
}

func (s *InMemoryIndexJobStateStore) Get(_ context.Context, jobID string) (IndexJobResult, bool, error) {
	trimmedJobID := strings.TrimSpace(jobID)
	if trimmedJobID == "" {
		return IndexJobResult{}, false, nil
	}

	s.mu.RLock()
	result, ok := s.jobs[trimmedJobID]
	s.mu.RUnlock()
	if !ok {
		return IndexJobResult{}, false, nil
	}

	return cloneJobResult(result), true, nil
}

var _ IndexJobStateStore = (*InMemoryIndexJobStateStore)(nil)
