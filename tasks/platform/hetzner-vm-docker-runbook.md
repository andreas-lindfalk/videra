# Hetzner VM Docker Runbook (Videra MCP)

Goal: run Videra in an EU-hosted Hetzner VM with persistent data, TLS, and a stable MCP endpoint.

Image profile guidance:

- Use `*-slim` image tags for minimal/default flows.
- Use `*-full` image tags when real-mode fallback tooling (Whisper/OCR) is required.

## 0) Baseline choices

- Target: Hetzner Cloud VM (Ubuntu 24.04 LTS recommended)
- Region: pick EU region (e.g. Falkenstein/Nuremberg/Helsinki)
- Exposure: HTTPS endpoint via Caddy reverse proxy
- MCP URL outcome: `https://<your-domain>/mcp`

## 1) Provision VM and DNS

1. Create VM in Hetzner Cloud:
   - Suggested starter size: 2 vCPU / 4 GB RAM
   - Add SSH key during creation
2. Create DNS A record:
   - `<your-domain>` -> VM public IP
3. In Hetzner firewall (or host firewall), allow inbound:
   - `22/tcp` (SSH)
   - `80/tcp` (HTTP for TLS challenge)
   - `443/tcp` (HTTPS)

## 2) Install Docker + Compose plugin

SSH into VM:

```bash
ssh root@<vm-ip>
```

Install Docker engine and compose plugin:

```bash
apt-get update
apt-get install -y ca-certificates curl gnupg
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo \"$VERSION_CODENAME\") stable" \
  > /etc/apt/sources.list.d/docker.list

apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
systemctl enable --now docker
```

## 3) Create deployment directory

```bash
mkdir -p /opt/videra/{data,videos,caddy}
cd /opt/videra
```

`data/` is persistent vector/runtime storage.
`videos/` is optional local ingest source folder.

## 4) Add runtime config

Create `/opt/videra/.env`:

```bash
cat > /opt/videra/.env <<'EOF'
VIDERA_TRANSPORT=http
VIDERA_HTTP_ADDR=:8080
VIDERA_DATA_DIR=/data
VIDERA_LOG_LEVEL=info
VIDERA_RUNTIME_MODE=prod
DOMAIN=<your-domain>
VIDERA_IMAGE=ghcr.io/andreas-lindfalk/videra:latest-full
EOF
```

If you do not publish to GHCR yet, replace `VIDERA_IMAGE` with your registry/tag.

## 5) Add compose and Caddy config

Create `/opt/videra/docker-compose.yml`:

```yaml
services:
  videra:
    image: ${VIDERA_IMAGE}
    restart: unless-stopped
    env_file:
      - .env
    environment:
      VIDERA_TRANSPORT: ${VIDERA_TRANSPORT}
      VIDERA_HTTP_ADDR: ${VIDERA_HTTP_ADDR}
      VIDERA_DATA_DIR: ${VIDERA_DATA_DIR}
      VIDERA_LOG_LEVEL: ${VIDERA_LOG_LEVEL}
      VIDERA_RUNTIME_MODE: ${VIDERA_RUNTIME_MODE}
    volumes:
      - ./data:/data
      - ./videos:/videos
    expose:
      - "8080"

  caddy:
    image: caddy:2-alpine
    restart: unless-stopped
    depends_on:
      - videra
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./caddy/Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config

volumes:
  caddy_data:
  caddy_config:
```

Create `/opt/videra/caddy/Caddyfile`:

```caddyfile
{$DOMAIN} {
  encode gzip
  reverse_proxy videra:8080
}
```

## 6) Start service

```bash
cd /opt/videra
docker compose up -d
docker compose ps
```

Inspect logs:

```bash
docker compose logs -f videra
```

## 7) Validate MCP endpoint

From your local machine (or any MCP client), point to:

- `https://<your-domain>/mcp`

Then run MCP checks:

1. `list_videos`
2. `index_video` with server-visible path (example: `/videos/IMG_3711.MOV`)
3. `search_video` with a test query
4. `read_resource` for `video://<videoId>/transcript`

If indexing local file from VM, upload file first:

```bash
scp ./IMG_3711.MOV root@<vm-ip>:/opt/videra/videos/
```

## 8) Operational basics

Restart service:

```bash
cd /opt/videra
docker compose restart
```

Upgrade image:

```bash
cd /opt/videra
docker compose pull
docker compose up -d
```

Backup runtime data:

```bash
tar -czf /root/videra-data-backup-$(date +%F).tgz -C /opt/videra data
```

## 9) Troubleshooting

- TLS not issuing:
  - verify DNS A record points to VM
  - verify ports `80/443` are open
- MCP connection errors:
  - verify endpoint uses `/mcp`
  - check `docker compose logs -f videra caddy`
- `index_video` path not found:
  - file must exist inside container path (`/videos/...`)

## 10) Parity note

This runbook is deployment-layer only. Keep MCP contracts and core runtime behavior identical to Cloud Run deployments.
