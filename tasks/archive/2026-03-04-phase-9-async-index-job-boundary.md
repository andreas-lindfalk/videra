# Todo

## Active Task

### Videra Phase 9 — Async Index Job Boundary (Cloud-Ready)

Reference:

- `AGENTS.md` (Cloud Run scaling path + orchestrator boundary)

### Definition of Done (Applied)

- [x] Scope is explicit (in/out) and aligned with current architecture decisions.
- [x] Required interfaces/contracts are updated without breaking existing MCP surface.
- [x] Fast tests pass (`make test`).
- [x] Integration tests pass (`make integration-test`).
- [x] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [x] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [x] Todo is fully checked and ready for archive.

### Scope

- [x] **In scope:** Introduce explicit async indexing job flow boundaries while keeping existing synchronous behavior as default.
- [x] **In scope:** Expose deterministic job lifecycle semantics (`pending`, `completed`, `failed`) for async indexing requests.
- [x] **In scope:** Keep cloud-provider neutrality (Cloud Run + Hetzner) in core runtime logic.
- [x] **Out of scope:** Concrete broker integrations (Pub/Sub/SQS/NATS), GPU worker rollout, and full distributed scheduler implementation.

### Deliverables

- [x] Define stable request/response contract for async `index_video` initiation (job-oriented output).
- [x] Add a job-status retrieval surface (tool or equivalent MCP-accessible contract) with explicit status/error semantics.
- [x] Implement bounded in-process job execution path behind existing orchestrator abstraction.
- [x] Preserve idempotent source behavior for both sync and async index paths.
- [x] Update runbooks/docs for async invocation and verification flow.

### Acceptance Criteria

- [x] Existing sync `index_video` behavior remains backward compatible by default.
- [x] Async initiation returns deterministic job metadata (`jobId`, status) and does not block on full indexing completion.
- [x] Job status polling yields explicit terminal states and meaningful error messages.
- [x] Retrieval/search contracts remain unchanged after successful async indexing.

### Testing (High-Rigor)

- [x] Add focused unit tests for orchestrator async lifecycle transitions and error propagation.
- [x] Add integration coverage for async happy path (initiate → poll complete → search/list/transcript verification).
- [x] Add integration coverage for at least one async failure path (invalid source or forced ingestion failure).
- [x] Re-run deterministic ordering checks to ensure no regression from async path introduction.

### Execution Notes

- [x] Keep async boundaries interface-first so external queue backends can be added later without MCP contract breakage.
- [x] Keep test fixtures deterministic and avoid timing-flaky assertions in polling tests.
- [x] Ensure runtime/log visibility is explicit for job start, completion, and failure transitions.

### Future Broker Direction (Post-Phase 9)

- [ ] Keep in-process execution as baseline/default for local and private self-hosted simplicity.
- [ ] Design a broker-agnostic `JobQueue` interface so backend choice is deployment concern, not MCP contract concern.
- [ ] Evaluate **NATS (JetStream)** first as the primary neutral broker candidate for private/on-prem setups.
- [ ] Define acceptance criteria for broker adoption: easy local bootstrap, HA option, clear retry/ack semantics, and no cloud lock-in.
- [ ] Add a mandatory vendor-choice checkpoint doc before implementation (candidate matrix, lock-in risks, and private-deployment fallback plan).

### Implementation Plan

- [x] Finalize MCP contract shape for async initiation and status retrieval.
- [x] Implement orchestrator/job store changes with deterministic state transitions.
- [x] Wire MCP handlers and preserve sync default behavior.
- [x] Add unit + integration tests for async lifecycle and failure semantics.
- [x] Update docs/runbooks/env-contract where needed and run full verification suite.
