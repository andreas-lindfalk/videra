package ingestion

import (
	"context"
	"errors"
	"testing"

	"github.com/andreas-lindfalk/videra/internal/storage"
	"github.com/stretchr/testify/require"
)

type fakeIngester struct {
	calls   int
	results []struct {
		video storage.Video
		err   error
	}
}

func (f *fakeIngester) IndexVideo(_ context.Context, _ string) (storage.Video, error) {
	f.calls++
	if len(f.results) == 0 {
		return storage.Video{}, errors.New("no fake result configured")
	}
	result := f.results[0]
	if len(f.results) > 1 {
		f.results = f.results[1:]
	}
	return result.video, result.err
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
	require.Equal(t, 1, ingester.calls)
}

func TestSyncIndexOrchestratorRunAsyncPending(t *testing.T) {
	lookup := &fakeLookup{videos: map[string]storage.Video{}}
	ingester := &fakeIngester{results: []struct {
		video storage.Video
		err   error
	}{{video: storage.Video{ID: "v1"}}}}
	orchestrator := NewSyncIndexOrchestrator(ingester, lookup)
	req := NewIndexJobRequest("https://example.com/v1.mp4", IndexModeAsync)

	result, err := orchestrator.Run(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, IndexJobStatusPending, result.Status)
	require.Equal(t, req.JobID, result.JobID)
	require.Nil(t, result.Video)
	require.Equal(t, 0, ingester.calls)
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
	require.Equal(t, 2, ingester.calls)
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
	require.Equal(t, 2, ingester.calls)
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
	require.Equal(t, 0, ingester.calls)
}
