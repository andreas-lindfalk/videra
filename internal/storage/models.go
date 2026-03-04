package storage

import "time"

type VideoStatus string

const (
	VideoStatusPending VideoStatus = "pending"
	VideoStatusIndexed VideoStatus = "indexed"
	VideoStatusFailed  VideoStatus = "failed"
)

type Video struct {
	ID             string      `json:"id"`
	FilePath       string      `json:"filePath"`
	Status         VideoStatus `json:"status"`
	Indexed        time.Time   `json:"indexedAt"`
	Duration       int64       `json:"durationMs"`
	AudioSegments  int         `json:"audioSegments"`
	VisualSegments int         `json:"visualSegments"`
	Modalities     []string    `json:"modalities"`
}

type SegmentType string

const (
	SegmentTypeAudio  SegmentType = "audio"
	SegmentTypeVisual SegmentType = "visual"
)

type Segment struct {
	VideoID    string      `json:"videoId"`
	StartMs    int64       `json:"startMs"`
	EndMs      int64       `json:"endMs"`
	Text       string      `json:"text"`
	Embedding  []float32   `json:"embedding,omitempty"`
	Type       SegmentType `json:"type"`
	SourcePath string      `json:"sourcePath,omitempty"`
}

type SearchResult struct {
	Segment Segment `json:"segment"`
	Score   float32 `json:"score"`
}

type SearchHit struct {
	VideoID       string      `json:"videoId"`
	StartMs       int64       `json:"startMs"`
	EndMs         int64       `json:"endMs"`
	Type          SegmentType `json:"type"`
	Snippet       string      `json:"snippet"`
	VisualContext string      `json:"visualContext,omitempty"`
	Similarity    float32     `json:"similarity"`
	SourcePath    string      `json:"sourcePath,omitempty"`
}
