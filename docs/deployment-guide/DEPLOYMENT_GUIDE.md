# Deployment Guide — Adisu Serategna

## 1. Document Metadata & Change Log

| Field | Value |
|-------|-------|
| **Version** | 1.0.0 |
| **Last Updated** | 2026-05-27 |
| **Target Environments** | Dev, Production |
| **Authors** | Yohannes Solomon ([solomonjohna21@gmail.com](mailto:solomonjohna21@gmail.com), [github.com/johna210](https://github.com/johna210)) |
| **Maintainers** | Yohannes Solomon, Dawit Belay |

### Deployment Log

| Date | Version | Service(s) | Change |
|------|---------|-----------|--------|
| 2026-05-27 | — | — | Initial deployment guide created |
| 2026-05-10 | core-backend v0.9.0 | core-backend | Library admin CRUD, notification templates, canonical events |
| 2026-05-10 | ai-service v0.5.1 | ai-service | Document ingestion fixes |
| 2026-04-29 | core-backend v0.8.0 | core-backend | Notification module (domain → delivery pipeline), campaigns |
| 2026-04-22 | core-backend v0.7.0 | core-backend | AI integration, conversation cache, documents endpoint |
| 2026-04-22 | ai-service v0.5.0 | ai-service | Ask stream, conversation service, feature flags |
| 2026-04-17 | core-backend v0.6.0 | core-backend | IAM, community, AI ingestion, guides, payments |
| 2026-04-17 | ai-service v0.4.0 | ai-service | Domain layer, SQLAlchemy repos, gRPC, ingestion orchestration |
| 2026-04-13 | core-backend v0.5.0 | core-backend | Dev sync |
| 2026-04-13 | ai-service v0.3.0 | ai-service | Dev sync |
| 2026-04-07 | core-backend v0.4.0 | core-backend | OTP, guide state, admin seeder, permissions |
| 2026-04-07 | ai-service v0.2.0 | ai-service | Application usecases, SQLAlchemy repos |
| 2026-04-07 | ai-service v0.1.0 | ai-service | Foundation, initial schema, domain contracts |
| 2026-03-18 | core-backend v0.3.0 | core-backend | Avatar upload, user profile, localization |
| 2026-03-12 | core-backend v0.2.0 | core-backend | Register/login, JWT, IAM domain, gin setup |
| 2026-03-?? | core-backend v0.1.0 | core-backend | Initial release |

---

## 2. Architecture & System Overview

### Architecture Diagram

```mermaid
architecture-beta
    group api(cloud)[API Layer]
    group infra(cloud)[Infrastructure]
    group ai(cloud)[AI Layer]

    service traefik(server)[Traefik] in api
    service core(internet)[core-backend] in api
    service web(server)[Web Admin] in api
    service mobile(device)[Mobile App] in api

    service postgres(database)[PostgreSQL :5432] in infra
    service pgvector(database)[pgvector :5433] in infra
    service redis(database)[Redis :6379] in infra
    service rabbitmq(database)[RabbitMQ :5672] in infra
    service seaweed(database)[SeaweedFS] in infra

    service aiservice(server)[ai-server :8000] in ai
    service aiworker(server)[ai-worker] in ai

    traefik:R --> L:core
    web:T --> B:traefik
    mobile:T --> B:traefik
    core:R --> L:aiservice
    core:B --> T:postgres
    core:B --> T:redis
    core:B --> T:rabbitmq
    core:B --> T:seaweed
    aiservice:B --> T:pgvector
    aiservice:B --> T:redis
    aiservice:B --> T:rabbitmq
    aiservice:B --> T:seaweed
    aiworker:B --> T:rabbitmq
    aiworker:B --> T:pgvector
    aiservice:R --> L:core
```

### Tech Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| **Main API** | Go (Gin + Huma + FX) | 1.25.3 |
| **AI Service** | Python (FastAPI + gRPC) | 3.11.9 |
| **Database (Core)** | PostgreSQL | 18-alpine |
| **Database (AI)** | PostgreSQL + pgvector | pg18 |
| **Cache** | Redis | 8-alpine |
| **Message Broker** | RabbitMQ | 3.13-management-alpine |
| **File Storage** | SeaweedFS (Master + Volume + Filer) | latest |
| **Web Admin** | Next.js | 16.1.6 |
| **Mobile** | Flutter | 3.11+ |

### Services & Ports

| Service | Internal Port | External | Protocol |
|---------|--------------|----------|----------|
| core-backend | 4000 | 4000 (prod) / 4001 (dev) | HTTP |
| core-backend gRPC | 50051 | — | gRPC |
| ai-server | 8000 | 8000 (prod) / 8001 (dev) | HTTP |
| ai-server gRPC | 50051 | — | gRPC |
| ai-server metrics | 9090 | — | Prometheus |
| PostgreSQL (core) | 5432 | — | TCP |
| PostgreSQL (AI) | 5433 | — | TCP |
| Redis | 6379 | — | TCP |
| RabbitMQ AMQP | 5672 | — | TCP |
| RabbitMQ Management | 15672 | — | HTTP |
| SeaweedFS Master | 9333 | — | HTTP |
| SeaweedFS Volume | 8080 | — | HTTP |
| SeaweedFS Filer | 8888 | 443 (via Traefik) | HTTP |

### Network & Security

- **Reverse Proxy**: Traefik (managed by Dokploy) with Let's Encrypt TLS
- **Public endpoints**: `api.johna.me` → core-backend, `files.johna.me` → SeaweedFS Filer
- **Internal-only**: All infrastructure services (DB, Redis, RabbitMQ, SeaweedFS internal)
- **AI service**: Internal-only; public API routes through core-backend
- **Network**: Custom Docker network `dokploy-network`
- **Container Registry**: `ghcr.io/final-year-project-g22/`

### Docker Compose Files

| File | Purpose |
|------|---------|
| `backend/docker-compose.yml` | Base infrastructure (not used directly in deploy) |
| `backend/docker-compose.dev.yml` | Dev stack (port offset +1, Traefik labels) |
| `backend/docker-compose.prod.yml` | Production stack (standard ports, no Traefik) |

### Image Naming

| Service | Registry Path |
|---------|--------------|
| core-backend | `ghcr.io/final-year-project-g22/core-backend` |
| ai-service | `ghcr.io/final-year-project-g22/ai-service` |

---

## 3. Prerequisites & Environment Setup

### Server Specification

| Resource | Dev & Prod |
|----------|-----------|
| **Provider** | Google Cloud Platform (GCP) |
| **Machine Type** | `e2-custom-4-8192` (4 vCPU, 8 GB RAM) |
| **Disk** | 80 GB balanced persistent disk |
| **OS** | Ubuntu Minimal 2604 LTS (resolute) |
| **Location** | `us-central1-a` |

> **Note**: 8 GB RAM is the minimum for this stack. Under load (document ingestion, LLM inference, concurrent users), consider upgrading to 16 GB.

### Access & Permissions

| Method | Purpose |
|--------|---------|
| **GitHub Actions** | CI/CD — build, push to ghcr.io, trigger Dokploy deploy |
| **Dokploy UI** | Service management, env vars, deploy logs |
| **SSH (key-only)** | Manual debugging, emergency rollback |
| **GitHub Container Registry** | Image storage (`ghcr.io`) |

### Required GitHub Actions Secrets

| Secret | Used By | Description |
|--------|---------|-------------|
| `GITHUB_TOKEN` | deploy-dev.yml, deploy-prod.yml | Auto-provided, requires `packages: write` |
| `DOKPLOY_URL` | deploy-dev.yml, deploy-prod.yml | Dokploy API endpoint |
| `DOKPLOY_COMPOSE_ID_DEV` | deploy-dev.yml | Dokploy dev compose project ID |
| `DOKPLOY_COMPOSE_ID_PROD` | deploy-prod.yml | Dokploy prod compose project ID |

### Dokploy Configuration

| Project | Compose File | Port Offset |
|---------|-------------|-------------|
| `adisu-dev` | `docker-compose.dev.yml` | +1 (4001, 8001) |
| `adisu-prod` | `docker-compose.prod.yml` | 0 (4000, 8000) |

Both projects must have all environment variables configured in the Dokploy UI (see Section 4).

### Dependencies

Pre-installed on the VPS (managed by Dokploy):

| Software | Purpose |
|----------|---------|
| Docker Engine 24+ | Container runtime |
| Docker Compose v2 | Service orchestration |
| Traefik | Reverse proxy + TLS (managed by Dokploy) |

---

## 4. Configuration & Secrets Management

### Environment Variables

Variables are set in the Dokploy UI per environment (dev/prod). Never commit production secrets to the repository.

#### Core Backend (`core-backend`)

| Variable | Description | Dev Example | Prod Example |
|----------|-------------|-------------|--------------|
| `APP_ENVIRONMENT` | Runtime environment | `development` | `production` |
| `CORE_TAG` | Image tag to deploy | `dev-abc1234` | `v0.9.0` |
| `CORE_PORT` | Host port for core-backend | `4001` | `4000` |
| `DATABASE_USER` | PostgreSQL user | `user_adisu` | `user_adisu` |
| `DATABASE_PASSWORD` | PostgreSQL password | — | — |
| `DATABASE_DBNAME` | Core database name | `adisu_db` | `adisu_db` |
| `JWT_SECRET` | JWT signing key | — | — |
| `JWT_ACCESS_TOKEN_TTL` | Access token TTL | `15m` | `15m` |
| `JWT_REFRESH_TOKEN_TTL` | Refresh token TTL | `168h` | `168h` |
| `RABBITMQ_USERNAME` | RabbitMQ user | `guest` | `guest` |
| `RABBITMQ_PASSWORD` | RabbitMQ password | — | — |
| `RABBITMQ_VHOST` | RabbitMQ vhost | `/` | `/` |
| `EMAIL_HOST` | SMTP host | `smtp.gmail.com` | `smtp.gmail.com` |
| `EMAIL_PORT` | SMTP port | `587` | `587` |
| `EMAIL_USERNAME` | SMTP username | — | — |
| `EMAIL_PASSWORD` | SMTP password | — | — |
| `EMAIL_FROM` | From address | `noreply@adisu-serategna.com` | `noreply@adisu-serategna.com` |
| `EMAIL_FROM_NAME` | From display name | `Adisu Serategna` | `Adisu Serategna` |
| `EMAIL_ENABLED` | Enable email sending | `false` | `true` |
| `OAUTH_ENCRYPTION_KEY` | OAuth state encryption key | — | — |
| `OAUTH_MOBILE_REDIRECT_BASE_URL` | Mobile OAuth redirect base | — | — |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | — | — |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | — | — |
| `OAUTH_GOOGLE_REDIRECT_URI` | Google OAuth redirect URI | — | — |
| `FACEBOOK_CLIENT_ID` | Facebook OAuth client ID | — | — |
| `FACEBOOK_CLIENT_SECRET` | Facebook OAuth client secret | — | — |
| `OAUTH_FACEBOOK_REDIRECT_URI` | Facebook OAuth redirect URI | — | — |
| `RESEND_API_KEY` | Resend email API key | — | — |
| `RESEND_FROM_EMAIL` | Resend from address | — | — |
| `RESEND_FROM_NAME` | Resend from name | — | — |
| `RESEND_ENABLED` | Enable Resend | `false` | `false` |
| `CHAPA_SECRET_KEY` | Chapa payment secret key | — | — |
| `CHAPA_PUBLIC_KEY` | Chapa payment public key | — | — |
| `CHAPA_WEBHOOK_SECRET` | Chapa webhook secret | — | — |
| `CHAPA_CALLBACK_URL` | Chapa callback URL | — | — |
| `CHAPA_RETURN_URL` | Chapa return URL | — | — |
| `CHAPA_ENABLED` | Enable Chapa payments | `false` | `true` |
| `FCM_CREDENTIALS_FILE` | Firebase Cloud Messaging credentials | — | — |
| `IAM_SUPER_ADMIN_EMAIL` | Super admin seed email | — | — |
| `IAM_SUPER_ADMIN_PASSWORD` | Super admin seed password | — | — |
| `INGESTION_SIGNING_ACTIVE_KEY_ID` | Ingestion signing key ID | `ingestion-v1` | `ingestion-v1` |
| `INGESTION_SIGNING_ACTIVE_KEY_SECRET` | Ingestion signing key secret | — | — |

#### AI Service (`ai-server`, `ai-worker`)

| Variable | Description | Notes |
|----------|-------------|-------|
| `AI_TAG` | Image tag to deploy | e.g., `dev-abc1234` or `v0.5.1` |
| `AI_PORT` | Host port for ai-server | `8001` (dev) / `8000` (prod) |
| `COHERE_API_KEY` | Cohere LLM API key | — |
| `GEMINI_API_KEY` | Gemini API key (non-Vertex) | — |
| `GOOGLE_APPLICATION_CREDENTIALS` | GCP service account path | Mounted as volume: `/app/gcp-credentials.json` |
| `GEMINI_VERTEX_PROJECT` | GCP project for Vertex AI | `mindful-backup-495108-n5` |

### Secrets Storage

| Secret Type | Storage Location |
|-------------|-----------------|
| API keys, DB passwords, JWT secrets | **Dokploy UI** — per-project environment variables |
| GitHub tokens, Dokploy webhook URLs | **GitHub Actions Secrets** |
| GCP service account key | **File on VPS** mounted as Docker volume (`mindful-backup-495108-n5-640600bd5b33.json`) |

### Environment Differences (Dev vs Prod)

| Aspect | Dev | Prod |
|--------|-----|------|
| Docker compose file | `docker-compose.dev.yml` | `docker-compose.prod.yml` |
| Port offset | +1 (4001, 8001) | 0 (4000, 8000) |
| Traefik exposure | Yes (`api.johna.me`) | Yes (`api.johna.me`) |
| Email sending | Usually disabled | Enabled |
| Payments (Chapa) | Usually disabled (test keys) | Enabled (live keys) |
| Image tag pattern | `dev-<sha>` | `vX.Y.Z` |
| Trigger | Push to `dev` branch | GitHub Release published |

### Web Frontend (Vercel)

| Variable | Dev | Prod |
|----------|-----|------|
| `NEXT_PUBLIC_API_URL` | `https://api-dev.johna.me` | `https://api.johna.me` |
| Deploy trigger | Push to `dev` | Push to `main` |
| Install command | `pnpm install --frozen-lockfile` | `pnpm install --frozen-lockfile` |

---

## 5. Step-by-Step Deployment Instructions

### Automated Deployment (CI/CD — Primary Path)

#### Dev Deployment

Triggered automatically on every push to the `dev` branch.

1. **Push to `dev`**:
   ```bash
   git push origin dev
   ```

2. **GitHub Actions** (`deploy-dev.yml`) runs:
   - Checkout code
   - Set up Docker Buildx (docker-container driver)
   - Log in to `ghcr.io` via `GITHUB_TOKEN`
   - Build and push `core-backend:dev-<sha>` + `core-backend:dev` (moving tag)
   - Build and push `ai-service:dev-<sha>` + `ai-service:dev` (moving tag)
   - POST to Dokploy API to trigger deploy

3. **Dokploy** pulls new images and restarts services:
   - `docker compose pull`
   - Runs migration containers (`migration-core`, `migration-ai`)
   - `docker compose up -d` for all services

#### Production Deployment

Triggered by publishing a GitHub Release.

1. **Create a GitHub Release** with tag format:
   - `core-backend-vX.Y.Z` for core-backend releases
   - `ai-service-vX.Y.Z` for ai-service releases

   Releases are managed via Release-Please (automated PRs + version bumps).

2. **GitHub Actions** (`deploy-prod.yml`) runs:
   - Parse release tag → determine service + version
   - Set up Docker Buildx
   - Log in to `ghcr.io`
   - Build and push only the released service:
     - `ghcr.io/.../core-backend:vX.Y.Z` + `:latest`
     - `ghcr.io/.../ai-service:vX.Y.Z` + `:latest`
   - POST to Dokploy API for production project

3. **Dokploy** pulls the new image and restarts the affected service.

### Manual Deployment (Fallback)

Use when CI/CD is unavailable or for debugging.

```bash
# SSH into the VPS
ssh deploy@<vps-ip>

# Navigate to deploy directory (managed by Dokploy)
cd /opt/adisu/prod  # or /opt/adisu/dev

# Pull the new images
CORE_TAG=v0.9.0 AI_TAG=v0.5.1 docker compose -f docker-compose.prod.yml pull

# Run migrations
docker compose -f docker-compose.prod.yml --profile migration run migration-core
docker compose -f docker-compose.prod.yml --profile migration run migration-ai

# Restart services
CORE_TAG=v0.9.0 AI_TAG=v0.5.1 docker compose -f docker-compose.prod.yml up -d
```

### Database Migrations

Migrations run automatically as part of the Dokploy deploy via one-shot containers:

| Service | Command | Tool |
|---------|---------|------|
| Core DB | `/app/schema -action=apply` | Atlas |
| AI DB | `alembic upgrade head` | Alembic |

To run migrations manually:

```bash
# Core migrations
docker compose --profile migration run --rm migration-core

# AI migrations
docker compose --profile migration run --rm migration-ai
```

> **Note**: Migrations require the target database to be healthy. The compose file enforces this via `depends_on: condition: service_healthy`.

### Pre-Deployment Steps

- [ ] Verify the target branch/tag is correct
- [ ] Check that upstream infrastructure (DB, Redis, RabbitMQ) is healthy
- [ ] For production: ensure a recent database backup exists
- [ ] Review migration SQL for breaking changes (if any)

---

## 6. Smoke Testing & Verification

### Health Check Endpoints

| Service | Endpoint | Expected Response |
|---------|----------|-------------------|
| core-backend | `GET https://api.johna.me/api/v1/health` | HTTP 200 |
| ai-service | `GET http://localhost:8000/health` | `{"status": "healthy", "service": "ai-service", "version": "..."}` |
| ai-service (root) | `GET http://localhost:8000/` | `{"service": "ai-service", "version": "..."}` |

### Container Health Checks

Infrastructure services have Docker-level health checks configured:

| Service | Check Command |
|---------|--------------|
| PostgreSQL (core) | `pg_isready -U $DATABASE_USER -d $DATABASE_DBNAME` |
| PostgreSQL (AI) | `pg_isready -U $DATABASE_USER -d adisu_ai` |
| Redis | `redis-cli ping` |
| RabbitMQ | `rabbitmq-diagnostics -q ping` |
| SeaweedFS Master | `wget -qO- http://localhost:9333/cluster/status` |
| SeaweedFS Filer | `wget -qO- http://localhost:8888/` |

### Verification Commands

```bash
# Check all container statuses
docker compose -f docker-compose.prod.yml ps

# Verify core-backend health
curl -s -o /dev/null -w "%{http_code}" https://api.johna.me/api/v1/health

# Verify ai-service health (from within the Docker network)
docker compose exec ai-server curl -s localhost:8000/health

# Check core-backend logs (look for startup success)
docker compose logs --tail 50 core-backend

# Check ai-server logs
docker compose logs --tail 50 ai-server

# Check ai-worker logs
docker compose logs --tail 50 ai-worker

# Verify infrastructure health
docker compose ps | grep -E "(healthy|unhealthy)"
```

### Successful Startup Indicators

| Service | Log Line |
|---------|----------|
| core-backend | `Starting HTTP server` + `Huma API initialized` |
| ai-server | `Uvicorn running on` |
| ai-worker | Worker starts consuming from RabbitMQ queues |
| Migrations (core) | `atlas: apply finished` |
| Migrations (AI) | `INFO  [alembic.runtime.migration] Running upgrade` |

### Critical User Paths (Manual Checks)

- [ ] Login to the admin dashboard (web)
- [ ] Navigate to dashboard — stats load
- [ ] List guides in the guide module
- [ ] Open AI Ask page — chat loads
- [ ] Upload a document to Knowledge Base
- [ ] Verify community categories load
- [ ] Check that payment plans/subscriptions are accessible

---

## 7. Rollback Plan

### Trigger Criteria

Abort the deployment and roll back if any of the following occur within 5 minutes of deploy:

1. **Health endpoint failure**: `GET /api/v1/health` returns non-200 for more than 2 minutes
2. **Elevated 5xx errors**: Noticeable spike in HTTP 5xx responses from API
3. **Uptime monitor alert**: Automated alert from uptime monitoring (e.g., Better Uptime, Pingdom)
4. **UI/UX failure**: Dashboard or critical page fails to load or shows errors
5. **Migration failure**: Migration container exits with non-zero code

### Rollback Steps

#### Option A: Dokploy UI (Recommended)

1. Go to Dokploy UI → project (`adisu-dev` or `adisu-prod`)
2. Navigate to the service deployment history
3. Find the previous working deployment
4. Click "Redeploy" to roll back to the previous image tag

#### Option B: Manual SSH Rollback

```bash
# SSH into the VPS
ssh deploy@<vps-ip>

# Set the previous known-good tags
CORE_TAG=<previous-version> AI_TAG=<previous-version>

# Pull the old images
docker compose -f docker-compose.prod.yml pull

# Run migrations (if reverting schema changes)
# docker compose --profile migration run migration-core
# docker compose --profile migration run migration-ai

# Restart with previous versions
CORE_TAG=$CORE_TAG AI_TAG=$AI_TAG docker compose -f docker-compose.prod.yml up -d
```

#### Option C: Image Tag Revert

```bash
# Re-tag and push the previous image as the current tag
docker pull ghcr.io/final-year-project-g22/core-backend:<previous-version>
docker tag ghcr.io/final-year-project-g22/core-backend:<previous-version> ghcr.io/final-year-project-g22/core-backend:latest
docker push ghcr.io/final-year-project-g22/core-backend:latest

# Trigger Dokploy redeploy
curl -X POST $DOKPLOY_URL/api/deploy/compose/$DOKPLOY_COMPOSE_ID
```

### Data Preservation

- **Database**: Deployments use forward-only migrations. Reverting a migration must be done manually. Document the migration SQL before deploying to enable safe reversion.
- **Uploaded files**: SeaweedFS data is persistent on Docker volumes. Rollback does not affect stored files.
- **Message queue**: RabbitMQ messages may be lost during rollback. Check queue depths after rollback and re-publish if necessary.

---

## 8. Troubleshooting & Common Failure Modes

### Known Issues

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| `Error: Connection Refused` (core-backend → DB) | PostgreSQL container not fully initialized | Wait for health check; check `docker compose logs postgres` |
| SeaweedFS upload fails | Filer not ready before volume; file permission issues | Restart SeaweedFS containers: `docker compose restart seaweed-filer seaweed-volume` |
| Migrations not running | Migration profile not executed; container exited early | Run manually: `docker compose --profile migration run --rm migration-core` |
| `GCP credentials not found` | Service account JSON file path incorrect or missing | Verify `./mindful-backup-495108-n5-640600bd5b33.json` exists in the backend directory and is mounted correctly |
| `Error: dial tcp: connect: connection refused` (ai-service → core-backend gRPC) | Core-backend gRPC server not ready when ai-service starts | Ensure `depends_on` conditions are correct; increase startup delay |
| AI ingestion worker not processing | RabbitMQ connection failure; worker crashed | Check `docker compose logs ai-worker`; verify RabbitMQ is healthy |
| RabbitMQ management UI requires login | Default guest credentials not matching environment | Verify `RABBITMQ_USERNAME` and `RABBITMQ_PASSWORD` match |
| Docker disk space full | Logs or images accumulating | Run `docker system prune -af` (caution: removes unused images and containers) |

### Resource Issues

| Issue | Signs | Mitigation |
|-------|-------|------------|
| High memory usage | Docker stats show >80% RAM, OOM kills | Consider upgrading to 16 GB; reduce AI service batch sizes |
| Disk space | Container logs growing, DB WAL files | Set up log rotation; monitor disk with `df -h` |
| CPU spikes | During document ingestion or LLM inference | Expected behavior; monitor duration |

### Log Inspection

```bash
# View real-time logs for a service
docker compose logs -f core-backend
docker compose logs -f ai-server
docker compose logs -f ai-worker

# View last N lines
docker compose logs --tail 100 core-backend

# Filter logs by error level
docker compose logs core-backend 2>&1 | grep -i error

# Check specific container
docker logs adisu_db --tail 50
```

### Contact Escalation

| Priority | Contact | Method |
|----------|---------|--------|
| Primary | Yohannes Solomon | solomonjohna21@gmail.com |
| Secondary | Dawit Belay | — |
| Emergency | GitHub Issues | https://github.com/Final-Year-Project-G22/backend/issues |
