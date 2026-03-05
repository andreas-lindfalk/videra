# Videra Phase 23 — Real Corpus Onboarding & Quality Gate (Archived 2026-03-05)

Status: GO

Reference:

- `README.md`
- `internal/proofpack/`
- `test/integration/index_video_real_mode_test.go`
- `tasks/platform/env-contract.md`
- `tasks/platform/domain-profile-evaluation-evidence-2026-03-05.md`
- `tasks/platform/real-corpus-onboarding-checklist-2026-03-05.md`
- `tasks/platform/real-corpus-quality-gate-2026-03-05.md`
- `tasks/platform/real-corpus-quality-gate-evidence-2026-03-05.md`

### Definition of Done (Target)

- [x] A reproducible real-corpus onboarding checklist is defined (inputs, paths, sidecars, expected outputs).
- [x] Objective quality gate metrics are defined for top-k evidence quality and deterministic behavior.
- [x] At least one real-mode integration validation path is documented and runnable locally.
- [x] MCP contract remains unchanged (`index_video`, `get_index_job`, `search_video`, `list_videos`, transcript resource).
- [x] Focused tests pass and evidence note is produced with GO/NO-GO recommendation.

### Scope

- [x] **In scope:** define real corpus intake requirements and acceptance criteria.
- [x] **In scope:** add/update fixtures or test scenarios needed for quality gate evaluation.
- [x] **In scope:** document operator/developer run steps for repeatable validation.
- [x] **Out of scope:** MCP schema changes, queue architecture changes, storage backend migration.

### Implementation Plan

- [x] Draft real-corpus onboarding checklist artifact with file/source constraints.
- [x] Define measurable quality thresholds and map them to existing/new test assertions.
- [x] Execute focused validation runs and record outputs in a phase evidence file.
- [x] Finalize GO/NO-GO snapshot and prepare archive-ready task state.
