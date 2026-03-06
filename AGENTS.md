# Videra RAG-Go MCP Server

A high-performance MCP server written in Go that provides deep searchability in video files using RAG (Retrieval-Augmented Generation).

## Tech Stack (Decided)

- **Go 1.25.4** — module path: `github.com/andreas-lindfalk/videra`
- **MCP SDK:** `github.com/mark3labs/mcp-go` (v0.44+) — pure MCP protocol, supports stdio + streamable HTTP
- **Vector Store:** `github.com/philippgille/chromem-go` — embedded, pure Go, zero deps, file-persistent
  - Default backend remains `chromem-go` for zero-dependency local/runtime portability.
  - `lancedb` backend is available behind runtime toggle; current path in repo uses adapter isolation to preserve portability.
- **Media Processing:** FFmpeg (in Docker) for audio/frame extraction
- **Deployment:** Docker (local), Cloud Run (future)

## Architecture

### Project Structure

```
cmd/videra/          — Entrypoint: wire deps, select transport, start server
internal/
  config/            — Configuration from VIDERA_ env vars
  mcpserver/         — MCP server setup, tool/resource registration
  ingestion/         — Pipeline: extract → transcribe → embed → store
  storage/           — VectorStore interface + chromem-go implementation
test/integration/    — Testcontainers-based integration tests
```

### Design Patterns

- **Interface-driven:** Core abstractions (`VectorStore`, `Ingester`, `FFmpegRunner`) are interfaces. Implementations are injected at startup.
- **No global state:** All dependencies wired in `main.go` and passed explicitly.
- **Transport-agnostic:** MCP server logic is independent of transport. Stdio or HTTP selected via `VIDERA_TRANSPORT` env var.

### Configuration

All config via environment variables with `VIDERA_` prefix:

- `VIDERA_TRANSPORT` — `stdio` (default) or `http`
- `VIDERA_HTTP_ADDR` — HTTP listen address (default `:8080`)
- `VIDERA_DATA_DIR` — Persistent storage path (default `./data`)
- `VIDERA_LOG_LEVEL` — `debug`, `info` (default), `warn`, `error`
- `VIDERA_REMOTE_FETCH_ENABLED` — allow/deny remote HTTP(S) media ingestion in real mode (default `true`)
- `VIDERA_REMOTE_FETCH_TIMEOUT_SEC` — remote fetch timeout in seconds (default `60`)
- `VIDERA_REMOTE_FETCH_MAX_MB` — max remote media payload size in MB (default `200`)

## Testing

Tests are critical. All code changes should be covered by proper tests with thorough assertions.

### Philosophy

- **Integration tests > unit tests.** The primary test suite uses testcontainers-go to spin up the real Dockerized server and exercises it via the MCP client.
- **Integration lifecycle should be efficient and isolated.** Prefer suite-level container startup with explicit per-test reset/cleanup verification to prevent data leakage.
- **Unit tests** exist only for pure logic (config parsing, model validation).
- **Build tag:** `//go:build integration` separates slow container tests from fast unit tests.
- **Framework:** `github.com/stretchr/testify/require` (fail-fast) + `github.com/testcontainers/testcontainers-go` (v0.40+).

### Test Rules

Never ignore or dismiss failing tests as "pre-existing" or "unrelated" failures. When a test fails after making changes:

1. Investigate whether the failure could be caused by your changes
2. Check error messages carefully - they often reveal issues with generated code, mappings, or other side effects
3. Only after thorough investigation can a test failure be considered unrelated
4. If truly unrelated, note the failure and suggest investigating separately

When testing remote-source ingestion behavior in integration tests:

- Use deterministic in-test HTTP fixtures (for example `httptest` server) instead of external URLs.
- Expose host fixture ports to containers and access via `host.testcontainers.internal` when needed.
- Assert explicit error semantics for bounded failures (timeout, invalid source, max-size exceeded).

### Test Quality Bar (Mandatory)

- **Depth over checkbox testing:** Tests must verify behavior details (result shape, ordering/ranking expectations, error messages, idempotency behavior), not only success status.
- **Coverage by behavior paths:** For each feature, cover happy path, malformed input, edge cases, and at least one failure/degradation path.
- **Integration assertions should be concrete:** Prefer explicit assertions on payload fields and side effects over broad `not nil` checks.
- **No silent regressions:** When fixing a bug, add or tighten a test that would have failed before the fix.
- **Deterministic tests:** Avoid flaky timing/randomness assumptions; keep test data and expected outputs stable.
- **Verification is required before completion:** A task is not done until relevant `make test` and `make integration-test` pass.

## Key Decisions

| Decision | Chosen | Over | Reason |
|---|---|---|---|
| MCP SDK | mcp-go | Genkit | Genkit wraps mcp-go anyway; adds unnecessary framework weight for a server that exposes tools, not calls LLMs |
| Vector store | chromem-go | LanceDB | chromem-go keeps default runtime zero-dep; LanceDB path exists but introduces native/runtime packaging considerations |
| Test strategy | Testcontainers integration | Mocked unit tests | Real Docker containers catch real bugs; matches production runtime |
| Transport | Stdio + HTTP | Stdio only | HTTP needed for testcontainers and future Cloud Run deployment |

## Platform Strategy (2026-03-04)

- **Enterprise front door:** AgentGateway is the preferred enterprise entry point for auth, RBAC, rate limiting, and auditing in multi-tenant deployments.
- **Federation model:** Videra should remain composable as one MCP service among many behind a single gateway.
- **Cloud deployment parity (Cloud Run + Hetzner):** Treat GCP/Cloud Run and Hetzner as equal-priority deployment targets.
- **EU data-residency path:** Preserve a first-class EU-hosted path (Hetzner/self-hosted) for customers that do not want data outside Europe.
- **Cloud Run scaling path:** Keep scale-to-zero-friendly search API behavior and separate heavy indexing workloads into async job boundaries.
- **GPU indexing path:** Design ingestion boundaries so Whisper/CLIP workloads can later run on GPU-backed jobs without changing MCP interfaces.
- **Storage portability:** Keep storage abstraction so local Docker and cloud backends can share the same domain logic.
- **Dual deployment modes:** Maintain both self-hosted (on-prem/local Docker) and SaaS deployment paths from the same codebase.
- **Queue portability path:** Keep async orchestration broker-agnostic; default to in-process execution first, with NATS/JetStream as the first candidate when external queueing is introduced.
- **Vendor decision checkpoint:** Before implementing any vendor-specific backend, run an explicit architecture checkpoint comparing at least one neutral/self-hosted alternative and document lock-in impact.
- **Protocol-native gateway expectations:** Plan for AgentGateway-native capabilities including MCP federation, JWT/RBAC/CEL policy controls, and observability at the edge.

Reference:
- AgentGateway docs: https://agentgateway.dev/docs/
- MCP gateway overview: https://agentgateway.dev/docs/mcp/

## Cloud Migration Constraints

- Keep `storage.VectorStore` backend-agnostic; avoid leaking local-only assumptions into ingestion/search packages.
- Ensure data path configuration can target local volume mounts and cloud object storage-backed persistence.
- Keep transport and auth concerns outside core tool logic so AgentGateway integration remains an edge concern.
- Favor idempotent indexing operations and explicit job boundaries to fit async Cloud Run Job execution.
- Favor idempotent indexing operations and explicit job boundaries to fit async execution models on both Cloud Run and Hetzner-based schedulers.
- Keep MCP transport compatibility for federation patterns (stdio/SSE/streamable HTTP) expected by AgentGateway MCP integrations.
- Keep ingestion source contract parity: local paths remain supported, and cloud parity indexing should use remote HTTP(S) sources with bounded fetch controls.
- Keep queueing concerns behind interface boundaries (for example `JobQueue`) so private deployments can run in-process or with a self-hosted broker without MCP API changes.
- Do not merge vendor-specific integrations without a documented fallback/portability path for private deployments.

## Change Hygiene (Agent Workflow)

- If `VIDERA_` runtime/env contract changes, update these docs in the same change set:
  - `README.md`
  - `tasks/platform/env-contract.md`
  - relevant deployment runbooks under `tasks/platform/`
- If ingestion behavior changes, update both focused unit tests and at least one integration scenario proving the contract.
- If queue/backplane architecture is under consideration, update/check these checkpoint artifacts before implementation:
  - `tasks/platform/queue-vendor-checkpoint.md`
  - `tasks/platform/jobqueue-interface-proposal.md`

## Hetzner Equivalency Note

- Hetzner does not offer a direct Cloud Run equivalent; plan baseline deployments as Docker containers on Hetzner VMs (or Kubernetes) behind a reverse proxy/load balancer.
- Keep provider differences in deployment manifests and ops runbooks, not in MCP/runtime business logic.

## GTM & Positioning Notes (2026-03-04)

- **Core positioning:** Privacy-native, context-injecting Video Memory for agents (Claude/Cursor) that works in local/on-prem and cloud modes.
- **Go + Docker thesis:** Lean static binaries, concurrency-oriented ingestion, and portable deployment image are strategic differentiators.
- **USP framing:**
  - Context Injection Edge: MCP-delivered video context lands directly in agent workflows.
  - Zero-Knowledge option: customer-controlled data path remains central for regulated teams.
  - Multimodal advantage: retrieval spans transcript + visual evidence.

### Customer Segments (Priority)

- **Engineering Teams:** Incident/debug review and architecture walkthrough retrieval.
- **Legal & Compliance:** Time-bound evidence lookup in recorded meetings/interviews.
- **Product Managers:** Research synthesis and fast recall from interviews/demos.

### Packaging Direction

- **Tier 1 (Community):** Free local MCP server for individual usage.
- **Tier 2 (Pro/Team):** Team-ready deployment with cloud object storage integration and faster indexing.
- **Tier 3 (Enterprise):** Managed SaaS or licensed self-hosted deployment with SSO and gateway controls.

### Competitive Proof Discipline

- Keep competitive positioning measurable and reproducible: benchmark claims should map to deterministic test fixtures and published result criteria.
- Prefer defensible language ("verified", "measured", "integration-proven") over absolute marketing claims ("by miles") unless externally validated.

## Strategic Clarifications

- External strategy ideas may reference LanceDB and `mcp-go-sdk`; implementation should still follow validated constraints in this repo:
  - `lancedb-go` exists, but requires native CGO artifacts (`include` + platform libs) and explicit build/runtime packaging discipline.
  - MCP package in use is `github.com/mark3labs/mcp-go`.

## Decision Log (Immutable)

Use this as a quick historical ledger of decisions that should not be silently changed.

| Date | Decision | Rationale | Change Requires |
|---|---|---|---|
| 2026-03-04 | MCP SDK = `github.com/mark3labs/mcp-go` | Most direct/control-focused Go MCP implementation; avoids framework indirection | Explicit architectural review + migration plan |
| 2026-03-04 | Vector store (MVP) = `chromem-go` | zero-dependency baseline for deterministic local/runtime portability | Validation of alternative backend + compatibility plan |
| 2026-03-04 | Transport strategy = stdio + streamable HTTP | Stdio for local MCP clients; HTTP for integration tests/cloud deployment | Backward-compatibility check for local tooling |
| 2026-03-04 | Test strategy = integration-first with testcontainers | Best production fidelity for this server architecture | Proof that alternate strategy catches equivalent failure modes |
| 2026-03-04 | Deployment direction = local/on-prem + SaaS parity | Supports privacy-native and enterprise procurement requirements | Product + platform alignment review |

## Code Style

- No `TODO` comments in code — fix it now or create an issue.
- Error handling: always return errors, never panic in library code.
- Naming: follow Go conventions (exported = public, unexported = internal).
- Packages: keep `internal/` — nothing is a public library yet.
