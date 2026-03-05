# Pilot Quality Gate Evidence — 2026-03-05

Date: 2026-03-05
Phase: 25 (Pilot Quality Gate Automation)
Branch: main

## Objective

Operationalize a single reproducible command that runs:

- pilot benchmark scorecard (`TestPilotBenchmarkScorecard`)
- real-mode ingestion guardrails (`RemotePathRespectsMaxSizeBound`, `RemotePathHonorsDisabledFetch`, `RequiresSidecarForLocalPath`)

## Command Executed

```bash
make pilot-quality-gate
```

Underlying test invocation (from Makefile target):

```bash
go test ./test/integration/... -v -tags=integration -run 'TestDefaultIntegrationSuite/TestPilotBenchmarkScorecard|TestIndexVideoRealMode(RemotePathRespectsMaxSizeBound|RemotePathHonorsDisabledFetch|RequiresSidecarForLocalPath)' -count=1
```

## Results

- Exit status: success (PASS)
- Integration package result: `ok github.com/andreas-lindfalk/videra/test/integration`
- Pilot benchmark scorecard log:
  - `scenarios=6`
  - `evidenceMatchRate=1.00`
  - `deterministicRate=1.00`
  - `topTwoQualityRate=1.00`
- Real-mode guardrails:
  - `TestIndexVideoRealModeRemotePathRespectsMaxSizeBound` PASS
  - `TestIndexVideoRealModeRemotePathHonorsDisabledFetch` PASS
  - `TestIndexVideoRealModeRequiresSidecarForLocalPath` PASS

## Contract/Scope Check

- MCP API surface unchanged.
- Existing release-gate commands unchanged.
- Change is operational only (`Makefile` + docs + evidence).

## GO/NO-GO Snapshot

- Status: **GO**
