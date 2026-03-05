# Storage Benchmark Harness — 2026-03-05

Purpose: provide a repeatable storage baseline benchmark command that can be reused in future backend comparison checkpoints.

Status: active (Phase 31).

## Canonical Command

Run from repo root:

```bash
make storage-benchmark-gate
```

Underlying command:

```bash
go test ./internal/storage -run '^$' -bench 'BenchmarkChromemStoreBaseline' -benchmem -count=1
```

## Benchmark Suite Scope

The benchmark suite is implemented in:

- `internal/storage/chromem_benchmark_test.go`

Current sub-benchmarks:

- `IndexVideo_8Segments`
- `SearchSegments_Top5_Corpus200x8`
- `ListVideos_Corpus200`
- `GetTranscript_8Segments`
- `Reset`

## Intended Decision Mapping (Phase 29 GO Criteria)

This harness contributes evidence to:

1. **Measured benefit** (criterion 1): establishes the current `chromem-go` baseline so alternative backends can be compared using the same benchmark names and workloads.
2. **Compatibility plan inputs** (criterion 3): keeps workload semantics explicit and stable for future dual-backend validation.
3. **Gate parity support** (criterion 5): provides an additional repeatable signal alongside release/promotion gates.

This harness does **not** by itself satisfy runtime-contract readiness or rollback proof criteria.

## Output Expectations

A valid benchmark run must include:

- exit code `0`,
- one output line per benchmark sub-case with `ns/op`,
- allocation metrics (`B/op`, `allocs/op`).

## Evidence Workflow

For each benchmark run:

1. execute `make storage-benchmark-gate`,
2. capture raw output,
3. summarize results using:
   - `tasks/platform/storage-benchmark-evidence-template.md`
4. store dated evidence as:
   - `tasks/platform/storage-benchmark-evidence-YYYY-MM-DD.md`

## Related Artifacts

- `tasks/platform/storage-decision-recheckpoint-2026-03-05.md`
- `tasks/platform/storage-decision-recheckpoint-evidence-2026-03-05.md`
- `tasks/platform/roadmap-end-state-2026-03-05.md`
