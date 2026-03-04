package ingestion

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/andreas-lindfalk/videra/internal/embedding"
	"github.com/andreas-lindfalk/videra/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestParseTimedCaptionSegmentsSRT(t *testing.T) {
	raw := "1\n00:00:01,000 --> 00:00:03,200\nHello world\n\n2\n00:00:03,500 --> 00:00:05,000\nBudget roadmap\n"

	segments, err := parseTimedCaptionSegments(raw, 5)
	require.NoError(t, err)
	require.Len(t, segments, 2)

	require.Equal(t, int64(1000), segments[0].StartMs)
	require.Equal(t, int64(3200), segments[0].EndMs)
	require.Equal(t, "Hello world", segments[0].Text)

	require.Equal(t, int64(3500), segments[1].StartMs)
	require.Equal(t, int64(5000), segments[1].EndMs)
	require.Equal(t, "Budget roadmap", segments[1].Text)
}

func TestRealIngesterRequiresSidecarTranscript(t *testing.T) {
	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "clip.mp4")
	require.NoError(t, createFile(videoPath, "placeholder"))

	store, err := storage.NewChromemStore(filepath.Join(tmpDir, "data"), embedding.NewDeterministicTextEmbedder())
	require.NoError(t, err)

	ingester := NewRealIngester(store, IndexOptions{FrameIntervalSec: 5, Concurrency: 1})
	_, err = ingester.IndexVideo(context.Background(), videoPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires a sidecar transcript")
}

func TestRealIngesterIndexesPlainTextSidecar(t *testing.T) {
	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "meeting.mp4")
	require.NoError(t, createFile(videoPath, "placeholder"))
	require.NoError(t, createFile(filepath.Join(tmpDir, "meeting.txt"), "hello team\nbudget roadmap\nnext steps"))

	store, err := storage.NewChromemStore(filepath.Join(tmpDir, "data"), embedding.NewDeterministicTextEmbedder())
	require.NoError(t, err)

	ingester := NewRealIngester(store, IndexOptions{FrameIntervalSec: 5, Concurrency: 1})
	video, err := ingester.IndexVideo(context.Background(), videoPath)
	require.NoError(t, err)
	require.Equal(t, 3, video.AudioSegments)
	require.GreaterOrEqual(t, video.VisualSegments, 1)

	transcript, err := store.GetTranscript(context.Background(), video.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transcript)
	require.Equal(t, "hello team", transcript[0].Text)
	require.NotContains(t, transcript[0].Text, "[simulated]")
}

func createFile(path, contents string) error {
	return os.WriteFile(path, []byte(contents), 0o644)
}
