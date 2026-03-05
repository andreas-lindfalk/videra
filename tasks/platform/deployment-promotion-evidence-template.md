# Deployment Promotion Evidence Template

Date: <YYYY-MM-DD>
Operator: <name>
Branch/Commit: <branch>/<sha>

## Command

```bash
make deployment-promotion-gate
```

## Segment Results

- release-gate: pass/fail
- release-gate-split: pass/fail
- real-corpus-promotion-gate: pass/fail

## Quality/Promotion Signals

- evidenceMatchRate: <value>
- deterministicRate: <value>
- topTwoQualityRate: <value>
- real-mode guardrails (max-size / disabled-fetch / sidecar): pass/fail

## Contract Checks

- MCP tool/resource compatibility unchanged: yes/no
- If no, describe regression:

## Decision

- GO / NO-GO
- rationale:
- mitigation + rerun plan (if NO-GO):
