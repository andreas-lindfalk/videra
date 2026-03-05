# Real Corpus Onboarding Checklist — 2026-03-05

Purpose: provide a repeatable, operator-friendly checklist for onboarding real corpus videos while preserving deterministic validation discipline.

## 1) Input Contract

- Video sources:
  - Local mounted paths (for local/self-hosted environments)
  - Remote HTTP(S) URLs (required for cloud parity where local paths are not stable)
- Transcript sidecars for local real-mode quality checks:
  - same basename as video file
  - supported extensions: `.srt`, `.vtt`, `.txt`
  - example:
    - `videos/session-01.mp4`
    - `videos/session-01.txt`

## 2) Runtime Configuration Baseline

Minimum real-mode env for local validation:

- `VIDERA_INGESTION_MODE=real`
- `VIDERA_TRANSPORT=http`
- `VIDERA_HTTP_ADDR=:8080`
- `VIDERA_DATA_DIR=./data`

Remote fetch controls (when source is URL):

- `VIDERA_REMOTE_FETCH_ENABLED` (`true` or `false`)
- `VIDERA_REMOTE_FETCH_TIMEOUT_SEC`
- `VIDERA_REMOTE_FETCH_MAX_MB`

## 3) Onboarding Procedure (Local)

1. Place video + sidecar transcript in mounted corpus directory.
2. Start service in real mode.
3. Call `index_video` with local file path.
4. Call `search_video` using at least 3 representative queries.
5. Call `list_videos` and transcript resource (`video://{id}/transcript`) to verify indexing completeness.

## 4) Acceptance Checks per Video Batch

- Indexing contract:
  - `index_video` returns non-error response with `videoId`.
  - `list_videos` includes indexed item.
- Transcript/resource contract:
  - `video://{id}/transcript` returns non-empty segments.
- Search contract:
  - query returns at least one result for known evidence phrases.

## 5) Failure Semantics to Verify Explicitly

- Local real-mode without sidecar transcript should fail with sidecar-required message.
- Remote real-mode with fetch disabled should fail with disabled-fetch message.
- Remote real-mode oversized payload should fail with max-size bounded message.

## 6) Repeatability Notes

- Keep deterministic query set per corpus domain (saved alongside onboarding evidence).
- Record the exact env values used during onboarding (including fetch bounds).
- Capture command outputs and pass/fail status in phase evidence file.
