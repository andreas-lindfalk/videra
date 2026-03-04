# Todo

## Active Task

### Videra Phase 12 — Queue Production Hardening (Worker Mode + Redis-First Rollout)

Reference:

- `tasks/platform/queue-vendor-checkpoint.md`
- `tasks/platform/jobqueue-interface-proposal.md`
- `tasks/platform/env-contract.md`
- `AGENTS.md` (queue portability + Cloud Run/Hetzner parity constraints)

### Definition of Done (Applied)

- [ ] Scope is explicit (in/out) and aligned with current architecture decisions.
- [ ] Required interfaces/contracts are updated without breaking existing MCP surface.
- [ ] Fast tests pass (`make test`).
- [ ] Integration tests pass (`make integration-test`) for changed behavior.
- [ ] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [ ] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [ ] Todo is fully checked and ready for archive.

### Scope

- [ ] **In scope:** Introduce explicit runtime role separation for async indexing (`api` enqueue path vs `worker` processing path) without MCP contract changes.
- [ ] **In scope:** Harden queue reliability semantics (retry budget, backoff strategy, terminal failure handling) for production-like operation.
- [ ] **In scope:** Validate and document a **Redis-first** rollout path when Redis is also used for adjacent key/value workloads, while preserving NATS fallback portability.
- [ ] **In scope:** Add observability hooks for async lifecycle (queue backend, worker role, job transitions, retry/failure outcomes).
- [ ] **Out of scope:** Autoscaling policy tuning, managed cloud queue vendor adoption, or auth/RBAC implementation inside core service.

### Go/No-Go Gate (Required Before Implementation)

- [ ] Confirm default local behavior remains safe/simple (`inprocess` path still available with no external dependency).
- [ ] Confirm rollback path to in-process mode with zero MCP API changes.
- [ ] Confirm selected default external backend criteria are explicit (operational footprint, reliability evidence, and portability fallback).

### Deliverables

- [ ] Add runtime role configuration contract (for example, API-only enqueue mode and worker mode) and wire startup behavior accordingly.
- [ ] Add worker execution loop boundary with clean shutdown/cancellation semantics.
- [ ] Add deterministic retry/failure policy wiring across async job lifecycle status updates.
- [ ] Add integration coverage for split-role operation (enqueue in API process, completion by worker process) using real queue backend containers.
- [ ] Update docs: `README.md`, `tasks/platform/env-contract.md`, and queue checkpoint notes with rollout guidance.

### Acceptance Criteria

- [ ] `index_video` async path remains backward compatible and returns quickly with `jobId`.
- [ ] `get_index_job` reflects deterministic state transitions under success, retry, and failure exhaustion paths.
- [ ] Split-role flow is proven in integration tests with at least one external backend.
- [ ] In-process mode remains default-compatible for local development and tests.
- [ ] Recommendation for default external backend is evidence-backed and includes fallback trigger conditions.

### Testing (High-Rigor)

- [ ] Add focused unit tests for worker loop control flow, retry budgeting, and terminal failure mapping.
- [ ] Add deterministic integration tests for API/worker role split and queue-backed lifecycle behavior.
- [ ] Run `make build`, `make test`, `make integration-test`, and relevant Docker build targets.

### Implementation Plan

- [ ] Define runtime role/env contract and startup wiring approach.
- [ ] Implement worker loop boundary with graceful shutdown and queue polling semantics.
- [ ] Implement retry/backoff + terminal failure lifecycle mapping.
- [ ] Add split-role integration scenarios with real backend containers.
- [ ] Update docs/lessons and close DoD checklist for archive readiness.

## Definition of Done Template

Use this checklist at the top of each new phase/feature todo.

- [ ] Scope is explicit (in/out) and aligned with current architecture decisions.
- [ ] Required interfaces/contracts are updated without breaking existing MCP surface.
- [ ] Fast tests pass (`make test`).
- [ ] Integration tests pass (`make integration-test`).
- [ ] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [ ] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [ ] Todo is fully checked and ready for archive.
