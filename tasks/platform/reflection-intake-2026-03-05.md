# Reflection Intake Log — 2026-03-05

Purpose: capture Gemini discussion inputs + internal synthesis so planning decisions are traceable and not lost.

Status: Intake complete; moved to dedicated LanceDB checkpoint phase.

## Captured Inputs (from discussion)

### Input 1 — User flow framing
- Product should be understood in two layers:
  - Management layer (ingestion/admin)
  - Agent layer (MCP usage through Claude/Cursor)
- User seeks clarity on:
  - how videos enter the system,
  - whether original videos are stored,
  - whether users can list indexed videos and search by spoken content and "feeling".

### Input 2 — Zero-copy / zero-upload direction
- Desired operating model: mount paths or grant bucket access; avoid upload-through-app requirement.
- System should index where data already lives and minimize data movement.

### Input 3 — Gemini architecture proposal summary
- Positioning: enterprise-ready via zero-copy/zero-upload indexer model.
- Discovery proposal: local FS watcher (`fsnotify`) and bucket events/polling for new files.
- Indexing proposal: store metadata/vectors, keep originals in-place.
- Retrieval proposal: hybrid across text + visual + potential audio sentiment for "vibe".
- Gateway proposal: AgentGateway as edge layer for auth/rate limits/audit.

### Input 4 — Vector backend question (LanceDB vs Chroma-like)
- Gemini argument: LanceDB is preferable for multimodal + cloud/object-storage workflows.
- User provided LanceDB Go client link:
  - https://github.com/lancedb/lancedb-go

### Input 5 — User desired runtime data model (final intake)
- User wants explicit two-path model:
  1) read-only source path(s) where videos already live (bucket/mount variants over time),
  2) separate read-write index/data path where system outputs are stored.
- User asks for alignment on this architecture and prioritizes clarifying LanceDB direction now.

## Current Repo Reality (high-level)
- Core MCP contract implemented and release-gated:
  - `index_video`, `get_index_job`, `search_video`, `list_videos`, transcript resource.
- Backend/storage currently aligned to `chromem-go` decisions in repo docs.
- Async + split-role queue architecture and RC2 release packaging are complete.
- Semantic quality gaps remain (text embedding baseline + visual stub behavior vs true multimodal target).

## Key Tensions to Resolve in planning
- Vision says "deeper multimodal/vibe retrieval" vs current deterministic baseline implementation.
- Original MVP spec references LanceDB + older package naming; implementation decisions diverged intentionally.
- New evidence (`lancedb-go` availability) re-opens storage decision checkpoint, but should not be mixed with retrieval-quality lift unless explicitly chosen.

## Planning Decision (locked for now)
- First priority target state: **Semantic quality lift** while preserving MCP contract stability.
- Defer storage backend migration decision to explicit checkpoint after quality-lift outcome unless blocked sooner.

## Alignment Check (current)
- **Aligned principle:** zero-upload ingestion from source locations + separate writable index/data location is the target operating model.
- **Current implementation status:** partial alignment exists (`index_video` path/URL ingestion + writable data dir), but watcher/event-driven source discovery is not implemented yet.
- **LanceDB status:** `lancedb-go` exists and is viable to evaluate, but introduces CGO/native build/runtime implications that must be compared against current static portability assumptions.

## Open Questions Queue (to resolve in this planning pass)
1. Transcript resource semantics: strict audio transcript vs all timeline segments?
2. Definition of "vibe" success: which measurable signals and acceptance metrics?
3. Auto-discovery scope: explicit `index_video` only vs watcher/event-based ingestion in near term?
4. AgentGateway requirement timing: immediate gate vs follow-up ops target?
5. Storage checkpoint timing: run before RC3 implementation or after quality baseline is improved?
6. Build/runtime contract choice if LanceDB is adopted: keep static-default via profile split, or accept CGO as default runtime requirement?

## Working Method for the remaining discussion
1. Continue collecting your incoming Q/A snippets into this file.
2. Tag each point as: requirement, assumption, risk, or deferred.
3. Produce a single consolidated comparison:
   - MVP spec intent
   - implemented reality
   - selected next target state
   - acceptance criteria + out-of-scope boundaries.
4. Convert final decision into a phase plan in `tasks/todo.md` before execution.

## Handoff to Decision Phase

- Active decision artifact: `tasks/platform/lancedb-storage-checkpoint-2026-03-05.md`
- Active planning tracker: `tasks/todo.md` (Phase 19)
