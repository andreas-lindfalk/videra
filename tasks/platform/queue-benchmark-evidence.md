# Queue Benchmark Evidence (Phase 13)

Goal: provide reproducible, implementation-aligned evidence for Redis-first queue rollout readiness while preserving NATS fallback portability.

## Environment Snapshot

- Date: 2026-03-04
- Runtime: local Docker Desktop + testcontainers
- Focus: queue lifecycle behavior and async indexing paths (not model-quality retrieval benchmarks)

## Reproducible Commands

Focused queue evidence sweep:

```bash
go test ./test/integration/... -v -tags=integration \
  -run 'TestNATSJetStreamJobQueueContractIntegration|TestRedisStreamsJobQueueContractIntegration|TestIndexVideoAsyncSplitRoleRedisLifecycle|TestDefaultIntegrationSuite/TestIndexVideoAsyncLifecycleSuccess|TestDefaultIntegrationSuite/TestIndexVideoAsyncLifecycleFailure' \
  -count=1
```

Full integration regression confirmation:

```bash
make integration-test
```

Fast unit/build confirmation:

```bash
make build && make test
```

## Baseline Results (Measured)

From the focused queue evidence run (`go test ... -run ... -count=1`):

| Scenario | Result | Duration |
|---|---|---:|
| `TestDefaultIntegrationSuite/TestIndexVideoAsyncLifecycleSuccess` (in-process baseline) | PASS | 0.06s |
| `TestDefaultIntegrationSuite/TestIndexVideoAsyncLifecycleFailure` (in-process terminal semantics) | PASS | 0.78s |
| `TestNATSJetStreamJobQueueContractIntegration` | PASS | 0.62s |
| `TestRedisStreamsJobQueueContractIntegration` | PASS | 0.62s |
| `TestIndexVideoAsyncSplitRoleRedisLifecycle` (api/worker split) | PASS | 14.72s |
| Focused evidence run total | PASS | 43.186s |

## Interpretation

- Redis and NATS both satisfy baseline queue contract semantics with near-identical lifecycle test durations in this environment.
- Split-role Redis flow is stable and deterministic; terminal retry exhaustion semantics are observable through `get_index_job`.
- In-process async baseline remains intact, supporting rollback confidence.

## Evidence Scope Notes

- This evidence covers correctness + operational readiness in local/containerized conditions.
- It does not represent global latency SLO commitments for production traffic.
- Retrieval/search latency in shared-storage cloud deployments should be measured separately once final data-plane topology is selected.

## Decision Utility

This benchmark evidence is intended to satisfy queue go/no-go checklist requirements for:

- candidate benchmarking,
- failure semantics verification,
- and rollback confidence in existing MCP contract behavior.
