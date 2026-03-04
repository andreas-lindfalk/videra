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

## 2026-03-04 — Container Runtime Profiles

- **Split runtime images by capability, not by API behavior.** Keep a slim default image for speed and reproducibility, and a full tool-complete image for fallback paths while preserving identical MCP contracts.
- **Capability visibility prevents deployment surprises.** Startup logging for ffmpeg/whisper/python/tesseract availability makes missing optional dependencies explicit before indexing requests fail.

## 2026-03-04 — Cloud-Ready Remote Ingestion Parity

- **Preserve idempotency via source identity, not temp-file identity.** For remote ingestion, keep the original URL as `sourcePath` while processing a fetched temporary file so retries remain deterministic.
- **Bounded fetch controls are a deployment contract.** Keep timeout and max payload size explicit (`VIDERA_REMOTE_FETCH_TIMEOUT_SEC`, `VIDERA_REMOTE_FETCH_MAX_MB`) to avoid environment-dependent failures.
- **Containerized remote-fetch integration is deterministic with host-port exposure.** Exposing host test fixture ports to integration containers enables reliable verification of remote error paths (e.g., oversize payload).

## 2026-03-04 — Vendor Lock-In Guardrail

- **Run a vendor decision checkpoint before implementation.** For any vendor-specific backend choice, compare at least one neutral/self-hosted alternative first.
- **Require a portability path up front.** Do not adopt vendor-specific infrastructure without documenting fallback options for private/self-hosted deployments.

## 2026-03-04 — Async Job Lifecycle Contract

- **Keep async initiation and status lookup explicitly separate.** `index_video` should return quickly in async mode with `jobId`, while `get_index_job` owns lifecycle polling (`pending` → `completed`/`failed`).
- **Preserve sync defaults for compatibility.** Existing clients should continue to get blocking `index_video` behavior unless they opt into `mode=async`.
- **Make failure semantics observable and deterministic.** Background failures should always persist to job state with explicit error strings rather than disappearing in logs.

## 2026-03-04 — Queue Decision Discipline

- **Finalize decision artifacts before adapter implementation.** A vendor checkpoint matrix + interface proposal sharply reduces lock-in drift and prevents premature broker coupling.
- **Treat go/no-go criteria as implementation gates.** Queue rollout should start only after reproducible evidence covers retry semantics, operational path, and rollback to in-process mode.

## 2026-03-04 — Queue Adapter Baseline Implementation

- **Build queue contracts before vendor adapters.** A shared `JobQueue` contract test suite catches lifecycle semantic drift (`enqueue/reserve/ack/retry/fail`) before introducing broker-specific complexity.
- **Keep async API behavior stable during internal rewiring.** Routing async orchestration through an in-process queue can preserve `index_video`/`get_index_job` compatibility when integration tests are run end-to-end.

## 2026-03-04 — Queue Payload Discipline

- **Queue instructions, not media bytes.** Async backplanes should carry compact job metadata (`jobId`, source reference, attempts), while workers fetch media from shared-access sources.
- **Distributed scaling requires shared source reachability.** Local node paths are fine for in-process/local mode, but multi-worker queue mode needs URLs or shared mounts visible to all workers.

## 2026-03-04 — Queue Backend Consolidation Heuristic

- **Prefer fewer moving parts when capabilities overlap.** If Redis is already required for key/value workloads, Redis Streams can be the pragmatic first queue backend to reduce operational surface area.
- **Keep portability through shared queue contracts.** Even when choosing Redis for consolidation, maintain parity-tested adapters so NATS remains a viable fallback path.

## 2026-03-04 — Split Role Async Hardening

- **Separate API and worker roles via explicit runtime contract.** A dedicated `VIDERA_JOBQUEUE_ROLE` (`all|api|worker`) avoids accidental coupling and clarifies deployment topology.
- **Persist async job status outside process memory for split deployments.** `get_index_job` must read from a shared backend (Redis prefix store or NATS KV) when API and worker run in separate processes.
- **Queue retry budget should be lifecycle-visible.** Terminal failure messages should include attempt exhaustion semantics so operators can distinguish transient retries from permanent failure.

## 2026-03-04 — Queue Go/No-Go Evidence Discipline

- **Benchmark artifacts should include exact commands and measured outputs, not qualitative claims.** Reproducibility matters more than broad performance wording.
- **Split-role validation must separate control-plane and data-plane assertions.** `get_index_job` can be shared-state correct even when API-visible indexed data depends on storage topology decisions.
- **Rollback confidence should be tested as a first-class path.** Keeping in-process async lifecycle tests passing is part of external-queue readiness, not an afterthought.

## 2026-03-04 — Queue Rollout Guardrails & Observability

- **Worker-only role needs explicit transport semantics.** Enforce `VIDERA_TRANSPORT=stdio` for `VIDERA_JOBQUEUE_ROLE=worker` so invalid API-mode assumptions fail at startup instead of at runtime.
- **Lifecycle logs need stable keys before rollout.** A fixed `queue_lifecycle` schema (`event`, `job_id`, `status`, `attempt`, `max_attempts`, `delay_ms`, `error`) makes ops dashboards and incident triage durable across refactors.
- **Observability should be integration-proven, not assumed.** Split-role tests should assert worker logs for both success and retry-exhausted paths to validate operator visibility under real container wiring.

## 2026-03-04 — Split-Role Data-Plane Parity

- **Control-plane success does not imply data-plane visibility.** `get_index_job=completed` can still coexist with empty `list_videos`/`search_video` unless API and worker share index storage explicitly.
- **Make shared-storage behavior explicit in runtime config.** A dedicated split-role data-plane flag (`VIDERA_SPLIT_SHARED_STORAGE`) reduces ambiguity and makes operator intent auditable.
- **Prove both sides in integration tests.** Keep one test for degraded non-shared semantics and one for shared-mount visibility success so regressions in either path are caught quickly.

## 2026-03-04 — MVP Exit Gate Discipline

- **A single release command reduces ambiguity.** A canonical gate (`make release-gate`) avoids fragmented validation across ad-hoc command chains.
- **Keep split-role checks explicit even when included in full integration runs.** A focused gate (`make release-gate-split`) makes control-plane/data-plane semantics visible to operators and release reviewers.
- **Treat parity evidence as a required artifact, not tribal knowledge.** A fixed release evidence template improves repeatability and go/no-go decisions.
