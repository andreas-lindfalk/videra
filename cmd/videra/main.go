package main

import (
	"fmt"
	"log"
	"os"

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

	store, err := storage.NewChromemStore(cfg.DataDir, embedding.NewDeterministicTextEmbedder())
	if err != nil {
		return fmt.Errorf("initialize store: %w", err)
	}

	ingester := ingestion.NewMockIngester(store, ingestion.IndexOptions{
		FrameIntervalSec: cfg.FrameIntervalSec,
		Concurrency:      cfg.IndexConcurrency,
	})
	orchestrator := ingestion.NewSyncIndexOrchestrator(ingester, store)
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
