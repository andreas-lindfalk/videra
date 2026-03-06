package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/andreas-lindfalk/videra/internal/embedding"
)

const lanceDBDataDirName = "lancedb"
const defaultLanceDBTable = "videra_segments"

const (
	lanceDBFieldDocID      = "doc_id"
	lanceDBFieldVideoID    = "video_id"
	lanceDBFieldFilePath   = "file_path"
	lanceDBFieldStartMs    = "start_ms"
	lanceDBFieldEndMs      = "end_ms"
	lanceDBFieldType       = "type"
	lanceDBFieldSourcePath = "source_path"
	lanceDBFieldText       = "text"
	lanceDBFieldEmbedding  = "embedding"
)

type LanceDBStoreOptions struct {
	SplitSharedStorage bool
	URI                string
	Region             string
	TableName          string
	Bridge             lanceDBBridge
}

type LanceDBStore struct {
	dataDir      string
	syncShared   bool
	embedder     embedding.TextEmbedder
	bridge       lanceDBBridge
	segmentCount map[string]int

	mu           sync.RWMutex
	videos       map[string]Video
	transcripts  map[string][]Segment
	videosByPath map[string]string
}

var _ VectorStore = (*LanceDBStore)(nil)

type lanceDBBridge interface {
	UpsertSegments(ctx context.Context, rows []lanceDBSegmentRow) error
	SearchSegments(ctx context.Context, queryEmbedding []float32, limit int) ([]map[string]any, error)
	Reset(ctx context.Context) error
}

type lanceDBSegmentRow struct {
	DocID      string    `json:"doc_id"`
	VideoID    string    `json:"video_id"`
	FilePath   string    `json:"file_path"`
	StartMs    int64     `json:"start_ms"`
	EndMs      int64     `json:"end_ms"`
	Type       string    `json:"type"`
	SourcePath string    `json:"source_path"`
	Text       string    `json:"text"`
	Embedding  []float32 `json:"embedding"`
}

func NewLanceDBStore(dataDir string, textEmbedder embedding.TextEmbedder) (*LanceDBStore, error) {
	return NewLanceDBStoreWithOptions(dataDir, textEmbedder, LanceDBStoreOptions{})
}

func NewLanceDBStoreWithOptions(dataDir string, textEmbedder embedding.TextEmbedder, options LanceDBStoreOptions) (*LanceDBStore, error) {
	normalizedDataDir := strings.TrimSpace(dataDir)
	if normalizedDataDir == "" {
		normalizedDataDir = "./data"
	}
	if textEmbedder == nil {
		textEmbedder = embedding.NewDeterministicTextEmbedder()
	}

	lanceDataDir := filepath.Join(normalizedDataDir, lanceDBDataDirName)
	if err := os.MkdirAll(lanceDataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create lancedb data dir: %w", err)
	}
	if options.SplitSharedStorage {
		if err := os.MkdirAll(filepath.Join(lanceDataDir, manifestDirName), 0o755); err != nil {
			return nil, fmt.Errorf("create lancedb manifest dir: %w", err)
		}
	}

	tableName := strings.TrimSpace(options.TableName)
	if tableName == "" {
		tableName = defaultLanceDBTable
	}

	region := strings.TrimSpace(options.Region)

	uri := strings.TrimSpace(options.URI)
	if uri == "" {
		uri = filepath.Join(lanceDataDir, "db")
	}
	if !strings.Contains(uri, "://") {
		if err := os.MkdirAll(uri, 0o755); err != nil {
			return nil, fmt.Errorf("create lancedb uri dir: %w", err)
		}
	}

	bridge := options.Bridge
	if bridge == nil {
		nativeBridge, err := newNativeLanceDBBridge(context.Background(), uri, tableName, region)
		if err != nil {
			return nil, fmt.Errorf("initialize native lancedb bridge: %w", err)
		}
		bridge = nativeBridge
	}

	store := &LanceDBStore{
		dataDir:      lanceDataDir,
		syncShared:   options.SplitSharedStorage,
		embedder:     textEmbedder,
		bridge:       bridge,
		segmentCount: map[string]int{},
		videos:       map[string]Video{},
		transcripts:  map[string][]Segment{},
		videosByPath: map[string]string{},
	}

	if store.syncShared {
		if err := store.loadManifestsFromDisk(); err != nil {
			return nil, fmt.Errorf("load persisted manifests: %w", err)
		}
	}

	return store, nil
}

func (s *LanceDBStore) IndexVideo(ctx context.Context, video Video, segments []Segment) error {
	if s.syncShared {
		if err := s.loadManifestsFromDisk(); err != nil {
			return fmt.Errorf("load persisted manifests: %w", err)
		}
	}

	s.mu.Lock()
	if existingID, ok := s.videosByPath[video.FilePath]; ok {
		if _, exists := s.videos[existingID]; exists {
			s.mu.Unlock()
			return nil
		}
	}
	nextID := s.segmentCount[video.ID]
	s.segmentCount[video.ID] = nextID + len(segments)
	s.mu.Unlock()

	rows := make([]lanceDBSegmentRow, 0, len(segments))
	for idx, segment := range segments {
		embeddingVector := segment.Embedding
		if len(embeddingVector) == 0 {
			computed, err := s.embedder.EmbedText(ctx, segment.Text)
			if err != nil {
				return fmt.Errorf("embed segment text: %w", err)
			}
			embeddingVector = computed
		}

		rows = append(rows, lanceDBSegmentRow{
			DocID:      fmt.Sprintf("%s:%d", video.ID, nextID+idx),
			VideoID:    video.ID,
			FilePath:   video.FilePath,
			StartMs:    segment.StartMs,
			EndMs:      segment.EndMs,
			Type:       string(segment.Type),
			SourcePath: segment.SourcePath,
			Text:       segment.Text,
			Embedding:  embeddingVector,
		})
	}

	if err := s.bridge.UpsertSegments(ctx, rows); err != nil {
		return fmt.Errorf("upsert lancedb segments: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.videos[video.ID] = video
	s.videosByPath[video.FilePath] = video.ID
	s.transcripts[video.ID] = append([]Segment(nil), segments...)
	if s.syncShared {
		if err := s.persistVideoManifest(video, segments); err != nil {
			return fmt.Errorf("persist video manifest: %w", err)
		}
	}

	return nil
}

func (s *LanceDBStore) SearchSegments(ctx context.Context, queryEmbedding []float32, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 5
	}

	if len(queryEmbedding) == 0 {
		computed, err := s.embedder.EmbedText(ctx, "default-query")
		if err != nil {
			return nil, fmt.Errorf("embed default query: %w", err)
		}
		queryEmbedding = computed
	}

	rows, err := s.bridge.SearchSegments(ctx, queryEmbedding, limit)
	if err == nil && len(rows) > 0 {
		mapped := mapLanceDBResults(rows)
		if len(mapped) > limit {
			return mapped[:limit], nil
		}
		return mapped, nil
	}

	if s.syncShared {
		_ = s.loadManifestsFromDisk()
	}

	if err != nil {
		return s.fallbackSegments(limit), nil
	}

	return s.fallbackSegments(limit), nil
}

func (s *LanceDBStore) EmbedQuery(_ context.Context, query string) []float32 {
	vector, err := s.embedder.EmbedText(context.Background(), query)
	if err != nil {
		fallback, _ := embedding.NewDeterministicTextEmbedder().EmbedText(context.Background(), query)
		return fallback
	}
	return vector
}

func (s *LanceDBStore) ListVideos(_ context.Context) ([]Video, error) {
	if s.syncShared {
		if err := s.loadManifestsFromDisk(); err != nil {
			return nil, fmt.Errorf("load persisted manifests: %w", err)
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Video, 0, len(s.videos))
	for _, video := range s.videos {
		out = append(out, video)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Indexed.Before(out[j].Indexed)
	})

	return out, nil
}

func (s *LanceDBStore) GetVideoBySourcePath(_ context.Context, sourcePath string) (Video, bool) {
	if s.syncShared {
		if err := s.loadManifestsFromDisk(); err != nil {
			return Video{}, false
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	videoID, ok := s.videosByPath[sourcePath]
	if !ok {
		return Video{}, false
	}

	video, ok := s.videos[videoID]
	return video, ok
}

func (s *LanceDBStore) GetTranscript(_ context.Context, videoID string) ([]Segment, error) {
	if s.syncShared {
		if err := s.loadManifestsFromDisk(); err != nil {
			return nil, fmt.Errorf("load persisted manifests: %w", err)
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	segments, ok := s.transcripts[videoID]
	if !ok {
		return nil, fmt.Errorf("video transcript not found: %s", videoID)
	}

	return append([]Segment(nil), segments...), nil
}

func (s *LanceDBStore) Reset(ctx context.Context) error {
	if err := s.bridge.Reset(ctx); err != nil {
		return fmt.Errorf("reset lancedb table: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.videos = map[string]Video{}
	s.transcripts = map[string][]Segment{}
	s.videosByPath = map[string]string{}
	s.segmentCount = map[string]int{}

	if s.syncShared {
		manifestDir := filepath.Join(s.dataDir, manifestDirName)
		if err := os.RemoveAll(manifestDir); err != nil {
			return fmt.Errorf("clear manifest dir: %w", err)
		}
		if err := os.MkdirAll(manifestDir, 0o755); err != nil {
			return fmt.Errorf("recreate manifest dir: %w", err)
		}
	}

	return nil
}

func (s *LanceDBStore) fallbackSegments(limit int) []SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]SearchResult, 0, limit)
	for _, segments := range s.transcripts {
		for _, segment := range segments {
			out = append(out, SearchResult{Segment: segment, Score: 0.1})
			if len(out) >= limit {
				return out
			}
		}
	}
	return out
}

func (s *LanceDBStore) loadManifestsFromDisk() error {
	manifestDir := filepath.Join(s.dataDir, manifestDirName)
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		return fmt.Errorf("create manifest dir: %w", err)
	}

	entries, err := os.ReadDir(manifestDir)
	if err != nil {
		return fmt.Errorf("read manifest dir: %w", err)
	}

	videos := map[string]Video{}
	transcripts := map[string][]Segment{}
	videosByPath := map[string]string{}
	segmentCounters := map[string]int{}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}

		manifestPath := filepath.Join(manifestDir, entry.Name())
		payload, readErr := os.ReadFile(manifestPath)
		if readErr != nil {
			return fmt.Errorf("read manifest %s: %w", entry.Name(), readErr)
		}

		var manifest persistedVideoManifest
		if decodeErr := json.Unmarshal(payload, &manifest); decodeErr != nil {
			return fmt.Errorf("decode manifest %s: %w", entry.Name(), decodeErr)
		}
		if strings.TrimSpace(manifest.Video.ID) == "" || strings.TrimSpace(manifest.Video.FilePath) == "" {
			continue
		}

		copiedSegments := append([]Segment(nil), manifest.Segments...)
		videos[manifest.Video.ID] = manifest.Video
		transcripts[manifest.Video.ID] = copiedSegments
		videosByPath[manifest.Video.FilePath] = manifest.Video.ID
		segmentCounters[manifest.Video.ID] = len(copiedSegments)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.videos = videos
	s.transcripts = transcripts
	s.videosByPath = videosByPath
	s.segmentCount = segmentCounters

	return nil
}

func (s *LanceDBStore) persistVideoManifest(video Video, segments []Segment) error {
	manifestDir := filepath.Join(s.dataDir, manifestDirName)
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		return fmt.Errorf("create manifest dir: %w", err)
	}

	copiedSegments := make([]Segment, 0, len(segments))
	for _, segment := range segments {
		copied := segment
		copied.Embedding = nil
		copiedSegments = append(copiedSegments, copied)
	}

	manifest := persistedVideoManifest{
		Video:    video,
		Segments: copiedSegments,
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	tempFile, err := os.CreateTemp(manifestDir, "manifest-*.json")
	if err != nil {
		return fmt.Errorf("create manifest temp file: %w", err)
	}
	tempPath := tempFile.Name()
	if _, err := tempFile.Write(payload); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		return fmt.Errorf("write manifest temp file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("close manifest temp file: %w", err)
	}

	manifestPath := filepath.Join(s.dataDir, manifestDirName, sanitizeManifestFileName(video.ID)+".json")
	if err := os.Rename(tempPath, manifestPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("commit manifest file: %w", err)
	}

	return nil
}

func mapLanceDBResults(rows []map[string]any) []SearchResult {
	out := make([]SearchResult, 0, len(rows))
	for _, row := range rows {
		typeName := strings.TrimSpace(valueAsString(row, lanceDBFieldType))
		segmentType := SegmentTypeAudio
		if typeName == string(SegmentTypeVisual) {
			segmentType = SegmentTypeVisual
		}

		out = append(out, SearchResult{
			Segment: Segment{
				VideoID:    valueAsString(row, lanceDBFieldVideoID),
				StartMs:    valueAsInt64(row, lanceDBFieldStartMs),
				EndMs:      valueAsInt64(row, lanceDBFieldEndMs),
				Text:       valueAsString(row, lanceDBFieldText),
				Type:       segmentType,
				SourcePath: valueAsString(row, lanceDBFieldSourcePath),
			},
			Score: valueAsScore(row),
		})
	}
	return out
}

func valueAsString(row map[string]any, key string) string {
	raw, ok := row[key]
	if !ok || raw == nil {
		return ""
	}
	switch value := raw.(type) {
	case string:
		return value
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case int:
		return strconv.Itoa(value)
	case int64:
		return strconv.FormatInt(value, 10)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func valueAsInt64(row map[string]any, key string) int64 {
	raw, ok := row[key]
	if !ok || raw == nil {
		return 0
	}
	switch value := raw.(type) {
	case int:
		return int64(value)
	case int64:
		return value
	case float64:
		return int64(value)
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err == nil {
			return parsed
		}
	}
	return 0
}

func valueAsScore(row map[string]any) float32 {
	for _, key := range []string{"score", "similarity", "_score"} {
		if score, ok := valueAsFloat64(row, key); ok {
			return float32(score)
		}
	}

	for _, key := range []string{"_distance", "distance"} {
		if distance, ok := valueAsFloat64(row, key); ok {
			if distance < 0 {
				distance = math.Abs(distance)
			}
			return float32(1.0 / (1.0 + distance))
		}
	}

	return 0.1
}

func valueAsFloat64(row map[string]any, key string) (float64, bool) {
	raw, ok := row[key]
	if !ok || raw == nil {
		return 0, false
	}
	switch value := raw.(type) {
	case float64:
		return value, true
	case float32:
		return float64(value), true
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err == nil {
			return parsed, true
		}
	}
	return 0, false
}
