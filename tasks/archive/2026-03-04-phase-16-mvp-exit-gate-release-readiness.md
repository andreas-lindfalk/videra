# Todo

## Active Task

### Videra Phase 16 — MVP Exit Gate (Release Candidate Readiness)

Reference:

- `AGENTS.md` (architecture + testing constraints)
- `README.md` (developer/operator workflows)
- `tasks/platform/parity-validation-checklist.md`
- `tasks/platform/env-contract.md`
- `tasks/platform/queue-redis-first-runbook.md`
- `tasks/platform/hetzner-gcp-parity-primer.md`

### Definition of Done (Target)

- [x] Scope is explicit (in/out) and aligned with current architecture decisions.
- [x] Required interfaces/contracts are updated without breaking existing MCP surface.
- [x] Fast tests pass (`make test`).
- [x] Integration tests pass (`make integration-test`) for changed behavior.
- [x] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [x] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [x] Todo is fully checked and archive-ready.

### Scope

- [x] **In scope:** Define and implement a bounded MVP release-readiness gate spanning local, split-role, and parity-critical runtime paths.
- [x] **In scope:** Consolidate verification evidence into a deterministic pass/fail matrix (build, test, integration, container build, key split-role flows).
- [x] **In scope:** Tighten operator guidance for local/private deployment and cloud-parity expectations so required topology assumptions are explicit.
- [x] **In scope:** Close any final contract-level gaps discovered by the parity checklist without changing existing MCP tool schemas.
- [x] **Out of scope:** New product features, auth/RBAC implementation, managed cloud service provisioning, and backend migration beyond current storage/queue decisions.

### Release-Readiness Gate

- [x] MVP contract paths are validated end-to-end: `index_video`, `get_index_job`, `search_video`, `list_videos`, transcript resource.
- [x] Split-role behavior is explicitly verified for both shared-storage-correct and degraded/non-shared semantics.
- [x] Local operator loop is reproducible via documented commands with no ambiguous setup assumptions.
- [x] Cloud/Hetzner parity notes clearly separate supported behavior, limitations, and required deployment topology.

### Deliverables

- [x] Add/update verification workflow artifacts (commands/checklist) that provide a single reproducible MVP go/no-go signal.
- [x] Add/adjust targeted tests only where parity checklist reveals contract-risk gaps.
- [x] Update `README.md` and relevant runbooks/checklists so rollout and validation steps are copy/paste-ready.
- [x] Document explicit release risks/deferred items (if any) without expanding scope into new feature tracks.

### Acceptance Criteria

- [x] Release-readiness matrix is green with reproducible evidence and no MCP schema changes.
- [x] Required split-role data/control-plane semantics are validated and documented for operators.
- [x] Local/private deployment guidance is clear enough for first-pass execution without tribal knowledge.
- [x] Remaining non-MVP work is explicitly captured as post-MVP backlog, not folded into this phase.

### Testing (High-Rigor)

- [x] Run targeted unit/integration tests for any gap-closing changes introduced in this phase.
- [x] Run full validation suite: `make build`, `make test`, `make integration-test`, `make docker-build`.
- [x] Re-run focused split-role integration checks used as release-critical signals.

### Implementation Plan

- [x] Translate parity checklist into an explicit MVP release gate and identify any failing/ambiguous checks.
- [x] Implement minimal code/doc/test changes to close gate failures while preserving current contracts.
- [x] Validate release-critical flows with deterministic evidence capture.
- [x] Update docs/runbooks/checklists to match verified runtime behavior and operator steps.
- [x] Record lessons and complete DoD for archive readiness.
