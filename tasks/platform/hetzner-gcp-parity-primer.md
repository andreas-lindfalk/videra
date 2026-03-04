# Hetzner + GCP Deployment Parity Primer

Goal: support EU data-residency-sensitive customers with Hetzner as a first-class deployment path, while keeping Cloud Run equally supported.

Related runbook:

- `tasks/platform/hetzner-vm-docker-runbook.md`
- `tasks/platform/cloud-run-runbook.md`
- `tasks/platform/parity-validation-checklist.md`

## Quick Mental Model

- Cloud Run is a serverless container runtime.
- Hetzner has no direct Cloud Run equivalent.
- Hetzner-first path is usually:
  - Docker containers on Hetzner Cloud VM(s), or
  - Kubernetes on Hetzner (self-managed or managed tooling around Hetzner infra).

## Service Mapping (Practical)

| GCP / Cloud Run Concept | Hetzner Practical Equivalent | Notes |
|---|---|---|
| Cloud Run service (HTTP container) | Docker container on Hetzner VM behind reverse proxy (Caddy/Traefik/Nginx) | Start simple with one VM and systemd/compose |
| Cloud Run Jobs (batch indexing) | Cron/systemd jobs on VM or K8s Jobs | Keep orchestrator contract unchanged |
| Cloud Storage bucket | Persistent volume / Storage Box / external EU S3-compatible object store | Choose based on throughput and API needs |
| Cloud Logging/Monitoring | Prometheus + Grafana + Loki/ELK (or hosted EU observability) | Keep app logs structured JSON |
| IAM/service account auth | Network controls + gateway/JWT at edge | AgentGateway can remain edge control plane |
| Managed autoscaling | VM horizontal scaling + load balancer or K8s HPA | Explicit capacity planning required |

## Recommended Sequence (Low Risk)

1. Keep Docker image and MCP HTTP endpoint identical.
2. Validate parity locally (`make local-e2e`) with fixed fixtures.
3. Deploy same image to Cloud Run and Hetzner VM target.
4. Run identical parity checklist on both:
   - `index_video`
   - `search_video` (deterministic ordering)
   - `list_videos`
   - `video://{id}/transcript`
   - restart + data persistence behavior
5. Record deltas as deployment-layer concerns only (not core MCP behavior changes).

## Architecture Guardrails

- No provider-specific logic in `internal/mcpserver`, `internal/ingestion`, or `internal/storage` interfaces.
- Keep auth/policy at gateway/edge level.
- Keep data path configurable via env vars so local, Hetzner, and Cloud Run can share runtime contracts.
- Prefer idempotent indexing boundaries to support both async job models.

## What to Decide Early

- Preferred Hetzner baseline: single VM + compose first, or K8s-first.
- Data persistence model on Hetzner (volume strategy and backup cadence).
- Edge/gateway placement for enterprise auth + audit requirements.
