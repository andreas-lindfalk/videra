package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"

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
		mcp.WithString(
			"mode",
			mcp.Description("Indexing mode: sync (default) or async"),
		),
	)

	s.mcp.AddTool(indexVideoTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := req.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		mode, err := parseIndexMode(req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		job := ingestion.NewIndexJobRequest(path, mode)
		result, err := s.orchestrator.Run(ctx, job)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if result.Status == ingestion.IndexJobStatusFailed {
			return mcp.NewToolResultError(result.Error), nil
		}

		if mode == ingestion.IndexModeAsync && result.Status == ingestion.IndexJobStatusPending {
			payload := map[string]any{
				"jobId":  result.JobID,
				"status": result.Status,
				"mode":   string(mode),
				"path":   strings.TrimSpace(path),
			}
			return mcp.NewToolResultStructured(payload, fmt.Sprintf("scheduled indexing job %s", result.JobID)), nil
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
			"mode":           string(mode),
		}
		return mcp.NewToolResultStructured(payload, fmt.Sprintf("indexed video %s", video.ID)), nil
	})

	getIndexJobTool := mcp.NewTool(
		"get_index_job",
		mcp.WithDescription("Returns status for an indexing job started by index_video in async mode."),
		mcp.WithString(
			"jobId",
			mcp.Required(),
			mcp.Description("Index job ID returned by index_video"),
		),
	)

	s.mcp.AddTool(getIndexJobTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		jobID, err := req.RequireString("jobId")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		reader, ok := s.orchestrator.(ingestion.IndexJobReader)
		if !ok {
			return mcp.NewToolResultError("index job status lookup is not supported by current orchestrator"), nil
		}

		result, found := reader.GetJob(ctx, jobID)
		if !found {
			return mcp.NewToolResultError(fmt.Sprintf("index job not found: %s", strings.TrimSpace(jobID))), nil
		}

		payload := map[string]any{
			"jobId":  result.JobID,
			"status": result.Status,
		}
		if result.Error != "" {
			payload["error"] = result.Error
		}
		if result.Video != nil {
			payload["videoId"] = result.Video.ID
			payload["filePath"] = result.Video.FilePath
			payload["audioSegments"] = result.Video.AudioSegments
			payload["visualSegments"] = result.Video.VisualSegments
			payload["modalities"] = result.Video.Modalities
		}

		return mcp.NewToolResultStructured(payload, fmt.Sprintf("index job %s is %s", result.JobID, result.Status)), nil
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

		hits := rerankHybridResults(results, query, limit, s.ranking, includeDebug)
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

func rerankHybridResults(results []storage.SearchResult, query string, limit int, ranking RankingOptions, includeDebug bool) []storage.SearchHit {
	if len(results) == 0 || limit <= 0 {
		return nil
	}
	if ranking.AudioWeight <= 0 {
		ranking.AudioWeight = 1.0
	}
	if ranking.VisualWeight <= 0 {
		ranking.VisualWeight = 1.0
	}
	normalizedQuery := normalizeSearchText(query)

	sorted := make([]storage.SearchResult, 0, len(results))
	sorted = append(sorted, results...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left := weightedScore(sorted[i], ranking) + lexicalMatchBoost(sorted[i].Segment.Text, normalizedQuery)
		right := weightedScore(sorted[j], ranking) + lexicalMatchBoost(sorted[j].Segment.Text, normalizedQuery)
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
	deferred := make([]storage.SearchResult, 0, len(sorted))
	seen := map[string]struct{}{}
	hasAudio := false
	hasVisual := false

	for _, result := range sorted {
		key := fmt.Sprintf("%s:%d:%d:%s", result.Segment.VideoID, result.Segment.StartMs, result.Segment.EndMs, result.Segment.Type)
		if _, ok := seen[key]; ok {
			continue
		}

		lexical := lexicalMatchBoost(result.Segment.Text, normalizedQuery)

		if len(selected) < 2 {
			if lexical == 0 && result.Segment.Type == storage.SegmentTypeAudio && hasAudio {
				deferred = append(deferred, result)
				continue
			}
			if lexical == 0 && result.Segment.Type == storage.SegmentTypeVisual && hasVisual {
				deferred = append(deferred, result)
				continue
			}
		}

		weighted := float32(weightedScore(result, ranking) + lexical)
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

	if len(selected) < limit && len(deferred) > 0 {
		for _, result := range deferred {
			key := fmt.Sprintf("%s:%d:%d:%s", result.Segment.VideoID, result.Segment.StartMs, result.Segment.EndMs, result.Segment.Type)
			if _, ok := seen[key]; ok {
				continue
			}

			weighted := float32(weightedScore(result, ranking) + lexicalMatchBoost(result.Segment.Text, normalizedQuery))
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
			}

			selected = append(selected, hit)
			seen[key] = struct{}{}
			if len(selected) >= limit {
				break
			}
		}
	}

	return selected
}

func lexicalMatchBoost(snippet, normalizedQuery string) float64 {
	if normalizedQuery == "" {
		return 0
	}
	normalizedSnippet := normalizeSearchText(snippet)
	if normalizedSnippet == "" {
		return 0
	}

	queryTokens := strings.Fields(normalizedQuery)
	snippetTokens := strings.Fields(normalizedSnippet)

	if len(queryTokens) == 0 || len(snippetTokens) == 0 {
		return 0
	}

	boost := 0.0
	if normalizedSnippet == normalizedQuery {
		boost += 5.0
	}
	if strings.Contains(normalizedSnippet, normalizedQuery) {
		boost += 2.0
	}

	querySet := makeTokenSet(queryTokens)
	overlap := tokenOverlapCount(querySet, snippetTokens)
	if overlap > 0 {
		boost += 2.0
		ratio := float64(overlap) / float64(len(querySet))
		boost += ratio * 4.0
		if overlap == len(querySet) {
			boost += 2.0
		}
	}

	queryBigrams := makeBigrams(queryTokens)
	if len(queryBigrams) > 0 {
		snippetBigrams := makeBigrams(snippetTokens)
		bigramOverlap := tokenSetOverlapCount(queryBigrams, snippetBigrams)
		if bigramOverlap > 0 {
			boost += (float64(bigramOverlap) / float64(len(queryBigrams))) * 1.5
		}
	}

	return boost
}

func normalizeSearchText(value string) string {
	tokens := tokenizeSearchText(value)
	if len(tokens) == 0 {
		return ""
	}
	return strings.Join(tokens, " ")
}

func tokenizeSearchText(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	var normalized strings.Builder
	normalized.Grow(len(value))
	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			normalized.WriteRune(r)
			continue
		}
		normalized.WriteRune(' ')
	}

	parts := strings.Fields(normalized.String())
	for i := range parts {
		parts[i] = normalizeSearchToken(parts[i])
	}
	return parts
}

func normalizeSearchToken(token string) string {
	switch token {
	case "cost", "price", "pricing", "spend", "expense", "expenses", "financial", "finance":
		return "budget"
	case "plan", "planning", "timeline", "milestone", "milestones":
		return "roadmap"
	case "step", "steps":
		return "actions"
	case "summary", "wrap", "wrapup":
		return "closing"
	case "introduction", "opening":
		return "intro"
	case "chat", "conversation", "talk":
		return "discussion"
	case "uneasy", "awkward", "hesitant", "hesitation", "uncertain", "uncertainty":
		return "tension"
	default:
		return token
	}
}

func makeTokenSet(tokens []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		set[token] = struct{}{}
	}
	return set
}

func tokenOverlapCount(expected map[string]struct{}, actual []string) int {
	if len(expected) == 0 || len(actual) == 0 {
		return 0
	}

	seen := map[string]struct{}{}
	count := 0
	for _, token := range actual {
		if _, ok := expected[token]; !ok {
			continue
		}
		if _, already := seen[token]; already {
			continue
		}
		seen[token] = struct{}{}
		count++
	}

	return count
}

func makeBigrams(tokens []string) map[string]struct{} {
	if len(tokens) < 2 {
		return nil
	}

	bigrams := make(map[string]struct{}, len(tokens)-1)
	for i := 1; i < len(tokens); i++ {
		bigrams[tokens[i-1]+"_"+tokens[i]] = struct{}{}
	}
	return bigrams
}

func tokenSetOverlapCount(left map[string]struct{}, right map[string]struct{}) int {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}

	count := 0
	for token := range right {
		if _, ok := left[token]; ok {
			count++
		}
	}
	return count
}

func weightedScore(result storage.SearchResult, ranking RankingOptions) float64 {
	base := float64(result.Score)
	if math.IsNaN(base) || math.IsInf(base, 0) {
		base = 0
	}
	if base < 0 {
		base = 0
	}
	if base > 1 {
		base = 1
	}
	if result.Segment.Type == storage.SegmentTypeVisual {
		return base * ranking.VisualWeight
	}
	return base * ranking.AudioWeight
}

func parseIndexMode(req mcp.CallToolRequest) (ingestion.IndexMode, error) {
	arguments, ok := req.Params.Arguments.(map[string]any)
	if !ok || arguments == nil {
		return ingestion.IndexModeSync, nil
	}

	rawMode, hasMode := arguments["mode"]
	if !hasMode || rawMode == nil {
		return ingestion.IndexModeSync, nil
	}

	modeValue, ok := rawMode.(string)
	if !ok {
		return "", fmt.Errorf("mode must be a string")
	}

	normalizedMode := ingestion.IndexMode(strings.ToLower(strings.TrimSpace(modeValue)))
	switch normalizedMode {
	case "", ingestion.IndexModeSync:
		return ingestion.IndexModeSync, nil
	case ingestion.IndexModeAsync:
		return ingestion.IndexModeAsync, nil
	default:
		return "", fmt.Errorf("unsupported mode: %s", modeValue)
	}
}
