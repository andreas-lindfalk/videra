# Todo

## Active Task

### Videra Phase 6 — Cloud Deployment Parity (Cloud Run + Hetzner)

### Definition of Done (Applied)

- [ ] Scope is explicit (in/out) and aligned with current architecture decisions.
- [ ] Required interfaces/contracts are updated without breaking existing MCP surface.
- [ ] Fast tests pass (`make test`).
- [ ] Integration tests pass (`make integration-test`).
- [ ] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [ ] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [ ] Todo is fully checked and ready for archive.

### Scope

- [ ] **In scope:** Define and implement deployment parity where Hetzner (EU-first self-hosted path) is equally supported and documented alongside GCP/Cloud Run.
- [ ] **In scope:** Keep the same MCP interface and core runtime behavior across both deployment paths.
- [ ] **Out of scope:** Provider-specific lock-in features that cannot be mapped to the other platform.

### Deliverables

- [x] Add a capability matrix mapping Cloud Run concepts to Hetzner equivalents and constraints.
- [x] Produce a minimal production-ready Hetzner deployment runbook (single-node Docker first, K8s optional path).
- [x] Produce a matching Cloud Run runbook with equivalent operational checkpoints.
- [x] Define a parity validation checklist (index/search/list/transcript + persistence + restart behavior) that passes on both platforms.

### Execution Guardrail (Agreed)

- [x] Phase 6 implementation work is timeboxed to planning/docs baseline only.
- [x] No deep platform-specific build-out on Cloud Run/Hetzner until local semantic ingestion is satisfactory.
- [x] Start execution with local semantic ingestion track before additional cloud platform expansion.

### Critical Near-Term Track (Address Next)

- [ ] Replace simulated transcript/visual placeholder content with real semantic ingestion so retrieval quality reflects actual video content.
- [ ] Keep this track as the next execution priority after baseline deployment parity artifacts are in place.

### Real Semantic Ingestion (Next Phase Candidate)

- [x] Add a configurable ingestion mode (`simulated` for tests/dev fixtures, `real` for actual content extraction) without breaking MCP contracts.
- [x] Implement interim real transcript path using sidecar transcript files (`.srt`, `.vtt`, `.txt`) in `real` mode for local quality validation.
- [ ] Implement real transcription path from extracted audio (FFmpeg output) and persist transcript segments with source timestamps.
- [x] Implement real visual-context extraction path for keyframes (captioning or equivalent textual representation) and persist visual segments.
- [x] Ensure `search_video` uses real segment text/embeddings in `real` mode and keeps deterministic ordering guarantees.
- [ ] Add a retrieval-quality validation checklist using a known local clip with expected phrase/object queries.

### Real Semantic Ingestion — Acceptance Criteria

- [ ] For at least one known clip, transcript output includes non-placeholder text grounded in the clip content.
- [ ] Querying a known spoken phrase returns a top hit within an expected timestamp window.
- [ ] Querying a known visual cue returns at least one plausible timestamped hit.
- [ ] Existing integration tests remain green; new tests cover both `simulated` and `real` mode behavior boundaries.

### Product/Tech Work

- [ ] Keep storage and ingestion orchestration boundaries backend-agnostic (no Cloud Run-only assumptions in core packages).
- [ ] Define environment/config contract that is identical across providers where possible.
- [ ] Add deployment artifacts/scripts for both targets without altering MCP tool/resource contracts.

### Testing (High-Rigor)

- [ ] Add platform verification steps that prove parity for identical fixtures and query expectations.
- [ ] Add regression guardrails to prevent platform-specific behavior drift in API responses.

### Verification

- [ ] `make build` passes.
- [ ] `make test` passes.
- [ ] `make integration-test` passes.
- [ ] `make docker-build` passes.

## Definition of Done Template

Use this checklist at the top of each new phase/feature todo.

- [ ] Scope is explicit (in/out) and aligned with current architecture decisions.
- [ ] Required interfaces/contracts are updated without breaking existing MCP surface.
- [ ] Fast tests pass (`make test`).
- [ ] Integration tests pass (`make integration-test`).
- [ ] Build succeeds (`make build`) and, when relevant, container build succeeds (`make docker-build`).
- [ ] `tasks/lessons.md` updated with at least one concrete learning (if applicable).
- [ ] Todo is fully checked and ready for archive.
