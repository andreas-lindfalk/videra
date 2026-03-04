# Archive — Videra MVP Phase 1 (Project Bootstrap)

Archived: 2026-03-04
Source: tasks/todo.md

# Videra MVP — Phase 1: Project Bootstrap

## Setup
- [x] Initialize Go 1.25.4 module (`github.com/andreas-lindfalk/videra`)
- [x] Add dependencies: `mcp-go`, `chromem-go`, `testify`, `testcontainers-go`
- [x] Create project directory structure (`cmd/`, `internal/`, `test/`)

## Configuration
- [x] Implement `internal/config/config.go` — Config struct loaded from `VIDERA_` env vars

## Storage Layer
- [x] Define `VectorStore` interface in `internal/storage/store.go`
- [x] Define domain models in `internal/storage/models.go` (Video, Segment, SearchResult)
- [x] Implement `ChromemStore` in `internal/storage/chromem.go`

## Ingestion
- [x] Define `Ingester` interface in `internal/ingestion/pipeline.go`
- [x] Implement `MockIngester` in `internal/ingestion/mock.go` (simulated transcription)
- [x] Stub `FFmpegRunner` interface + `ExecFFmpeg` in `internal/ingestion/ffmpeg.go`

## MCP Server
- [x] Implement `NewVideraServer()` in `internal/mcpserver/server.go`
- [x] `index_video` tool: calls Ingester, returns videoID + status
- [x] `search_video` tool: stub returning "not yet implemented"
- [x] `list_videos` tool: calls store.ListVideos()
- [x] `video://{id}/transcript` resource template

## Entrypoint
- [x] Implement `cmd/videra/main.go` — wire deps, transport switch, graceful shutdown

## Docker
- [x] Create multi-stage `Dockerfile` (builder → alpine + ffmpeg)
- [x] Create `.dockerignore`
- [x] Create `docker-compose.yml`

## Testing
- [x] Create `test/integration/testhelpers.go` — StartVideraContainer()
- [x] Create `test/integration/index_video_test.go` — full round-trip test
- [x] Error handling test: invalid path returns tool error
- [x] Add build tag `//go:build integration`

## Build & Dev
- [x] Create `Makefile` (build, test, integration-test, docker-build, run-stdio, run-http)

## Verification
- [x] `make build` compiles
- [x] `make test` passes unit tests
- [x] `make docker-build` produces image with ffmpeg
- [x] `make integration-test` passes full round-trip
