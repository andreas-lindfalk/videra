//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

func TestIndexVideoAsyncSplitRoleRedisSharedStorageVisibility(t *testing.T) {
	ctx := context.Background()

	redisContainer, _ := startRedisContainer(t, ctx)
	t.Cleanup(func() {
		_ = redisContainer.Terminate(ctx)
	})

	redisIP, err := redisContainer.ContainerIP(ctx)
	require.NoError(t, err)
	redisAddr := fmt.Sprintf("%s:6379", redisIP)

	workingDir, err := os.Getwd()
	require.NoError(t, err)
	repoRoot := filepath.Clean(filepath.Join(workingDir, "../.."))
	sharedBaseDir := filepath.Join(filepath.Dir(repoRoot), "videra-integration-tmp")
	require.NoError(t, os.MkdirAll(sharedBaseDir, 0o755))
	sharedDataDir, err := os.MkdirTemp(sharedBaseDir, "split-shared-data-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(sharedDataDir)
	})
	sharedMount := testcontainers.BindMount(sharedDataDir, testcontainers.ContainerMountTarget("/shared-data"))

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	sharedEnv := map[string]string{
		"VIDERA_DATA_DIR":                  "/shared-data",
		"VIDERA_SPLIT_SHARED_STORAGE":      "true",
		"VIDERA_JOBQUEUE_BACKEND":          "redis",
		"VIDERA_JOBQUEUE_ROLE":             "api",
		"VIDERA_JOBQUEUE_REDIS_ADDR":       redisAddr,
		"VIDERA_JOBQUEUE_REDIS_STREAM":     "videra:index:jobs:" + suffix,
		"VIDERA_JOBQUEUE_REDIS_GROUP":      "videra-index-workers-" + suffix,
		"VIDERA_JOBQUEUE_REDIS_CONSUMER":   "videra-index-worker-" + suffix,
		"VIDERA_JOBSTATE_REDIS_PREFIX":     "videra:index:jobstatus:" + suffix + ":",
		"VIDERA_JOBQUEUE_WORKER_POLL_MS":   "25",
		"VIDERA_JOBQUEUE_RETRY_BACKOFF_MS": "25",
	}

	workerEnv := cloneEnvMap(sharedEnv)
	workerEnv["VIDERA_JOBQUEUE_ROLE"] = "worker"
	_ = startVideraWorkerContainerWithEnvHostPortsAndMounts(t, ctx, workerEnv, nil, []testcontainers.ContainerMount{sharedMount})

	_, cli := startVideraContainerWithEnvHostPortsAndMounts(t, ctx, sharedEnv, nil, []testcontainers.ContainerMount{sharedMount})
	resetIndex(t, ctx, cli)

	initResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/split-role-shared.mp4",
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
	require.NotEmpty(t, jobID)

	completedJob := pollIndexJobStatus(t, ctx, cli, jobID, "completed")
	videoID, ok := completedJob["videoId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, videoID)

	require.Eventually(t, func() bool {
		listResult, listErr := cli.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "list_videos"}})
		if listErr != nil || listResult.IsError {
			return false
		}
		videos, castOK := listResult.StructuredContent.([]any)
		if !castOK || len(videos) == 0 {
			return false
		}

		for _, raw := range videos {
			entry, entryOK := raw.(map[string]any)
			if !entryOK {
				continue
			}
			candidateID, _ := entry["id"].(string)
			if candidateID == videoID {
				return true
			}
		}
		return false
	}, 5*time.Second, 150*time.Millisecond)

	searchResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search_video",
			Arguments: map[string]any{
				"query": "roadmap and budget",
				"limit": 5,
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

	hasIndexedVideo := false
	for _, raw := range rawResults {
		entry, entryOK := raw.(map[string]any)
		if !entryOK {
			continue
		}
		candidateVideoID, _ := entry["videoId"].(string)
		if candidateVideoID == videoID {
			hasIndexedVideo = true
			break
		}
	}
	require.True(t, hasIndexedVideo)
}
