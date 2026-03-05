# Todo

## Active Task

### Phase 34 — Cross-Environment Parity Execution (Hetzner + Cloud Run)

Status: IN PROGRESS

Why this is next:

- Phase 32 is closed (NO-GO) and parked by checkpoint decision.
- The highest remaining release confidence gap is pending parity evidence in real Hetzner/Cloud Run environments.

Primary artifacts:

- `tasks/platform/parity-validation-checklist.md`
- `tasks/platform/cross-environment-promotion-summary-2026-03-05.md`
- `tasks/platform/cross-environment-promotion-summary-runbook-2026-03-05.md`

Execution plan (current pass):

- [ ] Execute parity checklist in Hetzner with dated evidence capture.
- [ ] Execute parity checklist in Cloud Run with dated evidence capture.
- [ ] Update unified cross-environment summary from CONDITIONAL to GO/NO-GO with rationale.

## Definition of Done Template

Use this checklist at the top of each new phase/feature todo.

- [ ] Scope is explicit (in/out) and aligned with current architecture decisions.
- [ ] Required interfaces/contracts are updated without breaking existing MCP surface.
- [ ] Fast tests pass (`make test`).
- [ ] Integration tests pass (`make integration-test`).
- [ ] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [ ] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [ ] Todo is fully checked and ready for archive.
