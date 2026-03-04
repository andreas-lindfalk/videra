# Container Runtime Profiles (Slim + Full)

Goal: define a stable image strategy so external tool dependencies are explicit, reproducible, and not forgotten.

## Profiles

### `slim` (default)

Purpose:
- Fast local/dev and CI workflows.
- Minimal runtime footprint for deterministic default flows.

Contains:
- Videra binary
- `ffmpeg`
- CA certificates

Expected behavior:
- Works for `simulated` mode and sidecar-driven `real` mode.
- If no transcript sidecar exists in `real` mode, transcription fallback may fail clearly when Whisper runtime is unavailable.

### `full`

Purpose:
- Production-like real-mode fallback support when sidecars are not present.

Contains:
- Everything in `slim`
- `python3` + Whisper-capable runtime path
- `tesseract` (for OCR visual context extraction)

Expected behavior:
- Supports audio transcription fallback path and OCR-enhanced visual context path.

## Image Contract

- MCP contracts (`index_video`, `search_video`, `list_videos`, transcript resource) must remain identical across profiles.
- Differences are runtime capability only, not API shape.
- Missing optional capabilities must be surfaced by explicit startup/runtime logs.

## Tagging Convention (Proposed)

- `videra:dev-slim`
- `videra:dev-full`
- CI/release variants follow same suffix pattern.

## Update and Security Cadence

- Pin base dependencies where practical.
- Rebuild image variants regularly (e.g., weekly) for security patches.
- Run vulnerability scan and artifact metadata generation (SBOM/signing) in CI.

## Verification Checklist

- `slim`: build + default local smoke + tests remain green.
- `full`: build + real-mode fallback path validation + tests remain green.
- Deterministic ordering and existing integration behavior remain unchanged.
