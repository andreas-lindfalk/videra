# Phase 36 — Remove LanceDB Native Stub File Pair

Archived: 2026-03-06
Status at archive: COMPLETE

## Why this was next

- We wanted to delete `internal/storage/lancedb_native_bridge_stub.go` and its dedicated test file without breaking non-native default builds.

## Primary artifacts

- `internal/storage/lancedb_native_bridge.go`
- `internal/storage/lancedb_native_bridge_stub.go` (deleted)
- `internal/storage/lancedb_native_bridge_stub_test.go` (deleted)
- `internal/storage/lancedb_native_bridge_factory.go` (new)
- `tasks/todo.md`

## Progress snapshot

- Introduced an untagged native-bridge factory with non-native default guidance behavior.
- Refactored native-tag implementation to register itself into the shared factory.
- Deleted the native stub file and its dedicated test file.
- Validation completed:
  - fast tests passed (`make test`)
  - LanceDB integration target passed (`make integration-test-lancedb-native`)

## Execution plan (completed)

- [x] Introduce an untagged native-bridge factory with non-native default guidance behavior.
- [x] Refactor native-tag implementation to register itself into the factory.
- [x] Delete stub file and its dedicated test file.
- [x] Run fast tests and LanceDB integration target to verify no regressions.
