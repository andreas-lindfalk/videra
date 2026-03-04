# Todo

## Active Task

### Videra Phase 8 — Cloud-Ready Ingestion Parity

### Definition of Done (Applied)

- [x] Scope is explicit (in/out) and aligned with current architecture decisions.
- [x] Required interfaces/contracts are updated without breaking existing MCP surface.
- [x] Fast tests pass (`make test`).
- [x] Integration tests pass (`make integration-test`).
- [x] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [x] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [x] Todo is fully checked and ready for archive.

### Scope

- [x] **In scope:** Introduce cloud-ready media ingestion input support so indexing works without server-local file paths.
- [x] **In scope:** Keep MCP contracts stable (`index_video`, `search_video`, `list_videos`, transcript resource) while extending accepted source forms.
- [x] **In scope:** Preserve local-first workflows (`/videos/...` path indexing remains supported).
- [x] **Out of scope:** Full distributed job queue implementation, GPU inference rollout, and provider-specific lock-in logic.

### Deliverables

- [x] Define ingestion source contract for local path + remote URL/object-storage-backed source.
- [x] Implement bounded remote media fetch/prepare stage with explicit timeout/size limits and clear errors.
- [x] Add config/env contract for remote ingestion controls (timeouts, max size, allow/deny behavior).
- [x] Keep idempotent source identity behavior for retry-safe indexing across local and remote sources.
- [x] Update Cloud Run and Hetzner runbooks with the new indexing path and parity expectations.

### Acceptance Criteria

- [x] `index_video` can index at least one remote/cloud-reachable media source in integration coverage.
- [x] Existing local path indexing behavior remains unchanged.
- [x] Failure paths (network timeout, oversized payload, invalid source) produce explicit non-ambiguous errors.
- [x] Deterministic retrieval ordering and existing response compatibility remain green.

### Testing (High-Rigor)

- [x] Add focused tests for source contract parsing and validation.
- [x] Add integration coverage for remote-source indexing happy path and at least one bounded-failure path.
- [x] Re-run parity checklist assumptions and remove current Cloud Run indexing blocker where satisfied.

### Execution Notes

- [x] Keep storage/ingestion interfaces backend-agnostic and cloud-provider-neutral.
- [x] Avoid silent fallback behavior; capability and failure mode must be visible in logs/tool errors.

## Definition of Done Template

Use this checklist at the top of each new phase/feature todo.

- [ ] Scope is explicit (in/out) and aligned with current architecture decisions.
- [ ] Required interfaces/contracts are updated without breaking existing MCP surface.
- [ ] Fast tests pass (`make test`).
- [ ] Integration tests pass (`make integration-test`).
- [ ] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [ ] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [ ] Todo is fully checked and ready for archive.
