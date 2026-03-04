//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestWorkerRoleWithHTTPTransportFailsFastAtStartup(t *testing.T) {
	ctx := context.Background()
	testcontainers.SkipIfProviderIsNotHealthy(t)

	env := map[string]string{
		"VIDERA_RUNTIME_MODE":        "test",
		"VIDERA_JOBQUEUE_BACKEND":    "redis",
		"VIDERA_JOBQUEUE_ROLE":       "worker",
		"VIDERA_TRANSPORT":           "http",
		"VIDERA_JOBQUEUE_REDIS_ADDR": "127.0.0.1:6379",
	}

	container, err := testcontainers.Run(ctx, "",
		testcontainers.WithDockerfile(testcontainers.FromDockerfile{
			Context:    "../..",
			Dockerfile: "Dockerfile",
			Repo:       "videra-integration",
			Tag:        "latest",
			KeepImage:  true,
		}),
		testcontainers.WithEnv(env),
		testcontainers.WithWaitStrategy(
			wait.ForLog("server failed:").
				WithStartupTimeout(20*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("expected container startup to surface fail-fast error log: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	logs := readContainerLogs(t, ctx, container)
	if !strings.Contains(logs, "VIDERA_JOBQUEUE_ROLE=worker requires VIDERA_TRANSPORT=stdio") {
		t.Fatalf("expected fail-fast guardrail message in logs, got: %s", logs)
	}
}
