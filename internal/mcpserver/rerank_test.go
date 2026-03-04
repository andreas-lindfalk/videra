package mcpserver

import (
	"testing"

	"github.com/andreas-lindfalk/videra/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestRerankHybridResultsDeterministicOrdering(t *testing.T) {
	input := []storage.SearchResult{
		{Segment: storage.Segment{VideoID: "v2", StartMs: 5000, EndMs: 6000, Type: storage.SegmentTypeAudio, Text: "a"}, Score: 0.7},
		{Segment: storage.Segment{VideoID: "v1", StartMs: 1000, EndMs: 2000, Type: storage.SegmentTypeVisual, Text: "b"}, Score: 0.7},
		{Segment: storage.Segment{VideoID: "v1", StartMs: 0, EndMs: 1000, Type: storage.SegmentTypeAudio, Text: "c"}, Score: 0.7},
	}

	first := rerankHybridResults(input, "", 3, RankingOptions{AudioWeight: 1, VisualWeight: 1}, false)
	second := rerankHybridResults(input, "", 3, RankingOptions{AudioWeight: 1, VisualWeight: 1}, false)

	require.Equal(t, first, second)
	require.Len(t, first, 3)
	require.Equal(t, "v1", first[0].VideoID)
	require.EqualValues(t, 0, first[0].StartMs)
	require.Equal(t, storage.SegmentTypeAudio, first[0].Type)
}

func TestRerankHybridResultsModalityWeighting(t *testing.T) {
	input := []storage.SearchResult{
		{Segment: storage.Segment{VideoID: "v1", StartMs: 0, EndMs: 1000, Type: storage.SegmentTypeAudio, Text: "audio-hit"}, Score: 0.8},
		{Segment: storage.Segment{VideoID: "v1", StartMs: 1000, EndMs: 2000, Type: storage.SegmentTypeVisual, Text: "visual-hit"}, Score: 0.7},
	}

	audioFavored := rerankHybridResults(input, "", 2, RankingOptions{AudioWeight: 1.5, VisualWeight: 1.0}, true)
	require.Len(t, audioFavored, 2)
	require.Equal(t, storage.SegmentTypeAudio, audioFavored[0].Type)
	require.Greater(t, audioFavored[0].RawSimilarity, float32(0))

	visualFavored := rerankHybridResults(input, "", 2, RankingOptions{AudioWeight: 1.0, VisualWeight: 2.0}, true)
	require.Len(t, visualFavored, 2)
	require.Equal(t, storage.SegmentTypeVisual, visualFavored[0].Type)
	require.Equal(t, "visual-hit", visualFavored[0].Snippet)
	require.Greater(t, visualFavored[0].RawSimilarity, float32(0))
}

func TestRerankHybridResultsPrefersExactPhraseTopHit(t *testing.T) {
	input := []storage.SearchResult{
		{Segment: storage.Segment{VideoID: "v1", StartMs: 0, EndMs: 1000, Type: storage.SegmentTypeVisual, Text: "random chart"}, Score: 0.95},
		{Segment: storage.Segment{VideoID: "v1", StartMs: 5000, EndMs: 7000, Type: storage.SegmentTypeAudio, Text: "budget roadmap"}, Score: 0.30},
		{Segment: storage.Segment{VideoID: "v1", StartMs: 1000, EndMs: 3000, Type: storage.SegmentTypeAudio, Text: "hello team"}, Score: 0.94},
	}

	hits := rerankHybridResults(input, "budget roadmap", 3, RankingOptions{AudioWeight: 1.0, VisualWeight: 1.0}, false)
	require.Len(t, hits, 3)
	require.Equal(t, storage.SegmentTypeAudio, hits[0].Type)
	require.Equal(t, "budget roadmap", hits[0].Snippet)
	require.EqualValues(t, 5000, hits[0].StartMs)
	require.EqualValues(t, 7000, hits[0].EndMs)
}
