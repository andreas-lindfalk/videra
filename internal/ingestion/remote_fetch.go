package ingestion

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RemoteFetcher interface {
	Fetch(ctx context.Context, sourceURL string) (string, func(), error)
}

type RemoteFetchOptions struct {
	Disabled bool
	Timeout  time.Duration
	MaxBytes int64
}

type HTTPRemoteFetcher struct {
	options RemoteFetchOptions
	client  *http.Client
}

func NewHTTPRemoteFetcher(options RemoteFetchOptions) *HTTPRemoteFetcher {
	if options.Timeout <= 0 {
		options.Timeout = 60 * time.Second
	}
	if options.MaxBytes <= 0 {
		options.MaxBytes = 200 * 1024 * 1024
	}
	return &HTTPRemoteFetcher{
		options: options,
		client:  &http.Client{Timeout: options.Timeout},
	}
}

func (f *HTTPRemoteFetcher) Fetch(ctx context.Context, sourceURL string) (string, func(), error) {
	if f.options.Disabled {
		return "", nil, fmt.Errorf("remote ingestion is disabled by configuration")
	}

	parsed, err := url.Parse(strings.TrimSpace(sourceURL))
	if err != nil {
		return "", nil, fmt.Errorf("invalid remote source URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", nil, fmt.Errorf("unsupported remote source scheme: %s", parsed.Scheme)
	}

	requestCtx, cancel := context.WithTimeout(ctx, f.options.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return "", nil, fmt.Errorf("build remote fetch request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("fetch remote media: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("remote media request failed with status %d", resp.StatusCode)
	}
	if resp.ContentLength > 0 && resp.ContentLength > f.options.MaxBytes {
		return "", nil, fmt.Errorf("remote media exceeds maximum size (%d bytes)", f.options.MaxBytes)
	}

	ext := filepath.Ext(parsed.Path)
	if ext == "" {
		ext = ".bin"
	}

	tempFile, err := os.CreateTemp("", "videra-remote-*"+ext)
	if err != nil {
		return "", nil, fmt.Errorf("create temp file for remote media: %w", err)
	}
	tempPath := tempFile.Name()
	cleanup := func() {
		_ = os.Remove(tempPath)
	}

	limitedReader := io.LimitReader(resp.Body, f.options.MaxBytes+1)
	written, copyErr := io.Copy(tempFile, limitedReader)
	closeErr := tempFile.Close()
	if copyErr != nil {
		cleanup()
		return "", nil, fmt.Errorf("copy remote media to temp file: %w", copyErr)
	}
	if closeErr != nil {
		cleanup()
		return "", nil, fmt.Errorf("close temp file for remote media: %w", closeErr)
	}
	if written > f.options.MaxBytes {
		cleanup()
		return "", nil, fmt.Errorf("remote media exceeds maximum size (%d bytes)", f.options.MaxBytes)
	}

	return tempPath, cleanup, nil
}

var _ RemoteFetcher = (*HTTPRemoteFetcher)(nil)
