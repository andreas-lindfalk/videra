# Lessons Learned

## 2026-03-04 — Project Bootstrap

- **No Go SDK for LanceDB.** The VIDERA_MVP_SPEC.md references LanceDB, but as of March 2026 there is no Go client. Use `chromem-go` (pure Go, embedded, zero-dep) for the MVP. Migration path: LanceDB REST API or pgvector in Phase 3.
- **Genkit wraps mcp-go.** Google's Genkit Go MCP plugin (`github.com/firebase/genkit/go/plugins/mcp`) imports `mcp-go` under the hood. For a server that *exposes* tools (not calls LLMs), using mcp-go directly gives more control with fewer dependencies.
- **mcp-go package is `github.com/mark3labs/mcp-go`**, not `mcp-go-sdk` as referenced in the original spec.

## 2026-03-04 — Product/Platform Direction

- **AgentGateway is a first-class enterprise integration target.** Keep server internals auth-agnostic and let gateway handle enterprise controls (RBAC, auditing, rate limiting).
- **Cloud Run architecture split is required.** Interactive MCP/search path should remain lightweight while heavy video indexing is designed for asynchronous job execution.
- **Deployment parity matters.** Preserve a single codebase and interface contracts that work both for local/on-prem Docker and SaaS cloud deployment.

## 2026-03-04 — Phase 2 Retrieval Reliability

- **Hybrid search should degrade gracefully.** If vector query execution fails, return best-effort results from in-memory indexed segments instead of surfacing MCP tool errors.

## 2026-03-04 — Business Strategy Snapshot

- **Positioning that resonates:** "Privacy-native Video Memory for agents" is clearer than generic Video-RAG phrasing when discussing enterprise value.
- **GTM should stay three-track:** Community (free local), Pro/Team (cloud storage + faster indexing), Enterprise (managed or self-hosted with SSO/governance).
- **Segment-led roadmap works:** Engineering, Legal/Compliance, and Product each need different retrieval narratives; demos and benchmarks should be tailored per segment.
- **Keep strategy vs implementation separate:** External architecture proposals can inspire roadmap, but implementation choices must remain constrained by verified SDK/tooling realities in this codebase.

## 2026-03-04 — Testing Standard (Non-Negotiable)

- **Test rigor is first-priority.** Future work must optimize for deep verification quality, not just feature completion.
- **Every major change requires detailed assertions.** Validate payload structure, ranking/ordering behavior, error semantics, and persistence/idempotency effects.
- **Coverage must include failure paths.** Happy-path-only testing is insufficient for acceptance.

## 2026-03-04 — Ranking Determinism

- **Hybrid ranking must be explicitly deterministic.** When scores tie, apply stable tie-breakers (video ID, timestamps, segment type) so repeated queries return identical ordering.
- **Modality weighting should be configurable.** Keep audio/visual weight knobs in config so retrieval tuning can be adjusted without code changes.

## 2026-03-04 — Integration Test Technique

- **Use env-overridden containers for behavior verification.** For ranking/weighting behavior, start separate integration containers with different config env values and compare measurable outputs (e.g., top modality similarity), not just pass/fail state.

## 2026-03-04 — Integration Lifecycle Efficiency

- **Prefer suite-level container lifecycle for integration tests.** Start containers once in `SetupSuite` and reuse MCP clients across tests to reduce build/startup overhead.
- **Enforce per-test isolation with explicit reset hooks.** Use a test-only MCP `reset_index` tool in `VIDERA_RUNTIME_MODE=test` and verify `list_videos` is empty in `SetupTest`.
- **Do not rely on implicit cleanup.** Isolation must be asserted in test code, not assumed from container behavior.

## 2026-03-04 — Indexing Orchestration Boundary

- **Keep MCP indexing behind an orchestrator contract.** Route `index_video` through an `IndexOrchestrator` interface even for synchronous execution so Cloud Run Job orchestration can be introduced without changing MCP tool handlers.
- **Use explicit job request/result contracts now.** Defining `IndexJobRequest` and `IndexJobResult` early makes async migration incremental instead of a breaking refactor.
- **Protect response compatibility with integration regression tests.** When internal orchestration changes, assert required response fields for `index_video`, `search_video`, and `list_videos` remain stable.

## 2026-03-04 — Retry-Safe Partial Failure Handling

- **Always recover through source-identity lookup on failures.** If indexing returns an error, check `GetVideoBySourcePath` before failing the job to handle "persisted-but-errored" partial failure cases safely.
- **Retry logic should be bounded and deterministic.** A small fixed retry budget (sync path) catches transient failures without hiding persistent errors.
- **Use deterministic failure hooks in integration tests.** Path-scoped one-time failure triggers validate partial-failure recovery at MCP level without introducing flaky timing-based tests.

## 2026-03-04 — Transport/Auth Separation Guardrail

- **Keep auth concerns out of core server/config contracts.** MCP and ingestion boundaries should not require auth configuration to operate in local/integration mode.
- **Add explicit tests for uncoupled config loading.** Validating config behavior with unrelated auth env vars helps prevent accidental edge-concern leakage into core runtime paths.
