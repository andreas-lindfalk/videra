# Videra

Privacy-native multimodal video memory MCP server in Go.

## Current MVP Status

- The local MCP flow is working end-to-end (`index_video`, `search_video`, `list_videos`, transcript resource).
- Default mode is `VIDERA_INGESTION_MODE=simulated` (fixture-like placeholder segments for deterministic testing).
- A new local `real` ingestion mode is available: `VIDERA_INGESTION_MODE=real` uses sidecar transcript files (`.srt`, `.vtt`, `.txt`) next to the video file.
- In `real` mode, if no sidecar is found, Videra attempts FFmpeg audio extraction + Whisper CLI transcription (`whisper` or `python3 -m whisper`).
- In `real` mode, `index_video` now accepts remote HTTP(S) media URLs with bounded fetch controls (`VIDERA_REMOTE_FETCH_ENABLED`, `VIDERA_REMOTE_FETCH_TIMEOUT_SEC`, `VIDERA_REMOTE_FETCH_MAX_MB`).
- `index_video` supports `mode=async` for non-blocking indexing; poll status via `get_index_job` using returned `jobId`.
- Async queue backend is runtime-selectable via `VIDERA_JOBQUEUE_BACKEND` (`inprocess` default; `nats` and `redis` available), with runtime role split support via `VIDERA_JOBQUEUE_ROLE` (`all|api|worker`).
- Deployment parity planning (Cloud Run + Hetzner) and real semantic ingestion are tracked in `tasks/todo.md` and `tasks/platform/hetzner-gcp-parity-primer.md`.
- Container runtime profile strategy (`slim` default + `full` tool-complete) is tracked in `tasks/platform/container-runtime-profiles.md`.

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

By default, local compose builds the `runtime-slim` profile.
Use the full tool-complete runtime profile when you need real-mode fallback tooling:

```bash
VIDERA_DOCKER_TARGET=runtime-full make local-up
```

### 2) Run deterministic smoke test with a local file

Mount a local folder into the container as `/videos` by setting `VIDERA_VIDEO_DIR`.

Quickest default flow (uses first video in `./videos`):

```bash
make local-smoke-default QUERY="test query"
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

To test `real` ingestion mode locally, place a transcript sidecar next to the video file:

- `./videos/IMG_3711.MOV`
- `./videos/IMG_3711.txt` (or `.srt` / `.vtt`)

Then run:

```bash
VIDERA_INGESTION_MODE=real make local-up
make local-smoke VIDEO=/videos/videos/IMG_3711.MOV QUERY="test query"
make local-down
```

Optional Whisper settings for transcription fallback:

- `VIDERA_WHISPER_MODEL` (default: `tiny`)
- `VIDERA_WHISPER_LANGUAGE` (optional language hint)

Remote media fetch controls (used when `index_video` path is `http://` or `https://`):

- `VIDERA_REMOTE_FETCH_ENABLED` (default: `true`)
- `VIDERA_REMOTE_FETCH_TIMEOUT_SEC` (default: `60`)
- `VIDERA_REMOTE_FETCH_MAX_MB` (default: `200`)

```bash
VIDERA_VIDEO_DIR=/absolute/path/to/your/video/folder make local-up
make local-smoke VIDEO=/videos/your-file.mp4 QUERY="test query"
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
- Docker build (slim default): `make docker-build` or `make docker-build-slim`
- Docker build (full tool-complete): `make docker-build-full`
- Stdio run: `make run-stdio`
- HTTP run: `make run-http`
- Stdio run (full): `make run-stdio-full`
- HTTP run (full): `make run-http-full`

## Copilot / MCP Client Setup Notes

Videra supports both stdio and streamable HTTP patterns.

- **Stdio mode:** start with `make run-stdio` (or equivalent docker command) and configure your MCP client to launch that command.
- **HTTP mode:** start with `make local-up` and point your MCP client to `http://localhost:8080/mcp`.

For VS Code/Copilot MCP configuration, use the MCP server setup UI and provide either:

- a command-based server (stdio), or
- a URL-based server endpoint (HTTP)

Then verify by calling `list_videos` first.

## Async Indexing Flow

Use async mode when indexing should not block the MCP caller:

1. Call `index_video` with `mode="async"`.
2. Read returned `jobId`.
3. Poll `get_index_job` with that `jobId` until `status` is `completed` or `failed`.

Example initiation payload:

```json
{
	"path": "https://example.com/meeting.mp4",
	"mode": "async"
}
```

Queue runtime environment options (Phase 12):

- `VIDERA_JOBQUEUE_BACKEND` (`inprocess|nats|redis`; default `inprocess`)
- `VIDERA_JOBQUEUE_ROLE` (`all|api|worker`; default `all`)
- `VIDERA_JOBQUEUE_RETRY_MAX_ATTEMPTS` (default `3`)
- `VIDERA_JOBQUEUE_RETRY_BACKOFF_MS` (default `250`)
- `VIDERA_JOBQUEUE_WORKER_POLL_MS` (default `250`)
- NATS options: `VIDERA_JOBQUEUE_NATS_URL`, `VIDERA_JOBQUEUE_NATS_STREAM`, `VIDERA_JOBQUEUE_NATS_SUBJECT`, `VIDERA_JOBQUEUE_NATS_CONSUMER`
- NATS job-state option: `VIDERA_JOBSTATE_NATS_BUCKET`
- Redis options: `VIDERA_JOBQUEUE_REDIS_ADDR`, `VIDERA_JOBQUEUE_REDIS_PASSWORD`, `VIDERA_JOBQUEUE_REDIS_DB`, `VIDERA_JOBQUEUE_REDIS_STREAM`, `VIDERA_JOBQUEUE_REDIS_GROUP`, `VIDERA_JOBQUEUE_REDIS_CONSUMER`
- Redis job-state option: `VIDERA_JOBSTATE_REDIS_PREFIX`
- Split-role shared data-plane option: `VIDERA_SPLIT_SHARED_STORAGE` (`false` default)

Role notes:

- `all` runs MCP API + worker loop in one process.
- `api` runs MCP API only (enqueue + `get_index_job` reads shared job-state backend).
- `worker` runs queue processing only (no MCP endpoint).
- `worker` role is startup-validated to require `VIDERA_TRANSPORT=stdio` (fail-fast on invalid transport config).
- `api|worker` split requires an external queue backend (`nats` or `redis`); `inprocess` is all-in-one only.
- In split `api|worker` mode, API visibility of worker-indexed content (`list_videos` / `search_video` / transcript resource) requires shared `VIDERA_DATA_DIR` plus `VIDERA_SPLIT_SHARED_STORAGE=true` on both roles.

Split-role data-plane notes (Phase 15):

- With `VIDERA_SPLIT_SHARED_STORAGE=true`, the store persists/reloads index manifests from `VIDERA_DATA_DIR` so API reads can reflect worker indexing under a shared mount.
- With `VIDERA_SPLIT_SHARED_STORAGE=false`, `get_index_job` remains shared-state correct, but API content visibility can be degraded in non-shared topologies.
- Startup logs explicitly signal this mode to operators.

Queue lifecycle observability keys (Phase 14):

- Worker events are emitted as structured logs with prefix `queue_lifecycle`.
- Stable fields: `event`, `job_id`, `status`, `attempt`, `max_attempts`, `delay_ms`, `error`.
- Core events: `enqueued`, `reserved`, `retry_scheduled`, `completed`, `retry_exhausted`, `retry_failed_terminal`.

Queue payloads are job instructions (source reference + job metadata), not video bytes.

Redis-first rollout guidance:

- Keep `inprocess` as default unless explicit scale/SLO pressure requires external queueing.
- Prefer Redis as first external backend when stack consolidation is a priority.
- Prefer NATS when dedicated broker specialization is desired over consolidation.
- Use Phase 13 operational guidance and fallback drills in:
	- `tasks/platform/queue-redis-first-runbook.md`
	- `tasks/platform/queue-benchmark-evidence.md`

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