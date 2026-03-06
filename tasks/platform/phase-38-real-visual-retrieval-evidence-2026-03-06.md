# Phase 38 — Real Visual Retrieval Core Evidence (2026-03-06)

## Scope validated

- Real ingestion visual path no longer defaults to simulated visual fallback.
- OCR-driven visual text path is primary in real mode.
- Runtime profile/docs updated so local LanceDB-native path includes OCR capability (`tesseract`).

## Commands and outcomes

### Unit / focused ingestion tests

- `runTests` on:
  - `internal/ingestion/real_test.go`
  - `internal/ingestion/runtime_capabilities_test.go`
- Result: passed (`16 passed, 0 failed`).

- `runTests` on all tests (repo-wide unit baseline)
- Result: passed (`81 passed, 0 failed`).

### LanceDB-native targeted integration lane

- Command: `make integration-test-lancedb-native`
- Result: pass.
  - `TestLanceDBBackendOnDefaultRuntimeReturnsGuidanceError`: PASS
  - `TestLanceDBNativeBackendIndexesAndSearches`: SKIP (expected on non-amd64 host per test guardrail)

## Behavioral assertions covered

- Real ingestion plain-text sidecar path indexes audio without fabricating visual segments when keyframes/OCR are unavailable.
- Real ingestion skips visual segments on keyframe extraction failure (no synthetic visual fallback content).
- Existing real-mode ingestion contract checks remain green.

## Changed artifacts (phase slice)

- `internal/ingestion/clip.go`
- `internal/ingestion/mock.go`
- `internal/ingestion/real.go`
- `internal/ingestion/real_test.go`
- `Dockerfile`
- `README.md`
- `tasks/quality/local-retrieval-quality-checklist.md`
- `tasks/platform/container-runtime-profiles.md`
- `tasks/platform/env-contract.md`
- `tasks/lessons.md`
