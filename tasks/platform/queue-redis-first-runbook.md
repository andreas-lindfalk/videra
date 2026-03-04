# Queue Operations Runbook (Redis-First with NATS Fallback)

Goal: operate async indexing queue topology with Redis as the first external backend when moving beyond in-process mode, while preserving a reversible path to NATS and back to in-process fallback.

## 1) Topology Modes

- `all` role (default): one process runs MCP API + worker loop.
- `api` role: MCP API only (`index_video` enqueue + `get_index_job` status reads).
- `worker` role: queue processor only.

Contract reminder:

- `index_video` / `get_index_job` API contract must remain unchanged across topology modes.

## 2) Redis-First Recommended Baseline

Use when Redis is already part of the stack or minimizing moving parts is the priority.

Required env:

- `VIDERA_JOBQUEUE_BACKEND=redis`
- `VIDERA_JOBQUEUE_ROLE=all|api|worker`
- `VIDERA_JOBQUEUE_REDIS_ADDR=<host:port>`
- `VIDERA_JOBQUEUE_REDIS_STREAM=<stream>`
- `VIDERA_JOBQUEUE_REDIS_GROUP=<consumer-group>`
- `VIDERA_JOBQUEUE_REDIS_CONSUMER=<consumer-name>`
- `VIDERA_JOBSTATE_REDIS_PREFIX=<prefix>`
- `VIDERA_JOBQUEUE_RETRY_MAX_ATTEMPTS=<n>`
- `VIDERA_JOBQUEUE_RETRY_BACKOFF_MS=<ms>`
- `VIDERA_JOBQUEUE_WORKER_POLL_MS=<ms>`

Guardrails:

- `VIDERA_JOBQUEUE_ROLE=worker` must run with `VIDERA_TRANSPORT=stdio` (startup fail-fast if set to `http`).

Recommended starter values:

- Retry attempts: `3`
- Retry backoff: `250ms`
- Worker poll: `250ms`

## 3) Single-Host Private Deployment Shape (Hetzner-like)

- Run Redis + Videra in Docker Compose on the same host.
- Start with `role=all` for simplicity.
- Move to split `api`/`worker` only when concurrency isolation is needed.

Split-mode requirement:

- API and worker must share both queue backend and job-state backend configuration.
- If `search_video`/`list_videos` are expected to reflect worker-ingested data, ensure a shared storage/data-plane strategy is in place for indexed content.

## 4) Cloud Run-Aligned Split Mode

- API service: `VIDERA_JOBQUEUE_ROLE=api`
- Worker service/job: `VIDERA_JOBQUEUE_ROLE=worker`
- Both must point to same Redis stream/group + job-state prefix.

## 5) Operational Checks

Health checks after rollout:

1. `index_video` with `mode=async` returns `jobId` quickly.
2. `get_index_job` transitions `pending -> completed|failed`.
3. Failure case reaches deterministic terminal status with retry exhaustion message.
4. Queue contract checks pass for duplicate delivery/idempotency-safe behavior.
5. Worker logs contain structured `queue_lifecycle` events for enqueue/reserve/retry/completion/terminal-failure flow.

Verification commands:

```bash
make build && make test
go test ./test/integration/... -v -tags=integration -run 'TestRedisStreamsJobQueueContractIntegration|TestIndexVideoAsyncSplitRoleRedisLifecycle' -count=1

# startup guardrail (worker role + invalid transport should fail fast)
go test ./test/integration/... -v -tags=integration -run 'TestWorkerRoleWithHTTPTransportFailsFastAtStartup' -count=1
```

Structured worker log schema:

- Prefix: `queue_lifecycle`
- Keys: `event`, `job_id`, `status`, `attempt`, `max_attempts`, `delay_ms`, `error`
- Expected events during normal retry flow: `enqueued` → `reserved` → (`retry_scheduled`)* → `completed|retry_exhausted`

## 6) Security Model (Queue Layer)

- Transport/API authN/authZ belongs to the edge/gateway layer (not queue internals).
- Redis auth/TLS/secret handling is deployment responsibility:
  - Do not hardcode credentials.
  - Prefer secret injection via environment/secret manager.
  - Restrict network exposure to trusted private networks.
- Keep queue payloads metadata-only (no media bytes) to reduce sensitive data propagation.

## 7) Fallback Trigger Criteria (Redis -> NATS)

Consider switching queue backend to NATS when one or more apply:

- Redis operational load from mixed KV + queue workloads becomes a stability bottleneck.
- Queue semantics/operations require stronger broker specialization than Redis Streams conventions.
- Team prefers dedicated queue infrastructure and can absorb extra operational surface.

Fallback activation is config-only backend switch (contract remains unchanged).

## 8) Rollback Drill (External Backend -> In-Process)

Purpose: prove no MCP contract break under emergency rollback.

Rollback steps:

1. Set `VIDERA_JOBQUEUE_BACKEND=inprocess`.
2. Set `VIDERA_JOBQUEUE_ROLE=all`.
3. Remove/ignore external queue/job-state backend env vars.
4. Restart service.
5. Run async lifecycle checks:
   - `index_video` async success/failure
   - `get_index_job` polling

Rollback verification command:

```bash
go test ./test/integration/... -v -tags=integration -run 'TestDefaultIntegrationSuite/TestIndexVideoAsyncLifecycleSuccess|TestDefaultIntegrationSuite/TestIndexVideoAsyncLifecycleFailure' -count=1
```

Success criterion:

- Async behavior remains contract-compatible with no MCP schema/tool changes.

## 9) Decision Summary

- Default remains `inprocess` until explicit scale/SLO pressure.
- First external candidate is Redis when stack consolidation is prioritized.
- NATS remains a parity-tested portability fallback.
