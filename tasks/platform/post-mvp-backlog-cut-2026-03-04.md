# Post-MVP Backlog Cut — 2026-03-04

Purpose: isolate non-MVP work from RC1 stabilization/release scope.

Scope rule:

- Items in this file are intentionally deferred and must not be merged into RC1 stabilization unless they become release blockers.

## P1 — Next Release Window (High Priority)

1. Cloud parity execution evidence capture in real deployed environments
   - Why deferred: MVP gate validated locally; environment-side parity runs are operationally separate.
   - Risk if delayed: medium (slower enterprise confidence).
   - Suggested owner: platform/devops.

2. Release tagging + artifact publication workflow
   - Why deferred: RC1 stabilization focused on gate repeatability, not release automation.
   - Risk if delayed: medium (manual release overhead).
   - Suggested owner: maintainers/release engineering.

3. Expanded failure taxonomy for release gate outcomes
   - Why deferred: current troubleshooting covers primary Docker pressure path; broader taxonomy can iterate safely.
   - Risk if delayed: low-medium.
   - Suggested owner: platform + QA.

## P2 — Hardening / Scale Readiness (Medium Priority)

1. Optional CI execution of `release-gate` + split-role checks
   - Why deferred: local reproducibility was the immediate RC1 requirement.
   - Risk if delayed: medium (less automation confidence).
   - Suggested owner: platform/CI maintainers.

2. More granular queue backend SLO/error-budget dashboards
   - Why deferred: current structured lifecycle logs are sufficient for MVP operations.
   - Risk if delayed: medium in scaled deployments, low in current footprint.
   - Suggested owner: platform/observability.

3. Additional remote-ingestion stress fixtures (large/slow content matrix)
   - Why deferred: existing bounded-fetch paths are covered for MVP; stress expansion is iterative.
   - Risk if delayed: low-medium.
   - Suggested owner: ingestion/test.

## P3 — Strategic / Optional (Lower Priority)

1. Cloud object-storage backend proof spike behind `storage.VectorStore` contract
   - Why deferred: would expand architecture scope beyond RC stabilization.
   - Risk if delayed: low for current MVP.
   - Suggested owner: architecture/platform.

2. Managed queue vendor comparative rerun with updated production-like load
   - Why deferred: Redis-first + NATS fallback decision is already documented and validated.
   - Risk if delayed: low-medium.
   - Suggested owner: platform.

## Entry / Exit Rules

- Add a deferred item only if it is non-blocking for current release criteria.
- Promote to active phase only when explicitly prioritized and scoped with acceptance criteria.
- Remove/close item only after evidence is captured in runbooks or tests.
