//go:build integration

package integration

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const containerMCPPort = nat.Port("8080/tcp")
const integrationStartupTimeout = 2 * time.Minute
const defaultVideraDockerTarget = "runtime"
const lanceDBNativeDockerTarget = "runtime-lancedb-native"

func startVideraContainer(t *testing.T, ctx context.Context) (testcontainers.Container, *client.Client) {
	return startVideraContainerWithEnv(t, ctx, nil)
}

func startVideraContainerWithEnv(t *testing.T, ctx context.Context, envOverrides map[string]string) (testcontainers.Container, *client.Client) {
	return startVideraContainerWithEnvAndHostPorts(t, ctx, envOverrides, nil)
}

func startVideraContainerWithEnvAndHostPorts(t *testing.T, ctx context.Context, envOverrides map[string]string, hostPorts []int) (testcontainers.Container, *client.Client) {
	return startVideraContainerWithEnvHostPortsAndMounts(t, ctx, envOverrides, hostPorts, nil)
}

func startVideraContainerWithEnvHostPortsAndMounts(t *testing.T, ctx context.Context, envOverrides map[string]string, hostPorts []int, mounts []testcontainers.ContainerMount) (testcontainers.Container, *client.Client) {
	return startVideraContainerWithEnvHostPortsAndMountsAndDockerTarget(t, ctx, envOverrides, hostPorts, mounts, defaultVideraDockerTarget)
}

func startVideraContainerWithEnvHostPortsAndMountsAndDockerTarget(t *testing.T, ctx context.Context, envOverrides map[string]string, hostPorts []int, mounts []testcontainers.ContainerMount, dockerTarget string) (testcontainers.Container, *client.Client) {
	t.Helper()

	testcontainers.SkipIfProviderIsNotHealthy(t)
	if strings.TrimSpace(dockerTarget) == "" {
		dockerTarget = defaultVideraDockerTarget
	}

	env := map[string]string{
		"VIDERA_TRANSPORT":    "http",
		"VIDERA_HTTP_ADDR":    ":8080",
		"VIDERA_DATA_DIR":     "/data",
		"VIDERA_RUNTIME_MODE": "test",
	}
	for key, value := range envOverrides {
		env[key] = value
	}

	buildRepo := "videra-integration"
	buildTag := "latest"
	if dockerTarget == lanceDBNativeDockerTarget {
		buildRepo = "videra-integration-lancedb-native"
		buildTag = strings.NewReplacer("/", "-", " ", "-", ":", "-", ".", "-").Replace(strings.ToLower(t.Name()))
	}

	buildConfig := testcontainers.FromDockerfile{
		Context:    "../..",
		Dockerfile: "Dockerfile",
		Repo:       buildRepo,
		Tag:        buildTag,
		KeepImage:  true,
		BuildOptionsModifier: func(options *dockertypes.ImageBuildOptions) {
			options.Target = dockerTarget
			if dockerTarget == lanceDBNativeDockerTarget {
				options.Platform = "linux/amd64"
				options.NoCache = true
			}
		},
	}

	customizers := []testcontainers.ContainerCustomizer{
		testcontainers.WithDockerfile(buildConfig),
		testcontainers.WithExposedPorts(string(containerMCPPort)),
		testcontainers.WithEnv(env),
		testcontainers.WithWaitStrategyAndDeadline(integrationStartupTimeout, wait.ForHTTP("/mcp").WithPort(containerMCPPort).WithStartupTimeout(integrationStartupTimeout)),
	}
	if len(hostPorts) > 0 {
		customizers = append(customizers, testcontainers.WithHostPortAccess(hostPorts...))
	}
	if len(mounts) > 0 {
		customizers = append(customizers, testcontainers.WithMounts(mounts...))
	}

	ctr, err := testcontainers.Run(ctx, "", customizers...)
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

func startVideraContainerExpectStartupFailureWithEnvAndDockerTarget(t *testing.T, ctx context.Context, envOverrides map[string]string, dockerTarget string) string {
	t.Helper()

	testcontainers.SkipIfProviderIsNotHealthy(t)
	if strings.TrimSpace(dockerTarget) == "" {
		dockerTarget = defaultVideraDockerTarget
	}

	env := map[string]string{
		"VIDERA_TRANSPORT":    "http",
		"VIDERA_HTTP_ADDR":    ":8080",
		"VIDERA_DATA_DIR":     "/data",
		"VIDERA_RUNTIME_MODE": "test",
	}
	for key, value := range envOverrides {
		env[key] = value
	}

	buildRepo := "videra-integration"
	buildTag := "latest"
	if dockerTarget == lanceDBNativeDockerTarget {
		buildRepo = "videra-integration-lancedb-native"
		buildTag = strings.NewReplacer("/", "-", " ", "-", ":", "-", ".", "-").Replace(strings.ToLower(t.Name()))
	}

	buildConfig := testcontainers.FromDockerfile{
		Context:    "../..",
		Dockerfile: "Dockerfile",
		Repo:       buildRepo,
		Tag:        buildTag,
		KeepImage:  true,
		BuildOptionsModifier: func(options *dockertypes.ImageBuildOptions) {
			options.Target = dockerTarget
			if dockerTarget == lanceDBNativeDockerTarget {
				options.Platform = "linux/amd64"
				options.NoCache = true
			}
		},
	}

	ctr, err := testcontainers.Run(
		ctx,
		"",
		testcontainers.WithDockerfile(buildConfig),
		testcontainers.WithExposedPorts(string(containerMCPPort)),
		testcontainers.WithEnv(env),
		testcontainers.WithWaitStrategyAndDeadline(
			integrationStartupTimeout,
			wait.ForLog("server failed").WithStartupTimeout(integrationStartupTimeout),
		),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	return readContainerLogs(t, ctx, ctr)
}

func startVideraWorkerContainerWithEnvAndHostPorts(t *testing.T, ctx context.Context, envOverrides map[string]string, hostPorts []int) testcontainers.Container {
	return startVideraWorkerContainerWithEnvHostPortsAndMounts(t, ctx, envOverrides, hostPorts, nil)
}

func startVideraWorkerContainerWithEnvHostPortsAndMounts(t *testing.T, ctx context.Context, envOverrides map[string]string, hostPorts []int, mounts []testcontainers.ContainerMount) testcontainers.Container {
	return startVideraWorkerContainerWithEnvHostPortsAndMountsAndDockerTarget(t, ctx, envOverrides, hostPorts, mounts, defaultVideraDockerTarget)
}

func startVideraWorkerContainerWithEnvHostPortsAndMountsAndDockerTarget(t *testing.T, ctx context.Context, envOverrides map[string]string, hostPorts []int, mounts []testcontainers.ContainerMount, dockerTarget string) testcontainers.Container {
	t.Helper()

	testcontainers.SkipIfProviderIsNotHealthy(t)
	if strings.TrimSpace(dockerTarget) == "" {
		dockerTarget = defaultVideraDockerTarget
	}

	env := map[string]string{
		"VIDERA_RUNTIME_MODE":  "test",
		"VIDERA_TRANSPORT":     "stdio",
		"VIDERA_JOBQUEUE_ROLE": "worker",
	}
	for key, value := range envOverrides {
		env[key] = value
	}

	buildRepo := "videra-integration"
	buildTag := "latest"
	if dockerTarget == lanceDBNativeDockerTarget {
		buildRepo = "videra-integration-lancedb-native"
		buildTag = strings.NewReplacer("/", "-", " ", "-", ":", "-", ".", "-").Replace(strings.ToLower(t.Name()))
	}

	buildConfig := testcontainers.FromDockerfile{
		Context:    "../..",
		Dockerfile: "Dockerfile",
		Repo:       buildRepo,
		Tag:        buildTag,
		KeepImage:  true,
		BuildOptionsModifier: func(options *dockertypes.ImageBuildOptions) {
			options.Target = dockerTarget
			if dockerTarget == lanceDBNativeDockerTarget {
				options.Platform = "linux/amd64"
				options.NoCache = true
			}
		},
	}

	customizers := []testcontainers.ContainerCustomizer{
		testcontainers.WithDockerfile(buildConfig),
		testcontainers.WithEnv(env),
		testcontainers.WithWaitStrategyAndDeadline(integrationStartupTimeout, wait.ForLog("queue worker started").WithStartupTimeout(integrationStartupTimeout)),
	}
	if len(hostPorts) > 0 {
		customizers = append(customizers, testcontainers.WithHostPortAccess(hostPorts...))
	}
	if len(mounts) > 0 {
		customizers = append(customizers, testcontainers.WithMounts(mounts...))
	}

	ctr, err := testcontainers.Run(ctx, "", customizers...)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	return ctr
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

func readContainerLogs(t *testing.T, ctx context.Context, ctr testcontainers.Container) string {
	t.Helper()

	logReader, err := ctr.Logs(ctx)
	require.NoError(t, err)
	defer func() {
		_ = logReader.Close()
	}()

	content, err := io.ReadAll(logReader)
	require.NoError(t, err)
	return string(content)
}
