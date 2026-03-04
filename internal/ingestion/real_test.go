package ingestion

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreas-lindfalk/videra/internal/embedding"
	"github.com/andreas-lindfalk/videra/internal/storage"
	"github.com/stretchr/testify/require"
)

type fakeFFmpegRunner struct {
	extractAudioCalls int
	extractAudioErr   error
}

func (f *fakeFFmpegRunner) ExtractAudio(_ context.Context, _ string, outputPath string) error {
	f.extractAudioCalls++
	if f.extractAudioErr != nil {
		return f.extractAudioErr
	}
	return os.WriteFile(outputPath, []byte("audio"), 0o644)
}

func (f *fakeFFmpegRunner) ExtractKeyframes(_ context.Context, _ string, _ string, _ int) error {
	return nil
}

type fakeTranscriber struct {
	called   bool
	segments []storage.Segment
	err      error
}

func (f *fakeTranscriber) Transcribe(_ context.Context, _ string) ([]storage.Segment, error) {
	f.called = true
	if f.err != nil {
		return nil, f.err
	}
	return append([]storage.Segment(nil), f.segments...), nil
}

type noVisualEmbedder struct{}

func (e *noVisualEmbedder) EmbedFrame(_ context.Context, _ string) ([]float32, string, error) {
	return nil, "", os.ErrPermission
}

type cueVisualEmbedder struct{}

func (e *cueVisualEmbedder) EmbedFrame(_ context.Context, framePath string) ([]float32, string, error) {
	if strings.Contains(framePath, "00001") {
		return nil, "whiteboard architecture", nil
	}
	return nil, "team discussing timeline", nil
}

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

func TestRealIngesterFallsBackToAudioTranscriptionWhenNoSidecar(t *testing.T) {
	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "meeting.mp4")
	require.NoError(t, createFile(videoPath, "placeholder"))

	store, err := storage.NewChromemStore(filepath.Join(tmpDir, "data"), embedding.NewDeterministicTextEmbedder())
	require.NoError(t, err)

	ffmpeg := &fakeFFmpegRunner{}
	transcriber := &fakeTranscriber{segments: []storage.Segment{
		{StartMs: 0, EndMs: 2000, Text: "hello from transcription"},
		{StartMs: 2000, EndMs: 5000, Text: "second caption"},
	}}
	ingester := NewRealIngesterWithDeps(store, IndexOptions{FrameIntervalSec: 5, Concurrency: 1}, ffmpeg, transcriber, NewStubCLIPEmbedder())

	video, err := ingester.IndexVideo(context.Background(), videoPath)
	require.NoError(t, err)
	require.Equal(t, 2, video.AudioSegments)
	require.True(t, transcriber.called)
	require.Equal(t, 1, ffmpeg.extractAudioCalls)

	transcript, err := store.GetTranscript(context.Background(), video.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transcript)
	require.Equal(t, "hello from transcription", transcript[0].Text)
}

func TestRealIngesterReturnsErrorWhenNoSidecarAndTranscriptionFails(t *testing.T) {
	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "meeting.mp4")
	require.NoError(t, createFile(videoPath, "placeholder"))

	store, err := storage.NewChromemStore(filepath.Join(tmpDir, "data"), embedding.NewDeterministicTextEmbedder())
	require.NoError(t, err)

	ffmpeg := &fakeFFmpegRunner{}
	transcriber := &fakeTranscriber{err: os.ErrPermission}
	ingester := NewRealIngesterWithDeps(store, IndexOptions{FrameIntervalSec: 5, Concurrency: 1}, ffmpeg, transcriber, NewStubCLIPEmbedder())

	_, err = ingester.IndexVideo(context.Background(), videoPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "transcription failed")
}

func TestRealIngesterSpokenPhraseQueryReturnsExpectedWindow(t *testing.T) {
	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "speech.mp4")
	require.NoError(t, createFile(videoPath, "placeholder"))

	srt := "1\n00:00:01,000 --> 00:00:03,000\nhello team\n\n2\n00:00:05,000 --> 00:00:07,000\nbudget roadmap\n"
	require.NoError(t, createFile(filepath.Join(tmpDir, "speech.srt"), srt))

	store, err := storage.NewChromemStore(filepath.Join(tmpDir, "data"), embedding.NewDeterministicTextEmbedder())
	require.NoError(t, err)

	ingester := NewRealIngesterWithDeps(store, IndexOptions{FrameIntervalSec: 5, Concurrency: 1}, &fakeFFmpegRunner{}, &fakeTranscriber{}, &noVisualEmbedder{})
	_, err = ingester.IndexVideo(context.Background(), videoPath)
	require.NoError(t, err)

	queryEmbedding := store.EmbedQuery(context.Background(), "budget roadmap")
	results, err := store.SearchSegments(context.Background(), queryEmbedding, 5)
	require.NoError(t, err)
	require.NotEmpty(t, results)

	foundPhrase := false
	for _, result := range results {
		if result.Segment.Text == "budget roadmap" {
			foundPhrase = true
			require.Equal(t, int64(5000), result.Segment.StartMs)
			require.Equal(t, int64(7000), result.Segment.EndMs)
			break
		}
	}
	require.True(t, foundPhrase)
}

func TestRealIngesterVisualCueQueryReturnsVisualHit(t *testing.T) {
	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "visual.mp4")
	require.NoError(t, createFile(videoPath, "placeholder"))
	require.NoError(t, createFile(filepath.Join(tmpDir, "visual.txt"), "standup notes"))

	store, err := storage.NewChromemStore(filepath.Join(tmpDir, "data"), embedding.NewDeterministicTextEmbedder())
	require.NoError(t, err)

	ffmpeg := &fakeFFmpegRunnerWithFrames{}
	ingester := NewRealIngesterWithDeps(store, IndexOptions{FrameIntervalSec: 5, Concurrency: 1}, ffmpeg, &fakeTranscriber{}, &cueVisualEmbedder{})
	_, err = ingester.IndexVideo(context.Background(), videoPath)
	require.NoError(t, err)

	queryEmbedding := store.EmbedQuery(context.Background(), "whiteboard architecture")
	results, err := store.SearchSegments(context.Background(), queryEmbedding, 5)
	require.NoError(t, err)
	require.NotEmpty(t, results)

	foundVisual := false
	for _, result := range results {
		if result.Segment.Type == storage.SegmentTypeVisual && strings.Contains(result.Segment.Text, "whiteboard architecture") {
			foundVisual = true
			require.GreaterOrEqual(t, result.Segment.StartMs, int64(0))
			break
		}
	}
	require.True(t, foundVisual)
}

type fakeFFmpegRunnerWithFrames struct{}

func (f *fakeFFmpegRunnerWithFrames) ExtractAudio(_ context.Context, _ string, outputPath string) error {
	return os.WriteFile(outputPath, []byte("audio"), 0o644)
}

func (f *fakeFFmpegRunnerWithFrames) ExtractKeyframes(_ context.Context, _ string, outputDir string, _ int) error {
	if err := os.WriteFile(filepath.Join(outputDir, "frame-00001.jpg"), []byte("frame"), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "frame-00002.jpg"), []byte("frame"), 0o644); err != nil {
		return err
	}
	return nil
}

func createFile(path, contents string) error {
	return os.WriteFile(path, []byte(contents), 0o644)
}
