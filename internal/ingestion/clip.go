package ingestion

import (
	"context"
	"fmt"

	"github.com/andreas-lindfalk/videra/internal/storage"
)

type StubCLIPEmbedder struct{}

func NewStubCLIPEmbedder() *StubCLIPEmbedder {
	return &StubCLIPEmbedder{}
}

func (e *StubCLIPEmbedder) EmbedFrame(_ context.Context, framePath string) ([]float32, string, error) {
	if framePath == "" {
		return nil, "", fmt.Errorf("frame path is required")
	}

	embedding := []float32{0.35, 0.11, 0.27, 0.48, 0.16, 0.24, 0.08, 0.31}
	description := fmt.Sprintf("[simulated visual] keyframe context from %s", framePath)
	return embedding, description, nil
}

func BuildVisualSegments(videoID string, framePaths []string, intervalSec int, embedder VisualEmbedder) []storage.Segment {
	if intervalSec <= 0 {
		intervalSec = 5
	}

	segments := make([]storage.Segment, 0, len(framePaths))
	for idx, framePath := range framePaths {
		embedding, description, err := embedder.EmbedFrame(context.Background(), framePath)
		if err != nil {
			continue
		}

		startMs := int64(idx * intervalSec * 1000)
		endMs := int64((idx + 1) * intervalSec * 1000)
		segments = append(segments, storage.Segment{
			VideoID:    videoID,
			StartMs:    startMs,
			EndMs:      endMs,
			Text:       description,
			Type:       storage.SegmentTypeVisual,
			Embedding:  embedding,
			SourcePath: framePath,
		})
	}

	return segments
}

var _ VisualEmbedder = (*StubCLIPEmbedder)(nil)
