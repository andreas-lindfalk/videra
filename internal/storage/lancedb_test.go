package storage

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/andreas-lindfalk/videra/internal/embedding"
	"github.com/stretchr/testify/require"
)

func TestNewLanceDBStoreUsesBackendScopedDataDir(t *testing.T) {
	baseDir := t.TempDir()

	store, err := NewLanceDBStoreWithOptions(baseDir, embedding.NewDeterministicTextEmbedder(), LanceDBStoreOptions{Bridge: &fakeLanceDBBridge{}})
	require.NoError(t, err)
	require.Equal(t, filepath.Join(baseDir, lanceDBDataDirName), store.dataDir)
}

func TestLanceDBStoreIsolatedFromChromemStore(t *testing.T) {
	ctx := context.Background()
	baseDir := t.TempDir()
	fakeBridge := &fakeLanceDBBridge{}

	chromemStore, err := NewChromemStore(baseDir, embedding.NewDeterministicTextEmbedder())
	require.NoError(t, err)

	lanceStore, err := NewLanceDBStoreWithOptions(baseDir, embedding.NewDeterministicTextEmbedder(), LanceDBStoreOptions{Bridge: fakeBridge})
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

func TestLanceDBStoreFallsBackOnSearchBridgeError(t *testing.T) {
	ctx := context.Background()
	baseDir := t.TempDir()
	fakeBridge := &fakeLanceDBBridge{searchErr: errors.New("search failed")}

	store, err := NewLanceDBStoreWithOptions(baseDir, embedding.NewDeterministicTextEmbedder(), LanceDBStoreOptions{Bridge: fakeBridge})
	require.NoError(t, err)

	video := Video{
		ID:             "v1",
		FilePath:       "https://example.com/v1.mp4",
		Status:         VideoStatusIndexed,
		Indexed:        time.Now().UTC(),
		Duration:       5000,
		AudioSegments:  1,
		VisualSegments: 0,
		Modalities:     []string{"audio"},
	}
	require.NoError(t, store.IndexVideo(ctx, video, []Segment{{
		VideoID:    video.ID,
		StartMs:    0,
		EndMs:      5000,
		Text:       "fallback text",
		Type:       SegmentTypeAudio,
		SourcePath: video.FilePath,
	}}))

	queryEmbedding := store.EmbedQuery(ctx, "fallback")
	results, err := store.SearchSegments(ctx, queryEmbedding, 5)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	require.Equal(t, video.ID, results[0].Segment.VideoID)
}

func TestLanceDBStoreSearchDoesNotAppendFallbackWhenBridgeHasResults(t *testing.T) {
	ctx := context.Background()
	baseDir := t.TempDir()
	fakeBridge := &fakeLanceDBBridge{}

	store, err := NewLanceDBStoreWithOptions(baseDir, embedding.NewDeterministicTextEmbedder(), LanceDBStoreOptions{Bridge: fakeBridge})
	require.NoError(t, err)

	video := Video{
		ID:             "v2",
		FilePath:       "https://example.com/v2.mp4",
		Status:         VideoStatusIndexed,
		Indexed:        time.Now().UTC(),
		Duration:       5000,
		AudioSegments:  1,
		VisualSegments: 0,
		Modalities:     []string{"audio"},
	}
	require.NoError(t, store.IndexVideo(ctx, video, []Segment{{
		VideoID:    video.ID,
		StartMs:    0,
		EndMs:      5000,
		Text:       "primary result",
		Type:       SegmentTypeAudio,
		SourcePath: video.FilePath,
	}}))

	queryEmbedding := store.EmbedQuery(ctx, "primary")
	results, err := store.SearchSegments(ctx, queryEmbedding, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, video.ID, results[0].Segment.VideoID)
	require.Equal(t, "primary result", results[0].Segment.Text)
}

type fakeLanceDBBridge struct {
	rows       []lanceDBSegmentRow
	searchRows []map[string]any
	searchErr  error
	resetCalls int
}

func (b *fakeLanceDBBridge) UpsertSegments(_ context.Context, rows []lanceDBSegmentRow) error {
	b.rows = append(b.rows, rows...)
	if len(b.searchRows) == 0 {
		for _, row := range rows {
			b.searchRows = append(b.searchRows, map[string]any{
				"video_id":    row.VideoID,
				"start_ms":    row.StartMs,
				"end_ms":      row.EndMs,
				"type":        row.Type,
				"source_path": row.SourcePath,
				"text":        row.Text,
				"_distance":   0.01,
			})
		}
	}
	return nil
}

func (b *fakeLanceDBBridge) SearchSegments(_ context.Context, _ []float32, limit int) ([]map[string]any, error) {
	if b.searchErr != nil {
		return nil, b.searchErr
	}
	if limit <= 0 || limit >= len(b.searchRows) {
		return append([]map[string]any(nil), b.searchRows...), nil
	}
	return append([]map[string]any(nil), b.searchRows[:limit]...), nil
}

func (b *fakeLanceDBBridge) Reset(_ context.Context) error {
	b.resetCalls++
	b.rows = nil
	b.searchRows = nil
	return nil
}
