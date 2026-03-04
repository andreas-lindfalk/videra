# Todo

## Active Task

### Videra Phase 15 — Split-Role Data-Plane Parity (Shared Index Visibility)

Reference:

- `AGENTS.md` (storage portability + Cloud Run/Hetzner parity constraints)
- `tasks/platform/queue-vendor-checkpoint.md`
- `tasks/platform/queue-redis-first-runbook.md`
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

- [x] **In scope:** Close the split-role data-plane gap so API query/list/read paths can reflect worker-indexed data under an explicitly shared storage topology.
- [x] **In scope:** Add clear runtime contract for shared storage usage in split `api|worker` mode (local/private first), preserving current queue role semantics.
- [x] **In scope:** Add deterministic integration evidence for split-role indexed-data visibility after async completion.
- [x] **In scope:** Update operator docs with exact shared-storage deployment requirements, limitations, and fallback behavior.
- [x] **Out of scope:** Managed cloud vector database migration and changes to MCP tool schemas/contracts.

### Data-Plane Readiness Gate

- [x] Split-role async success path proves API can `list_videos` and `search_video` for worker-produced content under documented shared storage configuration.
- [x] Misconfiguration paths are explicit (actionable error or clear degraded-mode guidance) when shared storage prerequisites are not met.
- [x] `index_video` / `get_index_job` behavior remains backward compatible while visibility guarantees are strengthened.
- [x] Rollback to current all-in-one local mode remains config-only.

### Deliverables

- [x] Add/adjust storage/runtime wiring needed for explicit shared data-plane operation in split-role deployments.
- [x] Add integration test coverage for split-role shared-storage visibility (`index_video` async -> `get_index_job` completed -> `list_videos`/`search_video` visibility).
- [x] Add integration check for non-shared topology behavior semantics (clear operator signal, no silent ambiguity).
- [x] Update `README.md`, `tasks/platform/env-contract.md`, and runbook/checkpoint docs with shared-storage contract and fallback guidance.

### Acceptance Criteria

- [x] In documented split-role shared-storage mode, API-visible retrieval behavior matches worker indexing outcomes deterministically.
- [x] Existing clients remain fully compatible (`index_video`, `get_index_job`, `search_video`, `list_videos` schemas unchanged).
- [x] Operators can reliably distinguish shared-storage-correct mode from degraded/misaligned mode.
- [x] Deployment guidance is copy/paste-ready for local/private (Hetzner-like) paths and clearly states current cloud limitations.

### Testing (High-Rigor)

- [x] Add focused unit tests for any new storage/runtime guardrails.
- [x] Add/extend integration tests for split-role data-plane visibility and degraded-mode semantics.
- [x] Run `make build`, `make test`, `make integration-test`, and relevant Docker build targets.

### Implementation Plan

- [x] Define and implement explicit shared-storage runtime contract for split-role data-plane consistency.
- [x] Wire runtime/storage behavior to honor the contract without altering MCP API shapes.
- [x] Implement deterministic split-role integration coverage for visibility success + degraded/misaligned path.
- [x] Refresh docs/runbooks/checkpoint artifacts to reflect the finalized contract and operator steps.
- [x] Update lessons and complete DoD for archive readiness.
