# Phase 38 — Real Visual Retrieval Core (Replace CLIP Stub Path)

Archived: 2026-03-06
Status at archive: COMPLETE

## Why this was next

- We had stabilized LanceDB-default local flow, but real-mode visual retrieval still depended on simulated stub behavior.
- We needed real-mode visual behavior to be explicit and non-fabricated before proceeding to true CLIP integration and later cloud parity execution.

## Primary artifacts

- `internal/ingestion/clip.go`
- `internal/ingestion/mock.go`
- `internal/ingestion/real.go`
- `internal/ingestion/real_test.go`
- `Dockerfile`
- `README.md`
- `tasks/quality/local-retrieval-quality-checklist.md`
- `tasks/platform/container-runtime-profiles.md`
- `tasks/platform/env-contract.md`
- `tasks/platform/phase-38-real-visual-retrieval-evidence-2026-03-06.md`
- `tasks/lessons.md`
- `tasks/todo.md`

## Progress snapshot

- Removed simulated visual fallback from real ingestion path.
- Made OCR-driven visual extraction the primary real-mode visual path.
- Enforced explicit behavior when keyframes/OCR are unavailable: skip visual segments instead of generating fabricated context.
- Updated LanceDB-native runtime image to include OCR capability (`tesseract`).
- Synced docs/checklists/runtime contract with the new visual behavior.
- Validation completed:
  - focused ingestion tests passed (`internal/ingestion/real_test.go`, `internal/ingestion/runtime_capabilities_test.go`)
  - full unit baseline passed (`runTests`: 81 passed, 0 failed)
  - targeted LanceDB integration lane passed with expected architecture skip (`make integration-test-lancedb-native`)

## Execution plan (completed)

- [x] Define production visual-embedder contract and wire it into real ingestion as the primary path.
- [x] Implement first real visual backend path and explicit fallback/error semantics when dependencies are unavailable.
- [x] Add/strengthen unit + integration tests for visual segment generation, visual query relevance, and deterministic ordering.
- [x] Update local quality checklist and README run-path so visual validation is operator-reproducible.
- [x] Run focused verification (`make test`, targeted integration real-mode checks) and capture phase evidence.
