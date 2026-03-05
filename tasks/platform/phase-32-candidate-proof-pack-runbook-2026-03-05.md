# Phase 32 Candidate Proof Pack Runbook — 2026-03-05

Purpose: run prerequisites 1, 4, and 5 in one candidate-backend pass once candidate mode exists.

Status: active (candidate backend mode implemented).

## One-Pass Command

```bash
make phase32-candidate-proof-pack BACKEND=<candidate-backend> PREFIX=/tmp/videra_phase32_candidate
```

## One-Pass Summary

```bash
make phase32-candidate-proof-pack-summarize PREFIX=/tmp/videra_phase32_candidate
```

## Generated Artifacts

- `/tmp/videra_phase32_candidate_prereq1_benchmark.out`
- `/tmp/videra_phase32_candidate_prereq1_benchmark.exit`
- `/tmp/videra_phase32_candidate_prereq4_rollback.out`
- `/tmp/videra_phase32_candidate_prereq4_rollback.exit`
- `/tmp/videra_phase32_candidate_prereq5_gate_parity.out`
- `/tmp/videra_phase32_candidate_prereq5_gate_parity.exit`

## Evidence Update Targets

- `tasks/platform/phase-32-prereq-1-measured-benefit-evidence-2026-03-05.md`
- `tasks/platform/phase-32-prereq-4-rollback-proof-evidence-2026-03-05.md`
- `tasks/platform/phase-32-prereq-5-gate-parity-migration-mode-evidence-2026-03-05.md`

## Guardrails

- `BACKEND` must be set and must not be `chromem`.
- Candidate backend must be runtime-supported by `VIDERA_STORAGE_BACKEND` contract (`chromem|lancedb`).
- Do not mark prerequisites 1/4/5 as pass unless candidate run exits are all `0` and evidence fields are fully populated.

Failure semantics:

- exit `1`: invalid user input (missing backend or baseline backend specified).
- non-zero from sub-gates: candidate execution failed in benchmark, rollback rehearsal, or gate parity checks.
