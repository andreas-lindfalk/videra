# Semantic Quality Lift Evidence — 2026-03-05

Date: 2026-03-05
Phase: 20 (Semantic Quality Lift)
Branch: main

## Baseline Reference

- Pre-change baseline artifact: `tasks/platform/semantic-quality-baseline-2026-03-05.md`

## Implemented Changes

1. Deterministic text embedding upgraded from raw whole-string hash to token-aware semantic hashing (still deterministic and dependency-free):
   - token normalization/canonicalization,
   - token + bigram feature hashing,
   - normalized vectors.
2. Hybrid reranking strengthened:
   - synonym-aware query/snippet normalization,
   - token-overlap + bigram lexical boosting,
   - safer handling of NaN/Inf similarity inputs,
   - modality-diversity heuristic adjusted so strongly query-relevant same-modality hits are not incorrectly deferred.
3. Simulated ingestion audio segments now rely on store embedder computation (instead of hard-coded mock vectors), aligning query/segment embedding semantics in test mode.
4. Candidate recall hardening in storage search path:
   - vector-query candidates are enriched with transcript fallback candidates before reranking.
5. New quality guardrail test added:
   - `TestProofPackProductRecallPrioritizesTop2Evidence`.

## Quality Signals (Before/After)

- Before (baseline): proofpack coverage and determinism gates passed, but no explicit assertion that the `product_interview_recall` scenario places both "intro segment" and "discussion" evidence in top-2.
- After (Phase 20): new integration guardrail proves top-2 prioritization for that scenario while existing proofpack/determinism/backward-compat tests remain green.

## Validation Commands and Outcomes

Focused regression suite (unit + integration):

```bash
go test ./internal/embedding ./internal/mcpserver ./internal/storage ./internal/ingestion
go test ./test/integration/... -v -tags=integration -run 'TestDefaultIntegrationSuite/(TestProofPackScenariosEvidenceAndDeterminism|TestProofPackProductRecallPrioritizesTop2Evidence|TestSearchVideoDeterministicOrdering|TestSearchVideoIncludeDebugMetadata|TestToolResponseBackwardCompatFields)|TestWeightingIntegrationSuite/TestSearchVideoModalityWeightingBehavior' -count=1
```

Outcome: `PASS`

Release gates:

```bash
make release-gate
make release-gate-split
```

Recorded outcomes:
- `make release-gate`: pass (`/tmp/videra_phase20_release_gate.exit` = `0`)
- `make release-gate-split`: pass (`/tmp/videra_phase20_release_split.exit` = `0`)

## Contract and Scope Checks

- MCP contract drift: none detected (`index_video`, `get_index_job`, `search_video`, `list_videos`, transcript resource unchanged).
- Runtime dependency posture: no new required dependency introduced for default slim profile.
- LanceDB migration: not part of this phase (Phase 19 checkpoint decision preserved).

## Phase 20 Decision Snapshot

- Status: GO (for semantic quality lift implementation scope)
