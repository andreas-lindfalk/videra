# Real Corpus Promotion Evidence Template

Date: <YYYY-MM-DD>
Operator: <name>
Branch/Commit: <branch>/<sha>

## Command

```bash
make real-corpus-promotion-gate
```

## Criteria Outcomes

- P1 Real-mode guardrails: pass/fail
- P2 Deterministic replay (`deterministicRate >= 1.00`): pass/fail
- P3 Evidence quality:
  - evidenceMatchRate (`>= 0.90`): pass/fail + value
  - topTwoQualityRate (`>= 0.85`): pass/fail + value
- P4 Contract stability checks: pass/fail
- P5 Command reproducibility (`exit code 0`): pass/fail

## Captured Metrics

- evidenceMatchRate: <value>
- deterministicRate: <value>
- topTwoQualityRate: <value>

## Command Output Summary

- Integration package result: <result>
- Notable pass/fail lines:
  - <line>
  - <line>

## Contract Notes

- MCP schema/tool/resource changes introduced: yes/no
- If yes, list and explain:

## GO/NO-GO

- Decision: GO / NO-GO
- Rationale:
- If NO-GO, mitigation and rerun plan:
