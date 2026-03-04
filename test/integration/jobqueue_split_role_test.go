//go:build integration

package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func TestIndexVideoAsyncSplitRoleRedisLifecycle(t *testing.T) {
	ctx := context.Background()

	redisContainer, _ := startRedisContainer(t, ctx)
	t.Cleanup(func() {
		_ = redisContainer.Terminate(ctx)
	})

	redisIP, err := redisContainer.ContainerIP(ctx)
	require.NoError(t, err)
	redisAddr := fmt.Sprintf("%s:6379", redisIP)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	sharedEnv := map[string]string{
		"VIDERA_JOBQUEUE_BACKEND":            "redis",
		"VIDERA_JOBQUEUE_ROLE":               "api",
		"VIDERA_JOBQUEUE_REDIS_ADDR":         redisAddr,
		"VIDERA_JOBQUEUE_REDIS_STREAM":       "videra:index:jobs:" + suffix,
		"VIDERA_JOBQUEUE_REDIS_GROUP":        "videra-index-workers-" + suffix,
		"VIDERA_JOBQUEUE_REDIS_CONSUMER":     "videra-index-worker-" + suffix,
		"VIDERA_JOBSTATE_REDIS_PREFIX":       "videra:index:jobstatus:" + suffix + ":",
		"VIDERA_JOBQUEUE_RETRY_MAX_ATTEMPTS": "2",
		"VIDERA_JOBQUEUE_RETRY_BACKOFF_MS":   "25",
		"VIDERA_JOBQUEUE_WORKER_POLL_MS":     "25",
	}

	workerEnv := cloneEnvMap(sharedEnv)
	workerEnv["VIDERA_JOBQUEUE_ROLE"] = "worker"
	workerContainer := startVideraWorkerContainerWithEnvAndHostPorts(t, ctx, workerEnv, nil)

	_, cli := startVideraContainerWithEnvAndHostPorts(t, ctx, sharedEnv, nil)
	resetIndex(t, ctx, cli)

	initResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "https://example.com/split-role-success.mp4",
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

	failResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": "/path/does/not/exist-split-role.mp4",
				"mode": "async",
			},
		},
	})
	require.NoError(t, err)
	require.False(t, failResult.IsError)

	failPayload, ok := failResult.StructuredContent.(map[string]any)
	require.True(t, ok)
	failJobID, ok := failPayload["jobId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, failJobID)

	failedJobPayload := pollIndexJobStatus(t, ctx, cli, failJobID, "failed")
	require.Equal(t, "failed", failedJobPayload["status"])
	errorText, ok := failedJobPayload["error"].(string)
	require.True(t, ok)
	require.Contains(t, errorText, "failed after 2 attempts")

	time.Sleep(200 * time.Millisecond)

	failedStatusCheck, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_index_job",
			Arguments: map[string]any{
				"jobId": failJobID,
			},
		},
	})
	require.NoError(t, err)
	require.False(t, failedStatusCheck.IsError)
	failedStatusPayload, ok := failedStatusCheck.StructuredContent.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "failed", failedStatusPayload["status"])
	require.Equal(t, errorText, failedStatusPayload["error"])
	_, hasVideoID := failedStatusPayload["videoId"]
	require.False(t, hasVideoID)

	require.Eventually(t, func() bool {
		logs := readContainerLogs(t, ctx, workerContainer)
		return strings.Contains(logs, "queue_lifecycle event=completed job_id="+jobID) &&
			strings.Contains(logs, "queue_lifecycle event=retry_scheduled job_id="+failJobID) &&
			strings.Contains(logs, "queue_lifecycle event=retry_exhausted job_id="+failJobID)
	}, 5*time.Second, 100*time.Millisecond)
}

func cloneEnvMap(input map[string]string) map[string]string {
	cloned := make(map[string]string, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}
