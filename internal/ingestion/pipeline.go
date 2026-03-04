package ingestion

import (
	"context"
	"time"

	"github.com/andreas-lindfalk/videra/internal/storage"
	"github.com/google/uuid"
)

type Ingester interface {
	IndexVideo(ctx context.Context, path string) (storage.Video, error)
}

type Transcriber interface {
	Transcribe(ctx context.Context, audioPath string) ([]storage.Segment, error)
}

type VisualEmbedder interface {
	EmbedFrame(ctx context.Context, framePath string) ([]float32, string, error)
}

type IndexOptions struct {
	FrameIntervalSec      int
	Concurrency           int
	RemoteFetchDisabled   bool
	RemoteFetchTimeoutSec int
	RemoteFetchMaxMB      int
}

type IndexMode string

const (
	IndexModeSync  IndexMode = "sync"
	IndexModeAsync IndexMode = "async"
)

type IndexJobStatus string

const (
	IndexJobStatusPending   IndexJobStatus = "pending"
	IndexJobStatusCompleted IndexJobStatus = "completed"
	IndexJobStatusFailed    IndexJobStatus = "failed"
)

type IndexJobRequest struct {
	JobID       string
	Path        string
	RequestedAt time.Time
	Mode        IndexMode
}

func NewIndexJobRequest(path string, mode IndexMode) IndexJobRequest {
	if mode == "" {
		mode = IndexModeSync
	}

	return IndexJobRequest{
		JobID:       uuid.NewString(),
		Path:        path,
		RequestedAt: time.Now().UTC(),
		Mode:        mode,
	}
}

type IndexJobResult struct {
	JobID  string
	Status IndexJobStatus
	Video  *storage.Video
	Error  string
}

type IndexOrchestrator interface {
	Run(ctx context.Context, req IndexJobRequest) (IndexJobResult, error)
}

type SourceVideoLookup interface {
	GetVideoBySourcePath(ctx context.Context, sourcePath string) (storage.Video, bool)
}
