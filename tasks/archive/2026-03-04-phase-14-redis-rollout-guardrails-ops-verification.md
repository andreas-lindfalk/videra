# Todo

## Active Task

### Videra Phase 14 — Redis-First Rollout Guardrails + Ops Verification

Reference:

- `tasks/platform/queue-vendor-checkpoint.md`
- `tasks/platform/queue-redis-first-runbook.md`
- `tasks/platform/queue-benchmark-evidence.md`
- `tasks/platform/env-contract.md`
- `README.md` (async indexing + split role)

### Definition of Done (Target)

- [x] Scope is explicit (in/out) and aligned with current architecture decisions.
- [x] Required interfaces/contracts are updated without breaking existing MCP surface.
- [x] Fast tests pass (`make test`).
- [x] Integration tests pass (`make integration-test`) for changed behavior.
- [x] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [x] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [x] Todo is fully checked and archive-ready.

### Scope

- [x] **In scope:** Add runtime guardrails that make split-role external queue deployments safer by default (startup validation + explicit config error semantics).
- [x] **In scope:** Add operational observability for async queue lifecycle (enqueue/reserve/retry/terminal completion) through structured logs and deterministic integration assertions.
- [x] **In scope:** Add rollout verification flow for Redis-first API+worker deployment with clear health checks and rollback trigger confirmation.
- [x] **In scope:** Tighten docs so operator steps are copy/paste-ready for local/private environments.
- [x] **Out of scope:** New MCP tools, cloud-vendor managed services, autoscaling policy tuning.

### Rollout Safety Gate

- [x] Misconfiguration paths fail fast with actionable error messages (role/backend/state-store compatibility).
- [x] Async lifecycle emits operator-usable structured events for retries and terminal failures.
- [x] Redis split-role deployment is validated end-to-end with deterministic integration proof.
- [x] Rollback to `inprocess` remains config-only and verified after changes.

### Deliverables

- [x] Add/adjust config validation and startup wiring for stricter external queue deployment safety checks.
- [x] Add structured queue lifecycle logging at orchestrator + adapter boundary (without changing tool responses).
- [x] Add/extend integration coverage for rollout safety scenarios (misconfig fail-fast + retry/terminal observability assertions).
- [x] Update `README.md`, `tasks/platform/env-contract.md`, and runbook/checkpoint docs with finalized rollout procedure.

### Acceptance Criteria

- [x] Existing clients remain fully compatible (`index_video`, `get_index_job`, `search_video` contracts unchanged).
- [x] Operator can diagnose queue lifecycle from logs without reproducing locally.
- [x] Known bad role/backend combinations fail during startup, not at runtime job execution.
- [x] Redis-first rollout and rollback flows are both reproducible from documented commands.

### Testing (High-Rigor)

- [x] Add focused unit tests for new config/startup guardrails.
- [x] Add/extend integration tests for split-role rollout safety and terminal retry observability.
- [x] Run `make build`, `make test`, `make integration-test`, and relevant Docker build targets.

### Implementation Plan

- [x] Design and implement fail-fast runtime validation for queue role/backend/state-store combinations.
- [x] Implement structured async lifecycle logging with stable field keys for operator troubleshooting.
- [x] Extend integration coverage for rollout safety gates and rollback verification.
- [x] Refresh runbook/env/checkpoint docs to match implementation behavior.
- [x] Update lessons and complete DoD for archive readiness.
