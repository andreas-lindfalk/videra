# Real Corpus Quality Gate Definition — 2026-03-05

Purpose: define objective, testable gate metrics for real-corpus onboarding.

## Gate Metrics

### G1 — Top-k Evidence Quality

- For each selected scenario query, expected evidence terms must be present in the top-k result snippets.
- Baseline threshold:
  - `matchedEvidence >= len(expectedEvidence)` for scenario assertions designed as strict evidence checks.
- Existing proof path references:
  - `TestDefaultIntegrationSuite/TestProofPackScenariosEvidenceAndDeterminism`
  - `TestDefaultIntegrationSuite/TestProofPackProductRecallPrioritizesTop2Evidence`

### G2 — Deterministic Ordering

- Repeated identical query against same indexed corpus returns identical ordered results.
- Baseline threshold:
  - run A results exactly equal run B results.
- Existing proof path references:
  - `TestDefaultIntegrationSuite/TestSearchVideoDeterministicOrdering`

### G3 — Real-Mode Source Constraint Semantics

- Real-mode source constraints must fail safely and explicitly:
  - local path without sidecar transcript => explicit sidecar-required error
  - remote path with fetch disabled => explicit disabled-fetch error
  - remote path over max-size bound => explicit size-limit error
- Proof path references:
  - `TestIndexVideoRealModeRequiresSidecarForLocalPath`
  - `TestIndexVideoRealModeRemotePathHonorsDisabledFetch`
  - `TestIndexVideoRealModeRemotePathRespectsMaxSizeBound`

### G4 — Contract Stability

- MCP tool/resource schema unchanged:
  - `index_video`, `get_index_job`, `search_video`, `list_videos`, transcript resource.

## GO/NO-GO Rule

- **GO** when G1–G4 all pass in focused validation run with documented command outputs.
- **NO-GO** if any gate fails or if failures are non-deterministic/unreproducible.
