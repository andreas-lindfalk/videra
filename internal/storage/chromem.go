package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"

	"github.com/andreas-lindfalk/videra/internal/embedding"
	chromem "github.com/philippgille/chromem-go"
)

const segmentsCollection = "segments"

type ChromemStore struct {
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
	if dataDir == "" {
		dataDir = "./data"
	}
	if textEmbedder == nil {
		textEmbedder = embedding.NewDeterministicTextEmbedder()
	}

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
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

	return &ChromemStore{
		db:              db,
		collection:      collection,
		embedder:        textEmbedder,
		videos:          map[string]Video{},
		transcripts:     map[string][]Segment{},
		videosByPath:    map[string]string{},
		segmentCounters: map[string]int{},
	}, nil
}

func (s *ChromemStore) IndexVideo(ctx context.Context, video Video, segments []Segment) error {
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
	if err != nil {
		return s.fallbackSegments(limit), nil
	}

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

	return out, nil
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

	return nil
}

var _ VectorStore = (*ChromemStore)(nil)
