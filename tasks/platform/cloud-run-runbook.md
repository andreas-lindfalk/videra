# Cloud Run Runbook (Videra MCP)

Goal: deploy Videra as a managed HTTPS MCP endpoint on GCP Cloud Run with operational checks aligned to the Hetzner runbook.

## 0) Baseline choices

- Platform: Cloud Run service (managed)
- Region: choose one region and keep it fixed per environment
- Exposure: HTTPS endpoint managed by Cloud Run
- MCP URL outcome: `https://<service-url>/mcp`

Image profile guidance:

- Use `runtime-slim` for minimal/simulated or sidecar-driven flows.
- Use `runtime-full` when real-mode fallback tooling (Whisper/OCR) is required.

## 1) Prerequisites

- GCP project with billing enabled
- `gcloud` CLI installed and authenticated
- Docker installed locally for image build/push

Set baseline vars:

```bash
export PROJECT_ID="<your-gcp-project-id>"
export REGION="europe-west1"
export REPO="videra"
export SERVICE="videra-mcp"
export PROFILE="full" # or: slim
export IMAGE="${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO}/videra:$(date +%Y%m%d-%H%M%S)-${PROFILE}"
```

## 2) Enable required APIs

```bash
gcloud config set project "${PROJECT_ID}"
gcloud services enable \
  run.googleapis.com \
  artifactregistry.googleapis.com \
  cloudbuild.googleapis.com
```

## 3) Create Artifact Registry repository

```bash
gcloud artifacts repositories create "${REPO}" \
  --repository-format=docker \
  --location="${REGION}" \
  --description="Videra container images"
```

If it already exists, this command can be skipped.

Configure Docker auth for registry:

```bash
gcloud auth configure-docker "${REGION}-docker.pkg.dev"
```

## 4) Build and push image

From repo root:

```bash
docker build --target "runtime-${PROFILE}" -t "${IMAGE}" .
docker push "${IMAGE}"
```

## 5) Deploy Cloud Run service

```bash
gcloud run deploy "${SERVICE}" \
  --image "${IMAGE}" \
  --region "${REGION}" \
  --platform managed \
  --allow-unauthenticated \
  --port 8080 \
  --min-instances 1 \
  --max-instances 1 \
  --cpu 1 \
  --memory 2Gi \
  --set-env-vars "VIDERA_TRANSPORT=http,VIDERA_HTTP_ADDR=:8080,VIDERA_DATA_DIR=/tmp/videra-data,VIDERA_LOG_LEVEL=info,VIDERA_RUNTIME_MODE=prod"
```

Get service URL:

```bash
export SERVICE_URL="$(gcloud run services describe "${SERVICE}" --region "${REGION}" --format='value(status.url)')"
echo "${SERVICE_URL}"
```

MCP endpoint is:

```bash
echo "${SERVICE_URL}/mcp"
```

## 6) Validate MCP endpoint

From VS Code MCP config or MCP client, use:

- `https://<service-url>/mcp`

Run checks:

1. `list_videos`
2. `search_video`
3. `read_resource` for `video://<id>/transcript` (if indexed data exists)

## 7) Important current constraint (Phase 6)

`index_video` currently expects a server-visible filesystem path.

- On Hetzner VM, this maps naturally to mounted `/videos/...` paths.
- On Cloud Run, local files are not persisted/shared and your laptop path is not directly accessible.

For parity today, treat Cloud Run runbook as endpoint/runtime validation and keep full indexing validation on Hetzner/local until cloud ingestion path is finalized.

Planned parity path:

- move cloud indexing to explicit async jobs,
- source media from cloud object storage,
- keep MCP contracts unchanged.

## 8) Operational basics

Inspect logs:

```bash
gcloud run services logs read "${SERVICE}" --region "${REGION}" --limit=200
```

Update service to new image:

```bash
gcloud run deploy "${SERVICE}" \
  --image "${IMAGE}" \
  --region "${REGION}" \
  --platform managed
```

Describe revision and traffic:

```bash
gcloud run services describe "${SERVICE}" --region "${REGION}"
```

## 9) Troubleshooting

- `404` or connection issues:
  - verify MCP URL includes `/mcp`
  - verify service deployed in expected region/project
- Cold starts/latency spikes:
  - keep `--min-instances 1` during early validation
- Empty search/list results:
  - expected if no indexed data is persisted in this deployment mode

## 10) Parity note

This runbook is deployment-layer only. Keep MCP contracts and core runtime behavior aligned with Hetzner deployment and local validation flows.
