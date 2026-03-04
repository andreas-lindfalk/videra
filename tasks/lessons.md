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

## 2026-03-04 — AgentGateway Alignment (Source-Confirmed)

- **AgentGateway is protocol-native for MCP + A2A.** Treat Videra as a federated MCP backend behind a single enterprise gateway endpoint.
- **Edge controls should stay at the gateway.** JWT/RBAC/CEL policies, TLS, and observability belong at AgentGateway, not inside core ingestion/search logic.
- **Integration planning should assume MCP federation workflows.** Keep Videra transport/runtime behavior compatible with gateway-side MCP federation and discovery patterns.

## 2026-03-04 — Competitive Messaging Guardrails

- **Use only evidence-backed claims in competitor comparisons.** Positioning should be anchored to reproducible benchmarks (latency, determinism, retrieval quality) and integration proofs.
- **Do not promise backend choices that conflict with repo decisions.** Keep MVP messaging aligned with `chromem-go` unless architecture decisions are explicitly changed.
- **Sell workflow ownership, not just search quality.** The strongest defensible narrative is multimodal retrieval + MCP context injection + privacy-native deployment.

## 2026-03-04 — Proof Pack Benchmarking

- **Keep benchmark scenarios as deterministic fixtures in-repo.** Versioned JSON fixtures make competitive claims reproducible and auditable.
- **Measure repeatability, not just relevance.** Running the same query twice and asserting identical ordered results is a core proof artifact.
- **Encode evidence coverage expectations per scenario.** Benchmark checks should require explicit snippet/evidence matches, not only non-empty result lists.

## 2026-03-04 — Proof Metadata Compatibility

- **Expose benchmarking/debug metadata behind opt-in flags.** Optional search metadata (`includeDebug`) supports proof artifacts without breaking existing client flows.
- **Preserve primary response semantics.** Keep existing `similarity` behavior stable and add extra fields (e.g., `rawSimilarity`, debug block) as additive metadata only.

## 2026-03-04 — MCP Argument Typing Pitfall

- **Treat `req.Params.Arguments` as `any` unless typed helpers are used.** Do not index directly; type-assert to `map[string]any` before reading optional fields.
- **Additive tool args should fail safe.** Optional flags (e.g., `includeDebug`) should default cleanly when missing or malformed.

## 2026-03-04 — Local Validation UX Pattern

- **Provide a one-command local loop as first-class workflow.** `local-up`, `local-smoke`, and `local-down` style tasks reduce setup ambiguity and support repeatable pre-cloud validation.
- **Use explicit host-to-container path contracts for local files.** Mounting a configurable host video directory to `/videos` prevents path confusion during local indexing.
- **Treat smoke CLI as integration contract check.** A lightweight MCP client executable that runs index/search/list/transcript in sequence is a practical anti-blackbox guardrail.

## 2026-03-04 — Local Validation First (Pre-Cloud)

- **Avoid CloudRun-first trial-and-error loops.** Keep a simple local developer validation path (service spin-up, local file indexing, MCP client connection) as a required pre-cloud gate.
- **Treat local smoke tests as product UX.** If local setup/testing is hard, the platform feels like a black box regardless of backend quality.

## 2026-03-04 — EU Deployment Parity Requirement

- **Treat Hetzner support as equal-priority to Cloud Run in planning.** For Europe-first customers, deployment location is a product requirement, not a later ops detail.
- **Avoid Cloud-provider coupling in core runtime.** Keep provider differences in infra/runbooks while preserving identical MCP behavior and contracts.

## 2026-03-04 — Local-First Execution Guardrail

- **Do not over-invest in cloud platform engineering before local semantic quality is proven.** Keep Cloud Run/Hetzner work to parity planning/runbooks until real-content retrieval quality is satisfactory locally.
- **Use local acceptance criteria as the gate.** Real transcript/visual grounding and query-quality checks should pass before deeper provider-specific rollout work.

## 2026-03-04 — Real Mode Increment Strategy

- **Add a low-risk real-mode slice before full ASR integration.** A sidecar transcript path (`.srt`/`.vtt`/`.txt`) allows immediate local semantic validation without blocking on model/runtime integration complexity.
- **Keep deterministic testing defaults.** Preserve `simulated` mode as the default for repeatable integration tests while enabling `real` mode explicitly for quality checks.

## 2026-03-04 — Real Audio Transcription Fallback

- **Use sidecar-first, transcribe-second fallback for reliability.** In `real` mode, prefer deterministic sidecar transcripts, then fallback to FFmpeg audio extraction + Whisper CLI when sidecars are absent.
- **Protect portability with optional dependencies.** Whisper and Tesseract should remain optional runtime enhancements; clear error messaging is required when missing.
