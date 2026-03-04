package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	endpoint := flag.String("endpoint", "http://localhost:8080/mcp", "MCP streamable HTTP endpoint")
	video := flag.String("video", "", "Video path visible to the server (e.g. /videos/demo.mp4)")
	query := flag.String("query", "budget roadmap", "Search query")
	limit := flag.Int("limit", 5, "Search result limit")
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
				Name:    "videra-local-smoke",
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

	searchResult, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search_video",
			Arguments: map[string]any{
				"query":        *query,
				"limit":        *limit,
				"includeDebug": true,
			},
		},
	})
	if err != nil {
		log.Fatalf("search_video call failed: %v", err)
	}
	if searchResult.IsError {
		log.Fatalf("search_video returned tool error: %v", searchResult.Content)
	}

	listResult, err := cli.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "list_videos"}})
	if err != nil {
		log.Fatalf("list_videos call failed: %v", err)
	}
	if listResult.IsError {
		log.Fatalf("list_videos returned tool error: %v", listResult.Content)
	}

	resourceResult, err := cli.ReadResource(ctx, mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{URI: fmt.Sprintf("video://%s/transcript", videoID)},
	})
	if err != nil {
		log.Fatalf("read transcript resource failed: %v", err)
	}

	fmt.Fprintln(os.Stdout, "✅ Local smoke-test passed")
	fmt.Fprintf(os.Stdout, "- endpoint: %s\n", *endpoint)
	fmt.Fprintf(os.Stdout, "- indexed videoId: %s\n", videoID)
	fmt.Fprintf(os.Stdout, "- search response type: %T\n", searchResult.StructuredContent)
	fmt.Fprintf(os.Stdout, "- list response type: %T\n", listResult.StructuredContent)
	fmt.Fprintf(os.Stdout, "- transcript resources: %d\n", len(resourceResult.Contents))
}
