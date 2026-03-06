package ingestion

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andreas-lindfalk/videra/internal/storage"
)

const DefaultCLIPORTLibraryPath = "/usr/local/lib/libonnxruntime.so"

type CLIPVisualEmbedderOptions struct {
	ModelPath      string
	ORTLibraryPath string
	InputSize      int
	Runner         CLIPRunner
}

type CLIPRunner interface {
	Prepare(modelPath, ortLibraryPath string) error
	EmbedFrame(ctx context.Context, modelPath, framePath string, inputSize int) ([]float32, error)
}

type failoverVisualEmbedder struct {
	primary  VisualEmbedder
	fallback VisualEmbedder
}

func NewFailoverVisualEmbedder(primary, fallback VisualEmbedder) VisualEmbedder {
	if primary == nil {
		return fallback
	}
	if fallback == nil {
		return primary
	}
	return &failoverVisualEmbedder{primary: primary, fallback: fallback}
}

type CLIPVisualEmbedder struct {
	modelPath      string
	ortLibraryPath string
	inputSize      int
	runner         CLIPRunner
}

func NewCLIPVisualEmbedder(options CLIPVisualEmbedderOptions) (*CLIPVisualEmbedder, error) {
	modelPath := strings.TrimSpace(options.ModelPath)
	if modelPath == "" {
		return nil, fmt.Errorf("clip model path is required")
	}
	if _, err := os.Stat(modelPath); err != nil {
		return nil, fmt.Errorf("clip model path is not readable: %w", err)
	}

	ortLibraryPath := strings.TrimSpace(options.ORTLibraryPath)
	if ortLibraryPath == "" {
		ortLibraryPath = DefaultCLIPORTLibraryPath
	}

	runner := options.Runner
	if runner == nil {
		runner = NewNativeCLIPRunner()
	}

	if err := runner.Prepare(modelPath, ortLibraryPath); err != nil {
		return nil, fmt.Errorf("clip backend initialization failed: %w", err)
	}

	inputSize := options.InputSize
	if inputSize <= 0 {
		inputSize = 224
	}

	return &CLIPVisualEmbedder{
		modelPath:      modelPath,
		ortLibraryPath: ortLibraryPath,
		inputSize:      inputSize,
		runner:         runner,
	}, nil
}

func (e *CLIPVisualEmbedder) EmbedFrame(ctx context.Context, framePath string) ([]float32, string, error) {
	if strings.TrimSpace(framePath) == "" {
		return nil, "", fmt.Errorf("frame path is required")
	}

	embedding, err := e.runner.EmbedFrame(ctx, e.modelPath, framePath, e.inputSize)
	if err != nil {
		return nil, "", fmt.Errorf("clip embedding failed: %w", err)
	}
	if len(embedding) == 0 {
		return nil, "", fmt.Errorf("clip embedding failed: empty embedding")
	}

	base := filepath.Base(framePath)
	description := fmt.Sprintf("keyframe clip embedding: %s", base)
	return embedding, description, nil
}

func (e *failoverVisualEmbedder) EmbedFrame(ctx context.Context, framePath string) ([]float32, string, error) {
	embedding, description, err := e.primary.EmbedFrame(ctx, framePath)
	if err == nil {
		return embedding, description, nil
	}

	fallbackEmbedding, fallbackDescription, fallbackErr := e.fallback.EmbedFrame(ctx, framePath)
	if fallbackErr != nil {
		return nil, "", fmt.Errorf("primary visual embedding failed (%v), fallback failed (%w)", err, fallbackErr)
	}
	return fallbackEmbedding, fallbackDescription, nil
}

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

func BuildVisualSegments(ctx context.Context, videoID string, framePaths []string, intervalSec int, embedder VisualEmbedder) []storage.Segment {
	if intervalSec <= 0 {
		intervalSec = 5
	}
	if ctx == nil {
		ctx = context.Background()
	}

	segments := make([]storage.Segment, 0, len(framePaths))
	for idx, framePath := range framePaths {
		embedding, description, err := embedder.EmbedFrame(ctx, framePath)
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
var _ VisualEmbedder = (*CLIPVisualEmbedder)(nil)
var _ VisualEmbedder = (*failoverVisualEmbedder)(nil)
var _ CLIPRunner = (*NativeCLIPRunner)(nil)
