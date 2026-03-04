//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const containerMCPPort = nat.Port("8080/tcp")

func startVideraContainer(t *testing.T, ctx context.Context) (testcontainers.Container, *client.Client) {
	return startVideraContainerWithEnv(t, ctx, nil)
}

func startVideraContainerWithEnv(t *testing.T, ctx context.Context, envOverrides map[string]string) (testcontainers.Container, *client.Client) {
	t.Helper()

	testcontainers.SkipIfProviderIsNotHealthy(t)

	env := map[string]string{
		"VIDERA_TRANSPORT":    "http",
		"VIDERA_HTTP_ADDR":    ":8080",
		"VIDERA_DATA_DIR":     "/data",
		"VIDERA_RUNTIME_MODE": "test",
	}
	for key, value := range envOverrides {
		env[key] = value
	}

	ctr, err := testcontainers.Run(
		ctx,
		"",
		testcontainers.WithDockerfile(testcontainers.FromDockerfile{
			Context:    "../..",
			Dockerfile: "Dockerfile",
			Repo:       "videra-integration",
			Tag:        "latest",
			KeepImage:  true,
		}),
		testcontainers.WithExposedPorts(string(containerMCPPort)),
		testcontainers.WithEnv(env),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/mcp").WithPort(containerMCPPort)),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	host, err := ctr.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := ctr.MappedPort(ctx, containerMCPPort)
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s/mcp", host, mappedPort.Port())
	cli, err := client.NewStreamableHttpClient(baseURL)
	require.NoError(t, err)

	require.NoError(t, cli.Start(ctx))
	t.Cleanup(func() {
		_ = cli.Close()
	})

	_, err = cli.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "videra-integration-tests",
				Version: "0.1.0",
			},
		},
	})
	require.NoError(t, err)

	return ctr, cli
}

func resetIndex(t *testing.T, ctx context.Context, cli *client.Client) {
	t.Helper()

	resetResult, err := cli.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "reset_index"}})
	require.NoError(t, err)
	require.False(t, resetResult.IsError)

	listResult, err := cli.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "list_videos"}})
	require.NoError(t, err)
	require.False(t, listResult.IsError)

	videos, ok := listResult.StructuredContent.([]any)
	require.True(t, ok)
	require.Len(t, videos, 0)
}
