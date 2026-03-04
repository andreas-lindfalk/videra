package ingestion

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/andreas-lindfalk/videra/internal/storage"
	"github.com/stretchr/testify/require"
)

type fakeIngester struct {
	mu      sync.Mutex
	calls   int
	delay   time.Duration
	results []struct {
		video storage.Video
		err   error
	}
}

func (f *fakeIngester) IndexVideo(_ context.Context, _ string) (storage.Video, error) {
	if f.delay > 0 {
		time.Sleep(f.delay)
	}

	f.mu.Lock()
	f.calls++
	if len(f.results) == 0 {
		f.mu.Unlock()
		return storage.Video{}, errors.New("no fake result configured")
	}
	result := f.results[0]
	if len(f.results) > 1 {
		f.results = f.results[1:]
	}
	f.mu.Unlock()
	return result.video, result.err
}

func (f *fakeIngester) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

type fakeLookup struct {
	videos map[string]storage.Video
}

func (f *fakeLookup) GetVideoBySourcePath(_ context.Context, sourcePath string) (storage.Video, bool) {
	video, ok := f.videos[sourcePath]
	return video, ok
}

func TestSyncIndexOrchestratorRunSyncCompleted(t *testing.T) {
	lookup := &fakeLookup{videos: map[string]storage.Video{}}
	ingester := &fakeIngester{results: []struct {
		video storage.Video
		err   error
	}{{video: storage.Video{ID: "v1", FilePath: "https://example.com/v1.mp4"}}}}
	orchestrator := NewSyncIndexOrchestrator(ingester, lookup)
	req := NewIndexJobRequest("https://example.com/v1.mp4", IndexModeSync)

	result, err := orchestrator.Run(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, IndexJobStatusCompleted, result.Status)
	require.Equal(t, req.JobID, result.JobID)
	require.NotNil(t, result.Video)
	require.Equal(t, "v1", result.Video.ID)
	require.Equal(t, 1, ingester.Calls())
}

func TestSyncIndexOrchestratorRunAsyncPending(t *testing.T) {
	lookup := &fakeLookup{videos: map[string]storage.Video{}}
	ingester := &fakeIngester{results: []struct {
		video storage.Video
		err   error
	}{{video: storage.Video{ID: "v1"}}}, delay: 80 * time.Millisecond}
	orchestrator := NewSyncIndexOrchestrator(ingester, lookup)
	req := NewIndexJobRequest("https://example.com/v1.mp4", IndexModeAsync)

	result, err := orchestrator.Run(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, IndexJobStatusPending, result.Status)
	require.Equal(t, req.JobID, result.JobID)
	require.Nil(t, result.Video)

	stored, ok := orchestrator.GetJob(context.Background(), req.JobID)
	require.True(t, ok)
	require.Equal(t, IndexJobStatusPending, stored.Status)

	require.Eventually(t, func() bool {
		job, found := orchestrator.GetJob(context.Background(), req.JobID)
		return found && job.Status == IndexJobStatusCompleted && job.Video != nil
	}, 2*time.Second, 20*time.Millisecond)
	require.Equal(t, 1, ingester.Calls())
}

func TestSyncIndexOrchestratorRunRetriesAndSucceeds(t *testing.T) {
	lookup := &fakeLookup{videos: map[string]storage.Video{}}
	ingester := &fakeIngester{results: []struct {
		video storage.Video
		err   error
	}{
		{err: errors.New("transient failure")},
		{video: storage.Video{ID: "v1", FilePath: "https://example.com/retry.mp4"}},
	}}
	orchestrator := NewSyncIndexOrchestrator(ingester, lookup)
	req := NewIndexJobRequest("https://example.com/retry.mp4", IndexModeSync)

	result, err := orchestrator.Run(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, IndexJobStatusCompleted, result.Status)
	require.NotNil(t, result.Video)
	require.Equal(t, "v1", result.Video.ID)
	require.Equal(t, 2, ingester.Calls())
}

func TestSyncIndexOrchestratorRunPartialFailureRecoveryFromLookup(t *testing.T) {
	path := "https://example.com/partial.mp4"
	persisted := storage.Video{ID: "persisted", FilePath: path}

	lookup := &fakeLookup{videos: map[string]storage.Video{}}
	ingester := &fakeIngester{results: []struct {
		video storage.Video
		err   error
	}{{err: errors.New("error after persistence")}}}
	orchestrator := NewSyncIndexOrchestrator(ingester, lookup)
	req := NewIndexJobRequest(path, IndexModeSync)

	result, err := orchestrator.Run(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, IndexJobStatusFailed, result.Status)
	require.Equal(t, "error after persistence", result.Error)
	require.Nil(t, result.Video)

	lookup.videos[path] = persisted
	retried, err := orchestrator.Run(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, IndexJobStatusCompleted, retried.Status)
	require.NotNil(t, retried.Video)
	require.Equal(t, persisted.ID, retried.Video.ID)
}

func TestSyncIndexOrchestratorRunPersistentFailureStatus(t *testing.T) {
	lookup := &fakeLookup{videos: map[string]storage.Video{}}
	ingester := &fakeIngester{results: []struct {
		video storage.Video
		err   error
	}{
		{err: errors.New("boom-1")},
		{err: errors.New("boom-2")},
	}}
	orchestrator := NewSyncIndexOrchestrator(ingester, lookup)
	req := NewIndexJobRequest("https://example.com/v1.mp4", IndexModeSync)

	result, err := orchestrator.Run(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, IndexJobStatusFailed, result.Status)
	require.Equal(t, "boom-2", result.Error)
	require.Nil(t, result.Video)
	require.Equal(t, 2, ingester.Calls())
}

func TestSyncIndexOrchestratorRunExistingShortCircuits(t *testing.T) {
	path := "https://example.com/existing.mp4"
	existing := storage.Video{ID: "existing", FilePath: path}
	lookup := &fakeLookup{videos: map[string]storage.Video{path: existing}}
	ingester := &fakeIngester{results: []struct {
		video storage.Video
		err   error
	}{{video: storage.Video{ID: "new"}}}}
	orchestrator := NewSyncIndexOrchestrator(ingester, lookup)
	req := NewIndexJobRequest(path, IndexModeSync)

	result, err := orchestrator.Run(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, IndexJobStatusCompleted, result.Status)
	require.NotNil(t, result.Video)
	require.Equal(t, existing.ID, result.Video.ID)
	require.Equal(t, 0, ingester.Calls())
}

func TestSyncIndexOrchestratorGetJobUnknown(t *testing.T) {
	lookup := &fakeLookup{videos: map[string]storage.Video{}}
	ingester := &fakeIngester{results: []struct {
		video storage.Video
		err   error
	}{{video: storage.Video{ID: "v1"}}}}
	orchestrator := NewSyncIndexOrchestrator(ingester, lookup)

	_, ok := orchestrator.GetJob(context.Background(), "missing")
	require.False(t, ok)
}

func TestSyncIndexOrchestratorRunAsyncFailurePersistsStatus(t *testing.T) {
	lookup := &fakeLookup{videos: map[string]storage.Video{}}
	ingester := &fakeIngester{results: []struct {
		video storage.Video
		err   error
	}{
		{err: errors.New("boom")},
		{err: errors.New("boom")},
		{err: errors.New("boom")},
	}}
	orchestrator := NewSyncIndexOrchestrator(ingester, lookup)
	req := NewIndexJobRequest("/missing/video.mp4", IndexModeAsync)

	result, err := orchestrator.Run(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, IndexJobStatusPending, result.Status)

	require.Eventually(t, func() bool {
		job, ok := orchestrator.GetJob(context.Background(), req.JobID)
		return ok && job.Status == IndexJobStatusFailed && job.Error == "index job failed after 3 attempts: boom"
	}, 2*time.Second, 20*time.Millisecond)
	require.Equal(t, 3, ingester.Calls())
}

func TestSyncIndexOrchestratorRunWorkerProcessesQueuedJobs(t *testing.T) {
	lookup := &fakeLookup{videos: map[string]storage.Video{}}
	ingester := &fakeIngester{results: []struct {
		video storage.Video
		err   error
	}{{video: storage.Video{ID: "v-worker", FilePath: "https://example.com/worker.mp4"}}}}
	orchestrator := NewSyncIndexOrchestratorWithOptions(ingester, lookup, NewInProcessJobQueue(16), SyncIndexOrchestratorOptions{
		RunInlineAsyncWorker: false,
		AsyncRetryBackoff:    1 * time.Millisecond,
		WorkerPollInterval:   1 * time.Millisecond,
	})
	req := NewIndexJobRequest("https://example.com/worker.mp4", IndexModeAsync)

	result, err := orchestrator.Run(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, IndexJobStatusPending, result.Status)

	workerCtx, cancel := context.WithCancel(context.Background())
	workerDone := make(chan error, 1)
	go func() {
		workerDone <- orchestrator.RunWorker(workerCtx)
	}()

	require.Eventually(t, func() bool {
		job, ok := orchestrator.GetJob(context.Background(), req.JobID)
		return ok && job.Status == IndexJobStatusCompleted && job.Video != nil && job.Video.ID == "v-worker"
	}, 2*time.Second, 20*time.Millisecond)

	cancel()
	require.NoError(t, <-workerDone)
	require.Equal(t, 1, ingester.Calls())
}

func TestSyncIndexOrchestratorRunWorkerStopsOnCancel(t *testing.T) {
	orchestrator := NewSyncIndexOrchestratorWithOptions(
		&fakeIngester{results: []struct {
			video storage.Video
			err   error
		}{{video: storage.Video{ID: "unused"}}}},
		&fakeLookup{videos: map[string]storage.Video{}},
		NewInProcessJobQueue(4),
		SyncIndexOrchestratorOptions{RunInlineAsyncWorker: false, WorkerPollInterval: 5 * time.Millisecond},
	)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- orchestrator.RunWorker(ctx)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	require.NoError(t, <-done)
}
