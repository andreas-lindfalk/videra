package storage

import (
	"context"
	"testing"
	"time"

	"github.com/andreas-lindfalk/videra/internal/embedding"
	"github.com/stretchr/testify/require"
)

func TestChromemStoreSplitSharedStoragePersistsAcrossInstances(t *testing.T) {
	ctx := context.Background()
	dataDir := t.TempDir()

	storeWriter, err := NewChromemStoreWithOptions(dataDir, embedding.NewDeterministicTextEmbedder(), ChromemStoreOptions{SplitSharedStorage: true})
	require.NoError(t, err)

	video := Video{
		ID:             "video-shared-1",
		FilePath:       "https://example.com/shared-1.mp4",
		Status:         VideoStatusIndexed,
		Indexed:        time.Now().UTC(),
		Duration:       10000,
		AudioSegments:  1,
		VisualSegments: 1,
		Modalities:     []string{"audio", "visual"},
	}
	segments := []Segment{
		{VideoID: video.ID, StartMs: 0, EndMs: 5000, Text: "audio roadmap discussion", Type: SegmentTypeAudio, SourcePath: video.FilePath},
		{VideoID: video.ID, StartMs: 5000, EndMs: 10000, Text: "visual budget slide", Type: SegmentTypeVisual, SourcePath: video.FilePath},
	}

	require.NoError(t, storeWriter.IndexVideo(ctx, video, segments))

	storeReader, err := NewChromemStoreWithOptions(dataDir, embedding.NewDeterministicTextEmbedder(), ChromemStoreOptions{SplitSharedStorage: true})
	require.NoError(t, err)

	videos, err := storeReader.ListVideos(ctx)
	require.NoError(t, err)
	require.Len(t, videos, 1)
	require.Equal(t, video.ID, videos[0].ID)
	require.Equal(t, video.FilePath, videos[0].FilePath)

	byPath, ok := storeReader.GetVideoBySourcePath(ctx, video.FilePath)
	require.True(t, ok)
	require.Equal(t, video.ID, byPath.ID)

	transcript, err := storeReader.GetTranscript(ctx, video.ID)
	require.NoError(t, err)
	require.Len(t, transcript, 2)
	require.Equal(t, "audio roadmap discussion", transcript[0].Text)
}

func TestChromemStoreSharedSearchFallbackUsesPersistedManifests(t *testing.T) {
	ctx := context.Background()
	dataDir := t.TempDir()

	apiStore, err := NewChromemStoreWithOptions(dataDir, embedding.NewDeterministicTextEmbedder(), ChromemStoreOptions{SplitSharedStorage: true})
	require.NoError(t, err)
	require.NoError(t, apiStore.Reset(ctx))

	workerStore, err := NewChromemStoreWithOptions(dataDir, embedding.NewDeterministicTextEmbedder(), ChromemStoreOptions{SplitSharedStorage: true})
	require.NoError(t, err)

	video := Video{
		ID:             "video-shared-2",
		FilePath:       "https://example.com/shared-2.mp4",
		Status:         VideoStatusIndexed,
		Indexed:        time.Now().UTC(),
		Duration:       12000,
		AudioSegments:  2,
		VisualSegments: 0,
		Modalities:     []string{"audio"},
	}
	segments := []Segment{
		{VideoID: video.ID, StartMs: 0, EndMs: 6000, Text: "main discussion about roadmap and budget", Type: SegmentTypeAudio, SourcePath: video.FilePath},
		{VideoID: video.ID, StartMs: 6000, EndMs: 12000, Text: "closing action items", Type: SegmentTypeAudio, SourcePath: video.FilePath},
	}
	require.NoError(t, workerStore.IndexVideo(ctx, video, segments))

	queryEmbedding := apiStore.EmbedQuery(ctx, "roadmap and budget")
	results, err := apiStore.SearchSegments(ctx, queryEmbedding, 5)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	require.Equal(t, video.ID, results[0].Segment.VideoID)
}
