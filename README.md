# Videra

Privacy-native multimodal video memory MCP server in Go.

## Current MVP Status

- The local MCP flow is working end-to-end (`index_video`, `search_video`, `list_videos`, transcript resource).
- Current transcript/visual text content is simulated placeholder content; this validates pipeline/runtime behavior more than true semantic understanding.
- Deployment parity planning (Cloud Run + Hetzner) and real semantic ingestion are tracked in `tasks/todo.md` and `tasks/platform/hetzner-gcp-parity-primer.md`.

## Quick Start (Local, Non-Cloud)

This is the recommended pre-CloudRun validation path.

Fastest path (assuming a file exists in `./videos`):

```bash
make local-e2e QUERY="test clip"
```

### 1) Start local MCP HTTP service

```bash
make local-up
```

This starts the service on `http://localhost:8080/mcp`.

### 2) Run deterministic smoke test with a local file

Mount a local folder into the container as `/videos` by setting `VIDERA_VIDEO_DIR`.

Quickest default flow (uses first video in `./videos`):

```bash
make local-smoke-default QUERY="budget roadmap"
```

Practical first-test flow (repo-local `./videos` folder):

```bash
mkdir -p videos
# put a file in ./videos, for example: ./videos/IMG_3711.MOV
make local-up
make local-smoke-default QUERY="test clip"
make local-down
```

By default, local media in `videos/` is ignored by Git (`videos/*` with `videos/.gitkeep` allowed).

```bash
VIDERA_VIDEO_DIR=/absolute/path/to/your/video/folder make local-up
make local-smoke VIDEO=/videos/your-file.mp4 QUERY="budget roadmap"
```

If you do not set `VIDERA_VIDEO_DIR`, compose mounts the repo root (`.`) to `/videos`.
So a file at `./videos/IMG_3711.MOV` is visible in-container as `/videos/videos/IMG_3711.MOV`.

The smoke command validates this full flow:

- `index_video`
- `search_video`
- `list_videos`
- `video://{id}/transcript`

### 3) Stop local stack

```bash
make local-down
```

## VS Code MCP Setup (Simple)

A workspace config is included at `.vscode/mcp.json` pointing to `http://localhost:8080/mcp`.

Practical flow:

1. Run `make local-up` (or `make local-e2e`).
2. In VS Code/Copilot MCP settings, enable/use the workspace server config.
3. Test with MCP tools: `list_videos`, then `search_video`.

## Developer Commands

- Build: `make build`
- Fast tests: `make test`
- Integration tests: `make integration-test`
- Docker build: `make docker-build`
- Stdio run: `make run-stdio`
- HTTP run: `make run-http`

## Copilot / MCP Client Setup Notes

Videra supports both stdio and streamable HTTP patterns.

- **Stdio mode:** start with `make run-stdio` (or equivalent docker command) and configure your MCP client to launch that command.
- **HTTP mode:** start with `make local-up` and point your MCP client to `http://localhost:8080/mcp`.

For VS Code/Copilot MCP configuration, use the MCP server setup UI and provide either:

- a command-based server (stdio), or
- a URL-based server endpoint (HTTP)

Then verify by calling `list_videos` first.

## Troubleshooting

- `index_video` path not found:
	- ensure file path is visible inside runtime (container path vs host path).
	- for compose flow, use `/videos/...` and set `VIDERA_VIDEO_DIR`.
- MCP connection failure:
	- check service is running and endpoint matches (`/mcp`).
	- verify no port conflict on `8080`.
- Docker compose starts but smoke fails:
	- run `docker compose logs -f videra` and inspect startup/runtime errors.
- Integration tests flaky due to state:
	- use built-in test reset tooling and rerun with `-count=1` when debugging.