package ingestion

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeCLIPRunner struct {
	embedding []float32
	err       error
}

func (f *fakeCLIPRunner) Prepare(_ string, _ string) error {
	if f.err != nil {
		return f.err
	}
	return nil
}

func (f *fakeCLIPRunner) EmbedFrame(_ context.Context, _ string, _ string, _ int) ([]float32, error) {
	if f.err != nil {
		return nil, f.err
	}
	return append([]float32(nil), f.embedding...), nil
}

type fakeVisualEmbedder struct {
	embedding   []float32
	description string
	err         error
}

func (f *fakeVisualEmbedder) EmbedFrame(_ context.Context, _ string) ([]float32, string, error) {
	if f.err != nil {
		return nil, "", f.err
	}
	return append([]float32(nil), f.embedding...), f.description, nil
}

func TestNewCLIPVisualEmbedderRequiresModelPath(t *testing.T) {
	_, err := NewCLIPVisualEmbedder(CLIPVisualEmbedderOptions{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "clip model path is required")
}

func TestCLIPVisualEmbedderEmbedsFrameWithInjectedRunner(t *testing.T) {
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "clip.onnx")
	require.NoError(t, os.WriteFile(modelPath, []byte("model"), 0o644))

	embedder, err := NewCLIPVisualEmbedder(CLIPVisualEmbedderOptions{
		ModelPath:      modelPath,
		ORTLibraryPath: "/tmp/libonnxruntime.so",
		InputSize:      224,
		Runner:         &fakeCLIPRunner{embedding: []float32{0.1, 0.2, 0.3}},
	})
	require.NoError(t, err)

	embedding, description, err := embedder.EmbedFrame(context.Background(), "/tmp/frame-00001.jpg")
	require.NoError(t, err)
	require.Equal(t, []float32{0.1, 0.2, 0.3}, embedding)
	require.Contains(t, description, "keyframe clip embedding")
}

func TestCLIPVisualEmbedderRejectsEmptyFramePath(t *testing.T) {
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "clip.onnx")
	require.NoError(t, os.WriteFile(modelPath, []byte("model"), 0o644))

	embedder, err := NewCLIPVisualEmbedder(CLIPVisualEmbedderOptions{
		ModelPath:      modelPath,
		ORTLibraryPath: "/tmp/libonnxruntime.so",
		Runner:         &fakeCLIPRunner{embedding: []float32{1}},
	})
	require.NoError(t, err)

	_, _, err = embedder.EmbedFrame(context.Background(), "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "frame path is required")
}

func TestFailoverVisualEmbedderUsesFallbackWhenPrimaryFails(t *testing.T) {
	primary := &fakeVisualEmbedder{err: fmt.Errorf("clip inference failed")}
	fallback := &fakeVisualEmbedder{embedding: []float32{0.9, 0.1}, description: "keyframe text: architecture"}
	embedder := NewFailoverVisualEmbedder(primary, fallback)

	embedding, description, err := embedder.EmbedFrame(context.Background(), "/tmp/frame.jpg")
	require.NoError(t, err)
	require.Equal(t, []float32{0.9, 0.1}, embedding)
	require.Equal(t, "keyframe text: architecture", description)
}
