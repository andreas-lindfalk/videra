# Phase 32 Prerequisite 5 — Gate Parity in Migration Mode Evidence (Stub)

Date: 2026-03-05
Status: pass (migration-mode gate-parity capture completed)
Owner: Andreas (QA/Release lead, proposed)
Priority: P1
Target date (suggested): 2026-03-11

Goal: prove all required gates remain green with migration mode enabled.

## Required Commands

```bash
make release-gate
make release-gate-split
make pilot-quality-gate
make real-corpus-promotion-gate
```

## Results Matrix (Fill)

| Command | Exit code | Result | Evidence path |
|---|---:|---|---|
| `make release-gate` | 0 (candidate capture) | pass | `/tmp/videra_phase32_candidate_20260305_r3_prereq5_gate_parity.out` |
| `make release-gate-split` | 0 (candidate capture) | pass | `/tmp/videra_phase32_candidate_20260305_r3_prereq5_gate_parity.out` |
| `make pilot-quality-gate` | 0 (candidate capture) | pass | `/tmp/videra_phase32_candidate_20260305_r3_prereq5_gate_parity.out` |
| `make real-corpus-promotion-gate` | 0 (candidate capture) | pass | `/tmp/videra_phase32_candidate_20260305_r3_prereq5_gate_parity.out` |

## Quality Signal Snapshot (Fill)

- evidenceMatchRate: 1.00 (candidate capture)
- deterministicRate: 1.00 (candidate capture)
- topTwoQualityRate: 1.00 (candidate capture)
- real-mode guardrails: pass (candidate capture)

## Command Evidence

- capture command:
	- `make gate-parity-capture BACKEND=lancedb OUT=/tmp/videra_phase32_candidate_20260305_r3_prereq5_gate_parity.out EXIT_OUT=/tmp/videra_phase32_candidate_20260305_r3_prereq5_gate_parity.exit`
- summarize command:
	- `make gate-parity-summarize OUT=/tmp/videra_phase32_candidate_20260305_r3_prereq5_gate_parity.out`
- output files:
	- `/tmp/videra_phase32_candidate_20260305_r3_prereq5_gate_parity.out`
	- `/tmp/videra_phase32_candidate_20260305_r3_prereq5_gate_parity.exit`
- overall exit code: `0`

Transient note:

- Initial candidate parity capture under `/tmp/videra_phase32_candidate_20260305_r2_prereq5_gate_parity.out` failed due testcontainer startup timeout (`context deadline exceeded`) in one run of `TestIndexVideoRealModeRequiresSidecarForLocalPath`.
- Immediate rerun (`r3`) passed all four gates with exit markers at `0`.

## Candidate Exit Markers

- `release_gate_exit=0`
- `release_gate_split_exit=0`
- `pilot_quality_gate_exit=0`
- `real_corpus_promotion_gate_exit=0`

## Decision

- prerequisite status: pass
- rationale: all required promotion gates were re-run in candidate mode (`VIDERA_STORAGE_BACKEND=lancedb`) and completed green.

## Handoff Notes

- Execute last, after prerequisites 1 through 4 are completed.
- Execution runbook: `tasks/platform/phase-32-prereq-5-execution-runbook-2026-03-05.md`
