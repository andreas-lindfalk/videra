package storage

import "context"

type VectorStore interface {
	IndexVideo(ctx context.Context, video Video, segments []Segment) error
	SearchSegments(ctx context.Context, queryEmbedding []float32, limit int) ([]SearchResult, error)
	EmbedQuery(ctx context.Context, query string) []float32
	ListVideos(ctx context.Context) ([]Video, error)
	GetVideoBySourcePath(ctx context.Context, sourcePath string) (Video, bool)
	GetTranscript(ctx context.Context, videoID string) ([]Segment, error)
	Reset(ctx context.Context) error
}
