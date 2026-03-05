# Videra Phase 21 — Optional Domain Canonical Mapping (Archived 2026-03-05)

Status: GO

Reference:

- `internal/embedding/text.go`
- `internal/mcpserver/server.go`
- `internal/config/config.go`
- `README.md`
- `tasks/platform/env-contract.md`

### Definition of Done (Target)

- [x] Default normalization is domain-neutral (no hardcoded topic synonym map in core path).
- [x] Domain canonical mapping is optional and runtime-configurable.
- [x] Embedding and reranking normalization use the same optional mapping source.
- [x] Existing MCP contract remains unchanged.
- [x] Fast tests pass (`make test` or focused equivalent).
- [x] Todo is archive-ready with clear implementation notes.

### Scope

- [x] **In scope:** add optional canonical token mapping config and wire through embedding + reranking.
- [x] **In scope:** keep default behavior neutral for arbitrary domains.
- [x] **In scope:** update docs for any env contract changes.
- [x] **Out of scope:** storage backend changes, MCP tool schema changes, queue/runtime architecture changes.

### Implementation Plan

- [x] Add config parsing for optional canonical token mapping.
- [x] Refactor embedder normalization to use optional mapping (empty by default).
- [x] Refactor rerank lexical normalization to use optional mapping from runtime config.
- [x] Update/add tests and docs; run focused validations.
