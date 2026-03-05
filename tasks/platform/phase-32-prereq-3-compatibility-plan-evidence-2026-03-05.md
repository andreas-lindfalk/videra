# Phase 32 Prerequisite 3 — Compatibility Plan Evidence (Stub)

Date: 2026-03-05
Status: pass
Owner: Andreas (Core runtime lead, proposed)
Priority: P0
Target date (suggested): 2026-03-08

Goal: document staged migration plan while preserving MCP contract compatibility.

## Compatibility Invariants (Fill)

- `index_video` compatibility preserved: yes (required invariant)
- `get_index_job` compatibility preserved: yes (required invariant)
- `search_video` compatibility preserved: yes (required invariant)
- `list_videos` compatibility preserved: yes (required invariant)
- transcript resource compatibility preserved: yes (required invariant)

## Staged Plan (Fill)

1. stage 1: introduce backend-selection boundary with default behavior unchanged.
2. stage 2: implement candidate backend path behind explicit rollout control while keeping existing path as default.
3. stage 3: run full release/promotion/parity evidence flow under candidate mode before any default switch.
4. stage 4: perform rollback drill and record evidence before considering checkpoint GO.

## Rollout Controls (Fill)

- feature flag name: `VIDERA_STORAGE_BACKEND` (approved name for plan artifacts; implementation pending).
- default state: current backend path (`chromem-go`).
- enable path: explicit opt-in to candidate backend.
- disable path: immediate fallback to current backend path.

## Verification Plan

- test and evidence command set:
	- `make release-gate`
	- `make release-gate-split`
	- `make pilot-quality-gate`
	- `make real-corpus-promotion-gate`
	- cross-environment summary refresh via `tasks/platform/cross-environment-promotion-summary-template.md`
- expected outcomes:
	- no MCP contract regressions,
	- all required gates green under candidate mode,
	- rollback path documented and drillable.

## Approval Snapshot

- Decision: approved
- Approval date: 2026-03-05
- Guardrails:
  - no MCP schema/tool/resource contract changes,
  - default backend remains `chromem-go` until all Phase 32 prerequisites pass,
  - any candidate rollout is opt-in and reversible.

## Decision

- prerequisite status: pass
- rationale: staged compatibility plan, rollout controls, and gate-based verification path are now explicitly approved and actionable.

## Handoff Notes

- This prerequisite is a hard dependency for prerequisites 4 and 5.
