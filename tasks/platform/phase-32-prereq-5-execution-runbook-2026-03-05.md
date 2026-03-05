# Phase 32 Prerequisite 5 — Execution Runbook (Gate Parity)

Date: 2026-03-05
Scope: repeatable gate-parity capture for baseline and candidate backend modes.

## Commands

Baseline capture (stable backend):

```bash
make gate-parity-capture BACKEND=chromem OUT=/tmp/videra_phase32_prereq5_gate_parity_baseline.out EXIT_OUT=/tmp/videra_phase32_prereq5_gate_parity_baseline.exit
make gate-parity-summarize OUT=/tmp/videra_phase32_prereq5_gate_parity_baseline.out
```

Candidate capture (when candidate backend mode exists):

```bash
make gate-parity-capture BACKEND=<candidate> OUT=/tmp/videra_phase32_prereq5_gate_parity_candidate.out EXIT_OUT=/tmp/videra_phase32_prereq5_gate_parity_candidate.exit
make gate-parity-summarize OUT=/tmp/videra_phase32_prereq5_gate_parity_candidate.out
```

## What It Captures

1. `make release-gate`
2. `make release-gate-split`
3. `make pilot-quality-gate`
4. `make real-corpus-promotion-gate`

Output includes per-step exit markers:

- `release_gate_exit`
- `release_gate_split_exit`
- `pilot_quality_gate_exit`
- `real_corpus_promotion_gate_exit`

## Evidence Update Path

Update:

- `tasks/platform/phase-32-prereq-5-gate-parity-migration-mode-evidence-2026-03-05.md`

Required fields to fill:

- results matrix with capture output paths
- quality signal snapshot for candidate mode
- prerequisite decision status

## Decision Rule

Mark prerequisite 5 as pass only when all four commands exit `0` in candidate backend mode and quality metrics remain within accepted thresholds.

Related one-pass flow:

- `tasks/platform/phase-32-candidate-proof-pack-runbook-2026-03-05.md`
