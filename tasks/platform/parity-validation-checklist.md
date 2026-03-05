# Deployment Parity Validation Checklist (Hetzner + Cloud Run)

Goal: verify that Videra MCP behavior is consistent across Hetzner and Cloud Run for the same build/version and equivalent runtime config.

Use this checklist after each deployment update.

Release-candidate workflow reference:

- `tasks/platform/mvp-release-gate.md`

## Preconditions

- Same application image/version deployed to both environments.
- MCP endpoint reachable in both environments (`.../mcp`).
- Equivalent env settings for core behavior (`VIDERA_TRANSPORT`, log level, retrieval weights).
- Test fixture strategy chosen:
  - Hetzner: can use server-visible `/videos/...` path indexing or remote HTTP(S) URL indexing.
  - Cloud Run: use remote HTTP(S) URL indexing for parity checks.

## Execution Matrix

Run each check in both environments and record result.

| Check | Hetzner | Cloud Run | Expected |
|---|---|---|---|
| MCP connect to `/mcp` | [ ] pass / [ ] fail | [ ] pass / [ ] fail | Successful MCP handshake |
| `list_videos` | [ ] pass / [ ] fail | [ ] pass / [ ] fail | Stable response shape |
| `index_video` | [ ] pass / [ ] fail | [ ] pass / [ ] fail | Returns `videoId`, `status` |
| `search_video` deterministic repeat (same query twice) | [ ] pass / [ ] fail | [ ] pass / [ ] fail | Identical ordering for same inputs |
| `read_resource` transcript (`video://{id}/transcript`) | [ ] pass / [ ] fail | [ ] pass / [ ] fail | Resource resolves with expected format |
| Restart behavior | [ ] pass / [ ] fail | [ ] pass / [ ] fail | Service recovers without contract drift |
| Persistence across restart | [ ] pass / [ ] fail | [ ] pass / [ ] fail / [ ] n/a | Data state behavior matches environment design |

## Detailed Steps

### 1) MCP connectivity

- Connect MCP client to endpoint.
- Confirm tool/resource discovery works.

Pass criteria:
- No transport/protocol errors.

### 2) `list_videos` shape

- Call `list_videos`.
- Verify each item shape includes core fields used by clients (`id`/`filePath`/status metadata).

Pass criteria:
- Response is parseable and fields are stable across environments.

### 3) `index_video` behavior

- Attempt indexing using environment-appropriate source:
  - Hetzner: `/videos/...` path or remote HTTP(S) URL
  - Cloud Run: remote HTTP(S) URL
- Capture resulting `videoId`.

Pass criteria:
- MCP tool returns success and expected fields.

### 4) `search_video` determinism

- Run same query twice with same options (`limit`, `includeDebug`, weights).
- Compare ordered results by `(videoId, startMs, endMs, type)`.

Pass criteria:
- Ordering and top hits remain identical for repeated request.

### 5) Transcript resource behavior

- Read `video://<videoId>/transcript`.
- Verify response shape is consistent and content is non-empty when indexed data exists.

Pass criteria:
- Resource resolves; no schema mismatch.

### 6) Restart + persistence

- Restart service in each environment.
- Re-run `list_videos` and representative `search_video` query.

Pass criteria:
- Service recovers and behaves according to configured persistence model.

## Evidence Capture Template

Record one block per environment:

```text
Environment: Hetzner | Cloud Run
Image: <image-tag>
Endpoint: <mcp-url>
Date: <YYYY-MM-DD>

Checks:
- MCP connect: pass/fail
- list_videos: pass/fail
- index_video: pass/fail
- search deterministic repeat: pass/fail
- transcript resource: pass/fail
- restart behavior: pass/fail
- persistence behavior: pass/fail/n/a

Notes:
- Observed deltas:
- If any failed checks, mitigation plan:
```

## Exit Criteria

- No unexplained response-contract differences between environments.
- Determinism check passes where indexing/search is supported.

## Final Decision Summary (Phase 33)

After completing parity checks and local promotion gates, write one unified decision artifact using:

- `tasks/platform/cross-environment-promotion-summary-template.md`

This is the preferred final GO/NO-GO handoff format for cross-environment promotion review.

## Split-Role Release-Critical Add-on

Before final MVP go/no-go, also run:

```bash
make release-gate
make release-gate-split
make pilot-quality-gate
```

Pass criteria:

- Split-role lifecycle semantics pass (`pending -> completed|failed`) with explicit retry-exhausted failure behavior.
- Shared-storage split-role visibility path is proven (`list_videos` / `search_video` after async completion).
- Worker-role transport guardrail is proven (`VIDERA_JOBQUEUE_ROLE=worker` rejects HTTP transport).
- Pilot benchmark quality signal and real-mode guardrails are green in the same release cycle.

If local Docker pressure causes unstable gate execution, use:

```bash
make release-gate-preflight
make release-gate-clean
make release-gate
make release-gate-split
make pilot-quality-gate
```
