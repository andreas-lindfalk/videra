# Queue Vendor Checkpoint (Phase 10)

Goal: decide whether/when to introduce an external queue backend for async indexing, without breaking MCP contracts or private deployment portability.

## Scope and Guardrails

- In scope: vendor comparison, lock-in analysis, fallback path, and go/no-go criteria.
- Out of scope: implementing broker adapters or changing MCP tool contracts.
- Baseline remains in-process orchestration from Phase 9.

## Current Boundary (Must Hold)

- MCP contract remains `index_video` (sync/async initiation) + `get_index_job` (polling).
- Queueing is an internal orchestration concern behind a broker-agnostic interface.
- Private/self-hosted deployments must keep a first-class path with no cloud-specific dependency.

## Candidate Matrix

Scoring: 1 (weak) to 5 (strong), based on Videra constraints in `AGENTS.md`.

| Candidate | Local Bootstrap | Private/On-Prem Fit | Retry/Ack Semantics | HA Path | Lock-In Risk | Go Ecosystem Maturity | Notes |
|---|---:|---:|---:|---:|---:|---:|---|
| NATS JetStream | 5 | 5 | 5 | 4 | 5 | 5 | Strong neutral default for event/job workflows; lightweight ops footprint.
| Redis Streams | 4 | 4 | 4 | 4 | 4 | 5 | Good fallback option; broad ops familiarity; semantics are solid but stream/group operations require discipline.
| RabbitMQ (Quorum Queues) | 3 | 4 | 5 | 4 | 5 | 4 | Mature queue model; heavier operational profile for small teams.

## Recommendation

1. Keep in-process queueing as default baseline until explicit scale/SLO pressure appears.
2. Use **NATS JetStream** as first external broker candidate for Phase 11 spike.
3. Keep **Redis Streams** as fallback neutral/self-hosted alternative if NATS operational assumptions fail in target environments.

## Fallback and Portability Plan

- Keep `JobQueue` broker adapters internal and swappable.
- Preserve queue-agnostic payload schema (`jobId`, `sourcePath`, `mode`, timestamps, attempts).
- Require adapter parity tests for ack/nack/retry/dead-letter behavior before enabling any backend by default.
- Keep in-process adapter always available as a no-external-dependency mode.

## Go / No-Go Checklist (Before Broker Implementation)

All items must be satisfied before implementation begins:

- [ ] Candidate benchmarked in local Docker and private VM flow.
- [ ] Failure semantics proven: retry budget, backoff, poison message/dead-letter handling.
- [ ] Idempotency strategy documented and tested for duplicate delivery.
- [ ] Operational runbook exists for single-node and HA deployment.
- [ ] Security model documented (authN/authZ/TLS, secret handling, audit considerations).
- [ ] Rollback path documented to in-process mode with no MCP contract changes.

## Decision Record Entry Criteria

Any broker selection PR must include:

- Candidate comparison summary with rubric scores and rationale.
- Link to reproducible test evidence (latency, reliability, retry behavior).
- Explicit lock-in and portability analysis.
- Fallback trigger definition (when to switch adapter/back out).
- Runbook links for Cloud Run-aligned and Hetzner/private deployments.

## Phase Boundary Outcome

Phase 10 decision: **No broker integration yet**. Proceed with documentation + interface planning only.
