package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/andreas-lindfalk/videra/internal/embedding"
	"github.com/stretchr/testify/require"
)

func TestNewLanceDBStoreUsesBackendScopedDataDir(t *testing.T) {
	baseDir := t.TempDir()

	store, err := NewLanceDBStore(baseDir, embedding.NewDeterministicTextEmbedder())
	require.NoError(t, err)
	require.Equal(t, filepath.Join(baseDir, lanceDBCompatibilityDirName), store.dataDir)
}

func TestLanceDBStoreIsolatedFromChromemStore(t *testing.T) {
	ctx := context.Background()
	baseDir := t.TempDir()

	chromemStore, err := NewChromemStore(baseDir, embedding.NewDeterministicTextEmbedder())
	require.NoError(t, err)

	lanceStore, err := NewLanceDBStore(baseDir, embedding.NewDeterministicTextEmbedder())
	require.NoError(t, err)

	chromemVideo := Video{
		ID:             "chromem-video",
		FilePath:       "https://example.com/chromem.mp4",
		Status:         VideoStatusIndexed,
		Indexed:        time.Now().UTC(),
		Duration:       5000,
		AudioSegments:  1,
		VisualSegments: 0,
		Modalities:     []string{"audio"},
	}
	lanceVideo := Video{
		ID:             "lancedb-video",
		FilePath:       "https://example.com/lancedb.mp4",
		Status:         VideoStatusIndexed,
		Indexed:        time.Now().UTC().Add(time.Second),
		Duration:       5000,
		AudioSegments:  1,
		VisualSegments: 0,
		Modalities:     []string{"audio"},
	}

	require.NoError(t, chromemStore.IndexVideo(ctx, chromemVideo, []Segment{{
		VideoID:    chromemVideo.ID,
		StartMs:    0,
		EndMs:      5000,
		Text:       "chromem content",
		Type:       SegmentTypeAudio,
		SourcePath: chromemVideo.FilePath,
	}}))

	require.NoError(t, lanceStore.IndexVideo(ctx, lanceVideo, []Segment{{
		VideoID:    lanceVideo.ID,
		StartMs:    0,
		EndMs:      5000,
		Text:       "lancedb compatibility content",
		Type:       SegmentTypeAudio,
		SourcePath: lanceVideo.FilePath,
	}}))

	chromemVideos, err := chromemStore.ListVideos(ctx)
	require.NoError(t, err)
	require.Len(t, chromemVideos, 1)
	require.Equal(t, chromemVideo.ID, chromemVideos[0].ID)

	lanceVideos, err := lanceStore.ListVideos(ctx)
	require.NoError(t, err)
	require.Len(t, lanceVideos, 1)
	require.Equal(t, lanceVideo.ID, lanceVideos[0].ID)
}
