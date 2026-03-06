# Phase 39 — True CLIP Integration (ONNX Runtime Path)

Archived: 2026-03-06
Status at archive: COMPLETE

## Why this was next

- Phase 38 removed simulated real-mode visual fallback and made behavior explicit.
- We needed true CLIP-based visual embeddings so visual retrieval became semantic, not OCR-only.

## Targets achieved

- Introduced a production CLIP embedder implementation for frame embeddings in real mode.
- Established a clear runtime contract for model/runtime availability and deterministic fallback semantics.
- Validated end-to-end local real-mode behavior with native ONNX Runtime CLIP path.

## Scope (in) completed

- `internal/ingestion/clip.go` refactor from stub-only helper to real backend + explicit fallback strategy.
- Runtime dependency contract for ONNX/CLIP model assets and startup capability signaling.
- Docker/operator path for CLIP-capable runtime profile.
- Unit + integration tests and local quality checklist updates for CLIP-driven visual retrieval.

## Scope (out)

- Hetzner/Cloud Run parity execution.
- Queue/backplane and storage architecture changes.
- Non-visual ranking redesign beyond what is needed to validate CLIP signal ingestion.

## Primary artifacts

- `internal/ingestion/clip.go`
- `internal/ingestion/clip_native_runner.go`
- `internal/ingestion/clip_native_runner_nocgo.go`
- `internal/ingestion/real.go`
- `internal/ingestion/runtime_capabilities.go`
- `internal/ingestion/clip_test.go`
- `internal/ingestion/real_test.go`
- `internal/ingestion/runtime_capabilities_test.go`
- `test/integration/index_video_real_mode_test.go`
- `internal/config/config.go`
- `internal/config/config_test.go`
- `cmd/videra/main.go`
- `Dockerfile`
- `Makefile`
- `README.md`
- `tasks/quality/local-retrieval-quality-checklist.md`
- `tasks/platform/env-contract.md`
- `tasks/platform/container-runtime-profiles.md`
- `tasks/platform/phase-39-clip-onnx-evidence-2026-03-06.md`

## Verification summary

- Unit baseline: `runTests` repo-wide passed (`92 passed, 0 failed`).
- Focused integration startup contract: passed (`TestRealModeCLIPBackendMissingModelPathFailsStartup`, `TestRealModeCLIPBackendUnavailableFallsBackToOCRAtStartup`).
- Local native CLIP smoke: passed via `localsmoke` against real mode with native ORT, with indexed video showing `visualSegments: 21` and modalities including `visual`.

## Execution plan (completed)

- [x] Define CLIP runtime contract (model path/env + capability detection + explicit error/fallback behavior).
- [x] Implement ONNX-backed CLIP embedder and wire it as real-mode primary visual embedder.
- [x] Add CLIP-capable runtime image path and local operator commands for reproducible startup.
- [x] Add/strengthen tests for CLIP embedding path, visual retrieval relevance, and deterministic query ordering.
- [x] Update docs/checklists and run focused verification with captured evidence.
