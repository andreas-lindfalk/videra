# Phase 32 Prerequisite 4 — Execution Runbook (Rollback Proof)

Date: 2026-03-05
Scope: rollback rehearsal flow with explicit command capture and evidence update path.

## Current Capability Note

Candidate backend mode is available via `VIDERA_STORAGE_BACKEND=lancedb` (compatibility layer).
Use this rehearsal flow to prove candidate pre-check + rollback-to-stable gate rerun discipline.

## Commands (Current Rehearsal)

```bash
make rollback-rehearsal-capture OUT=/tmp/videra_phase32_prereq4_rollback_rehearsal.out EXIT_OUT=/tmp/videra_phase32_prereq4_rollback_rehearsal.exit
make rollback-rehearsal-summarize OUT=/tmp/videra_phase32_prereq4_rollback_rehearsal.out
```

What this does:

1. Runs `make release-gate-split` as pre-rollback gate.
2. Simulates rollback to stable backend (`VIDERA_STORAGE_BACKEND=chromem`) and reruns `make release-gate-split`.
3. Captures output + exit code files.

## Full Drill (Current)

Run pre-rollback checks in candidate mode, then switch to stable backend (`chromem`) and rerun required gates.

## Evidence Update Path

Update:

- `tasks/platform/phase-32-prereq-4-rollback-proof-evidence-2026-03-05.md`

Required fields to fill:

- command/output/exit evidence
- start state and rollback state
- verification checklist
- prerequisite decision status

## Decision Rule

Mark prerequisite 4 as pass only when:

- rollback drill is executed against real candidate mode,
- rollback to stable mode succeeds,
- required verification checks are green and evidence-backed.

Related one-pass flow:

- `tasks/platform/phase-32-candidate-proof-pack-runbook-2026-03-05.md`
