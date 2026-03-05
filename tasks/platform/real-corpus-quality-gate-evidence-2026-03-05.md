# Real Corpus Quality Gate Evidence — 2026-03-05

Date: 2026-03-05
Phase: 23 (Real Corpus Onboarding & Quality Gate)
Branch: main

## Gate Definition References

- `tasks/platform/real-corpus-onboarding-checklist-2026-03-05.md`
- `tasks/platform/real-corpus-quality-gate-2026-03-05.md`

## Implemented Additions

1. Added explicit real-mode source-constraint guardrail test:
   - `TestIndexVideoRealModeRemotePathHonorsDisabledFetch`
   - file: `test/integration/index_video_real_mode_test.go`
2. Added onboarding checklist artifact for repeatable real-corpus intake:
   - `tasks/platform/real-corpus-onboarding-checklist-2026-03-05.md`
3. Added objective quality gate definition with GO/NO-GO rules:
   - `tasks/platform/real-corpus-quality-gate-2026-03-05.md`
4. Added README runnable validation command:
   - section: `Real Corpus Quality Gate (Phase 23)`

## Validation Commands and Outcomes

Real-mode source-constraint focused integration:

```bash
go test ./test/integration/... -v -tags=integration -run 'TestIndexVideoRealMode(RemotePathRespectsMaxSizeBound|RemotePathHonorsDisabledFetch|RequiresSidecarForLocalPath)' -count=1
```

Outcome: `PASS`

Determinism + top-k evidence + backward-compat focused integration:

```bash
go test ./test/integration/... -v -tags=integration -run 'TestDefaultIntegrationSuite/(TestProofPackScenariosEvidenceAndDeterminism|TestProofPackProductRecallPrioritizesTop2Evidence|TestSearchVideoDeterministicOrdering|TestToolResponseBackwardCompatFields)' -count=1
```

Outcome: `PASS`

## Contract Check

- MCP schema unchanged:
  - `index_video`
  - `get_index_job`
  - `search_video`
  - `list_videos`
  - transcript resource

## GO/NO-GO Snapshot

- Gate G1 (top-k evidence quality): PASS
- Gate G2 (deterministic ordering): PASS
- Gate G3 (real-mode source constraint semantics): PASS
- Gate G4 (contract stability): PASS

Decision: **GO**
