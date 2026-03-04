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

Notes:
- In `real` mode, ingestion first looks for sidecar transcript files (`.srt`/`.vtt`/`.txt`).
- If no sidecar is found, it attempts FFmpeg audio extraction + Whisper CLI transcription.

## Image Profile Contract (Phase 7)

- `slim` profile is the default runtime baseline and should include ffmpeg-required paths for current default flows.
- `full` profile extends `slim` with runtime dependencies needed for fallback tooling (Whisper path + OCR tooling).
- MCP API/tool/resource behavior must remain identical across profiles; differences are runtime capability only.
- Missing optional tooling must be explicit in logs/errors rather than implicit/silent fallback.

## Provider-Specific Differences (Deployment Layer Only)

- Hetzner: server-visible local file paths can be mounted (`/videos/...`).
- Cloud Run: local file paths are not a reliable ingestion source; use endpoint/runtime validation until cloud media source path is implemented.

These differences must remain in deployment/runbook logic, not MCP contract behavior.
