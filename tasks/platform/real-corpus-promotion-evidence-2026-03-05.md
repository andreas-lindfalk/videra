# Real Corpus Promotion Evidence — 2026-03-05

Date: 2026-03-05
Operator: Copilot
Branch/Commit: main / working-tree

## Command

```bash
make real-corpus-promotion-gate
```

## Criteria Outcomes

- P1 Real-mode guardrails: PASS
- P2 Deterministic replay (`deterministicRate >= 1.00`): PASS
- P3 Evidence quality:
  - evidenceMatchRate (`>= 0.90`): PASS (`1.00`)
  - topTwoQualityRate (`>= 0.85`): PASS (`1.00`)
- P4 Contract stability checks: PASS
- P5 Command reproducibility (`exit code 0`): PASS

## Captured Metrics

- evidenceMatchRate: `1.00`
- deterministicRate: `1.00`
- topTwoQualityRate: `1.00`

## Command Output Summary

- Exit code file: `/tmp/videra_phase28_real_corpus_promotion_gate.exit` => `0`
- Integration package outputs captured:
  - `ok github.com/andreas-lindfalk/videra/test/integration 43.605s`
  - `ok github.com/andreas-lindfalk/videra/test/integration 11.852s`
- Notable lines:
  - `pilot benchmark scorecard: scenarios=6 evidenceMatchRate=1.00 deterministicRate=1.00 topTwoQualityRate=1.00`
  - `TestProofPackScenariosEvidenceAndDeterminism` PASS
  - `TestProofPackProductRecallPrioritizesTop2Evidence` PASS
  - `TestSearchVideoDeterministicOrdering` PASS
  - `TestToolResponseBackwardCompatFields` PASS

## Contract Notes

- MCP schema/tool/resource changes introduced: no

## GO/NO-GO

- Decision: **GO**
- Rationale: all promotion criteria (P1–P5) passed with reproducible command execution.
