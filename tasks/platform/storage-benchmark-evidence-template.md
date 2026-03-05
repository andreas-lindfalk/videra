# Storage Benchmark Evidence Template

Date: <YYYY-MM-DD>
Operator: <name>
Branch/Commit: <branch>/<sha>
Environment: <os/arch + go version>

## Command

```bash
make storage-benchmark-gate
```

## Raw Output Location

- output file: <path>
- exit code: <0/non-zero>

## Benchmark Summary

- IndexVideo_8Segments: <ns/op>, <B/op>, <allocs/op>
- SearchSegments_Top5_Corpus200x8: <ns/op>, <B/op>, <allocs/op>
- ListVideos_Corpus200: <ns/op>, <B/op>, <allocs/op>
- GetTranscript_8Segments: <ns/op>, <B/op>, <allocs/op>
- Reset: <ns/op>, <B/op>, <allocs/op>

## Decision Notes

- Baseline captured for future backend comparison: yes/no
- Any anomalous run characteristics: <none/details>
- Criteria impact:
  - measured benefit baseline (criterion 1): pass/fail
  - compatibility-plan input quality (criterion 3): pass/fail
  - gate-parity support signal (criterion 5): pass/fail

## Decision

- GO / NO-GO (for benchmark harness readiness)
- rationale:
