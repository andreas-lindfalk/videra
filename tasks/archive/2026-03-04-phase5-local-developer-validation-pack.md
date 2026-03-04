# Todo Archive

## Videra Phase 5 — Local Developer Validation Pack

### Definition of Done (Applied)

- [x] Scope is explicit (in/out) and aligned with current architecture decisions.
- [x] Required interfaces/contracts are updated without breaking existing MCP surface.
- [x] Fast tests pass (`make test`).
- [x] Integration tests pass (`make integration-test`).
- [x] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [x] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [x] Todo is fully checked and ready for archive.

### Scope

- [x] **In scope:** Local-first, low-friction validation workflow for running Videra, indexing a local video, querying via MCP, and connecting from developer tooling before Cloud Run.
- [x] **Out of scope:** Production cloud deployment hardening and full enterprise auth implementation.

### Deliverables

- [x] Create one-command local stack run/stop tasks for HTTP MCP endpoint.
- [x] Add a local smoke-test CLI flow (index local file, search query, list videos, read transcript).
- [x] Add Copilot/MCP setup quickstart for local usage (stdio + HTTP options).
- [x] Add troubleshooting guide for common local failures.

### Product/Tech Work

- [x] Add `cmd/localsmoke` executable for deterministic local MCP checks.
- [x] Add make targets for local up/down/smoke workflows.
- [x] Keep local workflow compatible with existing MCP schema and runtime flags.

### Testing (High-Rigor)

- [x] Add at least one integration test path that uses a local-file-like flow contract (path handling + search + transcript read).
- [x] Validate no regressions in existing integration suite.

### Verification

- [x] `make build` passes.
- [x] `make test` passes.
- [x] `make integration-test` passes.
- [x] `make docker-build` passes.
