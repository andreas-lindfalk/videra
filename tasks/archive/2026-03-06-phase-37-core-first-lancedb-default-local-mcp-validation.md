# Phase 37 — Core-First LanceDB Default + Local MCP Validation

Archived: 2026-03-06
Status at archive: COMPLETE

## Why this was next

- We postponed Hetzner/Cloud Run parity execution until core local behavior is validated end-to-end.
- We wanted LanceDB to be the default backend path while keeping Chromem available as explicit fallback.
- We needed a simple local workflow to index a folder of videos and interact with the running server via MCP.

## Primary artifacts

- `internal/config/config.go`
- `internal/config/config_test.go`
- `docker-compose.yml`
- `Makefile`
- `cmd/localindex/main.go`
- `README.md`
- `tasks/platform/env-contract.md`
- `tasks/platform/cloud-run-runbook.md`
- `tasks/platform/hetzner-vm-docker-runbook.md`
- `tasks/lessons.md`
- `tasks/todo.md`

## Progress snapshot

- Switched default storage backend to `lancedb` in runtime config and documented it.
- Aligned local Docker/compose and make defaults to `runtime-lancedb-native` with `linux/amd64` platform defaults.
- Added `local-index-folder` make target and `cmd/localindex` helper for structured folder indexing via MCP `index_video`.
- Corrected local path contract to `/videos/<file>` and updated docs/runbooks accordingly.
- Captured a regression lesson about immediate formatting/compile validation and safer Make recipe patterns.
- Validation completed:
  - config tests passed (`runTests` on `internal/config/config_test.go`)
  - `make -n local-index-folder` parse check passed
  - `go test ./cmd/localindex` compile check passed

## Execution plan (completed)

- [x] Switch default runtime/backend path to LanceDB for local operation (Chromem retained as non-default fallback).
- [x] Align local Docker/compose startup defaults with LanceDB native runtime requirements.
- [x] Add a structured local folder indexing workflow for quick MCP-ready validation.
- [x] Update docs/contracts to reflect the new default and local workflow.
- [x] Run focused verification for config + local workflow + LanceDB lane.
