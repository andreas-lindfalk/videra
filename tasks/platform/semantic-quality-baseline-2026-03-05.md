# Semantic Quality Baseline — 2026-03-05

Purpose: establish a measurable pre-change baseline for Phase 20 semantic quality lift.

## Scope of Baseline

Focused integration signals used as baseline:
- proofpack evidence matching + deterministic repeatability,
- deterministic ordering for repeated identical queries,
- debug metadata availability for search scoring diagnostics,
- backward-compatible response field contract,
- modality weighting behavior under extreme audio/visual weights.

## Command Executed

```bash
go test ./test/integration/... -v -tags=integration -run 'TestDefaultIntegrationSuite/(TestProofPackScenariosEvidenceAndDeterminism|TestSearchVideoDeterministicOrdering|TestSearchVideoIncludeDebugMetadata|TestToolResponseBackwardCompatFields)|TestWeightingIntegrationSuite/TestSearchVideoModalityWeightingBehavior' -count=1
```

## Result Snapshot

- Overall result: `PASS`
- Package result: `ok github.com/andreas-lindfalk/videra/test/integration 34.741s`

Passed test groups:
- `TestDefaultIntegrationSuite`
  - `TestProofPackScenariosEvidenceAndDeterminism`
    - `engineering_incident_review`
    - `legal_compliance_evidence_lookup`
    - `product_interview_recall`
  - `TestSearchVideoDeterministicOrdering`
  - `TestSearchVideoIncludeDebugMetadata`
  - `TestToolResponseBackwardCompatFields`
- `TestWeightingIntegrationSuite`
  - `TestSearchVideoModalityWeightingBehavior`

## Proofpack Fixture Baseline Set

From `internal/proofpack/fixtures/scenarios.json`:
- `engineering_incident_review`
  - query: `roadmap budget next actions`
  - expected evidence: `roadmap`, `budget`, `next actions`
  - min results: `3`
- `legal_compliance_evidence_lookup`
  - query: `closing remarks and next actions`
  - expected evidence: `closing remarks`, `next actions`
  - min results: `2`
- `product_interview_recall`
  - query: `intro segment discussion`
  - expected evidence: `intro segment`, `discussion`
  - min results: `2`

## Baseline Interpretation

- Current implementation satisfies deterministic and compatibility gates for selected semantic-quality signals.
- This file is the pre-change reference for Phase 20 improvement comparison.
- Improvement work must keep these tests green while increasing evidence quality in selected scenarios.

## Notes

- `runTests` tool did not execute build-tagged integration tests in this case (0 tests reported), so baseline execution used explicit `go test` with `-tags=integration`.
