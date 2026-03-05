# Pilot Corpus Benchmark Evidence — 2026-03-05

Date: 2026-03-05
Phase: 24 (Pilot Corpus Benchmark Pack)
Branch: main

## Inputs

- Fixture set: `internal/proofpack/fixtures/pilot_benchmark.json`
- Loader: `proofpack.LoadPilotBenchmarkScenarios()`
- Integration scorecard test: `TestDefaultIntegrationSuite/TestPilotBenchmarkScorecard`

Pilot scenarios (6 total):

- `pilot_engineering_planning`
- `pilot_legal_closeout`
- `pilot_product_recall`
- `pilot_animals_intro`
- `pilot_finance_review`
- `pilot_support_handoff`

## Commands Executed

Unit fixture validation:

```bash
go test ./internal/proofpack
```

Focused integration benchmark + real-mode guardrails:

```bash
go test ./test/integration/... -v -tags=integration -run 'TestDefaultIntegrationSuite/TestPilotBenchmarkScorecard|TestIndexVideoRealMode(RemotePathRespectsMaxSizeBound|RemotePathHonorsDisabledFetch|RequiresSidecarForLocalPath)' -count=1
```

## Baseline Metrics (Recorded)

From integration scorecard log:

- `scenarios=6`
- `evidenceMatchRate=1.00`
- `deterministicRate=1.00`
- `topTwoQualityRate=1.00`

Interpretation:

- Current pilot benchmark slice passes strict evidence matching.
- Repeated queries are deterministic across scenarios.
- Top-2 quality coverage is strong for this pilot scenario set.

## Contract/Scope Check

- MCP contract unchanged (`index_video`, `get_index_job`, `search_video`, `list_videos`, transcript resource).
- No storage backend or queue architecture changes introduced.

## Tuning Recommendation (Data-Backed)

Priority recommendation for next tuning pass:

- Preserve neutral default mapping and defer broad canonical expansion.
- Introduce domain mapping profiles only when pilot slices show measurable degradation (`topTwoQualityRate < 0.85` or evidence-match shortfall).
- Keep using paired OFF/ON benchmark comparisons before enabling any domain map in production-like configs.

Rationale:

- Current baseline already achieves `1.00` on all tracked metrics for this pilot slice, so expanding normalization heuristics now adds complexity without demonstrated benefit.

## GO/NO-GO Snapshot

- Status: **GO**
