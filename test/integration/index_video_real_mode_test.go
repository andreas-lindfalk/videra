//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func TestIndexVideoRealModeRejectsRemotePath(t *testing.T) {
	ctx := context.Background()
	_, cli := startVideraContainerWithEnv(t, ctx, map[string]string{"VIDERA_INGESTION_MODE": "real"})
	resetIndex(t, ctx, cli)

	result, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/real-mode.mp4",
			},
		},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)

	text, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	require.Contains(t, text.Text, "supports only local file paths")
}

func TestIndexVideoRealModeRequiresSidecarForLocalPath(t *testing.T) {
	ctx := context.Background()
	_, cli := startVideraContainerWithEnv(t, ctx, map[string]string{"VIDERA_INGESTION_MODE": "real"})
	resetIndex(t, ctx, cli)

	result, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "/etc/hosts",
			},
		},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)

	text, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	require.Contains(t, text.Text, "sidecar transcript")
}
