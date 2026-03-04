# Archive — Videra Phase 3 (Productionization Track A)

Archived: 2026-03-04
Source: tasks/todo.md

# Todo

## Videra Phase 3 — Productionization Track A (Retrieval Quality + Cloud Boundaries)

### Definition of Done (Applied)

- [x] Scope is explicit (in/out) and aligned with current architecture decisions.
- [x] Required interfaces/contracts are updated without breaking existing MCP surface.
- [x] Fast tests pass (`make test`).
- [x] Integration tests pass (`make integration-test`).
- [x] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [x] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [x] Todo is fully checked and ready for archive.

### Scope

- [x] **In scope:** Real embedding adapter boundaries, improved hybrid ranking behavior, and cloud-job-ready indexing boundaries.
- [x] **Out of scope:** Full production ML model deployment and full AgentGateway auth implementation.

### Architecture & Interfaces

- [x] Add explicit embedding provider interfaces for transcript and visual paths with swappable implementations.
- [x] Separate synchronous MCP request path from async indexing job orchestration boundary.
- [x] Preserve transport/auth separation so AgentGateway remains an edge concern.

### Retrieval Quality Improvements

- [x] Replace deterministic query embedding fallback with provider-driven embedding adapter path.
- [x] Improve hybrid reranking logic with deterministic tie-break rules and modality weighting config.
- [x] Ensure `search_video` returns stable, backward-compatible response schema.

### Cloud-Ready Indexing Boundaries

- [x] Introduce indexing job input/output contract for future Cloud Run Jobs.
- [x] Ensure indexing operations are idempotent by source identity and segment keying.
- [x] Add retry-safe persistence behavior for partial failures.

### Testing (High-Rigor)

- [x] Add integration tests for ranking determinism (same query + same data => same ordered results).
- [x] Add integration tests for modality weighting behavior (audio-prioritized vs visual-prioritized scenarios).
- [x] Add integration tests for idempotent re-index with partial-failure retry path.
- [x] Add malformed payload tests for `search_video` arguments and schema validation.
- [x] Add regression tests for backward-compat response fields in `index_video`, `search_video`, and `list_videos`.

### Verification

- [x] `make build` passes after interface and pipeline changes.
- [x] `make test` passes with added deterministic assertions.
- [x] `make integration-test` passes with new ranking/idempotency/failure-path coverage.
- [x] `make docker-build` passes with unchanged runtime behavior.

## Definition of Done Template

Use this checklist at the top of each new phase/feature todo.

- [ ] Scope is explicit (in/out) and aligned with current architecture decisions.
- [ ] Required interfaces/contracts are updated without breaking existing MCP surface.
- [ ] Fast tests pass (`make test`).
- [ ] Integration tests pass (`make integration-test`).
- [ ] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [ ] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [ ] Todo is fully checked and ready for archive.
