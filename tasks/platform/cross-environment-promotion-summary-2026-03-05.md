# Cross-Environment Promotion Summary — 2026-03-05

Date: 2026-03-05
Operator: GitHub Copilot
Branch/Commit: main / working-tree
Image/Version: local phase artifacts (no single cross-env image tag recorded)

## Source Evidence

- Local promotion evidence: `tasks/platform/deployment-promotion-evidence-2026-03-05.md`
- Local release evidence: `tasks/platform/rc2-release-evidence-2026-03-05.md`
- Real-corpus promotion evidence: `tasks/platform/real-corpus-promotion-evidence-2026-03-05.md`
- Hetzner parity evidence: pending
- Cloud Run parity evidence: pending

## Local Promotion Gate Status

- `make deployment-promotion-gate`: pass
- `release-gate`: pass
- `release-gate-split`: pass
- `real-corpus-promotion-gate`: pass

Quality metrics:

- evidenceMatchRate: `1.00`
- deterministicRate: `1.00`
- topTwoQualityRate: `1.00`
- real-mode guardrails: pass

## Environment Parity Matrix

| Check | Hetzner | Cloud Run | Notes |
|---|---|---|---|
| MCP connect (`/mcp`) | pending | pending | Not executed in this phase |
| `list_videos` contract shape | pending | pending | Not executed in this phase |
| `index_video` | pending | pending | Not executed in this phase |
| `search_video` deterministic repeat | pending | pending | Not executed in this phase |
| transcript resource (`video://{id}/transcript`) | pending | pending | Not executed in this phase |
| restart behavior | pending | pending | Not executed in this phase |
| persistence behavior | pending | pending | Not executed in this phase |

## Contract/Compatibility Checks

- MCP tool/resource schema changed: no
- Split-role lifecycle semantics status: pass (local evidence path)

## Risks and Mitigation

- Open risks:
  - Cross-environment parity is not yet evidenced in deployed Hetzner/Cloud Run environments.
- Mitigation plan:
  - Execute parity checklist in both environments and update this summary with concrete pass/fail outcomes.

## Decision

- Local promotion decision: GO
- Cross-environment promotion decision: CONDITIONAL
- Rationale: local promotion gates are green, but environment parity evidence is still pending.
- Required follow-up before final GO:
  - run `tasks/platform/parity-validation-checklist.md` in Hetzner and Cloud Run,
  - attach dated parity evidence paths,
  - re-evaluate this summary to GO/NO-GO.

Active execution owner phase:

- `tasks/todo.md` (Phase 34 — Cross-Environment Parity Execution)
