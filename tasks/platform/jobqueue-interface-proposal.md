# JobQueue Interface Proposal (Phase 10)

Goal: define a broker-agnostic queue boundary for async indexing jobs so broker choice is a deployment concern, not an MCP contract concern.

## Design Principles

- Preserve existing MCP behavior from Phase 9 (`index_video`, `get_index_job`).
- Keep in-process execution as first-class adapter.
- Make delivery semantics explicit (enqueue, lease/reserve, ack, retry, terminal failure).
- Keep payload idempotency-oriented around source identity.

## Proposed Domain Model

```go
type JobEnvelope struct {
	JobID       string
	SourcePath  string
	RequestedAt time.Time
	Attempt     int
	MaxAttempts int
}

type Lease struct {
	JobID      string
	Receipt    string
	LeasedUntil time.Time
}

type JobQueue interface {
	Enqueue(ctx context.Context, job JobEnvelope) error
	Reserve(ctx context.Context, wait time.Duration) (JobEnvelope, Lease, bool, error)
	Ack(ctx context.Context, lease Lease) error
	Retry(ctx context.Context, lease Lease, cause string, nextDelay time.Duration) error
	Fail(ctx context.Context, lease Lease, cause string) error
}
```

Notes:

- `Receipt` is opaque adapter-specific acknowledgment token.
- `Reserve` returns `ok=false` when no job is available in `wait` window.
- `Retry` is for bounded recoverable failures; `Fail` is terminal.

## Lifecycle Mapping

Queue lifecycle to existing index job statuses:

- Enqueued -> `pending`
- Reserved/processing -> `pending`
- Ack after successful ingestion -> `completed`
- Fail or retry budget exhausted -> `failed`

`get_index_job` remains the single status-read surface; queue internals are not exposed directly.

## Idempotency Contract

- Idempotency key: normalized `sourcePath` (same as current lookup strategy).
- Duplicate deliveries must be safe: worker checks `GetVideoBySourcePath` before ingesting.
- Ack happens only after terminal state write is persisted.

## Error and Retry Semantics

- Default bounded attempts: 2-3 (configurable per adapter).
- Retry path only for transient errors (network/storage/temporary runtime failures).
- Non-recoverable validation errors go directly to `Fail`.
- Last failure cause is persisted to job state for `get_index_job` responses.

## Adapter Strategy

- `InProcessJobQueue` (default): channel + in-memory lease tracking; used for local and baseline deployments.
- `NATSJetStreamJobQueue` (candidate): stream/consumer-backed reserve/ack/retry semantics.
- `RedisStreamsJobQueue` (fallback candidate): consumer-group semantics with bounded retries.

All adapters must satisfy the same contract tests.

## Migration Plan (No MCP Breakage)

1. Introduce `JobQueue` interface and in-process adapter behind orchestrator.
2. Add worker loop abstraction using `Reserve`/`Ack`/`Retry`/`Fail`.
3. Keep `index_video` async initiation unchanged.
4. Keep `get_index_job` as source of truth for lifecycle state.
5. Add adapter-specific wiring via config only after go/no-go checkpoint passes.

## Test Requirements for Future Implementation

- Contract tests shared across all adapters:
  - enqueue/reserve/ack happy path
  - retry then success path
  - retry budget exhausted -> terminal failed
  - duplicate delivery idempotency
  - lease timeout/visibility behavior (if adapter supports it)
- Integration scenario proving MCP response compatibility is unchanged.

## Non-Goals (Phase 10)

- No broker package imports in runtime code.
- No deployment manifest changes for broker infrastructure.
- No changes to MCP API shape.
