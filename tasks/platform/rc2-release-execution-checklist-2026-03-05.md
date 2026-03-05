# RC2 Release Execution Checklist — 2026-03-05

Purpose: provide a copy/paste-safe operator checklist for final RC2 execution handoff.

## 1) Preconditions

- [x] Working tree is clean or intentionally tracked.
- [x] Docker Desktop is healthy and running.
- [x] Required local disk headroom is available.
- [x] Operator has reviewed:
  - `tasks/platform/mvp-release-gate.md`
  - `tasks/platform/final-mvp-handoff-2026-03-04.md`
  - `tasks/platform/post-mvp-backlog-cut-2026-03-04.md`

## 2) Canonical Command Sequence

Run from repo root:

```bash
make release-gate
make release-gate-split
```

If Docker/cache pressure is suspected:

```bash
make release-gate-preflight
make release-gate-clean
make release-gate
make release-gate-split
```

## 3) Evidence Capture

- [x] Record command outcomes in `tasks/platform/rc2-release-evidence-2026-03-05.md`.
- [x] Confirm release-critical contract paths remain unchanged:
  - `index_video`
  - `get_index_job`
  - `search_video`
  - `list_videos`
  - transcript resource (`video://{id}/transcript`)

## 4) Release Notes (Concise)

Known limits:

- Cloud parity behavior still requires environment-run validation evidence outside local RC execution.
- Queue/backend stress validation beyond current gate remains post-MVP hardening scope.

Deferred items:

- Full deferred list and priorities live in `tasks/platform/post-mvp-backlog-cut-2026-03-04.md`.
- Do not pull deferred items into RC2 execution unless they become explicit release blockers.

## 5) Final Decision

- [x] GO — all gate commands pass and no unresolved contract regressions.
- [ ] NO-GO — any gate failure or unresolved ambiguity remains.

If NO-GO:

- [ ] Document blocker and mitigation path.
- [ ] Re-run canonical command sequence after mitigation.

## 6) Phase 27 Add-on (Release + Quality Signal Coupling)

For all new RC-style runs after Phase 27, include this additional required step:

```bash
make pilot-quality-gate
```

Required evidence additions:

- [ ] Record `make pilot-quality-gate` outcome in release evidence.
- [ ] Capture pilot scorecard metrics from logs:
  - [ ] `evidenceMatchRate`
  - [ ] `deterministicRate`
  - [ ] `topTwoQualityRate`
- [ ] Treat release decision as NO-GO if pilot quality gate fails.

## 7) Phase 30 Consolidated Operator Path

Preferred single-command promotion path for future runs:

```bash
make deployment-promotion-gate
```

This command composes:

- `make release-gate`
- `make release-gate-split`
- `make real-corpus-promotion-gate`
