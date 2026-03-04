# Todo

## Active Task

### Videra Phase 12 — Queue Production Hardening (Worker Mode + Redis-First Rollout)

Reference:

- `tasks/platform/queue-vendor-checkpoint.md`
- `tasks/platform/jobqueue-interface-proposal.md`
- `tasks/platform/env-contract.md`
- `AGENTS.md` (queue portability + Cloud Run/Hetzner parity constraints)

### Definition of Done (Applied)

- [x] Scope is explicit (in/out) and aligned with current architecture decisions.
- [x] Required interfaces/contracts are updated without breaking existing MCP surface.
- [x] Fast tests pass (`make test`).
- [x] Integration tests pass (`make integration-test`) for changed behavior.
- [x] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [x] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [x] Todo is fully checked and ready for archive.

### Scope

- [x] **In scope:** Introduce explicit runtime role separation for async indexing (`api` enqueue path vs `worker` processing path) without MCP contract changes.
- [x] **In scope:** Harden queue reliability semantics (retry budget, backoff strategy, terminal failure handling) for production-like operation.
- [x] **In scope:** Validate and document a **Redis-first** rollout path when Redis is also used for adjacent key/value workloads, while preserving NATS fallback portability.
- [x] **In scope:** Add observability hooks for async lifecycle (queue backend, worker role, job transitions, retry/failure outcomes).
- [x] **Out of scope:** Autoscaling policy tuning, managed cloud queue vendor adoption, or auth/RBAC implementation inside core service.

### Go/No-Go Gate (Required Before Implementation)

- [x] Confirm default local behavior remains safe/simple (`inprocess` path still available with no external dependency).
- [x] Confirm rollback path to in-process mode with zero MCP API changes.
- [x] Confirm selected default external backend criteria are explicit (operational footprint, reliability evidence, and portability fallback).

### Deliverables

- [x] Add runtime role configuration contract (for example, API-only enqueue mode and worker mode) and wire startup behavior accordingly.
- [x] Add worker execution loop boundary with clean shutdown/cancellation semantics.
- [x] Add deterministic retry/failure policy wiring across async job lifecycle status updates.
- [x] Add integration coverage for split-role operation (enqueue in API process, completion by worker process) using real queue backend containers.
- [x] Update docs: `README.md`, `tasks/platform/env-contract.md`, and queue checkpoint notes with rollout guidance.

### Acceptance Criteria

- [x] `index_video` async path remains backward compatible and returns quickly with `jobId`.
- [x] `get_index_job` reflects deterministic state transitions under success, retry, and failure exhaustion paths.
- [x] Split-role flow is proven in integration tests with at least one external backend.
- [x] In-process mode remains default-compatible for local development and tests.
- [x] Recommendation for default external backend is evidence-backed and includes fallback trigger conditions.

### Testing (High-Rigor)

- [x] Add focused unit tests for worker loop control flow, retry budgeting, and terminal failure mapping.
- [x] Add deterministic integration tests for API/worker role split and queue-backed lifecycle behavior.
- [x] Run `make build`, `make test`, `make integration-test`, and relevant Docker build targets.

### Implementation Plan

- [x] Define runtime role/env contract and startup wiring approach.
- [x] Implement worker loop boundary with graceful shutdown and queue polling semantics.
- [x] Implement retry/backoff + terminal failure lifecycle mapping.
- [x] Add split-role integration scenarios with real backend containers.
- [x] Update docs/lessons and close DoD checklist for archive readiness.
