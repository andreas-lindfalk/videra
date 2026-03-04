package ingestion

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type SyncIndexOrchestrator struct {
	ingester    Ingester
	lookup      SourceVideoLookup
	queue       JobQueue
	maxAttempts int

	mu   sync.RWMutex
	jobs map[string]IndexJobResult
}

func NewSyncIndexOrchestrator(ingester Ingester, lookup SourceVideoLookup) *SyncIndexOrchestrator {
	return NewSyncIndexOrchestratorWithQueue(ingester, lookup, NewInProcessJobQueue(128))
}

func NewSyncIndexOrchestratorWithQueue(ingester Ingester, lookup SourceVideoLookup, queue JobQueue) *SyncIndexOrchestrator {
	if queue == nil {
		queue = NewInProcessJobQueue(128)
	}

	return &SyncIndexOrchestrator{ingester: ingester, lookup: lookup, queue: queue, maxAttempts: 2, jobs: map[string]IndexJobResult{}}
}

func (o *SyncIndexOrchestrator) Run(ctx context.Context, req IndexJobRequest) (IndexJobResult, error) {
	if o.ingester == nil {
		return IndexJobResult{}, fmt.Errorf("ingester is required")
	}
	if o.lookup == nil {
		return IndexJobResult{}, fmt.Errorf("source lookup is required")
	}
	if o.queue == nil {
		o.queue = NewInProcessJobQueue(128)
	}
	if o.maxAttempts <= 0 {
		o.maxAttempts = 1
	}
	if req.Mode == "" {
		req.Mode = IndexModeSync
	}
	req.Path = strings.TrimSpace(req.Path)
	if req.Path == "" {
		return IndexJobResult{}, fmt.Errorf("path is required")
	}
	if existing, ok := o.lookup.GetVideoBySourcePath(ctx, req.Path); ok {
		result := IndexJobResult{JobID: req.JobID, Status: IndexJobStatusCompleted, Video: &existing}
		o.setJob(result)
		return result, nil
	}

	if req.Mode == IndexModeAsync {
		envelope := jobEnvelopeFromRequest(req)
		if err := o.queue.Enqueue(ctx, envelope); err != nil {
			return IndexJobResult{}, err
		}

		pending := IndexJobResult{JobID: req.JobID, Status: IndexJobStatusPending}
		o.setJob(pending)
		go o.processNextQueuedJob()
		return pending, nil
	}

	result := o.runSync(ctx, req)
	o.setJob(result)
	return result, nil
}

func (o *SyncIndexOrchestrator) GetJob(_ context.Context, jobID string) (IndexJobResult, bool) {
	trimmedJobID := strings.TrimSpace(jobID)
	if trimmedJobID == "" {
		return IndexJobResult{}, false
	}

	o.mu.RLock()
	result, ok := o.jobs[trimmedJobID]
	o.mu.RUnlock()
	if !ok {
		return IndexJobResult{}, false
	}

	return cloneJobResult(result), true
}

func (o *SyncIndexOrchestrator) processNextQueuedJob() {
	queued, lease, ok, err := o.queue.Reserve(context.Background(), 0)
	if err != nil || !ok {
		return
	}

	request := requestFromJobEnvelope(queued)
	request.Mode = IndexModeSync
	result := o.runSync(context.Background(), request)
	o.setJob(result)

	if result.Status == IndexJobStatusFailed {
		_ = o.queue.Fail(context.Background(), lease, result.Error)
		return
	}

	_ = o.queue.Ack(context.Background(), lease)
}

func (o *SyncIndexOrchestrator) runSync(ctx context.Context, req IndexJobRequest) IndexJobResult {
	if existing, ok := o.lookup.GetVideoBySourcePath(ctx, req.Path); ok {
		return IndexJobResult{JobID: req.JobID, Status: IndexJobStatusCompleted, Video: &existing}
	}

	var lastErr error
	for attempt := 1; attempt <= o.maxAttempts; attempt++ {
		video, err := o.ingester.IndexVideo(ctx, req.Path)
		if err == nil {
			return IndexJobResult{JobID: req.JobID, Status: IndexJobStatusCompleted, Video: &video}
		}

		if existing, ok := o.lookup.GetVideoBySourcePath(ctx, req.Path); ok {
			return IndexJobResult{JobID: req.JobID, Status: IndexJobStatusCompleted, Video: &existing}
		}

		lastErr = err
	}

	return IndexJobResult{
		JobID:  req.JobID,
		Status: IndexJobStatusFailed,
		Error:  lastErr.Error(),
	}
}

func (o *SyncIndexOrchestrator) setJob(result IndexJobResult) {
	if strings.TrimSpace(result.JobID) == "" {
		return
	}

	o.mu.Lock()
	if o.jobs == nil {
		o.jobs = map[string]IndexJobResult{}
	}
	o.jobs[result.JobID] = cloneJobResult(result)
	o.mu.Unlock()
}

func cloneJobResult(result IndexJobResult) IndexJobResult {
	cloned := result
	if result.Video != nil {
		copiedVideo := *result.Video
		cloned.Video = &copiedVideo
	}
	return cloned
}

func jobEnvelopeFromRequest(req IndexJobRequest) JobEnvelope {
	return JobEnvelope{
		JobID:       req.JobID,
		SourcePath:  req.Path,
		RequestedAt: req.RequestedAt,
		Attempt:     0,
		MaxAttempts: 2,
	}
}

func requestFromJobEnvelope(job JobEnvelope) IndexJobRequest {
	return IndexJobRequest{
		JobID:       job.JobID,
		Path:        job.SourcePath,
		RequestedAt: job.RequestedAt,
		Mode:        IndexModeAsync,
	}
}

var _ IndexOrchestrator = (*SyncIndexOrchestrator)(nil)
var _ IndexJobReader = (*SyncIndexOrchestrator)(nil)
