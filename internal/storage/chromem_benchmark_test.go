package storage

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/andreas-lindfalk/videra/internal/embedding"
)

func BenchmarkChromemStoreBaseline(b *testing.B) {
	b.Run("IndexVideo_8Segments", func(b *testing.B) {
		ctx := context.Background()
		store := newBenchmarkStore(b, false)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			video := benchmarkVideo(i)
			if err := store.IndexVideo(ctx, video, benchmarkSegments(video, 8)); err != nil {
				b.Fatalf("index video: %v", err)
			}
		}
	})

	b.Run("SearchSegments_Top5_Corpus200x8", func(b *testing.B) {
		ctx := context.Background()
		store := newBenchmarkStore(b, false)
		populateBenchmarkCorpus(b, ctx, store, 200, 8)
		queryEmbedding := store.EmbedQuery(ctx, "roadmap budget planning")

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			results, err := store.SearchSegments(ctx, queryEmbedding, 5)
			if err != nil {
				b.Fatalf("search segments: %v", err)
			}
			if len(results) == 0 {
				b.Fatal("search returned no results")
			}
		}
	})

	b.Run("ListVideos_Corpus200", func(b *testing.B) {
		ctx := context.Background()
		store := newBenchmarkStore(b, false)
		populateBenchmarkCorpus(b, ctx, store, 200, 8)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			videos, err := store.ListVideos(ctx)
			if err != nil {
				b.Fatalf("list videos: %v", err)
			}
			if len(videos) != 200 {
				b.Fatalf("expected 200 videos, got %d", len(videos))
			}
		}
	})

	b.Run("GetTranscript_8Segments", func(b *testing.B) {
		ctx := context.Background()
		store := newBenchmarkStore(b, false)
		video := benchmarkVideo(0)
		if err := store.IndexVideo(ctx, video, benchmarkSegments(video, 8)); err != nil {
			b.Fatalf("index video: %v", err)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			transcript, err := store.GetTranscript(ctx, video.ID)
			if err != nil {
				b.Fatalf("get transcript: %v", err)
			}
			if len(transcript) != 8 {
				b.Fatalf("expected 8 segments, got %d", len(transcript))
			}
		}
	})

	b.Run("Reset", func(b *testing.B) {
		ctx := context.Background()
		store := newBenchmarkStore(b, false)
		populateBenchmarkCorpus(b, ctx, store, 20, 8)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := store.Reset(ctx); err != nil {
				b.Fatalf("reset store: %v", err)
			}
		}
	})
}

func newBenchmarkStore(b *testing.B, splitSharedStorage bool) VectorStore {
	b.Helper()

	backend := strings.ToLower(strings.TrimSpace(os.Getenv("VIDERA_STORAGE_BACKEND")))
	if backend == "" {
		backend = "chromem"
	}

	dataDir := b.TempDir()
	textEmbedder := embedding.NewDeterministicTextEmbedder()

	switch backend {
	case "chromem":
		store, err := NewChromemStoreWithOptions(
			dataDir,
			textEmbedder,
			ChromemStoreOptions{SplitSharedStorage: splitSharedStorage},
		)
		if err != nil {
			b.Fatalf("create chromem benchmark store: %v", err)
		}
		return store
	case "lancedb":
		store, err := NewLanceDBStoreWithOptions(
			dataDir,
			textEmbedder,
			LanceDBStoreOptions{SplitSharedStorage: splitSharedStorage},
		)
		if err != nil {
			b.Fatalf("create lancedb benchmark store: %v", err)
		}
		return store
	default:
		b.Fatalf("unsupported VIDERA_STORAGE_BACKEND in benchmark: %s", backend)
		return nil
	}
}

func populateBenchmarkCorpus(b *testing.B, ctx context.Context, store VectorStore, videoCount int, segmentsPerVideo int) {
	b.Helper()

	for i := 0; i < videoCount; i++ {
		video := benchmarkVideo(i)
		if err := store.IndexVideo(ctx, video, benchmarkSegments(video, segmentsPerVideo)); err != nil {
			b.Fatalf("populate corpus index video %d: %v", i, err)
		}
	}
}

func benchmarkVideo(index int) Video {
	indexedAt := time.Unix(0, 0).UTC().Add(time.Duration(index) * time.Second)
	videoID := fmt.Sprintf("bench-video-%06d", index)
	filePath := fmt.Sprintf("https://example.com/%s.mp4", videoID)

	return Video{
		ID:             videoID,
		FilePath:       filePath,
		Status:         VideoStatusIndexed,
		Indexed:        indexedAt,
		Duration:       int64(8 * 5000),
		AudioSegments:  4,
		VisualSegments: 4,
		Modalities:     []string{"audio", "visual"},
	}
}

func benchmarkSegments(video Video, count int) []Segment {
	segments := make([]Segment, 0, count)
	for index := 0; index < count; index++ {
		startMs := int64(index * 5000)
		endMs := startMs + 5000
		typeName := SegmentTypeAudio
		textPrefix := "audio"
		if index%2 == 1 {
			typeName = SegmentTypeVisual
			textPrefix = "visual"
		}

		segments = append(segments, Segment{
			VideoID:    video.ID,
			StartMs:    startMs,
			EndMs:      endMs,
			Text:       fmt.Sprintf("%s segment %d for roadmap budget planning in %s", textPrefix, index, video.ID),
			Type:       typeName,
			SourcePath: video.FilePath,
		})
	}

	return segments
}
