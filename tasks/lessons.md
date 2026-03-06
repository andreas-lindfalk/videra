# Lessons Learned

## 2026-03-04 — Project Bootstrap

- **LanceDB Go SDK exists (`github.com/lancedb/lancedb-go`) but has native packaging requirements.** It requires CGO plus platform-specific native artifacts (`include/lancedb.h` + linked libs), so a repo can still choose `chromem-go` as the default zero-dependency baseline.
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

## 2026-03-04 — RC1 Integration Stability Hardening

- **Fresh integration runs expose real instability sooner.** Using `-count=1` in release-gate integration checks avoids cached false confidence.
- **Testcontainers deadline control should be explicit for containerized suites.** `WithWaitStrategyAndDeadline(...)` with a bounded startup window reduces intermittent `context deadline exceeded` failures in slow local environments.
- **Stabilization should pair code-level hardening with operator rerun discipline.** Preflight/cleanup + deterministic rerun steps prevent ambiguous release outcomes.

## 2026-03-05 — RC2 Release Packaging Discipline

- **Capture release evidence from command outputs + exit codes together.** Storing deterministic gate logs and explicit exit status per command makes RC go/no-go decisions auditable instead of narrative.
- **Keep split-role checks as a first-class gate artifact.** Recording focused split-role pass results alongside full release-gate output preserves control-plane/data-plane confidence for operators.

## 2026-03-05 — Semantic Reranking Guardrail

- **Modality diversity should not suppress strong lexical evidence.** Diversity heuristics for early result slots must yield when same-modality hits have clear query-token relevance, or proofpack evidence quality can regress.
- **Fallback recall + deterministic reranking is safer than vector-only candidate reliance in mixed-fidelity embeddings.** Enriching candidates with transcript fallback before reranking improves robustness without changing MCP contracts.

## 2026-03-05 — Domain-Neutral Normalization Default

- **Core normalization should be neutral by default.** Hardcoded topic synonym maps in the default path can bias retrieval for unrelated corpora.
- **Domain mapping belongs in explicit runtime config.** Optional canonical maps keep behavior adaptable without forcing assumptions across all datasets.

## 2026-03-05 — Domain Mapping Evaluation Discipline

- **Validate optional mappings with paired OFF/ON tests on the same fixture.** Comparing rank and similarity deltas for target evidence provides clear proof of domain profile value.
- **Keep neutral-default checks in the same phase.** Domain tuning should never replace baseline determinism and backward-compatibility guarantees.

## 2026-03-05 — Real Corpus Onboarding Gate Discipline

- **Treat real-mode source constraints as first-class quality gates.** Disabled remote fetch, sidecar requirements, and max-size bounds should have explicit integration assertions and stable error semantics.
- **Pair real-mode guardrails with deterministic/evidence checks.** Corpus onboarding is only release-ready when ingestion constraints and retrieval quality signals are both green.

## 2026-03-05 — Pilot Benchmark Scorecard Discipline

- **Keep pilot benchmark packs fixture-driven and measurable.** A small curated scenario slice with explicit evidence targets provides faster and more reliable tuning feedback than ad-hoc query checks.
- **Gate tuning changes on metric thresholds, not intuition.** Preserve neutral defaults unless benchmark scorecards show concrete degradation signals that justify domain-specific mapping.

## 2026-03-05 — Quality Gate Operationalization

- **Promote repeated validation flows to one-command targets.** Packaging stable integration subsets into a dedicated `make` target reduces execution drift and makes operator handoffs more reliable.
- **Keep composed gates contract-focused.** Combining benchmark quality checks with ingestion guardrails in one run provides a clearer GO/NO-GO signal without broad, expensive full-suite runs.

## 2026-03-05 — Roadmap Clarity Discipline

- **Keep a living North Star artifact in-repo.** Teams move faster when each phase maps directly to target-state outcomes instead of only to technical tasks.
- **Track phase purpose and endpoint together.** For each phase, make explicit what capability it adds and which final release criterion it advances.

## 2026-03-05 — Candidate Mode Execution Hardening

- **Default-value config tests must pin env-sensitive fields explicitly.** Candidate-mode gate runs can export `VIDERA_STORAGE_BACKEND`; tests expecting defaults should set `VIDERA_STORAGE_BACKEND=chromem` to avoid inherited-env false failures.
- **Composite Make recipes should use fail-fast shell behavior.** Multi-step capture targets that are intended as gates should include `set -e` so sub-step failures do not emit misleading success messages.

## 2026-03-05 — Release Decision Signal Coupling

- **Do not treat operational pass/fail as sufficient release proof.** Couple release gates with a compact retrieval-quality signal so GO decisions reflect both reliability and result quality.
- **Keep quality coupling lightweight and repeatable.** A focused quality command integrated into release evidence is more sustainable than ad-hoc manual quality checks.

## 2026-03-05 — Real-Corpus Promotion Discipline

- **Promotion gates need explicit thresholds and templates.** Defining measurable criteria (`evidenceMatchRate`, `topTwoQualityRate`, deterministic replay, guardrail semantics) prevents subjective rollout calls.
- **Compose promotion gates from existing proven checks.** Reusing stable focused tests via one command improves repeatability and lowers operational friction.

## 2026-03-05 — Storage Re-checkpoint Discipline

- **Run backend migration decisions through explicit GO criteria.** A weighted matrix plus hard preconditions prevents architecture churn driven by intuition.
- **Require benchmark and runtime-contract proof before migration.** If performance gains and operational readiness are not both demonstrated, defer migration and preserve the stable path.

## 2026-03-05 — Promotion Workflow Consolidation

- **One-command operator flows reduce decision drift.** Wrapping release, split-role, and promotion gates into a single command improves repeatability and handoff quality.
- **Evidence templates should match command composition.** Consolidated commands need consolidated evidence artifacts to keep GO/NO-GO review fast and auditable.

## 2026-03-05 — Storage Benchmark Harness Discipline

- **Benchmark workload names should be stable across backend candidates.** Reusing identical benchmark case names keeps cross-backend comparisons auditable and less interpretation-heavy.
- **Decision-refresh baselines need both raw logs and summarized metrics.** Capturing output files plus normalized evidence summaries improves future checkpoint reruns.

## 2026-03-05 — Cross-Environment Evidence Unification

- **Use one final decision artifact across environments.** Merging local promotion and deployment parity status into one summary reduces ambiguous GO/NO-GO handoffs.
- **Allow conditional status only with explicit follow-up actions.** Pending environment runs should be visible, bounded, and tied to concrete next steps.

## 2026-03-06 — Checkpoint Closure Discipline

- **Blocked phases need explicit activation verdict artifacts.** Recording a dated GO/NO-GO checkpoint avoids ambiguity between "in progress" and "intentionally paused" states.
- **A single unmet hard-gate criterion should be surfaced as the only blocker.** This keeps re-open conditions clear and prevents churn on already-passing prerequisites.

## 2026-03-06 — LanceDB Go Integration Pattern

- **Prefer native SDK path when available and compatible with platform constraints.** `lancedb-go` can be integrated directly while preserving `VectorStore` contract boundaries.
- **Treat CGO/native artifacts as first-class runtime dependencies.** Native headers/libs and linker flags must be explicit in operator docs and CI setup.

## 2026-03-06 — LanceDB SDK Reality Check

- **`lancedb-go` availability does not remove deployment complexity by itself.** The SDK currently depends on CGO/native artifacts, so builds and containers must package the matching headers/libs per platform.
- **Keep backend selection explicit and reversible.** Preserve runtime toggle boundaries so native and bridge-backed paths can be compared and rolled back without MCP contract changes.

## 2026-03-06 — Native Dependency Onboarding Guardrail

- **Do not force native SDK prerequisites on every contributor by default.** Keep native backends behind explicit build tags/runtime toggles so default local workflows remain zero-dependency.
- **Treat cloud-only connection fields as conditional, not global defaults.** For LanceDB, region should be required only for `db://` cloud URIs, not local file-backed paths.
- **Native mode must be an operator path, not a teammate prerequisite.** Provide first-class Docker/Make flows for `lancedb_native` so production usage is reproducible without burdening default local onboarding.

## 2026-03-06 — LanceDB Native Artifact Architecture Guardrail

- **Do not assume multi-arch Linux artifacts exist for every LanceDB native release.** Validate release bundle contents (`linux_amd64`/`linux_arm64`) before wiring default Docker build targets.
- **Pin native Docker builds to a known-good platform when artifact coverage is asymmetric.** For current releases, defaulting native Docker build/test lanes to `linux/amd64` avoids arm64 link failures on Apple Silicon hosts.

## 2026-03-06 — Build-Tag Stub Cleanup Pattern

- **Prefer build-tag registration over duplicate stub files when only one implementation should vary by tag.** A shared untagged factory entrypoint plus native `init()` registration keeps default behavior explicit while reducing file clutter.

## 2026-03-06 — Local Workflow Tooling Regression Guardrail

- **Validate new helper commands immediately with formatting + compile checks.** Running `gofmt` and a focused `go test ./cmd/<tool>` right after adding a new local CLI catches malformed files before downstream edits compound the failure.
- **Prefer simple line-oriented shell loops over heredoc-heavy Make recipes.** Complex heredoc blocks are fragile in tab-indented Make targets and can produce hard-to-read parser errors (`missing separator`).
