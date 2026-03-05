# Deployment Promotion Evidence — 2026-03-05

Date: 2026-03-05
Phase: 30 (Deployment Promotion Runbook Consolidation)
Branch: main

## Command

```bash
make deployment-promotion-gate
```

## Segment Results

- release-gate: PASS
- release-gate-split: PASS
- real-corpus-promotion-gate: PASS

## Quality/Promotion Signals

- evidenceMatchRate: `1.00`
- deterministicRate: `1.00`
- topTwoQualityRate: `1.00`
- real-mode guardrails: PASS (`max-size`, `disabled-fetch`, `sidecar-required`)

## Command Output Summary

- Exit code file: `/tmp/videra_phase30_deployment_promotion_gate.exit` => `0`
- Key integration package pass lines captured from `/tmp/videra_phase30_deployment_promotion_gate.out`:
  - `ok github.com/andreas-lindfalk/videra/test/integration 133.116s`
  - `ok github.com/andreas-lindfalk/videra/test/integration 25.070s`
  - `ok github.com/andreas-lindfalk/videra/test/integration 43.964s`
  - `ok github.com/andreas-lindfalk/videra/test/integration 12.016s`

## Contract Notes

- MCP schema/tool/resource changes introduced: no

## Decision

- GO
- Rationale: all composed segments passed and quality/promotion thresholds remained green.
