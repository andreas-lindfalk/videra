# Todo

## Active Task

### Videra Phase 7 — Container Runtime Profiles (Slim + Full)

Reference:

- `tasks/platform/container-runtime-profiles.md`

### Definition of Done (Applied)

- [x] Scope is explicit (in/out) and aligned with current architecture decisions.
- [x] Required interfaces/contracts are updated without breaking existing MCP surface.
- [x] Fast tests pass (`make test`).
- [x] Integration tests pass (`make integration-test`).
- [x] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [x] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [x] Todo is fully checked and ready for archive.

### Scope

- [x] **In scope:** Define and implement a stable container strategy with explicit runtime profiles:
	- **slim** (default): ffmpeg + current core runtime dependencies
	- **full**: slim + whisper-capable runtime deps (`python3` + whisper path) + OCR dependency (`tesseract`)
- [x] **In scope:** Keep MCP tool/resource contracts unchanged across image profiles.
- [x] **Out of scope:** GPU inference images, cloud-provider-specific image logic, and major ingestion pipeline redesign.

### Deliverables

- [x] Add Docker build targets (or equivalent) for `slim` and `full` images with clear tagging conventions.
- [x] Keep default local dev flow simple and deterministic (retain `slim` as default unless explicitly overridden).
- [x] Add runtime capability signaling/logging to make missing optional dependencies obvious at startup.
- [x] Document profile selection and expected behavior in README + deployment runbooks.

### Acceptance Criteria

- [x] `slim` image builds and runs current default flows without regression.
- [x] `full` image builds and supports real-mode fallback paths that require external tools.
- [x] Missing optional tooling behavior is explicit (clear logs/errors), not silent.
- [x] Existing integration behavior remains green and deterministic.

### Execution Notes

- [x] Prioritize clarity and reproducibility over optimization.
- [x] Prefer pinned versions/digests and update cadence notes to avoid security drift.

### Implementation Plan

- [x] Add multi-target Dockerfile stages for `runtime-slim` and `runtime-full` with compatible defaults.
- [x] Add Makefile build/run targets for `slim` and `full` image profiles while keeping existing developer commands working.
- [x] Add runtime capability detection/logging so missing optional dependencies are explicit at startup.
- [x] Update README and deployment runbooks with profile selection guidance.
- [x] Verify with build/test/integration plus both Docker profile builds.

## Definition of Done Template

Use this checklist at the top of each new phase/feature todo.

- [ ] Scope is explicit (in/out) and aligned with current architecture decisions.
- [ ] Required interfaces/contracts are updated without breaking existing MCP surface.
- [ ] Fast tests pass (`make test`).
- [ ] Integration tests pass (`make integration-test`).
- [ ] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [ ] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [ ] Todo is fully checked and ready for archive.
