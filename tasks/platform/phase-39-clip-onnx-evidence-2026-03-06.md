# Phase 39 — CLIP ONNX Integration Evidence (2026-03-06)

## Scope validated

- CLIP runtime contract added to config/runtime wiring (`VIDERA_VISUAL_BACKEND`, `VIDERA_CLIP_MODEL_PATH`, `VIDERA_CLIP_ORT_LIB_PATH`, `VIDERA_CLIP_INPUT_SIZE`).
- Real ingestion now resolves visual backend from options and falls back to OCR when CLIP setup is unavailable.
- Real ingestion now includes runtime failover (`clip` inference error -> OCR embedding) to avoid silent visual segment loss.
- ONNX-backed CLIP frame embedding path implemented directly in-process via native ONNX Runtime Go bindings.
- CLIP-capable native Docker runtime profile and Make targets added.

## Verification commands and outcomes

### Unit test baseline

- `runTests` (repo-wide)
- Result: `89 passed, 0 failed`

### Focused package checks

- `runTests` on `internal/config/config_test.go` + `internal/ingestion`
- Result: passed

### CLIP startup contract test coverage

- Added integration coverage in `test/integration/index_video_real_mode_test.go` for:
	- CLIP backend startup failure when `VIDERA_CLIP_MODEL_PATH` is missing
	- CLIP startup fallback warning path when configured model path is unavailable

### Formatting

- `gofmt -w` applied on all touched Go files.

## Assertions covered

- Config validation now rejects invalid visual backend values and invalid CLIP option combinations.
- CLIP embedder construction validates runtime prerequisites and frame embedding contract.
- Real ingester fallback behavior is deterministic at both initialization and runtime (`clip` unavailable/error -> OCR embedder).
- Startup capability logs now surface CLIP native shared-library readiness and fallback warning path.

## Phase closure note

- End-to-end semantic quality validation was completed in local real-mode smoke with native ONNX Runtime CLIP path.
- A local run indexed real video content with visual modality present (`visualSegments: 21`) and successful MCP smoke flow (`index_video`, `search_video`, `list_videos`, transcript resource).
