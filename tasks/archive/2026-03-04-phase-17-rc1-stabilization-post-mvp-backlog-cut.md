# Todo

## Active Task

### Videra Phase 17 — RC1 Stabilization + Post-MVP Backlog Cut

Reference:

- `AGENTS.md` (architecture and test rigor constraints)
- `README.md` (operator/developer workflows)
- `tasks/platform/mvp-release-gate.md`
- `tasks/platform/mvp-release-gate-evidence-2026-03-04.md`
- `tasks/platform/parity-validation-checklist.md`
- `tasks/platform/queue-redis-first-runbook.md`

### Definition of Done (Target)

- [x] Scope is explicit (in/out) and aligned with current architecture decisions.
- [x] Required interfaces/contracts are updated without breaking existing MCP surface.
- [x] Fast tests pass (`make test`).
- [x] Integration tests pass (`make integration-test`) for changed behavior.
- [x] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [x] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [x] Todo is fully checked and archive-ready.

### Scope

- [x] **In scope:** Address release-candidate stabilization issues found during repeated full-matrix runs (flaky execution, Docker-space sensitivity, command-path ambiguity).
- [x] **In scope:** Consolidate remaining actionable work into an explicit post-MVP backlog cut, separated by priority/risk.
- [x] **In scope:** Tighten release/operator guidance for routine RC re-validation and failure triage.
- [x] **In scope:** Keep all changes minimal and contract-safe (no MCP tool/schema changes).
- [x] **Out of scope:** New product capabilities, new external infrastructure dependencies, and major architecture shifts.

### RC1 Stability Gate

- [x] `make release-gate` and `make release-gate-split` are repeatable in normal local conditions with documented failure handling steps.
- [x] Any identified flaky path has either a deterministic fix or an explicit, reproducible mitigation documented for operators.
- [x] Release-critical contract paths remain unchanged (`index_video`, `get_index_job`, `search_video`, `list_videos`, transcript resource).
- [x] Post-MVP backlog entries are explicit, bounded, and not mixed into RC1 stabilization scope.

### Deliverables

- [x] Add/update docs for release validation troubleshooting and rerun discipline (especially Docker/testcontainers pressure cases).
- [x] Add/adjust targeted tests only for stabilization fixes that reduce nondeterminism risk.
- [x] Create a post-MVP backlog artifact capturing deferred items, ownership hints, and priority labels.
- [x] Produce an updated RC1 evidence snapshot after stabilization changes.

### Acceptance Criteria

- [x] RC1 stabilization changes are reproducible and verified without contract drift.
- [x] Operators can execute and troubleshoot the release gate using documented, copy/paste-safe steps.
- [x] Deferred work is clearly isolated into post-MVP backlog with no hidden scope creep.
- [x] Final status is ready for either tagged release prep or explicit no-go with concrete blockers.

### Testing (High-Rigor)

- [x] Run targeted tests for each stabilization fix.
- [x] Run full suite: `make release-gate` and `make release-gate-split`.
- [x] Re-run any previously unstable command path at least once to validate stability/mitigation.

### Implementation Plan

- [x] Identify and classify RC1 stabilization gaps from recent run history and current docs.
- [x] Implement minimal, high-value stabilization changes with focused verification.
- [x] Update runbooks/release docs with deterministic troubleshooting and rerun workflow.
- [x] Create post-MVP backlog cut (explicit deferred scope and priorities).
- [x] Capture updated evidence, update lessons, and complete DoD for archive readiness.
