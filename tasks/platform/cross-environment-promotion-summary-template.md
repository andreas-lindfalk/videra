# Cross-Environment Promotion Summary Template

Date: <YYYY-MM-DD>
Operator: <name>
Branch/Commit: <branch>/<sha>
Image/Version: <image-tag>

## Source Evidence

- Local promotion evidence: <path>
- Local release evidence: <path>
- Real-corpus promotion evidence: <path>
- Hetzner parity evidence: <path or pending>
- Cloud Run parity evidence: <path or pending>

## Local Promotion Gate Status

- `make deployment-promotion-gate`: pass/fail
- `release-gate`: pass/fail
- `release-gate-split`: pass/fail
- `real-corpus-promotion-gate`: pass/fail

Quality metrics:

- evidenceMatchRate: <value>
- deterministicRate: <value>
- topTwoQualityRate: <value>
- real-mode guardrails: pass/fail

## Environment Parity Matrix

| Check | Hetzner | Cloud Run | Notes |
|---|---|---|---|
| MCP connect (`/mcp`) | pass/fail/pending | pass/fail/pending | |
| `list_videos` contract shape | pass/fail/pending | pass/fail/pending | |
| `index_video` | pass/fail/pending | pass/fail/pending | |
| `search_video` deterministic repeat | pass/fail/pending | pass/fail/pending | |
| transcript resource (`video://{id}/transcript`) | pass/fail/pending | pass/fail/pending | |
| restart behavior | pass/fail/pending | pass/fail/pending | |
| persistence behavior | pass/fail/pending/n/a | pass/fail/pending/n/a | |

## Contract/Compatibility Checks

- MCP tool/resource schema changed: yes/no
- If yes, regression description:
- Split-role lifecycle semantics status: pass/fail/pending

## Risks and Mitigation

- Open risks:
- Mitigation plan:

## Decision

- Local promotion decision: GO / NO-GO
- Cross-environment promotion decision: GO / NO-GO / CONDITIONAL
- Rationale:
- Required follow-up before final GO (if any):
