# Storage Benchmark Evidence — 2026-03-05

Date: 2026-03-05
Phase: 31 (Storage Benchmark Harness)
Branch: main
Operator: GitHub Copilot
Environment: Darwin arm64, go version go1.25.4 darwin/arm64

## Command

```bash
make storage-benchmark-gate
```

## Raw Output Location

- output file: `/tmp/videra_phase31_storage_benchmark_gate.out`
- exit code file: `/tmp/videra_phase31_storage_benchmark_gate.exit`
- exit code: `0`

## Benchmark Summary

- `IndexVideo_8Segments`: `1338738 ns/op`, `6556715 B/op`, `646 allocs/op`
- `SearchSegments_Top5_Corpus200x8`: `470725 ns/op`, `20103 B/op`, `92 allocs/op`
- `ListVideos_Corpus200`: `27718 ns/op`, `24776 B/op`, `4 allocs/op`
- `GetTranscript_8Segments`: `165.8 ns/op`, `896 B/op`, `1 allocs/op`
- `Reset`: `244824 ns/op`, `818117 B/op`, `64 allocs/op`

## Decision Notes

- Baseline captured for future backend comparison: yes
- Any anomalous run characteristics: none
- Criteria impact:
  - measured benefit baseline (criterion 1): pass
  - compatibility-plan input quality (criterion 3): pass
  - gate-parity support signal (criterion 5): pass

## Decision

- GO (benchmark harness readiness)
- rationale: repeatable benchmark command passed and produced stable operation-level metrics across index/search/list/transcript/reset workloads.
