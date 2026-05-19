# Implementation Plan

Step-by-step checklist ordered by dependency. Each step must complete before the next begins.

---

## Phase 0: Preconditions

- [ ] 0.1 — Install Dokploy on VPS: `curl -sSL https://dokploy.com/install.sh | sh`
- [ ] 0.2 — Create Docker Hub / ghcr.io token with read/write: packages access for GitHub Actions
- [ ] 0.3 — Add `DOCKER_USERNAME`, `DOCKER_PASSWORD`, `DOKPLOY_URL`, `DOKPLOY_DEPLOY_TOKEN_DEV`, `DOKPLOY_DEPLOY_TOKEN_PROD` to GitHub Actions secrets
- [ ] 0.4 — Verify VPS firewall: expose ports 4000, 4001, 8000, 8001, 3000 (Dokploy admin only)
- [ ] 0.5 — Create Dokploy projects `adisu-dev` and `adisu-prod`
- [ ] 0.6 — Generate Dokploy deploy tokens for both projects, add to GitHub secrets

---

## Phase 1: Dockerfiles

- [ ] 1.1 — Create `core-backend/Dockerfile`
  - Multi-stage: `golang:1.25-alpine` builder → `alpine:3.20` runtime
  - Build `cmd/api` → binary `/app/app`
  - Build `cmd/schema` → binary `/app/schema`
  - OCI labels: `revision`, `version`, `created`

- [ ] 1.2 — Create `ai-service/Dockerfile`
  - Single stage: `python:3.11-slim` base
  - Install `uv`, copy `pyproject.toml` + `uv.lock`, run `uv sync --no-dev`
  - Copy source code
  - OCI labels: `revision`, `version`, `created`

---

## Phase 2: Docker Compose Files

- [ ] 2.1 — Update `docker-compose.yml` (local dev infra)
  - Keep existing infra services as-is
  - Add app services for local development (optional: commented or profiles)

- [ ] 2.2 — Create `docker-compose.dev.yml`
  - Full stack with +1 port offset
  - Image tags via variables: `${CORE_TAG}`, `${AI_TAG}`
  - Infra services: postgres, postgres-ai, redis, rabbitmq, seaweedfs (no host ports)
  - App services: core-backend, ai-server, ai-worker
  - Migration services with `profiles: ["migration"]`
  - Volumes mapped to `/opt/adisu/dev/...`

- [ ] 2.3 — Create `docker-compose.prod.yml`
  - Same structure as dev but standard ports
  - Volumes mapped to `/opt/adisu/prod/...`

---

## Phase 3: Deploy Script

- [ ] 3.1 — Create `scripts/deploy.sh`
  - Shared helper for GitHub Actions workflows
  - Functions:
    - `build_and_push <service> <tag>` — docker buildx + push
    - `call_dokploy <project> <service>` — POST to Dokploy API
  - Services: `core-backend`, `ai-service`

---

## Phase 4: GitHub Actions Workflows

- [ ] 4.1 — Create `.github/workflows/deploy-dev.yml`
  - Trigger: `push` on `dev` branch
  - Jobs:
    - `build-core`: build + push core-backend:dev-<sha>
    - `build-ai`: build + push ai-service:dev-<sha>
    - `deploy`: after both build jobs succeed, call Dokploy API
  - Steps use shared `scripts/deploy.sh`

- [ ] 4.2 — Create `.github/workflows/deploy-prod.yml`
  - Trigger: `release: published`
  - Parse release tag to determine which service (`core-backend-v*` or `ai-service-v*`)
  - Build + push only the released service
  - Call Dokploy API for adisu-prod

---

## Phase 5: Dokploy Configuration

- [ ] 5.1 — In Dokploy UI, configure `adisu-dev` project:
  - Upload `docker-compose.dev.yml`
  - Set environment variables: `CORE_TAG`, `AI_TAG`, `CORE_PORT`, `AI_PORT`, all DB secrets, API keys
  - Add deploy webhook/API token

- [ ] 5.2 — In Dokploy UI, configure `adisu-prod` project:
  - Upload `docker-compose.prod.yml`
  - Set environment variables (same shape, different values)
  - Add deploy webhook/API token

---

## Phase 6: First Deploy & Validation

- [ ] 6.1 — Push `dev` branch → verify GitHub Actions builds and pushes images
- [ ] 6.2 — Call Dokploy API manually to verify project pulls and starts
- [ ] 6.3 — Verify `http://<vps-ip>:4000` returns core-backend health
- [ ] 6.4 — Verify `http://<vps-ip>:4001` returns core-backend health (dev)
- [ ] 6.5 — Verify migrations run correctly
- [ ] 6.6 — Verify ai-worker connects to RabbitMQ and processes events

---

## Phase 7: Go Live

- [ ] 7.1 — Create first GitHub Release → verify prod deploy pipeline
- [ ] 7.2 — Verify prod stack is healthy
- [ ] 7.3 — Update DNS if domain is available later
- [ ] 7.4 — Clean up old deployment docs reference

---

## Files to Create/Modify

| File | Action |
|------|--------|
| `core-backend/Dockerfile` | Create |
| `ai-service/Dockerfile` | Create |
| `docker-compose.yml` | Modify (add dev app services) |
| `docker-compose.dev.yml` | Create |
| `docker-compose.prod.yml` | Create |
| `scripts/deploy.sh` | Create |
| `.github/workflows/deploy-dev.yml` | Create |
| `.github/workflows/deploy-prod.yml` | Create |
| `docs/deployment/README.md` | Modify |
| `docs/deployment/SIMPLE_PLAN.md` | Create |
