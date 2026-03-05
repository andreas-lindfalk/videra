# Videra Roadmap & End-State — 2026-03-05

Purpose: make the destination explicit, show how completed phases contribute, and define the next decisions.

Status: Active orientation artifact.

## North Star (Target State)

Videra should be a release-ready MCP video memory service where:

1. MCP contract is stable and backward-compatible (`index_video`, `get_index_job`, `search_video`, `list_videos`, transcript resource).
2. Retrieval quality is measurable and repeatable on pilot and real-corpus gates.
3. Operational readiness is one-command reproducible with evidence-backed GO/NO-GO artifacts.
4. Deployment path remains privacy-native and portable (local/on-prem + cloud parity).

## What "Done" Looks Like

For current roadmap horizon, "done" means all of the following are true:

- **Contract confidence:** no MCP breaking changes across release gates.
- **Quality confidence:** pilot quality gate remains green and real-corpus gate thresholds are met.
- **Ops confidence:** release evidence includes build/test/integration and quality-gate outcomes.
- **Decision confidence:** backend/storage checkpoint is resolved with benchmark + portability criteria, not intuition.

## Phase Map (Why each phase existed)

### Track A — Runtime Foundation (Phases 9–18)

Outcome delivered:

- Async indexing boundaries, split-role runtime behavior, queue hardening, release gate repeatability, and RC stabilization.

Why it matters:

- Establishes reliability and deployability before semantic tuning.

### Track B — Retrieval Quality Foundation (Phases 20–24)

Outcome delivered:

- Semantic reranking lift, neutral-default normalization, optional domain mapping, real-corpus guardrails, and pilot benchmark scorecard.

Why it matters:

- Converts retrieval quality from subjective assessment to measured signals.

### Track C — Quality Gate Operations (Phase 25)

Outcome delivered:

- One-command pilot quality gate (`make pilot-quality-gate`) combining benchmark and real-mode guardrails.

Why it matters:

- Gives fast, repeatable quality status without running the full suite.

## Current Position (2026-03-05)

- Runtime/release foundation: in place.
- Pilot quality gate: in place and green.
- Release evidence + quality signal coupling: in place (Phase 27).
- Real-corpus promotion criteria + evidence format: in place (Phase 28).
- Storage decision re-checkpoint: completed with NO-GO for immediate migration (Phase 29).
- Remaining gap: promotion workflow consolidation before any future backend migration decision refresh.

## Next Planned Phases

### Phase 30 — Deployment Promotion Runbook Consolidation

Goal:

- Consolidate release + quality + real-corpus promotion commands into one operator-facing promotion runbook.

Expected outcome:

- One concise promotion workflow for recurring release/deployment decisions.

### Phase 31 — Conditional Storage Migration Track (Only if a future checkpoint = GO)

Goal:

- Execute a controlled migration plan with rollback safety, preserving MCP contract compatibility.

Expected outcome:

- Measured migration outcome with parity/quality validation and documented rollback readiness.

### Phase 32 — Storage Benchmark Harness (Decision Refresh Input)

Goal:

- Produce repeatable backend comparison benchmarks required for a future storage checkpoint re-open.

Expected outcome:

- Measured artifact set that can satisfy migration GO criteria when/if benefits are proven.

## Guardrails for Upcoming Work

- Do not change MCP contract unless explicitly planned as a compatibility phase.
- Keep domain mapping optional and evidence-driven.
- Keep provider-specific deployment concerns outside core MCP/search logic.
- Do not start storage migration implementation before checkpoint decision.

## Primary References

- `tasks/platform/spec-implementation-alignment-2026-03-05.md`
- `tasks/platform/reflection-intake-2026-03-05.md`
- `tasks/platform/pilot-corpus-benchmark-evidence-2026-03-05.md`
- `tasks/platform/pilot-quality-gate-evidence-2026-03-05.md`