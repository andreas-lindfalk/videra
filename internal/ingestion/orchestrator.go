package ingestion

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

type SyncIndexOrchestratorOptions struct {
	JobStateStore        IndexJobStateStore
	RunInlineAsyncWorker bool
	SyncMaxAttempts      int
	AsyncRetryMax        int
	AsyncRetryBackoff    time.Duration
	WorkerPollInterval   time.Duration
}

type SyncIndexOrchestrator struct {
	ingester             Ingester
	lookup               SourceVideoLookup
	queue                JobQueue
	jobState             IndexJobStateStore
	runInlineAsyncWorker bool
	syncMaxAttempts      int
	asyncRetryMax        int
	asyncRetryBackoff    time.Duration
	workerPollInterval   time.Duration
}

func NewSyncIndexOrchestrator(ingester Ingester, lookup SourceVideoLookup) *SyncIndexOrchestrator {
	return NewSyncIndexOrchestratorWithOptions(ingester, lookup, NewInProcessJobQueue(128), SyncIndexOrchestratorOptions{
		RunInlineAsyncWorker: true,
	})
}

func NewSyncIndexOrchestratorWithQueue(ingester Ingester, lookup SourceVideoLookup, queue JobQueue) *SyncIndexOrchestrator {
	return NewSyncIndexOrchestratorWithOptions(ingester, lookup, queue, SyncIndexOrchestratorOptions{
		RunInlineAsyncWorker: true,
	})
}

func NewSyncIndexOrchestratorWithOptions(ingester Ingester, lookup SourceVideoLookup, queue JobQueue, options SyncIndexOrchestratorOptions) *SyncIndexOrchestrator {
	if queue == nil {
		queue = NewInProcessJobQueue(128)
	}
	if options.JobStateStore == nil {
		options.JobStateStore = NewInMemoryIndexJobStateStore()
	}
	if options.SyncMaxAttempts <= 0 {
		options.SyncMaxAttempts = 2
	}
	if options.AsyncRetryMax <= 0 {
		options.AsyncRetryMax = 3
	}
	if options.AsyncRetryBackoff < 0 {
		options.AsyncRetryBackoff = 0
	}
	if options.WorkerPollInterval <= 0 {
		options.WorkerPollInterval = 250 * time.Millisecond
	}

	return &SyncIndexOrchestrator{
		ingester:             ingester,
		lookup:               lookup,
		queue:                queue,
		jobState:             options.JobStateStore,
		runInlineAsyncWorker: options.RunInlineAsyncWorker,
		syncMaxAttempts:      options.SyncMaxAttempts,
		asyncRetryMax:        options.AsyncRetryMax,
		asyncRetryBackoff:    options.AsyncRetryBackoff,
		workerPollInterval:   options.WorkerPollInterval,
	}
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
	if o.jobState == nil {
		o.jobState = NewInMemoryIndexJobStateStore()
	}
	if o.syncMaxAttempts <= 0 {
		o.syncMaxAttempts = 2
	}
	if o.asyncRetryMax <= 0 {
		o.asyncRetryMax = 3
	}
	if req.Mode == "" {
		req.Mode = IndexModeSync
	}
	req.Path = strings.TrimSpace(req.Path)
	if req.Path == "" {
		return IndexJobResult{}, fmt.Errorf("path is required")
	}
	req.JobID = strings.TrimSpace(req.JobID)
	if req.JobID == "" {
		return IndexJobResult{}, fmt.Errorf("job id is required")
	}
	if existing, ok := o.lookup.GetVideoBySourcePath(ctx, req.Path); ok {
		result := IndexJobResult{JobID: req.JobID, Status: IndexJobStatusCompleted, Video: &existing}
		if err := o.setJob(ctx, result); err != nil {
			return IndexJobResult{}, err
		}
		return result, nil
	}

	if req.Mode == IndexModeAsync {
		envelope := jobEnvelopeFromRequest(req, o.asyncRetryMax)
		if err := o.queue.Enqueue(ctx, envelope); err != nil {
			return IndexJobResult{}, err
		}

		pending := IndexJobResult{JobID: req.JobID, Status: IndexJobStatusPending}
		if err := o.setJob(ctx, pending); err != nil {
			return IndexJobResult{}, err
		}
		if o.runInlineAsyncWorker {
			go o.processNextQueuedJob(context.Background())
		}
		return pending, nil
	}

	result := o.runSync(ctx, req)
	if err := o.setJob(ctx, result); err != nil {
		return IndexJobResult{}, err
	}
	return result, nil
}

func (o *SyncIndexOrchestrator) GetJob(ctx context.Context, jobID string) (IndexJobResult, bool) {
	if o.jobState == nil {
		o.jobState = NewInMemoryIndexJobStateStore()
	}

	trimmedJobID := strings.TrimSpace(jobID)
	if trimmedJobID == "" {
		return IndexJobResult{}, false
	}

	result, ok, err := o.jobState.Get(ctx, trimmedJobID)
	if err != nil || !ok {
		return IndexJobResult{}, false
	}

	return cloneJobResult(result), true
}

func (o *SyncIndexOrchestrator) RunWorker(ctx context.Context) error {
	if o.queue == nil {
		o.queue = NewInProcessJobQueue(128)
	}
	if o.jobState == nil {
		o.jobState = NewInMemoryIndexJobStateStore()
	}
	if o.workerPollInterval <= 0 {
		o.workerPollInterval = 250 * time.Millisecond
	}

	for {
		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return err
		}

		queued, lease, ok, err := o.queue.Reserve(ctx, o.workerPollInterval)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				if ctx.Err() != nil {
					return nil
				}
				continue
			}
			log.Printf("worker reserve error: %v", err)
			continue
		}
		if !ok {
			continue
		}

		o.processReservedJob(ctx, queued, lease)
	}
}

func (o *SyncIndexOrchestrator) processNextQueuedJob(ctx context.Context) {
	queued, lease, ok, err := o.queue.Reserve(ctx, 0)
	if err != nil || !ok {
		return
	}
	o.processReservedJob(ctx, queued, lease)
}

func (o *SyncIndexOrchestrator) processReservedJob(ctx context.Context, queued JobEnvelope, lease JobLease) {
	request := requestFromJobEnvelope(queued)
	request.Mode = IndexModeSync

	result := o.runSyncWithAttempts(ctx, request, 1)
	if result.Status == IndexJobStatusCompleted {
		if err := o.setJob(ctx, result); err != nil {
			log.Printf("job state set failed for completed job %s: %v", result.JobID, err)
		}
		if err := o.queue.Ack(ctx, lease); err != nil {
			log.Printf("job ack failed for %s: %v", lease.JobID, err)
		}
		log.Printf("async job completed job_id=%s attempt=%d", lease.JobID, lease.Attempt)
		return
	}

	attempt := lease.Attempt
	if attempt <= 0 {
		attempt = queued.Attempt + 1
	}

	maxAttempts := queued.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = o.asyncRetryMax
	}

	if attempt < maxAttempts {
		delay := o.retryDelay(attempt)
		if err := o.queue.Retry(ctx, lease, result.Error, delay); err != nil {
			terminal := IndexJobResult{
				JobID:  queued.JobID,
				Status: IndexJobStatusFailed,
				Error:  fmt.Sprintf("queue retry failed on attempt %d/%d: %v", attempt, maxAttempts, err),
			}
			if failErr := o.queue.Fail(ctx, lease, terminal.Error); failErr != nil {
				log.Printf("queue fail after retry failure also failed for job %s: %v", lease.JobID, failErr)
			}
			if setErr := o.setJob(ctx, terminal); setErr != nil {
				log.Printf("job state set failed for terminal retry failure %s: %v", terminal.JobID, setErr)
			}
			log.Printf("async job failed job_id=%s attempt=%d reason=%s", queued.JobID, attempt, terminal.Error)
			return
		}

		pending := IndexJobResult{JobID: queued.JobID, Status: IndexJobStatusPending}
		if err := o.setJob(ctx, pending); err != nil {
			log.Printf("job state set failed for pending retry job %s: %v", queued.JobID, err)
		}
		if o.runInlineAsyncWorker {
			go func(wait time.Duration) {
				if wait > 0 {
					timer := time.NewTimer(wait)
					defer timer.Stop()
					<-timer.C
				}
				o.processNextQueuedJob(context.Background())
			}(delay)
		}
		log.Printf("async job retry scheduled job_id=%s attempt=%d/%d delay=%s", queued.JobID, attempt, maxAttempts, delay)
		return
	}

	terminal := IndexJobResult{
		JobID:  queued.JobID,
		Status: IndexJobStatusFailed,
		Error:  fmt.Sprintf("index job failed after %d attempts: %s", maxAttempts, result.Error),
	}
	if err := o.queue.Fail(ctx, lease, terminal.Error); err != nil {
		log.Printf("queue fail failed for job %s: %v", queued.JobID, err)
	}
	if err := o.setJob(ctx, terminal); err != nil {
		log.Printf("job state set failed for terminal job %s: %v", queued.JobID, err)
	}
	log.Printf("async job exhausted retries job_id=%s attempts=%d", queued.JobID, maxAttempts)
}

func (o *SyncIndexOrchestrator) retryDelay(attempt int) time.Duration {
	if o.asyncRetryBackoff <= 0 {
		return 0
	}
	if attempt <= 1 {
		return o.asyncRetryBackoff
	}

	multiplier := 1 << (attempt - 1)
	if multiplier > 16 {
		multiplier = 16
	}

	return time.Duration(multiplier) * o.asyncRetryBackoff
}

func (o *SyncIndexOrchestrator) runSync(ctx context.Context, req IndexJobRequest) IndexJobResult {
	return o.runSyncWithAttempts(ctx, req, o.syncMaxAttempts)
}

func (o *SyncIndexOrchestrator) runSyncWithAttempts(ctx context.Context, req IndexJobRequest, attempts int) IndexJobResult {
	if attempts <= 0 {
		attempts = 1
	}

	if existing, ok := o.lookup.GetVideoBySourcePath(ctx, req.Path); ok {
		return IndexJobResult{JobID: req.JobID, Status: IndexJobStatusCompleted, Video: &existing}
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
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

func (o *SyncIndexOrchestrator) setJob(ctx context.Context, result IndexJobResult) error {
	if strings.TrimSpace(result.JobID) == "" {
		return nil
	}
	if o.jobState == nil {
		o.jobState = NewInMemoryIndexJobStateStore()
	}
	return o.jobState.Set(ctx, cloneJobResult(result))
}

func cloneJobResult(result IndexJobResult) IndexJobResult {
	cloned := result
	if result.Video != nil {
		copiedVideo := *result.Video
		cloned.Video = &copiedVideo
	}
	return cloned
}

func jobEnvelopeFromRequest(req IndexJobRequest, maxAttempts int) JobEnvelope {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	return JobEnvelope{
		JobID:       req.JobID,
		SourcePath:  req.Path,
		RequestedAt: req.RequestedAt,
		Attempt:     0,
		MaxAttempts: maxAttempts,
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
