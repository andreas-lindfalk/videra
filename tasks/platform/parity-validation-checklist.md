# Deployment Parity Validation Checklist (Hetzner + Cloud Run)

Goal: verify that Videra MCP behavior is consistent across Hetzner and Cloud Run for the same build/version and equivalent runtime config.

Use this checklist after each deployment update.

## Preconditions

- Same application image/version deployed to both environments.
- MCP endpoint reachable in both environments (`.../mcp`).
- Equivalent env settings for core behavior (`VIDERA_TRANSPORT`, log level, retrieval weights).
- Test fixture strategy chosen:
  - Hetzner: can use server-visible `/videos/...` path indexing.
  - Cloud Run: if server-local path indexing is unavailable, run endpoint/runtime checks and mark indexing-dependent checks as `blocked` with reason.

## Execution Matrix

Run each check in both environments and record result.

| Check | Hetzner | Cloud Run | Expected |
|---|---|---|---|
| MCP connect to `/mcp` | [ ] pass / [ ] fail | [ ] pass / [ ] fail | Successful MCP handshake |
| `list_videos` | [ ] pass / [ ] fail | [ ] pass / [ ] fail | Stable response shape |
| `index_video` (if supported in env) | [ ] pass / [ ] fail | [ ] pass / [ ] fail / [ ] blocked | Returns `videoId`, `status` |
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

- Attempt indexing using environment-appropriate source path.
- Capture resulting `videoId`.

Pass criteria:
- MCP tool returns success and expected fields.

Block condition (allowed today for Cloud Run):
- No server-visible local-file path strategy in environment.

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
- index_video: pass/fail/blocked (+ reason)
- search deterministic repeat: pass/fail
- transcript resource: pass/fail
- restart behavior: pass/fail
- persistence behavior: pass/fail/n/a

Notes:
- Observed deltas:
- If any blocked checks, mitigation plan:
```

## Exit Criteria

- No unexplained response-contract differences between environments.
- Any blocked check has explicit reason and mitigation tracked in phase plan.
- Determinism check passes where indexing/search is supported.
