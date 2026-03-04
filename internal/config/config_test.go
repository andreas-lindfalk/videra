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
