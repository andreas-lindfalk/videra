# Real Corpus Promotion Gate — 2026-03-05

Purpose: define promotion criteria for production-like real-corpus readiness using explicit thresholds and reproducible commands.

Status: active promotion criterion artifact.

## Promotion Scope

This gate is for promotion decisions (deploy/rollout readiness), not for introducing new retrieval algorithms or API changes.

## Required Command

Run from repo root:

```bash
make real-corpus-promotion-gate
```

This command composes:

- pilot benchmark + real-mode guardrails (`make pilot-quality-gate`)
- deterministic/evidence/contract-focused integration checks for proofpack scenarios

## Promotion Criteria (P1–P5)

### P1 — Real-mode Guardrail Semantics

All must pass:

- `TestIndexVideoRealModeRemotePathRespectsMaxSizeBound`
- `TestIndexVideoRealModeRemotePathHonorsDisabledFetch`
- `TestIndexVideoRealModeRequiresSidecarForLocalPath`

### P2 — Deterministic Replay

- Deterministic replay remains stable for gate queries.
- Threshold: `deterministicRate >= 1.00` on pilot scorecard set.

### P3 — Evidence Quality

- Pilot benchmark scorecard thresholds:
  - `evidenceMatchRate >= 0.90`
  - `topTwoQualityRate >= 0.85`

### P4 — Contract Stability

No regressions in contract-focused checks:

- `TestToolResponseBackwardCompatFields`
- transcript resource path compatibility remains unchanged

### P5 — Command Reproducibility

- `make real-corpus-promotion-gate` exits `0` on a fresh run.

## GO/NO-GO Decision Rule

- **GO** when P1–P5 are all satisfied.
- **NO-GO** when any criterion fails or results are non-reproducible.

## Evidence Requirement

Use and fill:

- `tasks/platform/real-corpus-promotion-evidence-template.md`

Store dated result artifacts as:

- `tasks/platform/real-corpus-promotion-evidence-YYYY-MM-DD.md`

## Non-Goals

- No MCP schema/tool/resource changes.
- No storage backend migration implementation.
- No ranking/embedding behavior changes in this gate definition phase.
