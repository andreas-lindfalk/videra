# Phase 32 Prerequisite 1 — Measured Backend Benefit Evidence (Stub)

Date: 2026-03-05
Status: pending (candidate comparison captured; material gain criterion not met)
Owner: Andreas (Storage lead, proposed)
Priority: P1
Target date (suggested): 2026-03-09

Goal: prove material backend benefit versus current `chromem-go` baseline.

## Required Inputs

- Baseline benchmark artifact:
  - `tasks/platform/storage-benchmark-evidence-2026-03-05.md`
- Baseline execution runbook:
  - `tasks/platform/phase-32-prereq-1-execution-runbook-2026-03-05.md`
- Comparison run(s) on candidate backend using equivalent benchmark workload names.

## Comparison Table (Fill)

| Benchmark case | Current baseline | Candidate backend | Delta | Material gain? |
|---|---|---|---|---|
| IndexVideo_8Segments | 1349774 ns/op, 6556919 B/op, 646 allocs/op | 1305814 ns/op, 6557518 B/op, 646 allocs/op | -3.26% ns/op (faster), +599 B/op, allocs unchanged | no |
| SearchSegments_Top5_Corpus200x8 | 462544 ns/op, 20118 B/op, 93 allocs/op | 466912 ns/op, 20108 B/op, 92 allocs/op | +0.94% ns/op (slower), -10 B/op, -1 alloc/op | no |
| ListVideos_Corpus200 | 27623 ns/op, 24776 B/op, 4 allocs/op | 28610 ns/op, 24776 B/op, 4 allocs/op | +3.57% ns/op (slower), memory unchanged | no |
| GetTranscript_8Segments | 167.9 ns/op, 896 B/op, 1 allocs/op | 176.4 ns/op, 896 B/op, 1 allocs/op | +5.06% ns/op (slower), memory unchanged | no |
| Reset | 247102 ns/op, 818091 B/op, 64 allocs/op | 251124 ns/op, 818237 B/op, 63 allocs/op | +1.63% ns/op (slower), +146 B/op, -1 alloc/op | no |

## Command Evidence

- baseline command: `make storage-benchmark-gate`
- baseline output file path(s):
  - `/tmp/videra_phase32_prereq1_baseline.out`
- baseline exit code(s): `0`
- baseline exit file: `/tmp/videra_phase32_prereq1_baseline.exit`
- baseline artifact(s):
  - `tasks/platform/storage-benchmark-evidence-2026-03-05.md`
  - `tasks/platform/phase-32-prereq-1-execution-runbook-2026-03-05.md`
- candidate command(s): `make storage-benchmark-capture BACKEND=lancedb OUT=/tmp/videra_phase32_candidate_20260305_r2_prereq1_benchmark.out EXIT_OUT=/tmp/videra_phase32_candidate_20260305_r2_prereq1_benchmark.exit`
- candidate output file path(s):
  - `/tmp/videra_phase32_candidate_20260305_r2_prereq1_benchmark.out`
  - `/tmp/videra_phase32_candidate_20260305_r2_prereq1_benchmark.exit`
- candidate exit code(s): `0`

## Decision

- prerequisite status: pending
- rationale: candidate-mode benchmark capture is complete, but measured deltas do not show a material backend benefit versus baseline.

## Handoff Notes

- Start only after prerequisite 2 and prerequisite 3 are completed.
