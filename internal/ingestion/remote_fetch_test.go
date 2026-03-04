package ingestion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHTTPRemoteFetcherRejectsUnsupportedScheme(t *testing.T) {
	fetcher := NewHTTPRemoteFetcher(RemoteFetchOptions{Disabled: false, Timeout: 2 * time.Second, MaxBytes: 1024})

	_, _, err := fetcher.Fetch(context.Background(), "ftp://example.com/video.mp4")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported remote source scheme")
}

func TestHTTPRemoteFetcherRespectsDisabledFlag(t *testing.T) {
	fetcher := NewHTTPRemoteFetcher(RemoteFetchOptions{Disabled: true, Timeout: 2 * time.Second, MaxBytes: 1024})

	_, _, err := fetcher.Fetch(context.Background(), "https://example.com/video.mp4")
	require.Error(t, err)
	require.Contains(t, err.Error(), "remote ingestion is disabled")
}

func TestHTTPRemoteFetcherRespectsMaxSize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(make([]byte, 2048))
	}))
	defer server.Close()

	fetcher := NewHTTPRemoteFetcher(RemoteFetchOptions{Disabled: false, Timeout: 2 * time.Second, MaxBytes: 1024})

	_, _, err := fetcher.Fetch(context.Background(), server.URL+"/clip.mp4")
	require.Error(t, err)
	require.Contains(t, err.Error(), "remote media exceeds maximum size")
}

func TestHTTPRemoteFetcherFetchesToTempFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("video-payload"))
	}))
	defer server.Close()

	fetcher := NewHTTPRemoteFetcher(RemoteFetchOptions{Disabled: false, Timeout: 2 * time.Second, MaxBytes: 1024})

	path, cleanup, err := fetcher.Fetch(context.Background(), server.URL+"/clip.mp4")
	require.NoError(t, err)
	require.NotEmpty(t, path)
	require.NotNil(t, cleanup)

	cleanup()
}
