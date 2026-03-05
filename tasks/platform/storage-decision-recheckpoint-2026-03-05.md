# Storage Decision Re-checkpoint — 2026-03-05 (Phase 29)

Purpose: re-evaluate storage direction now that quality gates (Phase 27/28) are stable, and decide whether to start a migration track.

Status: Completed

## Inputs Considered

- Phase 27 release-quality coupling is green.
- Phase 28 real-corpus promotion gate is green (`make real-corpus-promotion-gate`).
- Current build profile is static-first (`CGO_ENABLED=0`) in `Dockerfile`.
- Current store abstraction is stable (`internal/storage/store.go`, `VectorStore`).
- Current backend dependency is `github.com/philippgille/chromem-go` (`go.mod`).
- Previous checkpoint exists: `tasks/platform/lancedb-storage-checkpoint-2026-03-05.md`.

## Decision Options

1. **Option A:** Continue with `chromem-go` and defer migration.
2. **Option B:** Start controlled `lancedb-go` migration track now.

## Weighted Decision Matrix (1–5, higher is better)

| Criterion | Weight | Option A: chromem-go now | Option B: start lancedb-go now | Notes |
|---|---:|---:|---:|---|
| MCP compatibility risk | 20 | 5 | 3 | Migration increases near-term regression surface even with stable interface goals. |
| Build/runtime portability | 20 | 5 | 2 | Current runtime assumes CGO-free static build; Lance path introduces native/CGO operational complexity. |
| Delivery focus alignment | 15 | 5 | 2 | Current roadmap priority is low-churn reliability/ops consolidation after quality gates. |
| Ops/deploy complexity | 15 | 4 | 2 | Existing local/cloud runbooks are tuned to current profile and images. |
| Retrieval capability upside | 10 | 3 | 4 | Lance may provide backend flexibility, but benefit is not yet benchmark-proven in this repo. |
| Migration/rollback complexity | 10 | 5 | 2 | New backend requires migration semantics and rollback validation before safe adoption. |
| Data-path/cloud fit potential | 10 | 3 | 4 | Lance may improve future object-storage alignment if adopted with proper controls. |
| **Total (max 500)** | **100** | **445** | **255** | Decision signal favors Option A at current stage. |

## Migration GO Criteria (all required)

Only start migration track if all are satisfied:

1. **Measured benefit:** backend comparison benchmark proves material gain versus current backend on target workloads.
2. **Runtime contract readiness:** CGO/native artifact strategy is approved for local + deployment profiles.
3. **Compatibility plan:** staged migration approach is documented (feature flag/dual backend boundaries).
4. **Rollback proof:** rollback path is explicitly defined and testable.
5. **Gate parity:** release + pilot + promotion gates remain green with migration mode enabled.

## Phase 29 Verdict

- **Decision:** NO-GO for immediate migration track.
- **Selected path:** continue with `chromem-go` for now.

Rationale:

- Current weighted matrix strongly favors lower-risk continuity.
- Migration GO criteria are not yet fully satisfied (notably benchmark proof + runtime-contract readiness).
- Current roadmap objective is operational consolidation before backend transition risk.

## Follow-up Actions

- Keep storage migration out of active implementation until GO criteria are met.
- Use Phase 30 for promotion/runbook consolidation.
- Re-open storage migration only via an explicit checkpoint update with benchmark artifacts.
