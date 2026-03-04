# Project Specification: Videra RAG-Go MCP Server

> **Status note (2026-03-04):** This document is the original product/spec draft.
> For authoritative implementation decisions and current architecture constraints,
> use `AGENTS.md` first. Where this file conflicts with `AGENTS.md`, treat
> `AGENTS.md` as source of truth.

## 1. Vision & USP

**Product:** A high-performance MCP server written in Go that provides deep searchability in video files using RAG (Retrieval-Augmented Generation).
**The "Moat" (USP):** * **Multimodal RAG:** Searches both what is *said* (Whisper) and what is *seen* (CLIP).

* **Local-First / Cloud-Ready:** Runs in Docker. Data stays with the user (Enterprise-grade privacy).
* **Agentic Integration:** Designed specifically for MCP-compatible LLMs (Claude/Cursor) to "remember" and "see" video content.

---

## 2. Technical Stack

* **Language:** Golang (Efficiency, Concurrency, Static Binaries).
* **Protocol:** Model Context Protocol (MCP) via `github.com/mark3labs/mcp-go-sdk`.
* **Vector Database:** `LanceDB` (Embedded, Serverless, GCS/S3 compatible).
* **AI Models (Local/Edge):**
* **Audio:** `Whisper.cpp` (via CGO bindings) for transcription.
* **Visual:** `CLIP` (via ONNX Runtime Go) for visual embeddings every N seconds.


* **Processing:** `FFmpeg` (via Docker) for frame and audio extraction.
* **Deployment:** Dockerized for Local, Cloud Run, or AgentGateway.dev.

---

## 3. System Architecture (MVP)

### A. Ingestion Pipeline (The "Indexer")

1. **Extract:** FFmpeg splits video into an MP3 (audio) and Keyframes (images).
2. **Transcribe:** Whisper converts MP3 to text with millisecond timestamps.
3. **Embed Visuals:** CLIP generates vectors for each keyframe.
4. **Embed Text:** Generate vectors for transcript segments (using a lightweight embedding model like `bge-small-en`).
5. **Store:** Save all vectors and metadata in LanceDB.

### B. Query Pipeline (The "Searcher")

1. **Input:** User asks a question in Claude (e.g., "Where did they talk about the budget?").
2. **Search:** MCP Tool `search_video` triggers.
* Text query is embedded.
* Vector search in LanceDB (Hybrid: Text + Visual).


3. **Context Construction:** Top K matches (text + timestamps + visual descriptions) are returned to Claude.

---

## 4. MCP Tools & Resources (Interface)

| Type | Name | Description |
| --- | --- | --- |
| **Tool** | `index_video` | Takes a local path or URL, runs the ingestion pipeline. |
| **Tool** | `search_video` | Semantic search across indexed videos. Returns timestamps and snippets. |
| **Tool** | `list_videos` | Shows all currently indexed videos and their status. |
| **Resource** | `video://{id}/transcript` | Provides the full, timestamped transcript of a specific video. |

---

## 5. Development Roadmap (MVP Steps)

### Phase 1: The Core (Go + Docker)

* [ ] Setup Go project with MCP SDK.
* [ ] Create Dockerfile with FFmpeg, Whisper.cpp, and ONNX dependencies.
* [ ] Implement `index_video` tool (just audio-to-text first).
* [ ] Implement LanceDB storage for transcripts.

### Phase 2: Multimodal (The "See" part)

* [ ] Integrate CLIP for frame embedding.
* [ ] Update `search_video` to handle hybrid search (Visual + Text).
* [ ] Add "Visual Context" to the search results returned to the LLM.

### Phase 3: Scaling & Gateway

* [ ] Ensure LanceDB can use S3/GCS as a backend for Cloud Run.
* [ ] Test integration with `AgentGateway.dev` for API key management and RBAC.

---

## 6. Business Logic & Scalability

* **Cost Efficiency:** Using Cloud Run means we only pay for the Go-binary's uptime. Video processing can be offloaded to Cloud Run Jobs (with GPU support) for heavy lifting.
* **Data Gravity:** By keeping the index in LanceDB (files), we can easily move the "Brain" of the system between local laptops and enterprise cloud buckets without changing the code.

---
