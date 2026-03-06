# Container Runtime Profiles (Slim + Full + LanceDB Native)

Goal: define a stable image strategy so external tool dependencies are explicit, reproducible, and not forgotten.

## Profiles

### `runtime-lancedb-native` (local default)

Purpose:
- LanceDB-native local/default runtime path with first-class visual OCR support.

Contains:
- Videra binary built with `lancedb_native`
- `ffmpeg`
- `tesseract`
- CA certificates

Expected behavior:
- Works for `simulated` mode and sidecar-driven `real` mode.
- Supports OCR-derived visual segment generation in `real` mode.
- Does not include Whisper runtime fallback tooling (`whisper`/python-whisper).

### `runtime-lancedb-native-clip` (CLIP ONNX path)

Purpose:
- LanceDB-native runtime with CLIP ONNX visual embedding dependencies for real-mode visual semantics.

Contains:
- Everything in `runtime-lancedb-native`
- Native ONNX Runtime shared library (`libonnxruntime.so`)

Expected behavior:
- Supports `VIDERA_VISUAL_BACKEND=clip` when `VIDERA_CLIP_MODEL_PATH` is provided.
- Uses `VIDERA_CLIP_ORT_LIB_PATH` (default `/usr/local/lib/libonnxruntime.so`) for native model execution.
- On missing CLIP runtime/model dependencies, startup/runtime logs explicit fallback to OCR behavior.

### `slim` (minimal fallback)

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
- `videra:dev-lancedb-native`
- `videra:dev-lancedb-native-clip`
- CI/release variants follow same suffix pattern.

## Update and Security Cadence

- Pin base dependencies where practical.
- Rebuild image variants regularly (e.g., weekly) for security patches.
- Run vulnerability scan and artifact metadata generation (SBOM/signing) in CI.

## Verification Checklist

- `runtime-lancedb-native`: build + local LanceDB default smoke + real visual OCR validation remains green.
- `runtime-lancedb-native-clip`: build + CLIP backend startup validation + real visual indexing smoke remains green.
- `slim`: build + default local smoke + tests remain green.
- `full`: build + real-mode fallback path validation + tests remain green.
- Deterministic ordering and existing integration behavior remain unchanged.
