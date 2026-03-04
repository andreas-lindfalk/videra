# Todo

## Active Task

### Videra Phase 13 — Queue Go/No-Go Readiness (Redis-First Rollout)

Reference:

- `tasks/platform/queue-vendor-checkpoint.md`
- `tasks/platform/env-contract.md`
- `README.md` (async indexing + role split contract)
- `AGENTS.md` (Cloud Run + Hetzner parity and portability constraints)

### Definition of Done (Applied)

- [x] Scope is explicit (in/out) and aligned with current architecture decisions.
- [x] Required interfaces/contracts are updated without breaking existing MCP surface.
- [x] Fast tests pass (`make test`).
- [x] Integration tests pass (`make integration-test`) for changed behavior.
- [x] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [x] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [x] Todo is fully checked and ready for archive.

### Scope

- [x] **In scope:** Produce go/no-go evidence for selecting Redis Streams as first external queue backend under the existing portability model.
- [x] **In scope:** Add benchmark/profiling evidence for async indexing lifecycle in local Docker and private VM-like flow.
- [x] **In scope:** Add failure-path verification for retry exhaustion and poison-job terminal handling under split role mode.
- [x] **In scope:** Add operational + security runbooks for Redis-first deployment, with NATS fallback trigger criteria.
- [x] **Out of scope:** Autoscaling algorithm tuning, managed queue vendor integration, and new MCP tool contract changes.

### Go/No-Go Gate (Required Before External Backend Defaulting)

- [x] Candidate benchmarked with reproducible commands and recorded baseline metrics.
- [x] Failure semantics validated with deterministic tests (retry budget, terminal fail path, duplicate/idempotency behavior).
- [x] Operational runbook exists for single-node and HA-oriented deployment path.
- [x] Security model is documented (auth/TLS/secrets/audit responsibilities at runtime edge vs service internals).
- [x] Rollback drill to `inprocess` mode is documented and verified with no MCP API changes.

### Deliverables

- [x] Add/update benchmark artifact doc under `tasks/platform/` with reproducible command set and evidence format.
- [x] Add/update queue ops runbook(s) for Redis-first and NATS fallback deployment in local/private environments.
- [x] Add explicit rollback trigger conditions and fallback procedure in queue checkpoint docs.
- [x] Add deterministic integration coverage for retry exhaustion/terminal handling in split-role mode.
- [x] Update `README.md`, `tasks/platform/env-contract.md`, and `tasks/platform/queue-vendor-checkpoint.md` with finalized guidance.

### Acceptance Criteria

- [x] Recommendation for Redis-first defaulting is evidence-backed and reproducible.
- [x] NATS remains validated as portability fallback with clear trigger conditions.
- [x] `index_video` / `get_index_job` contracts remain unchanged and backward compatible.
- [x] Runbook + security notes are sufficient for operator handoff in Cloud Run and Hetzner-aligned environments.
- [x] Rollback path can be executed by config-only change and passes existing integration expectations.

### Testing (High-Rigor)

- [x] Add focused tests for newly covered failure/edge semantics introduced by this phase.
- [x] Run targeted integration coverage for split-role queue lifecycle + rollback verification.
- [x] Run `make build`, `make test`, `make integration-test`, and relevant Docker build targets.

### Implementation Plan

- [x] Define benchmark scenarios and evidence capture format.
- [x] Implement/extend tests for failure semantics + rollback expectations.
- [x] Draft/update operational and security runbook artifacts.
- [x] Refresh checkpoint recommendation and fallback trigger criteria from measured evidence.
- [x] Update lessons and complete DoD for archive readiness.
