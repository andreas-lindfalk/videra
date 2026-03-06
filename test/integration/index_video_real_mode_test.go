//go:build integration

package integration

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func TestIndexVideoSimulatedModeAcceptsRemotePath(t *testing.T) {
	ctx := context.Background()
	_, cli := startVideraContainer(t, ctx)
	resetIndex(t, ctx, cli)

	result, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/remote-source.mp4",
			},
		},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)
	structured, ok := result.StructuredContent.(map[string]any)
	require.True(t, ok)
	require.Contains(t, structured, "videoId")
}

func TestIndexVideoRealModeRemotePathRespectsMaxSizeBound(t *testing.T) {
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bytes.Repeat([]byte("v"), 2*1024*1024))
	}))
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	hostPort, err := strconv.Atoi(parsedURL.Port())
	require.NoError(t, err)

	containerURL := "http://host.testcontainers.internal:" + parsedURL.Port() + "/clip.mp4"
	_, cli := startVideraContainerWithEnvAndHostPorts(t, ctx, map[string]string{
		"VIDERA_INGESTION_MODE":           "real",
		"VIDERA_REMOTE_FETCH_ENABLED":     "true",
		"VIDERA_REMOTE_FETCH_TIMEOUT_SEC": "10",
		"VIDERA_REMOTE_FETCH_MAX_MB":      "1",
	}, []int{hostPort})
	resetIndex(t, ctx, cli)

	result, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": containerURL,
			},
		},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)

	text, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	require.Contains(t, text.Text, "remote media exceeds maximum size")
}

func TestIndexVideoRealModeRemotePathHonorsDisabledFetch(t *testing.T) {
	ctx := context.Background()
	_, cli := startVideraContainerWithEnv(t, ctx, map[string]string{
		"VIDERA_INGESTION_MODE":       "real",
		"VIDERA_REMOTE_FETCH_ENABLED": "false",
	})
	resetIndex(t, ctx, cli)

	result, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/disabled-fetch.mp4",
			},
		},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)

	text, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	require.Contains(t, text.Text, "remote ingestion is disabled by configuration")
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

func TestRealModeCLIPBackendMissingModelPathFailsStartup(t *testing.T) {
	ctx := context.Background()

	logs := startVideraContainerExpectStartupFailureWithEnvAndDockerTarget(t, ctx, map[string]string{
		"VIDERA_INGESTION_MODE": "real",
		"VIDERA_VISUAL_BACKEND": "clip",
	}, defaultVideraDockerTarget)

	require.Contains(t, logs, "VIDERA_CLIP_MODEL_PATH cannot be empty when VIDERA_VISUAL_BACKEND=clip")
}

func TestRealModeCLIPBackendUnavailableFallsBackToOCRAtStartup(t *testing.T) {
	ctx := context.Background()
	ctr, cli := startVideraContainerWithEnv(t, ctx, map[string]string{
		"VIDERA_INGESTION_MODE":  "real",
		"VIDERA_VISUAL_BACKEND":  "clip",
		"VIDERA_CLIP_MODEL_PATH": "/tmp/clip-image-model.onnx",
		"VIDERA_STORAGE_BACKEND": "chromem",
	})

	result, err := cli.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "list_videos"}})
	require.NoError(t, err)
	require.False(t, result.IsError)

	logs := readContainerLogs(t, ctx, ctr)
	require.True(t,
		strings.Contains(logs, "clip visual backend unavailable") && strings.Contains(logs, "falling back to ocr"),
		"expected startup fallback warning in logs, got: %s",
		logs,
	)
}
