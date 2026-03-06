package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	endpoint := flag.String("endpoint", "http://localhost:8080/mcp", "MCP streamable HTTP endpoint")
	video := flag.String("video", "", "Video path visible to the server (e.g. /videos/demo.mp4)")
	flag.Parse()

	if *video == "" {
		log.Fatal("--video is required")
	}

	ctx := context.Background()
	cli, err := client.NewStreamableHttpClient(*endpoint)
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	defer cli.Close()

	if err := cli.Start(ctx); err != nil {
		log.Fatalf("start client: %v", err)
	}

	_, err = cli.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "videra-local-index",
				Version: "0.1.0",
			},
		},
	})
	if err != nil {
		log.Fatalf("initialize client: %v", err)
	}

	indexResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "index_video",
			Arguments: map[string]any{
				"path": *video,
			},
		},
	})
	if err != nil {
		log.Fatalf("index_video call failed: %v", err)
	}
	if indexResult.IsError {
		log.Fatalf("index_video returned tool error: %v", indexResult.Content)
	}

	indexPayload, ok := indexResult.StructuredContent.(map[string]any)
	if !ok {
		log.Fatal("index_video payload was not an object")
	}
	videoID, _ := indexPayload["videoId"].(string)
	if videoID == "" {
		log.Fatal("index_video response missing videoId")
	}

	fmt.Printf("Indexed: %s (videoId=%s)\n", *video, videoID)
}
