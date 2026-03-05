package ingestion

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/andreas-lindfalk/videra/internal/storage"
	"github.com/google/uuid"
)

type MockIngester struct {
	store          storage.VectorStore
	ffmpeg         FFmpegRunner
	visualEmbedder VisualEmbedder
	options        IndexOptions
	failuresMu     sync.Mutex
	failedOnce     map[string]bool
}

func NewMockIngester(store storage.VectorStore, options IndexOptions) *MockIngester {
	if options.FrameIntervalSec <= 0 {
		options.FrameIntervalSec = 5
	}
	if options.Concurrency <= 0 {
		options.Concurrency = 4
	}

	return &MockIngester{
		store:          store,
		ffmpeg:         ExecFFmpeg{},
		visualEmbedder: NewStubCLIPEmbedder(),
		options:        options,
		failedOnce:     map[string]bool{},
	}
}

func (i *MockIngester) IndexVideo(ctx context.Context, path string) (storage.Video, error) {
	if strings.TrimSpace(path) == "" {
		return storage.Video{}, fmt.Errorf("path is required")
	}

	if !isURL(path) {
		if _, err := os.Stat(path); err != nil {
			return storage.Video{}, fmt.Errorf("video path not found: %w", err)
		}
	}

	if existing, ok := i.store.GetVideoBySourcePath(ctx, path); ok {
		return existing, nil
	}

	videoID := uuid.NewString()
	now := time.Now().UTC()
	video := storage.Video{
		ID:       videoID,
		FilePath: path,
		Status:   storage.VideoStatusIndexed,
		Indexed:  now,
		Duration: 15000,
	}

	baseName := filepath.Base(path)
	segments := []storage.Segment{
		{
			VideoID: videoID,
			StartMs: 0,
			EndMs:   5000,
			Text:    fmt.Sprintf("[simulated] Intro segment from %s", baseName),
			Type:    storage.SegmentTypeAudio,
		},
		{
			VideoID: videoID,
			StartMs: 5000,
			EndMs:   10000,
			Text:    "[simulated] Main discussion about roadmap and budget.",
			Type:    storage.SegmentTypeAudio,
		},
		{
			VideoID: videoID,
			StartMs: 10000,
			EndMs:   15000,
			Text:    "[simulated] Closing remarks and next actions.",
			Type:    storage.SegmentTypeAudio,
		},
	}

	visualSegments := i.buildVisualSegments(ctx, videoID, path)
	segments = append(segments, visualSegments...)

	video.AudioSegments = countSegmentsByType(segments, storage.SegmentTypeAudio)
	video.VisualSegments = countSegmentsByType(segments, storage.SegmentTypeVisual)
	video.Modalities = []string{"audio"}
	if video.VisualSegments > 0 {
		video.Modalities = append(video.Modalities, "visual")
	}

	if err := i.store.IndexVideo(ctx, video, segments); err != nil {
		return storage.Video{}, fmt.Errorf("index in store: %w", err)
	}

	if i.shouldFailAfterPersistOnce(path) {
		return storage.Video{}, fmt.Errorf("transient simulated failure after persistence")
	}

	return video, nil
}

func (i *MockIngester) shouldFailAfterPersistOnce(path string) bool {
	if !strings.Contains(path, "videra_fail_after_persist_once=1") {
		return false
	}

	i.failuresMu.Lock()
	defer i.failuresMu.Unlock()
	if i.failedOnce[path] {
		return false
	}
	i.failedOnce[path] = true
	return true
}

func (i *MockIngester) buildVisualSegments(ctx context.Context, videoID, path string) []storage.Segment {
	if isURL(path) {
		framePaths := []string{
			"remote/frame-00001.jpg",
			"remote/frame-00002.jpg",
			"remote/frame-00003.jpg",
		}
		return BuildVisualSegments(videoID, framePaths, i.options.FrameIntervalSec, i.visualEmbedder)
	}

	tmpDir, err := os.MkdirTemp("", "videra-frames-*")
	if err != nil {
		return nil
	}
	defer os.RemoveAll(tmpDir)

	err = i.ffmpeg.ExtractKeyframes(ctx, path, tmpDir, i.options.FrameIntervalSec)
	if err != nil {
		framePaths := []string{
			filepath.Join(tmpDir, "frame-fallback-00001.jpg"),
			filepath.Join(tmpDir, "frame-fallback-00002.jpg"),
		}
		return BuildVisualSegments(videoID, framePaths, i.options.FrameIntervalSec, i.visualEmbedder)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil
	}

	framePaths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".jpeg") || strings.HasSuffix(name, ".png") {
			framePaths = append(framePaths, filepath.Join(tmpDir, entry.Name()))
		}
	}

	if len(framePaths) == 0 {
		framePaths = []string{filepath.Join(tmpDir, "frame-empty-00001.jpg")}
	}

	return BuildVisualSegments(videoID, framePaths, i.options.FrameIntervalSec, i.visualEmbedder)
}

func countSegmentsByType(segments []storage.Segment, segmentType storage.SegmentType) int {
	count := 0
	for _, segment := range segments {
		if segment.Type == segmentType {
			count++
		}
	}
	return count
}

func isURL(input string) bool {
	u, err := url.Parse(input)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

var _ Ingester = (*MockIngester)(nil)
