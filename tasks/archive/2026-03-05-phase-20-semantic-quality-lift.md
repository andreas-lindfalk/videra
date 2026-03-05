# Videra Phase 20 — Semantic Quality Lift (Archived 2026-03-05)

Status: GO

Reference:

- `tasks/platform/spec-implementation-alignment-2026-03-05.md`
- `tasks/platform/lancedb-storage-checkpoint-2026-03-05.md`
- `tasks/platform/semantic-quality-baseline-2026-03-05.md`
- `tasks/platform/semantic-quality-lift-evidence-2026-03-05.md`
- `internal/embedding/text.go`
- `internal/ingestion/clip.go`
- `internal/proofpack/`

### Definition of Done (Target)

- [x] Scope is explicit and aligned with Phase 19 checkpoint outcome (no immediate LanceDB migration).
- [x] MCP contract remains unchanged (`index_video`, `get_index_job`, `search_video`, `list_videos`, transcript resource).
- [x] Semantic quality baseline is measured, then improved with documented before/after evidence.
- [x] Deterministic ordering guarantees remain intact for equal-score tie paths.
- [x] Fast tests pass (`make test`).
- [x] Integration tests pass (`make integration-test-fresh`) including proofpack/determinism paths.
- [x] Release gates pass (`make release-gate` + `make release-gate-split`).
- [x] Todo is archive-ready with clear GO/NO-GO note for Phase 20 outcome.

### Scope

- [x] **In scope:** improve retrieval quality in current backend by upgrading embedding/ranking behavior behind existing interfaces.
- [x] **In scope:** define measurable quality criteria for semantic and early "vibe-like" relevance (fixture-based).
- [x] **In scope:** keep data-flow invariant (RO source paths + RW index path) and current release discipline.
- [x] **Out of scope:** storage backend migration to LanceDB.
- [x] **Out of scope:** watcher/event-driven source discovery implementation.
- [x] **Out of scope:** AgentGateway rollout implementation.
- [x] **Out of scope:** MCP tool/resource schema changes.

### Deliverables

- [x] Quality baseline artifact capturing current retrieval behavior on selected fixtures.
- [x] Updated semantic retrieval implementation with stable deterministic ranking behavior.
- [x] Updated/added tests that prove improved relevance without contract drift.
- [x] Phase evidence note with before/after results and release-gate outcomes.

### Acceptance Criteria (Objective)

- [x] At least one documented proofpack scenario improves measured top-k evidence quality vs baseline.
- [x] Determinism tests continue to pass (same ordered results for repeated identical queries).
- [x] Backward compatibility tests for tool response fields continue to pass unchanged.
- [x] No new required runtime dependencies are introduced for default slim profile.

### Testing (High-Rigor)

- [x] Run focused unit/integration tests for embedding + ranking paths first.
- [x] Run proofpack and deterministic ordering integrations.
- [x] Run full release gates after focused validations.

### Implementation Plan

- [x] Capture and persist current semantic-quality baseline on selected fixtures.
- [x] Implement quality-lift changes in embedding/ranking paths with minimal surface impact.
- [x] Add/tighten tests for relevance and deterministic behavior.
- [x] Run gates, record evidence, and finalize GO/NO-GO for archive handoff.
