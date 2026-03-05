package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/andreas-lindfalk/videra/internal/embedding"
	chromem "github.com/philippgille/chromem-go"
)

const segmentsCollection = "segments"
const manifestDirName = "video-manifests"

type ChromemStoreOptions struct {
	SplitSharedStorage bool
}

type persistedVideoManifest struct {
	Video    Video     `json:"video"`
	Segments []Segment `json:"segments"`
}

type ChromemStore struct {
	dataDir    string
	syncShared bool

	db         *chromem.DB
	collection *chromem.Collection
	resetCount int
	embedder   embedding.TextEmbedder

	mu              sync.RWMutex
	videos          map[string]Video
	transcripts     map[string][]Segment
	videosByPath    map[string]string
	segmentCounters map[string]int
}

func NewChromemStore(dataDir string, textEmbedder embedding.TextEmbedder) (*ChromemStore, error) {
	return NewChromemStoreWithOptions(dataDir, textEmbedder, ChromemStoreOptions{})
}

func NewChromemStoreWithOptions(dataDir string, textEmbedder embedding.TextEmbedder, options ChromemStoreOptions) (*ChromemStore, error) {
	if dataDir == "" {
		dataDir = "./data"
	}
	if textEmbedder == nil {
		textEmbedder = embedding.NewDeterministicTextEmbedder()
	}

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	if options.SplitSharedStorage {
		if err := os.MkdirAll(filepath.Join(dataDir, manifestDirName), 0o755); err != nil {
			return nil, fmt.Errorf("create manifest dir: %w", err)
		}
	}

	dbPath := filepath.Join(dataDir, "chromem.gob")
	db, err := chromem.NewPersistentDB(dbPath, true)
	if err != nil {
		return nil, fmt.Errorf("create persistent chromem db: %w", err)
	}

	collection, err := db.GetOrCreateCollection(segmentsCollection, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get or create segments collection: %w", err)
	}

	store := &ChromemStore{
		dataDir:         dataDir,
		syncShared:      options.SplitSharedStorage,
		db:              db,
		collection:      collection,
		embedder:        textEmbedder,
		videos:          map[string]Video{},
		transcripts:     map[string][]Segment{},
		videosByPath:    map[string]string{},
		segmentCounters: map[string]int{},
	}

	if store.syncShared {
		if err := store.loadManifestsFromDisk(); err != nil {
			return nil, fmt.Errorf("load persisted manifests: %w", err)
		}
	}

	return store, nil
}

func (s *ChromemStore) IndexVideo(ctx context.Context, video Video, segments []Segment) error {
	if s.syncShared {
		if err := s.loadManifestsFromDisk(); err != nil {
			return fmt.Errorf("load persisted manifests: %w", err)
		}
	}

	s.mu.RLock()
	collection := s.collection
	s.mu.RUnlock()

	s.mu.Lock()
	if existingID, ok := s.videosByPath[video.FilePath]; ok {
		if _, exists := s.videos[existingID]; exists {
			s.mu.Unlock()
			return nil
		}
	}
	nextID := s.segmentCounters[video.ID]
	s.segmentCounters[video.ID] = nextID + len(segments)
	s.mu.Unlock()

	docs := make([]chromem.Document, 0, len(segments))
	for idx, segment := range segments {
		embedding := segment.Embedding
		if len(embedding) == 0 {
			computed, err := s.embedder.EmbedText(ctx, segment.Text)
			if err != nil {
				return fmt.Errorf("embed segment text: %w", err)
			}
			embedding = computed
		}

		doc := chromem.Document{
			ID: fmt.Sprintf("%s:%d", video.ID, nextID+idx),
			Metadata: map[string]string{
				"video_id":    video.ID,
				"file_path":   video.FilePath,
				"start_ms":    strconv.FormatInt(segment.StartMs, 10),
				"end_ms":      strconv.FormatInt(segment.EndMs, 10),
				"type":        string(segment.Type),
				"source_path": segment.SourcePath,
			},
			Embedding: embedding,
			Content:   segment.Text,
		}
		docs = append(docs, doc)
	}

	if len(docs) > 0 {
		if err := collection.AddDocuments(ctx, docs, 4); err != nil {
			return fmt.Errorf("add documents to chromem: %w", err)
		}
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

func (s *ChromemStore) SearchSegments(ctx context.Context, queryEmbedding []float32, limit int) ([]SearchResult, error) {
	s.mu.RLock()
	collection := s.collection
	s.mu.RUnlock()

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

	results, err := collection.QueryEmbedding(ctx, queryEmbedding, limit, nil, nil)
	if err == nil && len(results) > 0 {
		mapped := mapChromemResults(results)
		fallback := s.fallbackSegments(limit)
		combined := make([]SearchResult, 0, len(mapped)+len(fallback))
		combined = append(combined, mapped...)
		combined = append(combined, fallback...)
		return combined, nil
	}

	if s.syncShared {
		_ = s.loadManifestsFromDisk()
	}

	if err != nil {
		return s.fallbackSegments(limit), nil
	}
	return s.fallbackSegments(limit), nil
}

func (s *ChromemStore) fallbackSegments(limit int) []SearchResult {
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

func (s *ChromemStore) EmbedQuery(_ context.Context, query string) []float32 {
	vector, err := s.embedder.EmbedText(context.Background(), query)
	if err != nil {
		fallback, _ := embedding.NewDeterministicTextEmbedder().EmbedText(context.Background(), query)
		return fallback
	}
	return vector
}

func (s *ChromemStore) ListVideos(_ context.Context) ([]Video, error) {
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

func (s *ChromemStore) GetVideoBySourcePath(_ context.Context, sourcePath string) (Video, bool) {
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

func (s *ChromemStore) GetTranscript(_ context.Context, videoID string) ([]Segment, error) {
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

func (s *ChromemStore) Reset(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.resetCount++
	collectionName := fmt.Sprintf("%s_reset_%d", segmentsCollection, s.resetCount)
	collection, err := s.db.GetOrCreateCollection(collectionName, nil, nil)
	if err != nil {
		return fmt.Errorf("create reset collection: %w", err)
	}

	s.collection = collection
	s.videos = map[string]Video{}
	s.transcripts = map[string][]Segment{}
	s.videosByPath = map[string]string{}
	s.segmentCounters = map[string]int{}

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

func mapChromemResults(results []chromem.Result) []SearchResult {
	out := make([]SearchResult, 0, len(results))
	for _, result := range results {
		startMs, _ := strconv.ParseInt(result.Metadata["start_ms"], 10, 64)
		endMs, _ := strconv.ParseInt(result.Metadata["end_ms"], 10, 64)
		segmentType := SegmentType(result.Metadata["type"])

		out = append(out, SearchResult{
			Segment: Segment{
				VideoID:    result.Metadata["video_id"],
				StartMs:    startMs,
				EndMs:      endMs,
				Text:       result.Content,
				Type:       segmentType,
				SourcePath: result.Metadata["source_path"],
			},
			Score: result.Similarity,
		})
	}

	return out
}

func (s *ChromemStore) loadManifestsFromDisk() error {
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
	s.segmentCounters = segmentCounters

	return nil
}

func (s *ChromemStore) persistVideoManifest(video Video, segments []Segment) error {
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

	manifestPath := filepath.Join(manifestDir, sanitizeManifestFileName(video.ID)+".json")
	if err := os.Rename(tempPath, manifestPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("commit manifest file: %w", err)
	}

	return nil
}

func sanitizeManifestFileName(videoID string) string {
	trimmed := strings.TrimSpace(videoID)
	if trimmed == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", " ", "_")
	return replacer.Replace(trimmed)
}

var _ VectorStore = (*ChromemStore)(nil)
