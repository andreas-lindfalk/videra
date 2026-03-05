# Spec vs Implementation Alignment — 2026-03-05

Purpose: lock agreement points after reflection and define the next decision path.

Status: Phase 19 decision checkpoint initiated.

Checkpoint outcome: NO-GO for immediate LanceDB migration; keep `chromem-go` for next semantic-quality phase and revisit LanceDB after that phase.

## Confirmed Alignment

1. **Data-flow model (agreed):**
   - Source locations are treated as read-only inputs (local mounts and/or bucket-backed sources over time).
   - System writes only to a separate read-write index/data location.
   - Upload-through-app is not required for the core operating model.

2. **Contract stability (agreed):**
   - Keep MCP surface stable while improving retrieval quality.

3. **Priority (agreed):**
   - Next target state focuses on semantic quality lift first.

## MVP Intent vs Current Reality (delta snapshot)

- MVP spec intent includes LanceDB + multimodal retrieval depth.
- Current implementation is release-stable on MCP contract and async/split-role operations, but semantic quality is baseline in key embedding paths.
- Current storage backend is `chromem-go` by explicit repo decision history.

## LanceDB Clarification (current fact state)

- `lancedb-go` exists and is a valid candidate for evaluation.
- Integration is **not** a drop-in from current runtime assumptions:
  - requires CGO/native library handling,
  - requires explicit build/runtime profile decisions,
  - should be evaluated via checkpoint before migration commitment.

## Decision Checkpoint to Run Next

### Checkpoint Goal
Decide whether to keep `chromem-go` for the next quality-lift phase, or introduce a controlled LanceDB migration track.

### Checkpoint Scope
- Compare `chromem-go` vs `lancedb-go` against:
  - portability and build profile impact,
  - local + cloud deployment fit,
  - retrieval/query capability fit for planned quality lift,
  - rollback/migration complexity.

### Exit Criteria
- Produce a written go/no-go decision.
- If go:
  - define migration strategy (flagged dual backend or staged swap),
  - define acceptance tests and rollback path.
- If no-go:
  - continue quality-lift work on current backend without reopening storage during that phase.

## Explicit Non-Goals for this checkpoint

- No immediate storage migration.
- No MCP tool/resource contract changes.
- No watcher/event ingestion implementation in the same step.
