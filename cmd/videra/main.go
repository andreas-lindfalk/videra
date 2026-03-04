package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andreas-lindfalk/videra/internal/config"
	"github.com/andreas-lindfalk/videra/internal/embedding"
	"github.com/andreas-lindfalk/videra/internal/ingestion"
	"github.com/andreas-lindfalk/videra/internal/mcpserver"
	"github.com/andreas-lindfalk/videra/internal/storage"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "Videra"
	serverVersion = "0.1.0"
)

func main() {
	if err := run(); err != nil {
		log.Printf("server failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	runtimeCaps := ingestion.DetectRuntimeCapabilities()
	log.Printf("runtime capabilities: %s", runtimeCaps.Summary())
	if cfg.IngestionMode == config.IngestionModeReal {
		if !runtimeCaps.WhisperFallbackAvailable() {
			log.Printf("warning: real-mode transcription fallback is unavailable (install `whisper` CLI or python whisper module, or provide sidecar transcripts)")
		}
		if !runtimeCaps.Tesseract {
			log.Printf("info: tesseract not found; visual OCR enhancement is disabled")
		}
	}

	store, err := storage.NewChromemStore(cfg.DataDir, embedding.NewDeterministicTextEmbedder())
	if err != nil {
		return fmt.Errorf("initialize store: %w", err)
	}

	indexOptions := ingestion.IndexOptions{
		FrameIntervalSec:      cfg.FrameIntervalSec,
		Concurrency:           cfg.IndexConcurrency,
		RemoteFetchDisabled:   !cfg.RemoteFetchEnabled,
		RemoteFetchTimeoutSec: cfg.RemoteFetchTimeout,
		RemoteFetchMaxMB:      cfg.RemoteFetchMaxMB,
	}

	var ingester ingestion.Ingester
	switch cfg.IngestionMode {
	case config.IngestionModeSimulated:
		ingester = ingestion.NewMockIngester(store, indexOptions)
	case config.IngestionModeReal:
		ingester = ingestion.NewRealIngester(store, indexOptions)
	default:
		return fmt.Errorf("unsupported ingestion mode: %s", cfg.IngestionMode)
	}

	queue, err := newJobQueue(cfg)
	if err != nil {
		return fmt.Errorf("initialize job queue: %w", err)
	}
	jobStateStore, err := newIndexJobStateStore(cfg)
	if err != nil {
		return fmt.Errorf("initialize job state store: %w", err)
	}
	log.Printf("job queue backend: %s", cfg.JobQueueBackend)
	log.Printf("job queue role: %s", cfg.JobQueueRole)

	orchestrator := ingestion.NewSyncIndexOrchestratorWithOptions(ingester, store, queue, ingestion.SyncIndexOrchestratorOptions{
		JobStateStore:        jobStateStore,
		RunInlineAsyncWorker: false,
		SyncMaxAttempts:      2,
		AsyncRetryMax:        cfg.JobQueueRetryMax,
		AsyncRetryBackoff:    time.Duration(cfg.JobQueueRetryBackoff) * time.Millisecond,
		WorkerPollInterval:   time.Duration(cfg.JobQueueWorkerPollMS) * time.Millisecond,
	})

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.JobQueueRole == config.JobQueueRoleWorker {
		log.Printf("queue worker started (role=worker)")
		return orchestrator.RunWorker(shutdownCtx)
	}

	if cfg.JobQueueRole == config.JobQueueRoleAll {
		go func() {
			log.Printf("queue worker started (role=all)")
			if err := orchestrator.RunWorker(shutdownCtx); err != nil {
				log.Printf("queue worker stopped with error: %v", err)
			}
		}()
	}

	mcpSrv := mcpserver.New(serverName, serverVersion, orchestrator, store, cfg.DefaultSearchLimit, cfg.RuntimeMode, mcpserver.RankingOptions{
		AudioWeight:  cfg.SearchAudioWeight,
		VisualWeight: cfg.SearchVisualWeight,
	})

	switch cfg.Transport {
	case config.TransportStdio:
		log.Printf("starting Videra MCP server via stdio")
		return server.ServeStdio(mcpSrv.MCP())
	case config.TransportHTTP:
		log.Printf("starting Videra MCP server via streamable HTTP on %s", cfg.HTTPAddr)
		httpServer := server.NewStreamableHTTPServer(mcpSrv.MCP(), server.WithEndpointPath("/mcp"))
		return httpServer.Start(cfg.HTTPAddr)
	default:
		return fmt.Errorf("unsupported transport: %s", cfg.Transport)
	}
}

func newJobQueue(cfg config.Config) (ingestion.JobQueue, error) {
	switch cfg.JobQueueBackend {
	case config.JobQueueBackendInProcess:
		return ingestion.NewInProcessJobQueue(128), nil
	case config.JobQueueBackendNATS:
		return ingestion.NewNATSJetStreamJobQueue(ingestion.NATSJetStreamQueueConfig{
			URL:      cfg.JobQueueNATSURL,
			Stream:   cfg.JobQueueNATSStream,
			Subject:  cfg.JobQueueNATSSubject,
			Consumer: cfg.JobQueueNATSConsumer,
		})
	case config.JobQueueBackendRedis:
		return ingestion.NewRedisStreamsJobQueue(ingestion.RedisStreamsQueueConfig{
			Addr:     cfg.JobQueueRedisAddr,
			Password: cfg.JobQueueRedisPassword,
			DB:       cfg.JobQueueRedisDB,
			Stream:   cfg.JobQueueRedisStream,
			Group:    cfg.JobQueueRedisGroup,
			Consumer: cfg.JobQueueRedisConsumer,
		})
	default:
		return nil, fmt.Errorf("unsupported job queue backend: %s", cfg.JobQueueBackend)
	}
}

func newIndexJobStateStore(cfg config.Config) (ingestion.IndexJobStateStore, error) {
	switch cfg.JobQueueBackend {
	case config.JobQueueBackendInProcess:
		return ingestion.NewInMemoryIndexJobStateStore(), nil
	case config.JobQueueBackendNATS:
		return ingestion.NewNATSIndexJobStateStore(ingestion.NATSIndexJobStateStoreConfig{
			URL:    cfg.JobQueueNATSURL,
			Bucket: cfg.JobStateNATSBucket,
		})
	case config.JobQueueBackendRedis:
		return ingestion.NewRedisIndexJobStateStore(ingestion.RedisIndexJobStateStoreConfig{
			Addr:      cfg.JobQueueRedisAddr,
			Password:  cfg.JobQueueRedisPassword,
			DB:        cfg.JobQueueRedisDB,
			KeyPrefix: cfg.JobStateRedisPrefix,
		})
	default:
		return nil, fmt.Errorf("unsupported job queue backend for state store: %s", cfg.JobQueueBackend)
	}
}
