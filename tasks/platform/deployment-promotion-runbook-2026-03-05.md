# Deployment Promotion Runbook — 2026-03-05

Purpose: provide a single operator workflow for release/promotion decisions by combining release, quality, and real-corpus promotion gates.

Status: active consolidated runbook.

## Canonical Command

Run from repo root:

```bash
make deployment-promotion-gate
```

Composed flow:

1. `make release-gate`
2. `make release-gate-split`
3. `make real-corpus-promotion-gate`

## Pass Criteria

Promotion is **GO** only if all are true:

- All three composed command segments exit `0`.
- Release contract checks stay stable (MCP tool/resource compatibility).
- Pilot metrics satisfy thresholds (`evidenceMatchRate >= 0.90`, `topTwoQualityRate >= 0.85`, deterministic replay green).
- Real-mode guardrails remain green (`max-size`, `disabled fetch`, `sidecar required`).

## Failure Policy

Promotion is **NO-GO** if any command segment fails or outputs ambiguous/unreproducible results.

If NO-GO:

1. capture failing segment and key log lines,
2. document mitigation plan,
3. rerun `make deployment-promotion-gate` after mitigation.

## Evidence Workflow

For each promotion run, fill:

- `tasks/platform/deployment-promotion-evidence-template.md`

Store dated artifacts as:

- `tasks/platform/deployment-promotion-evidence-YYYY-MM-DD.md`

## Related Artifacts

- `tasks/platform/mvp-release-gate.md`
- `tasks/platform/real-corpus-promotion-gate-2026-03-05.md`
- `tasks/platform/release-quality-coupling-evidence-2026-03-05.md`
- `tasks/platform/real-corpus-promotion-evidence-2026-03-05.md`
