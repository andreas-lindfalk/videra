# Todo

## Active Task

### Phase 39 — True CLIP Integration (ONNX Runtime Path)

Status:

- [x] Planned
- [x] In progress
- [x] Complete

Why this is next:

- [x] Phase 38 removed simulated real-mode visual fallback and made behavior explicit.
- [x] We now need to implement actual CLIP-based visual embeddings so visual retrieval is semantic, not OCR-only.

Targets / goals:

- [x] Introduce a production CLIP embedder implementation for frame embeddings in real mode.
- [x] Keep a clear runtime contract for model/runtime availability and deterministic fallback semantics.
- [x] Validate end-to-end that visual semantic queries retrieve expected visual moments in local real-mode runs.

Scope (in):

- [x] `internal/ingestion/clip.go` refactor from stub-only helper to real backend + explicit fallback strategy.
- [x] Runtime dependency contract for ONNX/CLIP model assets and startup capability signaling.
- [x] Docker/operator path for CLIP-capable runtime profile.
- [x] Unit + integration tests and local quality checklist updates for CLIP-driven visual retrieval.

Scope (out):

- [x] Hetzner/Cloud Run parity execution.
- [x] Queue/backplane and storage architecture changes.
- [x] Non-visual ranking redesign beyond what is needed to validate CLIP signal ingestion.

Primary artifacts:

- [x] `internal/ingestion/clip.go`
- [x] `internal/ingestion/real.go`
- [x] `internal/ingestion/runtime_capabilities.go`
- [x] `internal/ingestion/real_test.go`
- [x] `test/integration/index_video_real_mode_test.go`
- [x] `Dockerfile`
- [x] `Makefile`
- [x] `README.md`
- [x] `tasks/quality/local-retrieval-quality-checklist.md`
- [x] `tasks/platform/env-contract.md`
- [x] `tasks/platform/phase-39-clip-onnx-evidence-2026-03-06.md`
- [x] `tasks/todo.md`

Execution plan:

- [x] Define CLIP runtime contract (model path/env + capability detection + explicit error/fallback behavior).
- [x] Implement ONNX-backed CLIP embedder and wire it as real-mode primary visual embedder.
- [x] Add CLIP-capable runtime image path and local operator commands for reproducible startup.
- [x] Add/strengthen tests for CLIP embedding path, visual retrieval relevance, and deterministic query ordering.
- [x] Update docs/checklists and run focused verification with captured evidence.

## Definition of Done Template

Use this checklist at the top of each new phase/feature todo.

- [ ] Scope is explicit (in/out) and aligned with current architecture decisions.
- [ ] Required interfaces/contracts are updated without breaking existing MCP surface.
- [ ] Fast tests pass (`make test`).
- [ ] Integration tests pass (`make integration-test`).
- [ ] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [ ] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [ ] Todo is fully checked and ready for archive.
