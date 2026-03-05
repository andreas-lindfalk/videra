# Cross-Environment Promotion Summary Runbook — 2026-03-05

Purpose: provide one final GO/NO-GO summary format that merges local promotion gate evidence with Hetzner/Cloud Run parity validation.

Status: active (Phase 33).

## Canonical Artifact

Use this template for final decision handoff:

- `tasks/platform/cross-environment-promotion-summary-template.md`

Store dated outputs as:

- `tasks/platform/cross-environment-promotion-summary-YYYY-MM-DD.md`

## Inputs Required

Local evidence inputs:

- `tasks/platform/deployment-promotion-evidence-YYYY-MM-DD.md`
- `tasks/platform/rc2-release-evidence-YYYY-MM-DD.md` (or current release evidence equivalent)
- `tasks/platform/real-corpus-promotion-evidence-YYYY-MM-DD.md`

Parity evidence inputs:

- Hetzner parity checklist run output
- Cloud Run parity checklist run output
- Checklist reference: `tasks/platform/parity-validation-checklist.md`

## Fill Order

1. Fill local promotion section from `deployment-promotion-evidence`.
2. Fill quality metrics (`evidenceMatchRate`, `deterministicRate`, `topTwoQualityRate`).
3. Fill environment parity matrix from Hetzner and Cloud Run runs.
4. Fill contract/split-role status.
5. Decide:
   - local promotion decision,
   - cross-environment promotion decision.

## Decision Policy

- **Cross-environment GO** requires:
  - local promotion gate GO,
  - no MCP contract regression,
  - parity matrix checks passed in both Hetzner and Cloud Run (or explicitly documented N/A semantics).
- **Cross-environment NO-GO** when any required parity or contract check fails.
- **Conditional** is allowed only for interim reporting when one or both environments are pending execution; must include explicit follow-up actions and owners.

## Related Artifacts

- `tasks/platform/deployment-promotion-runbook-2026-03-05.md`
- `tasks/platform/deployment-promotion-evidence-template.md`
- `tasks/platform/parity-validation-checklist.md`
- `tasks/platform/hetzner-gcp-parity-primer.md`
