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

type IngestionMode string

const (
	IngestionModeSimulated IngestionMode = "simulated"
	IngestionModeReal      IngestionMode = "real"
)

type JobQueueBackend string

const (
	JobQueueBackendInProcess JobQueueBackend = "inprocess"
	JobQueueBackendNATS      JobQueueBackend = "nats"
	JobQueueBackendRedis     JobQueueBackend = "redis"
)

type JobQueueRole string

const (
	JobQueueRoleAll    JobQueueRole = "all"
	JobQueueRoleAPI    JobQueueRole = "api"
	JobQueueRoleWorker JobQueueRole = "worker"
)

type StorageBackend string

const (
	StorageBackendChromem StorageBackend = "chromem"
	StorageBackendLanceDB StorageBackend = "lancedb"
)

type Config struct {
	Transport             Transport
	HTTPAddr              string
	DataDir               string
	StorageBackend        StorageBackend
	LogLevel              string
	RuntimeMode           string
	IngestionMode         IngestionMode
	RemoteFetchEnabled    bool
	RemoteFetchTimeout    int
	RemoteFetchMaxMB      int
	FrameIntervalSec      int
	DefaultSearchLimit    int
	IndexConcurrency      int
	SearchAudioWeight     float64
	SearchVisualWeight    float64
	SemanticCanonicalMap  map[string]string
	JobQueueBackend       JobQueueBackend
	JobQueueRole          JobQueueRole
	JobQueueRetryMax      int
	JobQueueRetryBackoff  int
	JobQueueWorkerPollMS  int
	JobQueueNATSURL       string
	JobQueueNATSStream    string
	JobQueueNATSSubject   string
	JobQueueNATSConsumer  string
	JobStateNATSBucket    string
	JobQueueRedisAddr     string
	JobQueueRedisPassword string
	JobQueueRedisDB       int
	JobQueueRedisStream   string
	JobQueueRedisGroup    string
	JobQueueRedisConsumer string
	JobStateRedisPrefix   string
	SplitSharedStorage    bool
}

func Load() (Config, error) {
	cfg := Config{
		Transport:             TransportStdio,
		HTTPAddr:              ":8080",
		DataDir:               "./data",
		StorageBackend:        StorageBackendChromem,
		LogLevel:              "info",
		RuntimeMode:           "local",
		IngestionMode:         IngestionModeSimulated,
		RemoteFetchEnabled:    true,
		RemoteFetchTimeout:    60,
		RemoteFetchMaxMB:      200,
		FrameIntervalSec:      5,
		DefaultSearchLimit:    5,
		IndexConcurrency:      4,
		SearchAudioWeight:     1.0,
		SearchVisualWeight:    1.0,
		SemanticCanonicalMap:  map[string]string{},
		JobQueueBackend:       JobQueueBackendInProcess,
		JobQueueRole:          JobQueueRoleAll,
		JobQueueRetryMax:      3,
		JobQueueRetryBackoff:  250,
		JobQueueWorkerPollMS:  250,
		JobQueueNATSURL:       "nats://127.0.0.1:4222",
		JobQueueNATSStream:    "videra_index_jobs",
		JobQueueNATSSubject:   "videra.index.jobs",
		JobQueueNATSConsumer:  "videra-index-worker",
		JobStateNATSBucket:    "videra_index_job_status",
		JobQueueRedisAddr:     "127.0.0.1:6379",
		JobQueueRedisPassword: "",
		JobQueueRedisDB:       0,
		JobQueueRedisStream:   "videra:index:jobs",
		JobQueueRedisGroup:    "videra-index-workers",
		JobQueueRedisConsumer: "videra-index-worker",
		JobStateRedisPrefix:   "videra:index:jobstatus:",
		SplitSharedStorage:    false,
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
	if value, ok := os.LookupEnv("VIDERA_STORAGE_BACKEND"); ok {
		cfg.StorageBackend = StorageBackend(strings.ToLower(strings.TrimSpace(value)))
	}
	if value, ok := os.LookupEnv("VIDERA_LOG_LEVEL"); ok {
		cfg.LogLevel = strings.ToLower(strings.TrimSpace(value))
	}
	if value, ok := os.LookupEnv("VIDERA_RUNTIME_MODE"); ok {
		cfg.RuntimeMode = strings.ToLower(strings.TrimSpace(value))
	}
	if value, ok := os.LookupEnv("VIDERA_INGESTION_MODE"); ok {
		cfg.IngestionMode = IngestionMode(strings.ToLower(strings.TrimSpace(value)))
	}
	if value, ok := os.LookupEnv("VIDERA_REMOTE_FETCH_ENABLED"); ok {
		normalized := strings.ToLower(strings.TrimSpace(value))
		switch normalized {
		case "true", "1", "yes", "on":
			cfg.RemoteFetchEnabled = true
		case "false", "0", "no", "off":
			cfg.RemoteFetchEnabled = false
		default:
			return Config{}, fmt.Errorf("invalid VIDERA_REMOTE_FETCH_ENABLED: %s", value)
		}
	}
	if value, ok := os.LookupEnv("VIDERA_REMOTE_FETCH_TIMEOUT_SEC"); ok {
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &cfg.RemoteFetchTimeout); err != nil {
			return Config{}, fmt.Errorf("invalid VIDERA_REMOTE_FETCH_TIMEOUT_SEC: %w", err)
		}
	}
	if value, ok := os.LookupEnv("VIDERA_REMOTE_FETCH_MAX_MB"); ok {
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &cfg.RemoteFetchMaxMB); err != nil {
			return Config{}, fmt.Errorf("invalid VIDERA_REMOTE_FETCH_MAX_MB: %w", err)
		}
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
	if value, ok := os.LookupEnv("VIDERA_SEMANTIC_CANONICAL_MAP"); ok {
		parsed, err := parseCanonicalMap(value)
		if err != nil {
			return Config{}, fmt.Errorf("invalid VIDERA_SEMANTIC_CANONICAL_MAP: %w", err)
		}
		cfg.SemanticCanonicalMap = parsed
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_BACKEND"); ok {
		cfg.JobQueueBackend = JobQueueBackend(strings.ToLower(strings.TrimSpace(value)))
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_ROLE"); ok {
		cfg.JobQueueRole = JobQueueRole(strings.ToLower(strings.TrimSpace(value)))
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_RETRY_MAX_ATTEMPTS"); ok {
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &cfg.JobQueueRetryMax); err != nil {
			return Config{}, fmt.Errorf("invalid VIDERA_JOBQUEUE_RETRY_MAX_ATTEMPTS: %w", err)
		}
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_RETRY_BACKOFF_MS"); ok {
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &cfg.JobQueueRetryBackoff); err != nil {
			return Config{}, fmt.Errorf("invalid VIDERA_JOBQUEUE_RETRY_BACKOFF_MS: %w", err)
		}
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_WORKER_POLL_MS"); ok {
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &cfg.JobQueueWorkerPollMS); err != nil {
			return Config{}, fmt.Errorf("invalid VIDERA_JOBQUEUE_WORKER_POLL_MS: %w", err)
		}
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_NATS_URL"); ok {
		cfg.JobQueueNATSURL = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_NATS_STREAM"); ok {
		cfg.JobQueueNATSStream = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_NATS_SUBJECT"); ok {
		cfg.JobQueueNATSSubject = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_NATS_CONSUMER"); ok {
		cfg.JobQueueNATSConsumer = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("VIDERA_JOBSTATE_NATS_BUCKET"); ok {
		cfg.JobStateNATSBucket = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_REDIS_ADDR"); ok {
		cfg.JobQueueRedisAddr = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_REDIS_PASSWORD"); ok {
		cfg.JobQueueRedisPassword = value
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_REDIS_DB"); ok {
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &cfg.JobQueueRedisDB); err != nil {
			return Config{}, fmt.Errorf("invalid VIDERA_JOBQUEUE_REDIS_DB: %w", err)
		}
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_REDIS_STREAM"); ok {
		cfg.JobQueueRedisStream = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_REDIS_GROUP"); ok {
		cfg.JobQueueRedisGroup = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("VIDERA_JOBQUEUE_REDIS_CONSUMER"); ok {
		cfg.JobQueueRedisConsumer = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("VIDERA_JOBSTATE_REDIS_PREFIX"); ok {
		cfg.JobStateRedisPrefix = strings.TrimSpace(value)
	}
	if value, ok := os.LookupEnv("VIDERA_SPLIT_SHARED_STORAGE"); ok {
		normalized := strings.ToLower(strings.TrimSpace(value))
		switch normalized {
		case "true", "1", "yes", "on":
			cfg.SplitSharedStorage = true
		case "false", "0", "no", "off":
			cfg.SplitSharedStorage = false
		default:
			return Config{}, fmt.Errorf("invalid VIDERA_SPLIT_SHARED_STORAGE: %s", value)
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
	if cfg.StorageBackend != StorageBackendChromem && cfg.StorageBackend != StorageBackendLanceDB {
		return Config{}, fmt.Errorf("unsupported VIDERA_STORAGE_BACKEND: %s", cfg.StorageBackend)
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
	if cfg.IngestionMode != IngestionModeSimulated && cfg.IngestionMode != IngestionModeReal {
		return Config{}, fmt.Errorf("unsupported VIDERA_INGESTION_MODE: %s", cfg.IngestionMode)
	}
	if cfg.RemoteFetchTimeout <= 0 {
		return Config{}, fmt.Errorf("VIDERA_REMOTE_FETCH_TIMEOUT_SEC must be > 0")
	}
	if cfg.RemoteFetchMaxMB <= 0 {
		return Config{}, fmt.Errorf("VIDERA_REMOTE_FETCH_MAX_MB must be > 0")
	}
	if cfg.SearchAudioWeight <= 0 {
		return Config{}, fmt.Errorf("VIDERA_SEARCH_AUDIO_WEIGHT must be > 0")
	}
	if cfg.SearchVisualWeight <= 0 {
		return Config{}, fmt.Errorf("VIDERA_SEARCH_VISUAL_WEIGHT must be > 0")
	}
	if cfg.JobQueueBackend != JobQueueBackendInProcess && cfg.JobQueueBackend != JobQueueBackendNATS && cfg.JobQueueBackend != JobQueueBackendRedis {
		return Config{}, fmt.Errorf("unsupported VIDERA_JOBQUEUE_BACKEND: %s", cfg.JobQueueBackend)
	}
	if cfg.JobQueueRole != JobQueueRoleAll && cfg.JobQueueRole != JobQueueRoleAPI && cfg.JobQueueRole != JobQueueRoleWorker {
		return Config{}, fmt.Errorf("unsupported VIDERA_JOBQUEUE_ROLE: %s", cfg.JobQueueRole)
	}
	if cfg.JobQueueRetryMax <= 0 {
		return Config{}, fmt.Errorf("VIDERA_JOBQUEUE_RETRY_MAX_ATTEMPTS must be > 0")
	}
	if cfg.JobQueueRetryBackoff < 0 {
		return Config{}, fmt.Errorf("VIDERA_JOBQUEUE_RETRY_BACKOFF_MS must be >= 0")
	}
	if cfg.JobQueueWorkerPollMS <= 0 {
		return Config{}, fmt.Errorf("VIDERA_JOBQUEUE_WORKER_POLL_MS must be > 0")
	}
	if cfg.JobQueueNATSURL == "" {
		return Config{}, fmt.Errorf("VIDERA_JOBQUEUE_NATS_URL cannot be empty")
	}
	if cfg.JobQueueNATSStream == "" {
		return Config{}, fmt.Errorf("VIDERA_JOBQUEUE_NATS_STREAM cannot be empty")
	}
	if cfg.JobQueueNATSSubject == "" {
		return Config{}, fmt.Errorf("VIDERA_JOBQUEUE_NATS_SUBJECT cannot be empty")
	}
	if cfg.JobQueueNATSConsumer == "" {
		return Config{}, fmt.Errorf("VIDERA_JOBQUEUE_NATS_CONSUMER cannot be empty")
	}
	if cfg.JobStateNATSBucket == "" {
		return Config{}, fmt.Errorf("VIDERA_JOBSTATE_NATS_BUCKET cannot be empty")
	}
	if cfg.JobQueueRedisAddr == "" {
		return Config{}, fmt.Errorf("VIDERA_JOBQUEUE_REDIS_ADDR cannot be empty")
	}
	if cfg.JobQueueRedisDB < 0 {
		return Config{}, fmt.Errorf("VIDERA_JOBQUEUE_REDIS_DB must be >= 0")
	}
	if cfg.JobQueueRedisStream == "" {
		return Config{}, fmt.Errorf("VIDERA_JOBQUEUE_REDIS_STREAM cannot be empty")
	}
	if cfg.JobQueueRedisGroup == "" {
		return Config{}, fmt.Errorf("VIDERA_JOBQUEUE_REDIS_GROUP cannot be empty")
	}
	if cfg.JobQueueRedisConsumer == "" {
		return Config{}, fmt.Errorf("VIDERA_JOBQUEUE_REDIS_CONSUMER cannot be empty")
	}
	if cfg.JobStateRedisPrefix == "" {
		return Config{}, fmt.Errorf("VIDERA_JOBSTATE_REDIS_PREFIX cannot be empty")
	}
	if cfg.JobQueueBackend == JobQueueBackendInProcess && cfg.JobQueueRole != JobQueueRoleAll {
		return Config{}, fmt.Errorf("VIDERA_JOBQUEUE_ROLE=%s requires an external queue backend", cfg.JobQueueRole)
	}
	if cfg.JobQueueRole == JobQueueRoleWorker && cfg.Transport != TransportStdio {
		return Config{}, fmt.Errorf("VIDERA_JOBQUEUE_ROLE=worker requires VIDERA_TRANSPORT=stdio")
	}

	return cfg, nil
}

func parseCanonicalMap(value string) (map[string]string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return map[string]string{}, nil
	}

	entries := strings.Split(trimmed, ",")
	out := make(map[string]string, len(entries))
	for _, entry := range entries {
		pair := strings.TrimSpace(entry)
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("entry %q must use key=value format", pair)
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		canonical := strings.ToLower(strings.TrimSpace(parts[1]))
		if key == "" || canonical == "" {
			return nil, fmt.Errorf("entry %q has empty key or value", pair)
		}

		out[key] = canonical
	}

	return out, nil
}
