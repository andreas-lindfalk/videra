# Todo

## Active Task

### Videra Phase 13 — Queue Go/No-Go Readiness (Redis-First Rollout)

Reference:

- `tasks/platform/queue-vendor-checkpoint.md`
- `tasks/platform/env-contract.md`
- `README.md` (async indexing + role split contract)
- `AGENTS.md` (Cloud Run + Hetzner parity and portability constraints)

### Definition of Done (Applied)

- [ ] Scope is explicit (in/out) and aligned with current architecture decisions.
- [ ] Required interfaces/contracts are updated without breaking existing MCP surface.
- [ ] Fast tests pass (`make test`).
- [ ] Integration tests pass (`make integration-test`) for changed behavior.
- [ ] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [ ] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [ ] Todo is fully checked and ready for archive.

### Scope

- [ ] **In scope:** Produce go/no-go evidence for selecting Redis Streams as first external queue backend under the existing portability model.
- [ ] **In scope:** Add benchmark/profiling evidence for async indexing lifecycle in local Docker and private VM-like flow.
- [ ] **In scope:** Add failure-path verification for retry exhaustion and poison-job terminal handling under split role mode.
- [ ] **In scope:** Add operational + security runbooks for Redis-first deployment, with NATS fallback trigger criteria.
- [ ] **Out of scope:** Autoscaling algorithm tuning, managed queue vendor integration, and new MCP tool contract changes.

### Go/No-Go Gate (Required Before External Backend Defaulting)

- [ ] Candidate benchmarked with reproducible commands and recorded baseline metrics.
- [ ] Failure semantics validated with deterministic tests (retry budget, terminal fail path, duplicate/idempotency behavior).
- [ ] Operational runbook exists for single-node and HA-oriented deployment path.
- [ ] Security model is documented (auth/TLS/secrets/audit responsibilities at runtime edge vs service internals).
- [ ] Rollback drill to `inprocess` mode is documented and verified with no MCP API changes.

### Deliverables

- [ ] Add/update benchmark artifact doc under `tasks/platform/` with reproducible command set and evidence format.
- [ ] Add/update queue ops runbook(s) for Redis-first and NATS fallback deployment in local/private environments.
- [ ] Add explicit rollback trigger conditions and fallback procedure in queue checkpoint docs.
- [ ] Add deterministic integration coverage for retry exhaustion/terminal handling in split-role mode.
- [ ] Update `README.md`, `tasks/platform/env-contract.md`, and `tasks/platform/queue-vendor-checkpoint.md` with finalized guidance.

### Acceptance Criteria

- [ ] Recommendation for Redis-first defaulting is evidence-backed and reproducible.
- [ ] NATS remains validated as portability fallback with clear trigger conditions.
- [ ] `index_video` / `get_index_job` contracts remain unchanged and backward compatible.
- [ ] Runbook + security notes are sufficient for operator handoff in Cloud Run and Hetzner-aligned environments.
- [ ] Rollback path can be executed by config-only change and passes existing integration expectations.

### Testing (High-Rigor)

- [ ] Add focused tests for newly covered failure/edge semantics introduced by this phase.
- [ ] Run targeted integration coverage for split-role queue lifecycle + rollback verification.
- [ ] Run `make build`, `make test`, `make integration-test`, and relevant Docker build targets.

### Implementation Plan

- [ ] Define benchmark scenarios and evidence capture format.
- [ ] Implement/extend tests for failure semantics + rollback expectations.
- [ ] Draft/update operational and security runbook artifacts.
- [ ] Refresh checkpoint recommendation and fallback trigger criteria from measured evidence.
- [ ] Update lessons and complete DoD for archive readiness.

## Definition of Done Template

Use this checklist at the top of each new phase/feature todo.

- [ ] Scope is explicit (in/out) and aligned with current architecture decisions.
- [ ] Required interfaces/contracts are updated without breaking existing MCP surface.
- [ ] Fast tests pass (`make test`).
- [ ] Integration tests pass (`make integration-test`).
- [ ] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [ ] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [ ] Todo is fully checked and ready for archive.
