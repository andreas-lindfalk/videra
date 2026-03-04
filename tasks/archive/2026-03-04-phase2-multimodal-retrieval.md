# Archive — Videra MVP Phase 2 (Multimodal Retrieval)

Archived: 2026-03-04
Source: tasks/todo.md

# Videra MVP — Phase 2: Multimodal Retrieval (See + Say)

## Goal
- [x] Add visual indexing and hybrid retrieval while preserving current MCP interface and cloud-ready boundaries.

## Architecture & Interfaces
- [x] Extend storage contracts for multimodal segments and hybrid query execution (text + visual embeddings).
- [x] Keep ingestion orchestrator backend-agnostic and idempotent for future Cloud Run Job execution.
- [x] Define clear ingestion stage boundaries (extract, transcribe, frame-embed, text-embed, persist).

## Ingestion (Visual)
- [x] Implement keyframe extraction workflow using FFmpeg runner with configurable frame interval.
- [x] Add CLIP embedding interface and initial adapter (stub/local placeholder for MVP).
- [x] Persist visual segments with timestamps and metadata alongside transcript segments.

## Query Pipeline (Hybrid)
- [x] Implement `search_video` MVP: embed query, run text + visual retrieval, merge and rerank top-k.
- [x] Return search response with timestamps, transcript snippets, and visual context fields.
- [x] Keep response schema stable for MCP clients and backward compatibility.

## MCP Surface
- [x] Upgrade `search_video` tool from stub to production MVP handler.
- [x] Ensure `list_videos` includes indexing modality/status details useful for operations.
- [x] Keep `video://{id}/transcript` behavior unchanged while enabling future visual resource endpoints.

## Configuration & Ops
- [x] Add config knobs for frame interval, top-k defaults, and indexing concurrency.
- [x] Add explicit runtime mode notes for local Docker vs cloud execution paths.

## Testing (Integration-First)
- [x] Add integration tests for hybrid retrieval path (text hit + visual hit).
- [x] Add integration tests for indexing idempotency (re-index same source safely).
- [x] Add integration tests for malformed/unavailable media handling.

## Verification
- [x] `make build` passes with new multimodal code paths.
- [x] `make test` passes fast suite.
- [x] `make integration-test` passes hybrid retrieval scenarios.
