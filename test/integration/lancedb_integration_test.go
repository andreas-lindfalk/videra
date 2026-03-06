//go:build integration

package integration

import (
	"context"
	"runtime"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func TestLanceDBNativeBackendIndexesAndSearches(t *testing.T) {
	if runtime.GOARCH != "amd64" {
		t.Skip("native LanceDB integration currently requires amd64 host runner because upstream Linux arm64 artifacts are not published")
	}

	ctx := context.Background()
	_, cli := startVideraContainerWithEnvHostPortsAndMountsAndDockerTarget(
		t,
		ctx,
		map[string]string{"VIDERA_STORAGE_BACKEND": "lancedb"},
		nil,
		nil,
		lanceDBNativeDockerTarget,
	)
	resetIndex(t, ctx, cli)

	indexResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/lancedb-native.mp4",
			},
		},
	})
	require.NoError(t, err)
	require.False(t, indexResult.IsError)

	listResult, err := cli.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "list_videos"}})
	require.NoError(t, err)
	require.False(t, listResult.IsError)

	videos, ok := listResult.StructuredContent.([]any)
	require.True(t, ok)
	require.Len(t, videos, 1)

	searchResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search_video",
			Arguments: map[string]any{
				"query": "budget and roadmap",
				"limit": 3,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, searchResult.IsError)

	payload, ok := searchResult.StructuredContent.(map[string]any)
	require.True(t, ok)
	results, ok := payload["results"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, results)
}

func TestLanceDBBackendOnDefaultRuntimeReturnsGuidanceError(t *testing.T) {
	ctx := context.Background()
	logs := startVideraContainerExpectStartupFailureWithEnvAndDockerTarget(
		t,
		ctx,
		map[string]string{"VIDERA_STORAGE_BACKEND": "lancedb"},
		defaultVideraDockerTarget,
	)
	require.Contains(t, logs, "lancedb native backend is disabled in this build")
}
