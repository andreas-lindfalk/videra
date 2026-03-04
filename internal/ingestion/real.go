package ingestion

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/andreas-lindfalk/videra/internal/storage"
	"github.com/google/uuid"
)

var timeRangePattern = regexp.MustCompile(`(\d{2}:\d{2}:\d{2}[\.,]\d{3})\s*-->\s*(\d{2}:\d{2}:\d{2}[\.,]\d{3})`)

type RealIngester struct {
	store          storage.VectorStore
	ffmpeg         FFmpegRunner
	transcriber    Transcriber
	visualEmbedder VisualEmbedder
	options        IndexOptions
}

func NewRealIngester(store storage.VectorStore, options IndexOptions) *RealIngester {
	return NewRealIngesterWithDeps(store, options, ExecFFmpeg{}, NewWhisperCLITranscriber(), NewOCRVisualEmbedder(NewStubCLIPEmbedder()))
}

func NewRealIngesterWithDeps(store storage.VectorStore, options IndexOptions, ffmpeg FFmpegRunner, transcriber Transcriber, visualEmbedder VisualEmbedder) *RealIngester {
	if options.FrameIntervalSec <= 0 {
		options.FrameIntervalSec = 5
	}
	if options.Concurrency <= 0 {
		options.Concurrency = 4
	}
	if ffmpeg == nil {
		ffmpeg = ExecFFmpeg{}
	}
	if transcriber == nil {
		transcriber = NewWhisperCLITranscriber()
	}
	if visualEmbedder == nil {
		visualEmbedder = NewOCRVisualEmbedder(NewStubCLIPEmbedder())
	}

	return &RealIngester{
		store:          store,
		ffmpeg:         ffmpeg,
		transcriber:    transcriber,
		visualEmbedder: visualEmbedder,
		options:        options,
	}
}

func (i *RealIngester) IndexVideo(ctx context.Context, path string) (storage.Video, error) {
	if strings.TrimSpace(path) == "" {
		return storage.Video{}, fmt.Errorf("path is required")
	}
	if isURL(path) {
		return storage.Video{}, fmt.Errorf("real ingestion mode currently supports only local file paths")
	}
	if _, err := os.Stat(path); err != nil {
		return storage.Video{}, fmt.Errorf("video path not found: %w", err)
	}

	if existing, ok := i.store.GetVideoBySourcePath(ctx, path); ok {
		return existing, nil
	}

	videoID := uuid.NewString()
	audioSegments, err := i.loadAudioSegments(ctx, path)
	if err != nil {
		return storage.Video{}, err
	}
	for index := range audioSegments {
		audioSegments[index].VideoID = videoID
		audioSegments[index].Type = storage.SegmentTypeAudio
	}

	visualSegments := i.buildVisualSegments(ctx, videoID, path)
	segments := make([]storage.Segment, 0, len(audioSegments)+len(visualSegments))
	segments = append(segments, audioSegments...)
	segments = append(segments, visualSegments...)

	durationMs := inferDurationMs(segments)
	if durationMs <= 0 {
		durationMs = int64(i.options.FrameIntervalSec * 1000)
	}

	video := storage.Video{
		ID:             videoID,
		FilePath:       path,
		Status:         storage.VideoStatusIndexed,
		Indexed:        time.Now().UTC(),
		Duration:       durationMs,
		AudioSegments:  countSegmentsByType(segments, storage.SegmentTypeAudio),
		VisualSegments: countSegmentsByType(segments, storage.SegmentTypeVisual),
		Modalities:     []string{"audio"},
	}
	if video.VisualSegments > 0 {
		video.Modalities = append(video.Modalities, "visual")
	}

	if err := i.store.IndexVideo(ctx, video, segments); err != nil {
		return storage.Video{}, fmt.Errorf("index in store: %w", err)
	}

	return video, nil
}

func (i *RealIngester) buildVisualSegments(ctx context.Context, videoID, path string) []storage.Segment {
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
	sort.Strings(framePaths)

	return BuildVisualSegments(videoID, framePaths, i.options.FrameIntervalSec, i.visualEmbedder)
}

func (i *RealIngester) loadAudioSegments(ctx context.Context, videoPath string) ([]storage.Segment, error) {
	segments, err := loadAudioSegmentsFromSidecar(videoPath, i.options.FrameIntervalSec)
	if err == nil {
		return segments, nil
	}

	tmpDir, tmpErr := os.MkdirTemp("", "videra-audio-*")
	if tmpErr != nil {
		return nil, fmt.Errorf("sidecar unavailable (%v) and failed to create temp dir for transcription: %w", err, tmpErr)
	}
	defer os.RemoveAll(tmpDir)

	audioPath := filepath.Join(tmpDir, "audio.mp3")
	if extractErr := i.ffmpeg.ExtractAudio(ctx, videoPath, audioPath); extractErr != nil {
		return nil, fmt.Errorf("sidecar unavailable (%v) and audio extraction failed: %w", err, extractErr)
	}

	if i.transcriber == nil {
		return nil, fmt.Errorf("sidecar unavailable (%v) and no transcriber configured", err)
	}

	segments, transcribeErr := i.transcriber.Transcribe(ctx, audioPath)
	if transcribeErr != nil {
		return nil, fmt.Errorf("sidecar unavailable (%v) and transcription failed: %w", err, transcribeErr)
	}
	if len(segments) == 0 {
		return nil, fmt.Errorf("transcription returned no segments")
	}

	for idx := range segments {
		segments[idx].Type = storage.SegmentTypeAudio
	}
	return segments, nil
}

func loadAudioSegmentsFromSidecar(videoPath string, intervalSec int) ([]storage.Segment, error) {
	sidecarPath, err := resolveTranscriptSidecarPath(videoPath)
	if err != nil {
		return nil, err
	}

	payload, err := os.ReadFile(sidecarPath)
	if err != nil {
		return nil, fmt.Errorf("read transcript sidecar: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(sidecarPath))
	switch ext {
	case ".srt", ".vtt":
		segments, parseErr := parseTimedCaptionSegments(string(payload), intervalSec)
		if parseErr != nil {
			return nil, parseErr
		}
		return segments, nil
	case ".txt":
		segments := parsePlainTextSegments(string(payload), intervalSec)
		if len(segments) == 0 {
			return nil, fmt.Errorf("transcript sidecar contains no usable lines: %s", sidecarPath)
		}
		return segments, nil
	default:
		return nil, fmt.Errorf("unsupported transcript sidecar extension: %s", ext)
	}
}

func resolveTranscriptSidecarPath(videoPath string) (string, error) {
	base := strings.TrimSuffix(videoPath, filepath.Ext(videoPath))
	candidates := []string{base + ".srt", base + ".vtt", base + ".txt"}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("real ingestion mode requires a sidecar transcript file next to the video (.srt, .vtt, or .txt)")
}

func parseTimedCaptionSegments(raw string, fallbackIntervalSec int) ([]storage.Segment, error) {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	blocks := strings.Split(normalized, "\n\n")

	segments := make([]storage.Segment, 0, len(blocks))
	for _, block := range blocks {
		lines := strings.Split(block, "\n")
		clean := make([]string, 0, len(lines))
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if strings.EqualFold(trimmed, "WEBVTT") {
				continue
			}
			clean = append(clean, trimmed)
		}
		if len(clean) == 0 {
			continue
		}

		timeLineIndex := -1
		for idx, line := range clean {
			if strings.Contains(line, "-->") {
				timeLineIndex = idx
				break
			}
		}
		if timeLineIndex == -1 {
			continue
		}

		match := timeRangePattern.FindStringSubmatch(clean[timeLineIndex])
		if len(match) != 3 {
			continue
		}

		startMs, err := parseTimecodeToMs(match[1])
		if err != nil {
			continue
		}
		endMs, err := parseTimecodeToMs(match[2])
		if err != nil {
			continue
		}
		if endMs <= startMs {
			endMs = startMs + int64(fallbackIntervalSec*1000)
		}

		textLines := clean[timeLineIndex+1:]
		text := strings.TrimSpace(strings.Join(textLines, " "))
		if text == "" {
			continue
		}

		segments = append(segments, storage.Segment{
			StartMs: startMs,
			EndMs:   endMs,
			Text:    text,
			Type:    storage.SegmentTypeAudio,
		})
	}

	if len(segments) == 0 {
		return nil, fmt.Errorf("transcript sidecar contains no parseable timed captions")
	}

	return segments, nil
}

func parsePlainTextSegments(raw string, intervalSec int) []storage.Segment {
	if intervalSec <= 0 {
		intervalSec = 5
	}
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	segments := make([]storage.Segment, 0, len(lines))
	cursor := int64(0)
	stepMs := int64(intervalSec * 1000)
	for _, line := range lines {
		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}
		segments = append(segments, storage.Segment{
			StartMs: cursor,
			EndMs:   cursor + stepMs,
			Text:    text,
			Type:    storage.SegmentTypeAudio,
		})
		cursor += stepMs
	}
	return segments
}

func parseTimecodeToMs(input string) (int64, error) {
	normalized := strings.ReplaceAll(strings.TrimSpace(input), ",", ".")
	parts := strings.Split(normalized, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid timecode: %s", input)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	secParts := strings.Split(parts[2], ".")
	if len(secParts) != 2 {
		return 0, fmt.Errorf("invalid timecode seconds: %s", input)
	}
	seconds, err := strconv.Atoi(secParts[0])
	if err != nil {
		return 0, err
	}
	millis, err := strconv.Atoi(secParts[1])
	if err != nil {
		return 0, err
	}
	if millis < 10 {
		millis *= 100
	} else if millis < 100 {
		millis *= 10
	}

	total := int64(hours*3600*1000 + minutes*60*1000 + seconds*1000 + millis)
	return total, nil
}

func inferDurationMs(segments []storage.Segment) int64 {
	var maxEnd int64
	for _, segment := range segments {
		if segment.EndMs > maxEnd {
			maxEnd = segment.EndMs
		}
	}
	return maxEnd
}

type OCRVisualEmbedder struct {
	fallback VisualEmbedder
}

func NewOCRVisualEmbedder(fallback VisualEmbedder) *OCRVisualEmbedder {
	if fallback == nil {
		fallback = NewStubCLIPEmbedder()
	}
	return &OCRVisualEmbedder{fallback: fallback}
}

func (e *OCRVisualEmbedder) EmbedFrame(ctx context.Context, framePath string) ([]float32, string, error) {
	if strings.TrimSpace(framePath) == "" {
		return nil, "", fmt.Errorf("frame path is required")
	}

	embedding, _, err := e.fallback.EmbedFrame(ctx, framePath)
	if err != nil {
		return nil, "", err
	}

	description := fmt.Sprintf("keyframe from %s", filepath.Base(framePath))
	if _, err := exec.LookPath("tesseract"); err != nil {
		return embedding, description, nil
	}

	command := exec.CommandContext(ctx, "tesseract", framePath, "stdout")
	output, err := command.CombinedOutput()
	if err != nil {
		return embedding, description, nil
	}
	text := strings.Join(strings.Fields(string(output)), " ")
	if text == "" {
		return embedding, description, nil
	}
	if len(text) > 180 {
		text = text[:180]
	}

	return embedding, fmt.Sprintf("keyframe text: %s", text), nil
}

var _ Ingester = (*RealIngester)(nil)
var _ VisualEmbedder = (*OCRVisualEmbedder)(nil)
