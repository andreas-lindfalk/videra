package config

import (
	"fmt"
	"os"
	"strings"
)

type Transport string

const (
	TransportStdio Transport = "stdio"
	TransportHTTP  Transport = "http"
)

type Config struct {
	Transport          Transport
	HTTPAddr           string
	DataDir            string
	LogLevel           string
	RuntimeMode        string
	FrameIntervalSec   int
	DefaultSearchLimit int
	IndexConcurrency   int
	SearchAudioWeight  float64
	SearchVisualWeight float64
}

func Load() (Config, error) {
	cfg := Config{
		Transport:          TransportStdio,
		HTTPAddr:           ":8080",
		DataDir:            "./data",
		LogLevel:           "info",
		RuntimeMode:        "local",
		FrameIntervalSec:   5,
		DefaultSearchLimit: 5,
		IndexConcurrency:   4,
		SearchAudioWeight:  1.0,
		SearchVisualWeight: 1.0,
	}

	if value, ok := os.LookupEnv("VIDERA_TRANSPORT"); ok {
		cfg.Transport = Transport(strings.ToLower(strings.TrimSpace(value)))
	}
	if value, ok := os.LookupEnv("VIDERA_HTTP_ADDR"); ok {
		cfg.HTTPAddr = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("VIDERA_DATA_DIR"); ok {
		cfg.DataDir = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("VIDERA_LOG_LEVEL"); ok {
		cfg.LogLevel = strings.ToLower(strings.TrimSpace(value))
	}
	if value, ok := os.LookupEnv("VIDERA_RUNTIME_MODE"); ok {
		cfg.RuntimeMode = strings.ToLower(strings.TrimSpace(value))
	}
	if value, ok := os.LookupEnv("VIDERA_FRAME_INTERVAL_SEC"); ok {
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &cfg.FrameIntervalSec); err != nil {
			return Config{}, fmt.Errorf("invalid VIDERA_FRAME_INTERVAL_SEC: %w", err)
		}
	}
	if value, ok := os.LookupEnv("VIDERA_DEFAULT_SEARCH_LIMIT"); ok {
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &cfg.DefaultSearchLimit); err != nil {
			return Config{}, fmt.Errorf("invalid VIDERA_DEFAULT_SEARCH_LIMIT: %w", err)
		}
	}
	if value, ok := os.LookupEnv("VIDERA_INDEX_CONCURRENCY"); ok {
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &cfg.IndexConcurrency); err != nil {
			return Config{}, fmt.Errorf("invalid VIDERA_INDEX_CONCURRENCY: %w", err)
		}
	}
	if value, ok := os.LookupEnv("VIDERA_SEARCH_AUDIO_WEIGHT"); ok {
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%f", &cfg.SearchAudioWeight); err != nil {
			return Config{}, fmt.Errorf("invalid VIDERA_SEARCH_AUDIO_WEIGHT: %w", err)
		}
	}
	if value, ok := os.LookupEnv("VIDERA_SEARCH_VISUAL_WEIGHT"); ok {
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%f", &cfg.SearchVisualWeight); err != nil {
			return Config{}, fmt.Errorf("invalid VIDERA_SEARCH_VISUAL_WEIGHT: %w", err)
		}
	}

	if cfg.Transport != TransportStdio && cfg.Transport != TransportHTTP {
		return Config{}, fmt.Errorf("unsupported VIDERA_TRANSPORT: %s", cfg.Transport)
	}
	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("VIDERA_HTTP_ADDR cannot be empty")
	}
	if cfg.DataDir == "" {
		return Config{}, fmt.Errorf("VIDERA_DATA_DIR cannot be empty")
	}
	if cfg.FrameIntervalSec <= 0 {
		return Config{}, fmt.Errorf("VIDERA_FRAME_INTERVAL_SEC must be > 0")
	}
	if cfg.DefaultSearchLimit <= 0 {
		return Config{}, fmt.Errorf("VIDERA_DEFAULT_SEARCH_LIMIT must be > 0")
	}
	if cfg.IndexConcurrency <= 0 {
		return Config{}, fmt.Errorf("VIDERA_INDEX_CONCURRENCY must be > 0")
	}
	if cfg.RuntimeMode == "" {
		return Config{}, fmt.Errorf("VIDERA_RUNTIME_MODE cannot be empty")
	}
	if cfg.SearchAudioWeight <= 0 {
		return Config{}, fmt.Errorf("VIDERA_SEARCH_AUDIO_WEIGHT must be > 0")
	}
	if cfg.SearchVisualWeight <= 0 {
		return Config{}, fmt.Errorf("VIDERA_SEARCH_VISUAL_WEIGHT must be > 0")
	}

	return cfg, nil
}
