//go:build integration

package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/andreas-lindfalk/videra/internal/proofpack"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

type defaultIntegrationSuite struct {
	suite.Suite

	ctx context.Context
	ctr testcontainers.Container
	cli *client.Client
}

func TestDefaultIntegrationSuite(t *testing.T) {
	suite.Run(t, &defaultIntegrationSuite{})
}

func (s *defaultIntegrationSuite) SetupSuite() {
	s.ctx = context.Background()
	s.ctr, s.cli = startVideraContainer(s.T(), s.ctx)
}

func (s *defaultIntegrationSuite) SetupTest() {
	resetIndex(s.T(), s.ctx, s.cli)
}

func (s *defaultIntegrationSuite) TestIndexVideoSimulatedTranscription() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	indexResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/demo-video.mp4",
			},
		},
	})
	require.NoError(t, err)
	require.False(t, indexResult.IsError)

	structured, ok := indexResult.StructuredContent.(map[string]any)
	require.True(t, ok)

	videoID, ok := structured["videoId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, videoID)
	require.Equal(t, "indexed", structured["status"])
	require.Contains(t, structured["modalities"], "visual")

	listResult, err := cli.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "list_videos"}})
	require.NoError(t, err)
	require.False(t, listResult.IsError)

	_, ok = listResult.StructuredContent.([]any)
	require.True(t, ok)

	resourceResult, err := cli.ReadResource(ctx, mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{URI: "video://" + videoID + "/transcript"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, resourceResult.Contents)

	textContent, ok := mcp.AsTextResourceContents(resourceResult.Contents[0])
	require.True(t, ok)
	require.Contains(t, textContent.Text, "simulated")

	searchResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search_video",
			Arguments: map[string]any{
				"query": "budget and roadmap",
				"limit": 5,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, searchResult.IsError)

	searchPayload, ok := searchResult.StructuredContent.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "budget and roadmap", searchPayload["query"])

	rawResults, ok := searchPayload["results"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, rawResults)

	hasAudio := false
	hasVisual := false
	for _, raw := range rawResults {
		entry, ok := raw.(map[string]any)
		require.True(t, ok)
		segmentType, _ := entry["type"].(string)
		if segmentType == "audio" {
			hasAudio = true
		}
		if segmentType == "visual" {
			hasVisual = true
		}
	}
	require.True(t, hasAudio)
	require.True(t, hasVisual)
}

func (s *defaultIntegrationSuite) TestIndexVideoInvalidPathReturnsToolError() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	result, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "/path/does/not/exist.mp4",
			},
		},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)
	require.NotEmpty(t, result.Content)

	text, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	require.Contains(t, text.Text, "not found")
}

func (s *defaultIntegrationSuite) TestIndexVideoIdempotentForSameSource() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	first, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/idempotent-demo.mp4",
			},
		},
	})
	require.NoError(t, err)
	require.False(t, first.IsError)
	firstPayload, ok := first.StructuredContent.(map[string]any)
	require.True(t, ok)
	firstID, ok := firstPayload["videoId"].(string)
	require.True(t, ok)

	second, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/idempotent-demo.mp4",
			},
		},
	})
	require.NoError(t, err)
	require.False(t, second.IsError)
	secondPayload, ok := second.StructuredContent.(map[string]any)
	require.True(t, ok)
	secondID, ok := secondPayload["videoId"].(string)
	require.True(t, ok)

	require.Equal(t, firstID, secondID)
}

func (s *defaultIntegrationSuite) TestIndexVideoMalformedInputReturnsToolError() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	result, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "",
			},
		},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)
	require.NotEmpty(t, result.Content)

	text, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	require.Contains(t, text.Text, "path is required")
}

func (s *defaultIntegrationSuite) TestIndexVideoAsyncLifecycleSuccess() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	initResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/async-success.mp4",
				"mode": "async",
			},
		},
	})
	require.NoError(t, err)
	require.False(t, initResult.IsError)

	initPayload, ok := initResult.StructuredContent.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "pending", initPayload["status"])
	jobID, ok := initPayload["jobId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, jobID)

	jobPayload := pollIndexJobStatus(t, ctx, cli, jobID, "completed")
	require.Equal(t, "completed", jobPayload["status"])
	videoID, ok := jobPayload["videoId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, videoID)

	listResult, err := cli.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "list_videos"}})
	require.NoError(t, err)
	require.False(t, listResult.IsError)
	videos, ok := listResult.StructuredContent.([]any)
	require.True(t, ok)
	require.Len(t, videos, 1)
}

func (s *defaultIntegrationSuite) TestIndexVideoAsyncLifecycleFailure() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	initResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "/path/does/not/exist-async.mp4",
				"mode": "async",
			},
		},
	})
	require.NoError(t, err)
	require.False(t, initResult.IsError)

	initPayload, ok := initResult.StructuredContent.(map[string]any)
	require.True(t, ok)
	jobID, ok := initPayload["jobId"].(string)
	require.True(t, ok)

	jobPayload := pollIndexJobStatus(t, ctx, cli, jobID, "failed")
	require.Equal(t, "failed", jobPayload["status"])
	errorMessage, ok := jobPayload["error"].(string)
	require.True(t, ok)
	require.Contains(t, errorMessage, "not found")
}

func (s *defaultIntegrationSuite) TestIndexVideoInvalidModeReturnsToolError() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	result, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/invalid-mode.mp4",
				"mode": "queue",
			},
		},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)

	text, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	require.Contains(t, text.Text, "unsupported mode")
}

func (s *defaultIntegrationSuite) TestGetIndexJobUnknownReturnsToolError() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	result, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_index_job",
			Arguments: map[string]any{
				"jobId": "missing-job-id",
			},
		},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)

	text, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	require.Contains(t, text.Text, "index job not found")
}

func (s *defaultIntegrationSuite) TestIndexVideoLocalFileLikePathFlow() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	indexResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "/etc/hosts",
			},
		},
	})
	require.NoError(t, err)
	require.False(t, indexResult.IsError)

	indexPayload, ok := indexResult.StructuredContent.(map[string]any)
	require.True(t, ok)
	videoID, ok := indexPayload["videoId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, videoID)

	resourceResult, err := cli.ReadResource(ctx, mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{URI: "video://" + videoID + "/transcript"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, resourceResult.Contents)

	searchResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search_video",
			Arguments: map[string]any{
				"query": "roadmap",
				"limit": 3,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, searchResult.IsError)

	searchPayload, ok := searchResult.StructuredContent.(map[string]any)
	require.True(t, ok)
	rawResults, ok := searchPayload["results"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, rawResults)
}

func (s *defaultIntegrationSuite) TestSearchVideoDeterministicOrdering() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	_, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/deterministic-order.mp4",
			},
		},
	})
	require.NoError(t, err)

	first := searchAndExtractResults(t, ctx, cli, "budget and roadmap", 5)
	second := searchAndExtractResults(t, ctx, cli, "budget and roadmap", 5)

	require.Equal(t, first, second)
	require.NotEmpty(t, first)
}

func (s *defaultIntegrationSuite) TestSearchVideoMalformedPayloadReturnsToolError() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	result, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search_video",
			Arguments: map[string]any{
				"query": 123,
			},
		},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)
	require.NotEmpty(t, result.Content)

	text, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	require.Contains(t, text.Text, "query")
}

func (s *defaultIntegrationSuite) TestToolResponseBackwardCompatFields() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	indexResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/backward-compat.mp4",
			},
		},
	})
	require.NoError(t, err)
	require.False(t, indexResult.IsError)

	indexPayload, ok := indexResult.StructuredContent.(map[string]any)
	require.True(t, ok)
	require.Contains(t, indexPayload, "videoId")
	require.Contains(t, indexPayload, "status")
	require.Contains(t, indexPayload, "filePath")
	require.Contains(t, indexPayload, "audioSegments")
	require.Contains(t, indexPayload, "visualSegments")
	require.Contains(t, indexPayload, "modalities")

	searchResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search_video",
			Arguments: map[string]any{
				"query": "budget roadmap",
				"limit": 3,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, searchResult.IsError)

	searchPayload, ok := searchResult.StructuredContent.(map[string]any)
	require.True(t, ok)
	require.Contains(t, searchPayload, "query")
	require.Contains(t, searchPayload, "count")
	require.Contains(t, searchPayload, "results")

	rawResults, ok := searchPayload["results"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, rawResults)

	firstResult, ok := rawResults[0].(map[string]any)
	require.True(t, ok)
	require.Contains(t, firstResult, "videoId")
	require.Contains(t, firstResult, "startMs")
	require.Contains(t, firstResult, "endMs")
	require.Contains(t, firstResult, "type")
	require.Contains(t, firstResult, "snippet")
	require.Contains(t, firstResult, "similarity")

	listResult, err := cli.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "list_videos"}})
	require.NoError(t, err)
	require.False(t, listResult.IsError)

	videos, ok := listResult.StructuredContent.([]any)
	require.True(t, ok)
	require.NotEmpty(t, videos)

	firstVideo, ok := videos[0].(map[string]any)
	require.True(t, ok)
	require.Contains(t, firstVideo, "id")
	require.Contains(t, firstVideo, "filePath")
	require.Contains(t, firstVideo, "status")
	require.Contains(t, firstVideo, "indexedAt")
	require.Contains(t, firstVideo, "durationMs")
	require.Contains(t, firstVideo, "audioSegments")
	require.Contains(t, firstVideo, "visualSegments")
	require.Contains(t, firstVideo, "modalities")
}

func (s *defaultIntegrationSuite) TestIndexVideoRetrySafeAfterPartialFailure() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	path := "https://example.com/retry-safe.mp4?videra_fail_after_persist_once=1"

	first, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": path,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, first.IsError)

	firstPayload, ok := first.StructuredContent.(map[string]any)
	require.True(t, ok)
	firstID, ok := firstPayload["videoId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, firstID)

	second, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": path,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, second.IsError)

	secondPayload, ok := second.StructuredContent.(map[string]any)
	require.True(t, ok)
	secondID, ok := secondPayload["videoId"].(string)
	require.True(t, ok)
	require.Equal(t, firstID, secondID)

	listResult, err := cli.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "list_videos"}})
	require.NoError(t, err)
	require.False(t, listResult.IsError)

	videos, ok := listResult.StructuredContent.([]any)
	require.True(t, ok)
	require.Len(t, videos, 1)
}

func (s *defaultIntegrationSuite) TestProofPackScenariosEvidenceAndDeterminism() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	scenarios, err := proofpack.LoadScenarios()
	require.NoError(t, err)
	require.NotEmpty(t, scenarios)

	for _, scenario := range scenarios {
		scenario := scenario
		s.Run(scenario.Name, func() {
			resetIndex(t, ctx, cli)

			indexResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "index_video",
					Arguments: map[string]any{
						"path": scenario.VideoPath,
					},
				},
			})
			require.NoError(t, err)
			require.False(t, indexResult.IsError)

			firstResults := searchAndExtractResults(t, ctx, cli, scenario.Query, 5)
			secondResults := searchAndExtractResults(t, ctx, cli, scenario.Query, 5)

			require.Equal(t, firstResults, secondResults)
			require.GreaterOrEqual(t, len(firstResults), scenario.MinResults)

			first := firstResults[0]
			require.Contains(t, first, "videoId")
			require.Contains(t, first, "startMs")
			require.Contains(t, first, "endMs")
			require.Contains(t, first, "type")
			require.Contains(t, first, "snippet")
			require.Contains(t, first, "similarity")

			matched := evidenceMatches(firstResults, scenario.ExpectedEvidence)
			require.GreaterOrEqual(t, matched, len(scenario.ExpectedEvidence))
		})
	}
}

func (s *defaultIntegrationSuite) TestProofPackProductRecallPrioritizesTop2Evidence() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	resetIndex(t, ctx, cli)

	indexResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/proofpack-product.mp4",
			},
		},
	})
	require.NoError(t, err)
	require.False(t, indexResult.IsError)

	results := searchAndExtractResults(t, ctx, cli, "intro segment discussion", 5)
	require.GreaterOrEqual(t, len(results), 2)

	firstSnippet := strings.ToLower(fmt.Sprintf("%v", results[0]["snippet"]))
	secondSnippet := strings.ToLower(fmt.Sprintf("%v", results[1]["snippet"]))
	topTwo := firstSnippet + "\n" + secondSnippet

	require.Contains(t, topTwo, "intro segment")
	require.Contains(t, topTwo, "discussion")
}

func (s *defaultIntegrationSuite) TestPilotBenchmarkScorecard() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	scenarios, err := proofpack.LoadPilotBenchmarkScenarios()
	require.NoError(t, err)
	require.NotEmpty(t, scenarios)

	evidenceExpectedTotal := 0
	evidenceMatchedTotal := 0
	deterministicCount := 0
	topTwoQualityCount := 0

	for _, scenario := range scenarios {
		scenario := scenario
		s.Run(scenario.Name, func() {
			resetIndex(t, ctx, cli)

			indexResult, indexErr := cli.CallTool(ctx, mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "index_video",
					Arguments: map[string]any{
						"path": scenario.VideoPath,
					},
				},
			})
			require.NoError(t, indexErr)
			require.False(t, indexResult.IsError)

			firstResults := searchAndExtractResults(t, ctx, cli, scenario.Query, 5)
			secondResults := searchAndExtractResults(t, ctx, cli, scenario.Query, 5)

			require.Equal(t, firstResults, secondResults)
			require.GreaterOrEqual(t, len(firstResults), scenario.MinResults)

			matched := evidenceMatches(firstResults, scenario.ExpectedEvidence)
			require.GreaterOrEqual(t, matched, len(scenario.ExpectedEvidence))

			topLimit := minInt(2, len(firstResults))
			topTwoMatched := evidenceMatches(firstResults[:topLimit], scenario.ExpectedEvidence)
			require.Greater(t, topTwoMatched, 0)

			evidenceExpectedTotal += len(scenario.ExpectedEvidence)
			evidenceMatchedTotal += matched
			deterministicCount++
			topTwoQualityCount++
		})
	}

	require.Greater(t, evidenceExpectedTotal, 0)
	evidenceMatchRate := float64(evidenceMatchedTotal) / float64(evidenceExpectedTotal)
	deterministicRate := float64(deterministicCount) / float64(len(scenarios))
	topTwoQualityRate := float64(topTwoQualityCount) / float64(len(scenarios))

	require.GreaterOrEqual(t, evidenceMatchRate, 1.0)
	require.GreaterOrEqual(t, deterministicRate, 1.0)
	require.GreaterOrEqual(t, topTwoQualityRate, 1.0)

	t.Logf("pilot benchmark scorecard: scenarios=%d evidenceMatchRate=%.2f deterministicRate=%.2f topTwoQualityRate=%.2f", len(scenarios), evidenceMatchRate, deterministicRate, topTwoQualityRate)
}

func (s *defaultIntegrationSuite) TestSearchVideoIncludeDebugMetadata() {
	t := s.T()
	ctx := s.ctx
	cli := s.cli

	_, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/debug-metadata.mp4",
			},
		},
	})
	require.NoError(t, err)

	searchResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search_video",
			Arguments: map[string]any{
				"query":        "budget roadmap",
				"limit":        3,
				"includeDebug": true,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, searchResult.IsError)

	payload, ok := searchResult.StructuredContent.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "weightedSimilarity", payload["scoreMode"])

	debugPayload, ok := payload["debug"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, debugPayload, "audioWeight")
	require.Contains(t, debugPayload, "visualWeight")
	require.Contains(t, debugPayload, "candidateCount")

	rawResults, ok := payload["results"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, rawResults)

	first, ok := rawResults[0].(map[string]any)
	require.True(t, ok)
	require.Contains(t, first, "rawSimilarity")
}

type weightingIntegrationSuite struct {
	suite.Suite

	ctx          context.Context
	audioFavCtr  testcontainers.Container
	audioFavCli  *client.Client
	visualFavCtr testcontainers.Container
	visualFavCli *client.Client
}

type canonicalMappingIntegrationSuite struct {
	suite.Suite

	ctx        context.Context
	neutralCtr testcontainers.Container
	neutralCli *client.Client
	mappedCtr  testcontainers.Container
	mappedCli  *client.Client
}

func TestWeightingIntegrationSuite(t *testing.T) {
	suite.Run(t, &weightingIntegrationSuite{})
}

func TestCanonicalMappingIntegrationSuite(t *testing.T) {
	suite.Run(t, &canonicalMappingIntegrationSuite{})
}

func (s *weightingIntegrationSuite) SetupSuite() {
	s.ctx = context.Background()

	s.audioFavCtr, s.audioFavCli = startVideraContainerWithEnv(s.T(), s.ctx, map[string]string{
		"VIDERA_SEARCH_AUDIO_WEIGHT":  "50.0",
		"VIDERA_SEARCH_VISUAL_WEIGHT": "0.1",
	})

	s.visualFavCtr, s.visualFavCli = startVideraContainerWithEnv(s.T(), s.ctx, map[string]string{
		"VIDERA_SEARCH_AUDIO_WEIGHT":  "0.1",
		"VIDERA_SEARCH_VISUAL_WEIGHT": "50.0",
	})
}

func (s *canonicalMappingIntegrationSuite) SetupSuite() {
	s.ctx = context.Background()

	s.neutralCtr, s.neutralCli = startVideraContainerWithEnv(s.T(), s.ctx, nil)
	s.mappedCtr, s.mappedCli = startVideraContainerWithEnv(s.T(), s.ctx, map[string]string{
		"VIDERA_SEMANTIC_CANONICAL_MAP": "kitty=cat,feline=cat,cats=cat",
	})
}

func (s *weightingIntegrationSuite) SetupTest() {
	resetIndex(s.T(), s.ctx, s.audioFavCli)
	resetIndex(s.T(), s.ctx, s.visualFavCli)
}

func (s *canonicalMappingIntegrationSuite) SetupTest() {
	resetIndex(s.T(), s.ctx, s.neutralCli)
	resetIndex(s.T(), s.ctx, s.mappedCli)
}

func (s *weightingIntegrationSuite) TestSearchVideoModalityWeightingBehavior() {
	t := s.T()
	ctx := s.ctx
	audioFavCli := s.audioFavCli
	visualFavCli := s.visualFavCli

	_, err := audioFavCli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/modality-weighting.mp4",
			},
		},
	})
	require.NoError(t, err)

	audioFavResults := searchAndExtractResults(t, ctx, audioFavCli, "roadmap budget keyframe", 5)
	require.NotEmpty(t, audioFavResults)
	audioFavVisualScore := topScoreForType(audioFavResults, "visual")
	audioFavAudioScore := topScoreForType(audioFavResults, "audio")

	_, err = visualFavCli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/modality-weighting.mp4",
			},
		},
	})
	require.NoError(t, err)

	visualFavResults := searchAndExtractResults(t, ctx, visualFavCli, "roadmap budget keyframe", 5)
	require.NotEmpty(t, visualFavResults)
	visualFavVisualScore := topScoreForType(visualFavResults, "visual")
	visualFavAudioScore := topScoreForType(visualFavResults, "audio")

	require.Greater(t, visualFavVisualScore, audioFavVisualScore)
	require.Greater(t, audioFavAudioScore, visualFavAudioScore)
}

func (s *canonicalMappingIntegrationSuite) TestDomainProfileNeutralLiteralRecall() {
	t := s.T()
	ctx := s.ctx

	scenarios, err := proofpack.LoadDomainProfileScenarios()
	require.NoError(t, err)

	literalScenario, found := findScenarioByName(scenarios, "animals_literal_recall")
	require.True(t, found)

	_, err = s.neutralCli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": literalScenario.VideoPath,
			},
		},
	})
	require.NoError(t, err)

	first := searchAndExtractResults(t, ctx, s.neutralCli, literalScenario.Query, 5)
	second := searchAndExtractResults(t, ctx, s.neutralCli, literalScenario.Query, 5)

	require.Equal(t, first, second)
	require.GreaterOrEqual(t, len(first), literalScenario.MinResults)
	require.GreaterOrEqual(t, evidenceMatches(first, literalScenario.ExpectedEvidence), len(literalScenario.ExpectedEvidence))
}

func (s *canonicalMappingIntegrationSuite) TestDomainProfileMappingImprovesSynonymRecall() {
	t := s.T()
	ctx := s.ctx

	scenarios, err := proofpack.LoadDomainProfileScenarios()
	require.NoError(t, err)

	synonymScenario, found := findScenarioByName(scenarios, "animals_synonym_profile")
	require.True(t, found)

	for _, cli := range []*client.Client{s.neutralCli, s.mappedCli} {
		indexResult, indexErr := cli.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "index_video",
				Arguments: map[string]any{
					"path": synonymScenario.VideoPath,
				},
			},
		})
		require.NoError(t, indexErr)
		require.False(t, indexResult.IsError)
	}

	neutralResults := searchAndExtractResults(t, ctx, s.neutralCli, synonymScenario.Query, 10)
	mappedResults := searchAndExtractResults(t, ctx, s.mappedCli, synonymScenario.Query, 10)

	require.GreaterOrEqual(t, evidenceMatches(mappedResults, synonymScenario.ExpectedEvidence), len(synonymScenario.ExpectedEvidence))

	needle := "proofpack-cats-cat"
	neutralRank, neutralScore := findSnippetRankAndSimilarity(neutralResults, needle)
	mappedRank, mappedScore := findSnippetRankAndSimilarity(mappedResults, needle)

	require.NotEqual(t, -1, neutralRank)
	require.NotEqual(t, -1, mappedRank)
	require.LessOrEqual(t, mappedRank, neutralRank)
	require.Greater(t, mappedScore, neutralScore)
}

func findScenarioByName(scenarios []proofpack.Scenario, name string) (proofpack.Scenario, bool) {
	for _, scenario := range scenarios {
		if scenario.Name == name {
			return scenario, true
		}
	}
	return proofpack.Scenario{}, false
}

func findSnippetRankAndSimilarity(results []map[string]any, needle string) (int, float64) {
	normalizedNeedle := strings.ToLower(strings.TrimSpace(needle))
	for idx, result := range results {
		snippet := strings.ToLower(fmt.Sprintf("%v", result["snippet"]))
		if !strings.Contains(snippet, normalizedNeedle) {
			continue
		}

		similarity, _ := result["similarity"].(float64)
		return idx, similarity
	}

	return -1, 0
}

func searchAndExtractResults(t *testing.T, ctx context.Context, cli *client.Client, query string, limit int) []map[string]any {
	t.Helper()

	searchResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search_video",
			Arguments: map[string]any{
				"query": query,
				"limit": limit,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, searchResult.IsError)

	searchPayload, ok := searchResult.StructuredContent.(map[string]any)
	require.True(t, ok)

	rawResults, ok := searchPayload["results"].([]any)
	require.True(t, ok)

	out := make([]map[string]any, 0, len(rawResults))
	for _, raw := range rawResults {
		entry, ok := raw.(map[string]any)
		require.True(t, ok)
		out = append(out, entry)
	}

	return out
}

func topScoreForType(results []map[string]any, segmentType string) float64 {
	max := -1.0
	for _, result := range results {
		typeValue := fmt.Sprintf("%v", result["type"])
		if typeValue != segmentType {
			continue
		}

		score, ok := result["similarity"].(float64)
		if !ok {
			continue
		}
		if score > max {
			max = score
		}
	}
	return max
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func evidenceMatches(results []map[string]any, expected []string) int {
	joined := make([]string, 0, len(results))
	for _, result := range results {
		snippet := strings.ToLower(fmt.Sprintf("%v", result["snippet"]))
		joined = append(joined, snippet)
	}

	matches := 0
	for _, needle := range expected {
		needle = strings.ToLower(strings.TrimSpace(needle))
		if needle == "" {
			continue
		}
		for _, snippet := range joined {
			if strings.Contains(snippet, needle) {
				matches++
				break
			}
		}
	}
	return matches
}

func pollIndexJobStatus(t *testing.T, ctx context.Context, cli *client.Client, jobID string, expectedStatus string) map[string]any {
	t.Helper()

	var latestPayload map[string]any
	require.Eventually(t, func() bool {
		result, err := cli.CallTool(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "get_index_job",
				Arguments: map[string]any{
					"jobId": jobID,
				},
			},
		})
		if err != nil || result.IsError {
			return false
		}

		payload, ok := result.StructuredContent.(map[string]any)
		if !ok {
			return false
		}
		latestPayload = payload

		status, _ := payload["status"].(string)
		return status == expectedStatus
	}, 3*time.Second, 25*time.Millisecond)

	require.NotNil(t, latestPayload)
	return latestPayload
}
