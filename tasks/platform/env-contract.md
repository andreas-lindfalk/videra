# Runtime Environment Contract (Parity Baseline)

Goal: keep a shared runtime contract across local, Hetzner, and Cloud Run where possible.

## Core Variables (Shared)

- `VIDERA_TRANSPORT` (`stdio` | `http`)
- `VIDERA_HTTP_ADDR` (e.g. `:8080`)
- `VIDERA_DATA_DIR` (persistent data path)
- `VIDERA_LOG_LEVEL` (`debug|info|warn|error`)
- `VIDERA_RUNTIME_MODE` (`local|test|prod`)
- `VIDERA_INGESTION_MODE` (`simulated|real`)
- `VIDERA_FRAME_INTERVAL_SEC` (default `5`)
- `VIDERA_DEFAULT_SEARCH_LIMIT` (default `5`)
- `VIDERA_INDEX_CONCURRENCY` (default `4`)
- `VIDERA_SEARCH_AUDIO_WEIGHT` (default `1.0`)
- `VIDERA_SEARCH_VISUAL_WEIGHT` (default `1.0`)

## Real-Ingestion Extras

- `VIDERA_WHISPER_MODEL` (default `tiny`)
- `VIDERA_WHISPER_LANGUAGE` (optional; empty means auto)
- `VIDERA_REMOTE_FETCH_ENABLED` (default `true`)
- `VIDERA_REMOTE_FETCH_TIMEOUT_SEC` (default `60`)
- `VIDERA_REMOTE_FETCH_MAX_MB` (default `200`)

## Async Queue Runtime Contract (Phase 12)

- `VIDERA_JOBQUEUE_BACKEND` (`inprocess|nats|redis`, default `inprocess`)
- `VIDERA_JOBQUEUE_ROLE` (`all|api|worker`, default `all`)
- `VIDERA_JOBQUEUE_RETRY_MAX_ATTEMPTS` (default `3`)
- `VIDERA_JOBQUEUE_RETRY_BACKOFF_MS` (default `250`)
- `VIDERA_JOBQUEUE_WORKER_POLL_MS` (default `250`)

NATS JetStream options:

- `VIDERA_JOBQUEUE_NATS_URL` (default `nats://127.0.0.1:4222`)
- `VIDERA_JOBQUEUE_NATS_STREAM` (default `videra_index_jobs`)
- `VIDERA_JOBQUEUE_NATS_SUBJECT` (default `videra.index.jobs`)
- `VIDERA_JOBQUEUE_NATS_CONSUMER` (default `videra-index-worker`)
- `VIDERA_JOBSTATE_NATS_BUCKET` (default `videra_index_job_status`)

Redis Streams options:

- `VIDERA_JOBQUEUE_REDIS_ADDR` (default `127.0.0.1:6379`)
- `VIDERA_JOBQUEUE_REDIS_PASSWORD` (default empty)
- `VIDERA_JOBQUEUE_REDIS_DB` (default `0`)
- `VIDERA_JOBQUEUE_REDIS_STREAM` (default `videra:index:jobs`)
- `VIDERA_JOBQUEUE_REDIS_GROUP` (default `videra-index-workers`)
- `VIDERA_JOBQUEUE_REDIS_CONSUMER` (default `videra-index-worker`)
- `VIDERA_JOBSTATE_REDIS_PREFIX` (default `videra:index:jobstatus:`)

Notes:

- Queue payloads are lightweight job instructions (`jobId`, source reference, attempts), not media bytes.
- `all` role runs MCP API + worker in one process; `api` and `worker` roles split responsibilities across processes.
- `api|worker` split requires an external backend (`nats` or `redis`); `inprocess` remains all-in-one local default.
- `worker` role requires `VIDERA_TRANSPORT=stdio`; startup fails fast if transport is configured as HTTP.
- Async worker lifecycle emits structured `queue_lifecycle` logs with stable keys: `event`, `job_id`, `status`, `attempt`, `max_attempts`, `delay_ms`, `error`.
- Redis-first rollout operations and rollback drills are documented in `tasks/platform/queue-redis-first-runbook.md`.
- Reproducible queue evidence baselines are tracked in `tasks/platform/queue-benchmark-evidence.md`.

Notes:
- In `real` mode, ingestion first looks for sidecar transcript files (`.srt`/`.vtt`/`.txt`).
- If no sidecar is found, it attempts FFmpeg audio extraction + Whisper CLI transcription.
- In `real` mode, remote HTTP(S) media sources are supported when fetch is enabled and within timeout/size limits.

## Image Profile Contract (Phase 7)

- `slim` profile is the default runtime baseline and should include ffmpeg-required paths for current default flows.
- `full` profile extends `slim` with runtime dependencies needed for fallback tooling (Whisper path + OCR tooling).
- MCP API/tool/resource behavior must remain identical across profiles; differences are runtime capability only.
- Missing optional tooling must be explicit in logs/errors rather than implicit/silent fallback.

## Provider-Specific Differences (Deployment Layer Only)

- Hetzner: server-visible local file paths can be mounted (`/videos/...`) and remote HTTP(S) URLs are also supported.
- Cloud Run: local file paths are not a reliable ingestion source; use remote HTTP(S) URLs for indexing parity.

These differences must remain in deployment/runbook logic, not MCP contract behavior.
