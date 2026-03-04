# Todo

## Active Task

### Videra Phase 11 — Queue Adapter Spike (NATS vs Redis)

Reference:

- `tasks/platform/queue-vendor-checkpoint.md`
- `tasks/platform/jobqueue-interface-proposal.md`
- `AGENTS.md` (queue portability + vendor checkpoint guardrails)

### Definition of Done (Applied)

- [x] Scope is explicit (in/out) and aligned with current architecture decisions.
- [x] Required interfaces/contracts are updated without breaking existing MCP surface.
- [x] Fast tests pass (`make test`).
- [x] Integration tests pass (`make integration-test`) for changed behavior.
- [x] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [x] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [x] Todo is fully checked and ready for archive.

### Scope

- [x] **In scope:** Implement a broker-agnostic `JobQueue` interface in code with an in-process adapter.
- [x] **In scope:** Add a minimal adapter spike for **NATS JetStream** and a comparison spike for **Redis Streams**.
- [x] **In scope:** Add contract tests that both adapters must pass (`enqueue`, `reserve`, `ack`, `retry`, `fail`, idempotency-safe duplicate handling).
- [x] **In scope:** Keep `index_video` and `get_index_job` MCP contracts unchanged.
- [x] **Out of scope:** Production-grade distributed worker rollout, autoscaling policy tuning, and migration to external queue as default.

### Go/No-Go Gate (Required Before Implementation)

- [x] Validate Phase 10 checklist items are explicitly satisfied for the adapters under test.
- [x] Confirm rollback path to in-process mode with zero MCP contract changes.
- [x] Confirm lock-in analysis and fallback criteria remain documented and current.

### Deliverables

- [x] Add `internal/ingestion/jobqueue.go` (or equivalent) with `JobQueue` abstraction.
- [x] Add in-process adapter implementation and wiring behind orchestrator boundaries.
- [x] Add NATS JetStream spike adapter behind build/runtime flag (non-default).
- [x] Add Redis Streams comparison adapter behind build/runtime flag (non-default).
- [x] Add adapter contract test suite shared by all queue implementations.
- [x] Add implementation notes/results to `tasks/platform/queue-vendor-checkpoint.md`.

### Acceptance Criteria

- [x] MCP API compatibility remains unchanged (existing integration tests continue to pass).
- [x] In-process mode remains default and stable.
- [x] NATS and Redis spikes both satisfy baseline queue contract tests.
- [x] Documented recommendation is updated with measured evidence from the spike.

### Testing (High-Rigor)

- [x] Add focused unit tests for queue lifecycle and failure semantics.
- [x] Add deterministic integration coverage proving MCP behavior parity under in-process queue mode.
- [x] Run `make test`, `make integration-test`, `make build`, and relevant Docker build target(s).

### Implementation Plan

- [x] Implement `JobQueue` abstraction and in-process adapter first.
- [x] Build shared adapter contract tests.
- [x] Implement NATS JetStream spike adapter.
- [x] Implement Redis Streams comparison adapter.
- [x] Execute test matrix and update recommendation docs + lessons.
