# Phase 32 Prerequisite 2 — Runtime Contract Readiness Evidence (Stub)

Date: 2026-03-05
Status: pass
Owner: Andreas (Platform lead, proposed)
Priority: P0
Target date (suggested): 2026-03-08

Goal: approve runtime/profile strategy for CGO/native artifacts across local and deployment targets.

## Runtime Contract Proposal (Fill)

- default local profile: CGO-free baseline remains default for all non-migration workflows.
- default container profile: `runtime-slim` remains default (`VIDERA_DOCKER_TARGET=runtime-slim`).
- optional profile(s): `runtime-full` remains tool-complete optional path; migration-candidate runtime profile must be explicit opt-in only.
- CGO expectations: CGO is allowed only in migration-candidate mode, never as default in this phase.
- native artifact distribution/loading strategy: package native artifacts only in candidate runtime profile; baseline profile remains static-first and unchanged.

## Environment Matrix (Fill)

| Environment | Expected profile | CGO required | Native artifacts available | Decision |
|---|---|---|---|---|
| Local dev | baseline default + optional candidate profile | no (default), yes (candidate only) | candidate-only | pass |
| Docker local | runtime-slim default / runtime-full optional + candidate profile | no (default), yes (candidate only) | candidate-only | pass |
| Hetzner target | baseline profile unchanged; candidate profile opt-in | no (default), yes (candidate only) | candidate-only | pass |
| Cloud Run target | baseline profile unchanged; candidate profile opt-in | no (default), yes (candidate only) | candidate-only | pass |

## Risks and Mitigation

- risks:
	- migration candidate may require CGO + native artifact packaging, increasing build/deploy complexity.
	- runtime contract divergence across local/Hetzner/Cloud Run could reduce parity confidence.
- mitigation:
	- approve one explicit profile strategy before any migration implementation.
	- validate target environments with the same contract matrix and dated evidence.

## Approval Snapshot

- Decision: approved
- Approval date: 2026-03-05
- Contract rule: static-first baseline is immutable for this phase; any CGO/native path is candidate-only and opt-in.

## Decision

- prerequisite status: pass
- rationale: runtime contract strategy is explicitly approved with baseline invariants preserved and candidate profile constrained to opt-in mode.

## Handoff Notes

- This prerequisite is a hard dependency for prerequisites 1, 4, and 5.
