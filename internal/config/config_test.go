package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadWorksWithoutAuthCoupling(t *testing.T) {
	t.Setenv("VIDERA_TRANSPORT", "stdio")
	t.Setenv("VIDERA_HTTP_ADDR", ":8080")
	t.Setenv("VIDERA_DATA_DIR", "./data")
	t.Setenv("VIDERA_LOG_LEVEL", "info")
	t.Setenv("VIDERA_RUNTIME_MODE", "local")
	t.Setenv("VIDERA_INGESTION_MODE", "simulated")
	t.Setenv("VIDERA_REMOTE_FETCH_ENABLED", "true")
	t.Setenv("VIDERA_REMOTE_FETCH_TIMEOUT_SEC", "60")
	t.Setenv("VIDERA_REMOTE_FETCH_MAX_MB", "200")
	t.Setenv("VIDERA_FRAME_INTERVAL_SEC", "5")
	t.Setenv("VIDERA_DEFAULT_SEARCH_LIMIT", "5")
	t.Setenv("VIDERA_INDEX_CONCURRENCY", "4")
	t.Setenv("VIDERA_SEARCH_AUDIO_WEIGHT", "1.0")
	t.Setenv("VIDERA_SEARCH_VISUAL_WEIGHT", "1.0")
	t.Setenv("VIDERA_AUTH_TOKEN", "ignored")
	t.Setenv("VIDERA_AUTH_MODE", "ignored")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, TransportStdio, cfg.Transport)
	require.Equal(t, ":8080", cfg.HTTPAddr)
	require.Equal(t, "./data", cfg.DataDir)
	require.Equal(t, "local", cfg.RuntimeMode)
	require.Equal(t, IngestionModeSimulated, cfg.IngestionMode)
	require.True(t, cfg.RemoteFetchEnabled)
	require.Equal(t, 60, cfg.RemoteFetchTimeout)
	require.Equal(t, 200, cfg.RemoteFetchMaxMB)
	require.Equal(t, JobQueueBackendInProcess, cfg.JobQueueBackend)
	require.Equal(t, JobQueueRoleAll, cfg.JobQueueRole)
	require.Equal(t, 3, cfg.JobQueueRetryMax)
	require.Equal(t, 250, cfg.JobQueueRetryBackoff)
	require.Equal(t, 250, cfg.JobQueueWorkerPollMS)
	require.False(t, cfg.SplitSharedStorage)
}

func TestLoadInvalidTransportStillFails(t *testing.T) {
	t.Setenv("VIDERA_TRANSPORT", "invalid")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported VIDERA_TRANSPORT")
}

func TestLoadInvalidIngestionModeFails(t *testing.T) {
	t.Setenv("VIDERA_INGESTION_MODE", "nonsense")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported VIDERA_INGESTION_MODE")
}

func TestLoadInvalidRemoteFetchEnabledFails(t *testing.T) {
	t.Setenv("VIDERA_REMOTE_FETCH_ENABLED", "maybe")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid VIDERA_REMOTE_FETCH_ENABLED")
}

func TestLoadInvalidRemoteFetchMaxMBFails(t *testing.T) {
	t.Setenv("VIDERA_REMOTE_FETCH_MAX_MB", "0")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "VIDERA_REMOTE_FETCH_MAX_MB must be > 0")
}

func TestLoadInvalidJobQueueBackendFails(t *testing.T) {
	t.Setenv("VIDERA_JOBQUEUE_BACKEND", "kafka")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported VIDERA_JOBQUEUE_BACKEND")
}

func TestLoadInvalidJobQueueRoleFails(t *testing.T) {
	t.Setenv("VIDERA_JOBQUEUE_ROLE", "scheduler")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported VIDERA_JOBQUEUE_ROLE")
}

func TestLoadSplitRoleWithInProcessBackendFails(t *testing.T) {
	t.Setenv("VIDERA_JOBQUEUE_BACKEND", "inprocess")
	t.Setenv("VIDERA_JOBQUEUE_ROLE", "api")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires an external queue backend")
}

func TestLoadWorkerRoleWithHTTPTransportFails(t *testing.T) {
	t.Setenv("VIDERA_JOBQUEUE_BACKEND", "redis")
	t.Setenv("VIDERA_JOBQUEUE_ROLE", "worker")
	t.Setenv("VIDERA_TRANSPORT", "http")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "VIDERA_JOBQUEUE_ROLE=worker requires VIDERA_TRANSPORT=stdio")
}

func TestLoadParsesQueueRoleAndRetrySettings(t *testing.T) {
	t.Setenv("VIDERA_JOBQUEUE_BACKEND", "redis")
	t.Setenv("VIDERA_JOBQUEUE_ROLE", "worker")
	t.Setenv("VIDERA_JOBQUEUE_RETRY_MAX_ATTEMPTS", "5")
	t.Setenv("VIDERA_JOBQUEUE_RETRY_BACKOFF_MS", "125")
	t.Setenv("VIDERA_JOBQUEUE_WORKER_POLL_MS", "50")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, JobQueueRoleWorker, cfg.JobQueueRole)
	require.Equal(t, 5, cfg.JobQueueRetryMax)
	require.Equal(t, 125, cfg.JobQueueRetryBackoff)
	require.Equal(t, 50, cfg.JobQueueWorkerPollMS)
}

func TestLoadParsesRedisJobQueueDB(t *testing.T) {
	t.Setenv("VIDERA_JOBQUEUE_BACKEND", "redis")
	t.Setenv("VIDERA_JOBQUEUE_REDIS_DB", "3")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, JobQueueBackendRedis, cfg.JobQueueBackend)
	require.Equal(t, 3, cfg.JobQueueRedisDB)
}

func TestLoadParsesSplitSharedStorageFlag(t *testing.T) {
	t.Setenv("VIDERA_SPLIT_SHARED_STORAGE", "true")

	cfg, err := Load()
	require.NoError(t, err)
	require.True(t, cfg.SplitSharedStorage)
}

func TestLoadParsesSemanticCanonicalMap(t *testing.T) {
	t.Setenv("VIDERA_SEMANTIC_CANONICAL_MAP", "cats=felines,kitty=felines")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"cats":  "felines",
		"kitty": "felines",
	}, cfg.SemanticCanonicalMap)
}

func TestLoadInvalidSemanticCanonicalMapFails(t *testing.T) {
	t.Setenv("VIDERA_SEMANTIC_CANONICAL_MAP", "cats")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid VIDERA_SEMANTIC_CANONICAL_MAP")
}

func TestLoadInvalidSplitSharedStorageFlagFails(t *testing.T) {
	t.Setenv("VIDERA_SPLIT_SHARED_STORAGE", "maybe")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid VIDERA_SPLIT_SHARED_STORAGE")
}
