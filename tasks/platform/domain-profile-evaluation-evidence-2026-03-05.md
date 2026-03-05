# Domain Profile Evaluation Evidence — 2026-03-05

Date: 2026-03-05
Phase: 22 (Domain Profile Evaluation Pack)
Branch: main

## Goal

Validate that:

1. neutral default behavior (`VIDERA_SEMANTIC_CANONICAL_MAP` unset) remains deterministic and usable on non-business/domain-diverse vocabulary,
2. optional domain mapping improves synonym-driven recall when explicitly enabled.

## Added Evaluation Fixtures

- `internal/proofpack/fixtures/domain_profiles.json`
  - `animals_literal_recall`
    - path: `https://example.com/proofpack-cats-cat.mp4`
    - query: `intro segment proofpack cats`
    - expected evidence: `intro segment`, `proofpack-cats`
  - `animals_synonym_profile`
    - path: `https://example.com/proofpack-cats-cat.mp4`
    - query: `kitty feline`
    - expected evidence: `proofpack-cats-cat`

Loader:

- `proofpack.LoadDomainProfileScenarios()` in `internal/proofpack/fixtures.go`

## Added/Updated Tests

- Unit fixture loader validation:
  - `internal/proofpack/harness_test.go`
  - `TestLoadDomainProfileScenariosFixture`

- Integration ON/OFF mapping comparison suite:
  - `test/integration/index_video_test.go`
  - `TestCanonicalMappingIntegrationSuite/TestDomainProfileNeutralLiteralRecall`
  - `TestCanonicalMappingIntegrationSuite/TestDomainProfileMappingImprovesSynonymRecall`

Comparison method for synonym profile:

- Starts two containers:
  - neutral default (`VIDERA_SEMANTIC_CANONICAL_MAP` unset)
  - mapped (`VIDERA_SEMANTIC_CANONICAL_MAP="kitty=cat,feline=cat,cats=cat"`)
- Indexes same video path in both.
- Searches same synonym query (`kitty feline`) in both.
- Asserts mapped result rank is better/equal and mapped similarity is strictly higher for the target evidence snippet.

## Validation Commands and Outcomes

Unit:

```bash
go test ./internal/proofpack
```

Outcome: `PASS`

Focused integration:

```bash
go test ./test/integration/... -v -tags=integration -run 'TestCanonicalMappingIntegrationSuite|TestDefaultIntegrationSuite/TestProofPackScenariosEvidenceAndDeterminism|TestDefaultIntegrationSuite/TestProofPackProductRecallPrioritizesTop2Evidence|TestDefaultIntegrationSuite/TestSearchVideoDeterministicOrdering|TestDefaultIntegrationSuite/TestToolResponseBackwardCompatFields' -count=1
```

Outcome: `PASS`

## Contract and Risk Check

- MCP tool/resource schemas unchanged.
- Determinism and backward-compat integration checks remain green.
- Domain mapping remains optional and explicitly configured; neutral default remains active when unset.

## Phase 22 Decision Snapshot

- Status: GO (evaluation pack complete, no contract drift)
