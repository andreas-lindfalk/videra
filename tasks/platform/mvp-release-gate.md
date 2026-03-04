# MVP Release Gate (Phase 16)

Goal: provide a single reproducible go/no-go workflow for Videra MVP release-candidate readiness.

Related docs:

- `README.md`
- `tasks/platform/parity-validation-checklist.md`
- `tasks/platform/env-contract.md`
- `tasks/platform/queue-redis-first-runbook.md`
- `tasks/platform/hetzner-gcp-parity-primer.md`
- `tasks/platform/mvp-release-gate-evidence-2026-03-04.md`
- `tasks/platform/rc1-stabilization-evidence-2026-03-04.md`
- `tasks/platform/post-mvp-backlog-cut-2026-03-04.md`
- `tasks/platform/final-mvp-handoff-2026-03-04.md`

## Required Command Gate

Run from repo root:

```bash
make release-gate
make release-gate-split
```

Optional preflight-only check:

```bash
make release-gate-preflight
```

Expected outcome:

- Both commands exit with code `0`.
- No MCP contract/tool schema changes are required to pass.

## What `make release-gate` covers

- `make build`
- `make test`
- `make integration-test`
- `make docker-build`

## What `make release-gate-split` covers

Focused split-role release-critical semantics:

- `TestIndexVideoAsyncSplitRoleRedisLifecycle`
- `TestIndexVideoAsyncSplitRoleRedisSharedStorageVisibility`
- `TestWorkerRoleWithHTTPTransportFailsFastAtStartup`

## Release Evidence Template

```text
Date: <YYYY-MM-DD>
Branch/Commit: <branch> / <sha>
Operator: <name>

Command Results:
- make release-gate: pass/fail
- make release-gate-split: pass/fail

Contract Checks:
- index_video / get_index_job compatibility: pass/fail
- search_video / list_videos compatibility: pass/fail
- transcript resource compatibility: pass/fail
- split-role shared-storage semantics verified: pass/fail
- split-role degraded semantics + operator signal verified: pass/fail

Deployment Notes:
- Local/private assumptions validated: <yes/no + notes>
- Hetzner/Cloud Run parity notes reviewed: <yes/no + notes>

Open Risks / Deferred Items:
- <item or 'none'>

Go/No-Go:
- GO / NO-GO
```

## Decision Policy

- **GO** when both gate commands pass and there are no unresolved contract regressions.
- **NO-GO** when any gate command fails or parity/split-role checks are ambiguous.
- If **NO-GO**, document mitigation and rerun both commands before re-evaluation.

## Troubleshooting and Rerun Discipline (RC1 Stabilization)

When release-gate runs fail intermittently due to local Docker pressure (for example no-space/image-cache pressure), use this deterministic rerun flow:

```bash
make release-gate-preflight
make release-gate-clean
make release-gate
make release-gate-split
```

Notes:

- `release-gate` now executes `integration-test-fresh` (`-count=1`) to avoid cached false-confidence runs.
- `release-gate-clean` intentionally prunes builder cache + dangling images only (no volume prune).
