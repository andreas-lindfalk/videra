# Local Retrieval Quality Checklist (Real Ingestion)

Goal: verify that local `real` ingestion returns content-grounded results for one known clip before deeper cloud rollout work.

## Preconditions

- `VIDERA_INGESTION_MODE=real`
- `VIDERA_VISUAL_BACKEND` set to expected path (`ocr` baseline or `clip` for CLIP validation)
- Video file and transcript sidecar exist next to each other (for current local validation path):
  - `./videos/<clip>.MOV`
  - `./videos/<clip>.srt` (or `.vtt` / `.txt`)
- Runtime has visual OCR capability (`tesseract`) and keyframe extraction capability (`ffmpeg`).
- For CLIP validation, runtime has native ONNX Runtime library access (`VIDERA_CLIP_ORT_LIB_PATH`) and a readable `VIDERA_CLIP_MODEL_PATH`.
- Service running via local MCP endpoint (`http://localhost:8080/mcp`)

## Known Clip Setup

Record one known clip and expected signals.

- Clip name: `<clip-name>`
- Expected spoken phrase: `<exact phrase>`
- Expected spoken phrase timestamp window: `<start-end>`
- Expected visual cue text/object: `<cue text>`
- Expected visual cue approximate window: `<start-end>`

## Checks

### 1) Transcript grounding

- Index clip with `index_video`.
- Read `video://<videoId>/transcript`.
- Verify transcript text is non-placeholder and aligned with the known clip.

Pass criteria:
- No `[simulated]` placeholder content for audio transcript entries.

### 2) Spoken phrase retrieval quality

- Run `search_video` with the exact spoken phrase.
- Use `limit=5` and compare top hit timestamp.

Pass criteria:
- At least one top result is within expected phrase window.
- Prefer top-1 hit match; otherwise record rationale.

### 3) Visual cue retrieval quality

- Run `search_video` with visual cue text/object query.
- Inspect results with `type=visual` and timestamps.

Pass criteria:
- At least one plausible visual hit exists within expected window.
- Visual snippets should reflect backend-derived content (OCR or CLIP), never simulated placeholders.

### 4) Determinism check

- Repeat the exact same query twice (spoken phrase and visual cue).
- Compare ordered tuples `(videoId,startMs,endMs,type)`.

Pass criteria:
- Ordering is identical for repeated requests.

## Evidence Template

```text
Date: <YYYY-MM-DD>
Clip: <clip-name>
Mode: real
Endpoint: http://localhost:8080/mcp

Transcript grounding:
- Pass/Fail
- Notes:

Spoken phrase query:
- Query: <phrase>
- Expected window: <start-end>
- Top hit: <timestamp/type/snippet>
- Pass/Fail

Visual cue query:
- Query: <cue>
- Expected window: <start-end>
- Best visual hit: <timestamp/snippet>
- Pass/Fail

Determinism:
- Spoken phrase repeat: Pass/Fail
- Visual cue repeat: Pass/Fail

Open issues:
- <list>
```

## Exit Criteria

- Transcript grounding passes.
- Spoken phrase retrieval window passes.
- Visual cue retrieval passes.
- Determinism passes for repeated queries.
- No simulated visual fallback content is observed in real mode.
