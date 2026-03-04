package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/andreas-lindfalk/videra/internal/ingestion"
	"github.com/andreas-lindfalk/videra/internal/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	orchestrator       ingestion.IndexOrchestrator
	store              storage.VectorStore
	defaultSearchLimit int
	runtimeMode        string
	ranking            RankingOptions

	mcp *server.MCPServer
}

type RankingOptions struct {
	AudioWeight  float64
	VisualWeight float64
}

func New(name, version string, orchestrator ingestion.IndexOrchestrator, store storage.VectorStore, defaultSearchLimit int, runtimeMode string, ranking RankingOptions) *Server {
	if defaultSearchLimit <= 0 {
		defaultSearchLimit = 5
	}
	if ranking.AudioWeight <= 0 {
		ranking.AudioWeight = 1.0
	}
	if ranking.VisualWeight <= 0 {
		ranking.VisualWeight = 1.0
	}

	mcpServer := server.NewMCPServer(
		name,
		version,
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, false),
		server.WithRecovery(),
	)

	s := &Server{
		orchestrator:       orchestrator,
		store:              store,
		defaultSearchLimit: defaultSearchLimit,
		runtimeMode:        strings.ToLower(strings.TrimSpace(runtimeMode)),
		ranking:            ranking,
		mcp:                mcpServer,
	}

	s.registerTools()
	s.registerResources()

	return s
}

func (s *Server) MCP() *server.MCPServer {
	return s.mcp
}

func (s *Server) registerTools() {
	indexVideoTool := mcp.NewTool(
		"index_video",
		mcp.WithDescription("Takes a local path or URL and indexes the video into the vector store."),
		mcp.WithString(
			"path",
			mcp.Required(),
			mcp.Description("Local path or URL to the video"),
		),
	)

	s.mcp.AddTool(indexVideoTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		job := ingestion.NewIndexJobRequest(path, ingestion.IndexModeSync)
		result, err := s.orchestrator.Run(ctx, job)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if result.Status == ingestion.IndexJobStatusFailed {
			return mcp.NewToolResultError(result.Error), nil
		}
		if result.Video == nil {
			return mcp.NewToolResultError("indexing did not produce a video"), nil
		}

		video := result.Video

		payload := map[string]any{
			"videoId":        video.ID,
			"status":         video.Status,
			"filePath":       video.FilePath,
			"audioSegments":  video.AudioSegments,
			"visualSegments": video.VisualSegments,
			"modalities":     video.Modalities,
			"jobId":          result.JobID,
		}
		return mcp.NewToolResultStructured(payload, fmt.Sprintf("indexed video %s", video.ID)), nil
	})

	searchVideoTool := mcp.NewTool(
		"search_video",
		mcp.WithDescription("Hybrid semantic search across transcript and visual segments."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query text")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results")),
		mcp.WithBoolean("includeDebug", mcp.Description("Optional: include ranking/debug metadata in response")),
	)

	s.mcp.AddTool(searchVideoTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		limit := req.GetInt("limit", s.defaultSearchLimit)
		if limit <= 0 {
			limit = s.defaultSearchLimit
		}

		queryEmbedding := s.store.EmbedQuery(ctx, query)
		includeDebug := false
		if args, ok := req.Params.Arguments.(map[string]any); ok {
			if rawDebug, ok := args["includeDebug"]; ok {
				if parsed, ok := rawDebug.(bool); ok {
					includeDebug = parsed
				}
			}
		}

		results, err := s.store.SearchSegments(ctx, queryEmbedding, limit*3)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		hits := rerankHybridResults(results, limit, s.ranking, includeDebug)
		payload := map[string]any{
			"query":     query,
			"count":     len(hits),
			"results":   hits,
			"scoreMode": "weightedSimilarity",
		}
		if includeDebug {
			payload["debug"] = map[string]any{
				"audioWeight":    s.ranking.AudioWeight,
				"visualWeight":   s.ranking.VisualWeight,
				"candidateCount": len(results),
			}
		}
		fallback, _ := json.Marshal(payload)
		return mcp.NewToolResultStructured(payload, string(fallback)), nil
	})

	listVideosTool := mcp.NewTool(
		"list_videos",
		mcp.WithDescription("Lists currently indexed videos and their status."),
	)

	s.mcp.AddTool(listVideosTool, func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		videos, err := s.store.ListVideos(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		fallback, _ := json.Marshal(videos)
		return mcp.NewToolResultStructured(videos, string(fallback)), nil
	})

	if s.runtimeMode == "test" {
		resetTool := mcp.NewTool(
			"reset_index",
			mcp.WithDescription("Clears all indexed videos and segments. Available only in test runtime mode."),
		)

		s.mcp.AddTool(resetTool, func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if err := s.store.Reset(ctx); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			payload := map[string]any{"status": "reset"}
			return mcp.NewToolResultStructured(payload, "reset index"), nil
		})
	}
}

func (s *Server) registerResources() {
	template := mcp.NewResourceTemplate(
		"video://{id}/transcript",
		"video transcript",
	)

	s.mcp.AddResourceTemplate(template, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		videoID := parseVideoID(req.Params.URI)
		if videoID == "" {
			return nil, fmt.Errorf("invalid transcript URI: %s", req.Params.URI)
		}

		segments, err := s.store.GetTranscript(ctx, videoID)
		if err != nil {
			return nil, err
		}

		body, err := json.MarshalIndent(segments, "", "  ")
		if err != nil {
			return nil, err
		}

		content := mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(body),
		}

		return []mcp.ResourceContents{content}, nil
	})
}

func parseVideoID(uri string) string {
	if !strings.HasPrefix(uri, "video://") || !strings.HasSuffix(uri, "/transcript") {
		return ""
	}

	videoID := strings.TrimPrefix(uri, "video://")
	videoID = strings.TrimSuffix(videoID, "/transcript")
	return strings.TrimSpace(videoID)
}

func rerankHybridResults(results []storage.SearchResult, limit int, ranking RankingOptions, includeDebug bool) []storage.SearchHit {
	if len(results) == 0 || limit <= 0 {
		return nil
	}
	if ranking.AudioWeight <= 0 {
		ranking.AudioWeight = 1.0
	}
	if ranking.VisualWeight <= 0 {
		ranking.VisualWeight = 1.0
	}

	sorted := make([]storage.SearchResult, 0, len(results))
	sorted = append(sorted, results...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left := weightedScore(sorted[i], ranking)
		right := weightedScore(sorted[j], ranking)
		if left != right {
			return left > right
		}
		if sorted[i].Segment.VideoID != sorted[j].Segment.VideoID {
			return sorted[i].Segment.VideoID < sorted[j].Segment.VideoID
		}
		if sorted[i].Segment.StartMs != sorted[j].Segment.StartMs {
			return sorted[i].Segment.StartMs < sorted[j].Segment.StartMs
		}
		if sorted[i].Segment.EndMs != sorted[j].Segment.EndMs {
			return sorted[i].Segment.EndMs < sorted[j].Segment.EndMs
		}
		return sorted[i].Segment.Type < sorted[j].Segment.Type
	})

	selected := make([]storage.SearchHit, 0, limit)
	seen := map[string]struct{}{}
	hasAudio := false
	hasVisual := false

	for _, result := range sorted {
		key := fmt.Sprintf("%s:%d:%d:%s", result.Segment.VideoID, result.Segment.StartMs, result.Segment.EndMs, result.Segment.Type)
		if _, ok := seen[key]; ok {
			continue
		}

		if len(selected) < 2 {
			if result.Segment.Type == storage.SegmentTypeAudio && hasAudio {
				continue
			}
			if result.Segment.Type == storage.SegmentTypeVisual && hasVisual {
				continue
			}
		}

		weighted := float32(weightedScore(result, ranking))
		hit := storage.SearchHit{
			VideoID:       result.Segment.VideoID,
			StartMs:       result.Segment.StartMs,
			EndMs:         result.Segment.EndMs,
			Type:          result.Segment.Type,
			Snippet:       result.Segment.Text,
			Similarity:    weighted,
			SourcePath:    result.Segment.SourcePath,
			VisualContext: "",
		}
		if includeDebug {
			hit.RawSimilarity = result.Score
		}
		if result.Segment.Type == storage.SegmentTypeVisual {
			hit.VisualContext = result.Segment.Text
			hasVisual = true
		}
		if result.Segment.Type == storage.SegmentTypeAudio {
			hasAudio = true
		}

		selected = append(selected, hit)
		seen[key] = struct{}{}
		if len(selected) >= limit {
			break
		}
	}

	return selected
}

func weightedScore(result storage.SearchResult, ranking RankingOptions) float64 {
	base := float64(result.Score)
	if result.Segment.Type == storage.SegmentTypeVisual {
		return base * ranking.VisualWeight
	}
	return base * ranking.AudioWeight
}
