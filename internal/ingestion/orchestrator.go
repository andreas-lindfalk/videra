package ingestion

import (
	"context"
	"fmt"
)

type SyncIndexOrchestrator struct {
	ingester    Ingester
	lookup      SourceVideoLookup
	maxAttempts int
}

func NewSyncIndexOrchestrator(ingester Ingester, lookup SourceVideoLookup) *SyncIndexOrchestrator {
	return &SyncIndexOrchestrator{ingester: ingester, lookup: lookup, maxAttempts: 2}
}

func (o *SyncIndexOrchestrator) Run(ctx context.Context, req IndexJobRequest) (IndexJobResult, error) {
	if o.ingester == nil {
		return IndexJobResult{}, fmt.Errorf("ingester is required")
	}
	if o.lookup == nil {
		return IndexJobResult{}, fmt.Errorf("source lookup is required")
	}
	if o.maxAttempts <= 0 {
		o.maxAttempts = 1
	}
	if req.Mode == "" {
		req.Mode = IndexModeSync
	}
	if req.Path == "" {
		return IndexJobResult{}, fmt.Errorf("path is required")
	}
	if existing, ok := o.lookup.GetVideoBySourcePath(ctx, req.Path); ok {
		return IndexJobResult{JobID: req.JobID, Status: IndexJobStatusCompleted, Video: &existing}, nil
	}

	if req.Mode == IndexModeAsync {
		return IndexJobResult{JobID: req.JobID, Status: IndexJobStatusPending}, nil
	}

	var lastErr error
	for attempt := 1; attempt <= o.maxAttempts; attempt++ {
		video, err := o.ingester.IndexVideo(ctx, req.Path)
		if err == nil {
			return IndexJobResult{JobID: req.JobID, Status: IndexJobStatusCompleted, Video: &video}, nil
		}

		if existing, ok := o.lookup.GetVideoBySourcePath(ctx, req.Path); ok {
			return IndexJobResult{JobID: req.JobID, Status: IndexJobStatusCompleted, Video: &existing}, nil
		}

		lastErr = err
	}

	return IndexJobResult{
		JobID:  req.JobID,
		Status: IndexJobStatusFailed,
		Error:  lastErr.Error(),
	}, nil
}

var _ IndexOrchestrator = (*SyncIndexOrchestrator)(nil)
