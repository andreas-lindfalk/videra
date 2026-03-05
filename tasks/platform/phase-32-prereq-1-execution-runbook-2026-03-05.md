# Phase 32 Prerequisite 1 — Execution Runbook (Measured Benefit)

Date: 2026-03-05
Scope: repeatable benchmark evidence capture for baseline and candidate backend comparison.

## Commands

Baseline capture:

```bash
make storage-benchmark-capture OUT=/tmp/videra_phase32_prereq1_baseline.out EXIT_OUT=/tmp/videra_phase32_prereq1_baseline.exit
make storage-benchmark-summarize OUT=/tmp/videra_phase32_prereq1_baseline.out
```

Candidate capture (when candidate backend mode exists):

```bash
VIDERA_STORAGE_BACKEND=<candidate> make storage-benchmark-capture OUT=/tmp/videra_phase32_prereq1_candidate.out EXIT_OUT=/tmp/videra_phase32_prereq1_candidate.exit
VIDERA_STORAGE_BACKEND=<candidate> make storage-benchmark-summarize OUT=/tmp/videra_phase32_prereq1_candidate.out
```

## Evidence Update Path

Update:

- `tasks/platform/phase-32-prereq-1-measured-benefit-evidence-2026-03-05.md`

Required fields to fill:

- benchmark table `Candidate backend`, `Delta`, `Material gain?`
- candidate command/output/exit evidence
- prerequisite decision status

## Decision Rule

Mark prerequisite 1 as pass only when:

- both baseline and candidate captures exit `0`, and
- candidate shows material gain vs baseline on target workloads with rationale recorded.

Related one-pass flow:

- `tasks/platform/phase-32-candidate-proof-pack-runbook-2026-03-05.md`
