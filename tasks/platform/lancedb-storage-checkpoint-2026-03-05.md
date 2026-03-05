# LanceDB Storage Checkpoint — 2026-03-05

Purpose: decide storage direction for the next target state without mixing in full migration scope.

Status: Completed

Superseded by: `tasks/platform/storage-decision-recheckpoint-2026-03-05.md` (Phase 29)

## Decision Goal

Decide whether the next phase should:
- keep `chromem-go` during semantic-quality lift, or
- start a controlled LanceDB migration track using `lancedb-go`.

## Locked Planning Invariants

1. Preserve MCP contract stability:
   - `index_video`, `get_index_job`, `search_video`, `list_videos`, transcript resource.
2. Preserve two-path data model:
   - Read-only source path(s) for media inputs.
   - Separate read-write path for index/state outputs.
3. Do not merge watcher/event-ingestion scope into this decision.

## Current Fact Baseline

- Repo currently uses `chromem-go` as storage backend.
- Existing release discipline (RC2) is green and should remain the quality gate baseline.
- `lancedb-go` exists and supports Go integration, but requires CGO and native artifact handling.

## Evaluation Matrix

| Criterion | Keep `chromem-go` now | Start LanceDB migration now | Notes |
|---|---|---|---|
| MCP contract stability risk | Low | Medium | Migration pressure increases contract-adjacent regression risk even if API is intended to remain unchanged. |
| Build portability risk (static vs CGO) | Low | High | Current Docker build uses `CGO_ENABLED=0`; LanceDB Go requires CGO + native artifacts and linker flags. |
| Local dev setup complexity | Low | High | Current `go build` flow is simple; LanceDB introduces per-platform artifact/bootstrap requirements. |
| Cloud deploy complexity | Medium | High | Requires deterministic native artifact packaging and runtime profile handling across targets. |
| Retrieval capability fit (next phase) | Medium | Medium/High | LanceDB can improve backend flexibility, but semantic quality lift depends primarily on embedding/model path changes. |
| Migration/rollback complexity | Low | High | Existing persistent data path and store behavior would require migration/compat strategy and rollback plan. |
| Time-to-value for semantic quality lift | High | Low | Staying on current backend accelerates quality work; migration now delays model-quality objectives. |

## Required Evidence Inputs

1. Current backend coupling points and migration surface under `internal/storage/` and ingestion/query call sites.
2. Build/runtime contract impact for LanceDB adoption:
   - CGO requirement,
   - native library distribution/loading,
   - profile implications for local and container builds.
3. Operational impact for local + cloud deployment paths.
4. Risk profile for running storage migration concurrently with semantic-quality lift.

## Go/No-Go Decision Template

Decision:
- [ ] GO — start controlled LanceDB migration track now.
- [x] NO-GO — keep `chromem-go` for next quality-lift phase and revisit after that phase.

Rationale:
- Priority for the next target state is semantic quality lift with stable MCP contract and minimal operational churn.
- Immediate LanceDB migration introduces build/runtime complexity (CGO + native artifact management) that competes with that priority.
- LanceDB remains a valid candidate and will be re-evaluated after the quality-lift phase with a controlled migration strategy.

If GO, required guardrails:
- [ ] Backend abstraction remains intact; no MCP schema changes.
- [ ] Migration path is staged (feature flag or dual backend strategy).
- [ ] Rollback path is documented and testable.
- [ ] CI/build profile updates are documented and reproducible.

If NO-GO, required follow-up:
- [x] Record explicit revisit trigger conditions.
- [x] Proceed with semantic-quality lift on current backend.

### Revisit Trigger Conditions (LanceDB)

Re-open LanceDB migration checkpoint when one or more conditions are true:
1. Semantic quality-lift phase is completed and release-gated.
2. Current storage backend becomes a measured retrieval bottleneck under target workload.
3. Feature needs require Lance-specific index/storage capabilities not practical in current backend.
4. Team accepts CGO/native artifact build profile as part of standard runtime contract.

## Explicit Non-Goals

- Immediate production migration to LanceDB.
- Watcher/event source discovery implementation.
- AgentGateway policy rollout implementation.
