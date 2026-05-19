# Simple Deployment Plan

## Architecture

A single VPS runs both **dev** and **prod** environments via Dokploy, each as a separate Docker Compose project. GitHub Actions builds Docker images on push, pushes to GitHub Container Registry (ghcr.io), then calls the Dokploy API to trigger a redeploy.

### Environments

| | Port Offset | Core | AI Server | AI Worker | Postgres | Postgres-AI | Redis | RabbitMQ | SeaweedFS |
|---|---|---|---|---|---|---|---|---|---|
| **prod** | +0 | :4000 | :8000 | internal | internal | internal | internal | internal | internal |
| **dev** | +1 | :4001 | :8001 | internal | internal | internal | internal | internal | internal |

### Services

- **core-backend** — Go HTTP+gRPC API server (port 4000/4001)
- **ai-server** — Python FastAPI+gRPC server (port 8000/8001)
- **ai-worker** — Python RabbitMQ ingestion consumer (no ports)

Both `ai-server` and `ai-worker` come from the same Docker image; only the `command` differs.

### Deploy Flow

#### Dev (automatic on push to `dev`)

```
git push to dev
  → GitHub Actions (deploy-dev.yml)
      1. Build 3 images: core-backend, ai-service (for server + worker)
      2. Tag as dev-<sha>
      3. Push to ghcr.io
      4. Call Dokploy API → trigger adisu-dev redeploy
  → Dokploy pulls images, runs migrations, composes up
```

#### Prod (manual via GitHub Release)

```
Release-please creates GitHub Release for core-backend/ai-service
  → GitHub Actions (deploy-prod.yml)
      1. Determine which service(s) were released
      2. Build + push only released service(s) with vX.Y.Z tag
      3. Call Dokploy API → trigger adisu-prod redeploy
  → Dokploy pulls images, runs migrations, composes up
```

### Image Tags

| Env | Pattern | Example |
|-----|---------|---------|
| Dev | `dev-<shortsha>` | `ghcr.io/org/core-backend:dev-a1b2c3d` |
| Prod | `v<semver>` | `ghcr.io/org/core-backend:v0.10.0` |

### Migration Strategy

Migrations run as one-shot containers before the API containers restart:

- **core-backend**: `/app/schema -action=apply`
- **ai-service**: `alembic upgrade head`

Both are defined in compose with `profiles: ["migration"]` — they only run when explicitly triggered by the deploy script.

### Dokploy Setup

On the VPS:

1. Install Dokploy via `curl -sSL https://dokploy.com/install.sh | sh`
2. Create two Dokploy projects: `adisu-dev` and `adisu-prod`
3. Each project gets its own Docker Compose YAML uploaded to Dokploy
4. Configure a **Deploy Token** per project for API-based triggers

### Required Repo Artifacts

| File | Purpose |
|------|---------|
| `core-backend/Dockerfile` | Multi-stage Go build (produces `app` + `schema` binaries) |
| `ai-service/Dockerfile` | Python image with uv |
| `docker-compose.dev.yml` | Full dev stack |
| `docker-compose.prod.yml` | Full prod stack |
| `.github/workflows/deploy-dev.yml` | Dev CI/CD |
| `.github/workflows/deploy-prod.yml` | Prod CI/CD |
| `scripts/deploy.sh` | Shared script for build + push + Dokploy API call |

### What's NOT in this plan

- Cosign signing, SBOM, vulnerability scans
- Separate deploy repository
- sops-encrypted secrets
- Compatibility matrix between services
- Caddy/Traefik reverse proxy
- Slack notifications
- Backup/restore tooling
- Rollback scripts
- Hot-reload or blue/green
