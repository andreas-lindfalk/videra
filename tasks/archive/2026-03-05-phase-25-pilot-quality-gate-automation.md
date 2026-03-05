# Videra Phase 25 — Pilot Quality Gate Automation (Archived 2026-03-05)

Status: GO

Reference:

- `Makefile`
- `README.md`
- `tasks/platform/pilot-corpus-benchmark-evidence-2026-03-05.md`
- `tasks/platform/pilot-quality-gate-evidence-2026-03-05.md`
- `test/integration/index_video_test.go`
- `test/integration/index_video_real_mode_test.go`

### Definition of Done (Target)

- [x] A single command exists to run pilot benchmark scorecard plus real-mode guardrails.
- [x] The command is documented in `README.md` with intended usage.
- [x] Focused validation of the new command path is recorded in a Phase 25 evidence artifact.
- [x] MCP contract and existing release-gate commands remain unchanged.
- [x] Phase archive note is prepared with GO/NO-GO outcome.

### Scope

- [x] **In scope:** add minimal make target(s) that compose existing integration checks.
- [x] **In scope:** document quality-gate command and phase artifact references.
- [x] **In scope:** run and record a reproducible validation command.
- [x] **Out of scope:** search/ranking algorithm changes, new MCP tools, storage or queue architecture changes.

### Implementation Plan

- [x] Add `make` target for pilot quality-gate composition.
- [x] Update `README.md` with the new command and intent.
- [x] Execute focused validation and capture results in an evidence note.
- [x] Close phase with lessons update and archive snapshot.
