# Todo

## Active Task

### Videra Phase 10 — Broker Decision Checkpoint (Queue Portability)

Reference:

- `AGENTS.md` (Queue portability path + vendor decision checkpoint)
- `tasks/archive/2026-03-04-phase-9-async-index-job-boundary.md` (post-phase broker direction)

### Definition of Done (Applied)

- [x] Scope is explicit (in/out) and aligned with current architecture decisions.
- [x] Required interfaces/contracts are updated without breaking existing MCP surface.
- [x] Fast tests pass (`make test`) when code changes are introduced.
- [x] Integration tests pass (`make integration-test`) when behavioral changes are introduced.
- [x] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [x] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [x] Todo is fully checked and ready for archive.

### Scope

- [x] **In scope:** Produce a mandatory vendor-choice checkpoint document before any broker implementation work.
- [x] **In scope:** Define a broker-agnostic `JobQueue` interface proposal and contract boundaries.
- [x] **In scope:** Evaluate **NATS (JetStream)** first, plus at least one neutral/self-hosted alternative, using explicit criteria.
- [x] **In scope:** Define acceptance criteria for future broker adoption (local bootstrap, HA path, retry/ack semantics, lock-in risk).
- [x] **Out of scope:** Implementing broker wiring in runtime, adding distributed workers, or changing MCP tool contracts.

### Deliverables

- [x] Add `tasks/platform/queue-vendor-checkpoint.md` with candidate matrix, lock-in risks, and fallback plan.
- [x] Add `tasks/platform/jobqueue-interface-proposal.md` with interface boundary and migration notes.
- [x] Add explicit recommendation and decision record entry criteria (what evidence is required before implementation).
- [x] Cross-link final checkpoint outputs from `AGENTS.md` and `tasks/todo.md`.

### Acceptance Criteria

- [x] Candidate comparison is explicit, reproducible, and includes NATS/JetStream + one neutral/self-hosted alternative.
- [x] Portability/fallback path is documented for private self-hosted deployments.
- [x] Proposed `JobQueue` interface does not require MCP contract changes.
- [x] A clear “go/no-go” checklist exists for starting broker implementation in the next phase.

### Testing (High-Rigor)

- [ ] If code prototypes are added, include focused unit tests for interface behavior and failure semantics.
- [x] If no code is added, validate all references/paths/docs are internally consistent and actionable.

### Implementation Plan

- [x] Draft vendor-checkpoint matrix and evaluation rubric.
- [x] Draft `JobQueue` interface proposal with lifecycle/state mapping.
- [x] Capture recommendation with fallback and lock-in analysis.
- [x] Run any relevant verification (tests/build) if code changes occur.
- [x] Update lessons and mark phase complete/ready for archive.
