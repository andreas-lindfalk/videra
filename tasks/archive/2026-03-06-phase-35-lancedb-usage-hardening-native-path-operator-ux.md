# Phase 35 — LanceDB Usage Hardening (Native Path + Operator UX)

Archived: 2026-03-06
Status at archive: COMPLETE

## Why this was next

- This phase needed a production-grade LanceDB workflow, not just local code-path support.
- New contributors should keep zero-dependency defaults while operators can reliably enable native LanceDB in Docker/CI.

## Primary artifacts

- `internal/storage/lancedb.go`
- `internal/storage/lancedb_native_bridge.go`
- `internal/storage/lancedb_native_bridge_stub.go`
- `internal/storage/lancedb_native_bridge_stub_test.go`
- `internal/storage/lancedb_test.go`
- `internal/config/config.go`
- `internal/config/config_test.go`
- `Dockerfile`
- `Makefile`
- `README.md`
- `tasks/platform/env-contract.md`

## Progress snapshot

- Native LanceDB was isolated behind build tag `lancedb_native` so default builds remain zero-dependency.
- `VIDERA_LANCEDB_REGION` semantics are cloud-only (required only for `db://` URIs), avoiding local default noise.
- Added first-class native operator path (`runtime-lancedb-native` Docker target + native Make commands).
- Added explicit non-native coverage (`lancedb_native`-off guidance error test).
- Validation completed:
	- focused storage/config tests passed (`go test` via targeted suite)
	- full fast suite passed (`go test ./...`)
	- LanceDB-focused integration target passed on current host (`make integration-test-lancedb-native`), with native-index/search intentionally skipped on non-amd64 runners.

## Execution plan (completed)

- [x] Add first-class native LanceDB Docker/Make targets (build + run) with deterministic artifact download and build-tag wiring.
- [x] Add explicit non-native behavior test coverage (clear guidance error when LanceDB backend is requested without native build tag).
- [x] Tighten docs/runtime contract for operator usage (local default vs native mode in Docker/CI).
- [x] Run focused and full fast test suites; record verification results.
