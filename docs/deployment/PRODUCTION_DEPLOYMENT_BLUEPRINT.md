# Production Deployment Blueprint (Core + AI)

## 1) Goal

Build a repeatable, production-grade deployment pipeline for `core-backend` and `ai-service` that:

- supports independent and joint service releases,
- uses local image builds/pushes to Docker Hub (phase 1),
- deploys to separate `dev` and `prod` servers,
- can bootstrap a brand-new server in minutes,
- includes rollback, backups, and operational safeguards.

---

## 2) Final Decisions (Locked)

## Platform and Environments

- Deployment model: single-VM Docker Compose (phase 1), not Kubernetes.
- Separate servers for `dev` and `prod` from day one.
- Same OS baseline on both servers: Ubuntu 24.04 LTS.
- Brief restart downtime during deploy is acceptable in phase 1.
- Infrastructure is self-hosted in phase 1; future migration to managed infra is expected.

## Release and Image Strategy

- Docker Hub repos: one per service.
  - `<dockerhub-org>/core-backend`
  - `<dockerhub-org>/ai-service`
- `prod` images come from release tags only.
- `dev` images use immutable commit tags: `dev-<shortsha>`.
- Build architecture: `linux/amd64` only (phase 1).
- Build and push run locally via scripts in this repo.
- Quality gates run before build/push; fail fast on test/lint/check failures.
- Build metadata labels required (`git_sha`, `build_time`, `version`).
- Cosign signing required for produced images.
- SBOM and vulnerability scan reports required per build.

## Deployment and Safety

- Deploy artifacts are immutable image tags; servers only pull and run.
- `prod` deploys are manual and access-restricted.
- `dev` deploys are automatic after local `dev` release script succeeds.
- Deploy scripts use server-side lock (`flock`) to prevent concurrent runs.
- Migration strategy: explicit pre-deploy step (never implicit app startup migration).
- Breaking migrations require explicit `--allow-breaking-migration` flag.
- Deploy includes automated smoke checks and rollback on failure.
- Deploy history must be logged and sent to Slack.

## Networking and Security

- Caddy is ingress proxy with automatic TLS.
- Only `80/443` publicly exposed.
- Internal service ports remain private Docker network only.
- `ai-service` is internal-only; public API routes through core.
- Dedicated `deploy` user, key-only SSH, no password auth.
- Host hardening baseline: UFW, fail2ban, unattended security updates.
- Containers run as non-root with reduced privileges where possible.

## Secrets and Config

- Runtime secrets are server-side only; no plaintext secrets in git.
- Deploy repo stores encrypted env files with `sops` + `age`.
- Separate webhook and secret values per environment (`dev` and `prod`).
- Runtime version must come from deployed image tag/env injection.

## Data, Backups, and DR

- Separate PostgreSQL databases for each service (`core` and `ai`).
- `ai-api` and `ai-worker` run as separate containers from the same image.
- Backups are off-server only (S3-compatible target) for DB/config-critical data.
- File storage backup is deferred (explicitly out of phase 1 scope).
- Backup tooling: `pg_dump` + `restic`.
- Backup cadence:
  - nightly full backup,
  - pre-deploy backup,
  - retain daily 14 days,
  - retain weekly 8 weeks,
  - retain monthly 6 months.
- Weekly restore test in dev is scripted pass/fail and posts to Slack.
- Recovery targets (phase 1): RPO 24h, RTO 60m.

## Governance

- Deploy repo uses single `main` branch + PR review.
- Release operators limited to 1-2 authorized maintainers.
- Controlled first prod cutover window is mandatory.

---

## 3) Target Topology

```text
Internet
   |
   v
Caddy (TLS, 80/443)
   |
   v
core-backend (public API)
   |
   +--> ai-api (internal gRPC only)
   |
   +--> Postgres (core DB)
   +--> Redis
   +--> RabbitMQ
   +--> SeaweedFS

ai-worker (separate container, same image as ai-api)
   |
   +--> RabbitMQ (consume ingestion events)
   +--> Postgres (ai DB)
```

---

## 4) Repository Boundaries

## App Repository (current repo)

Owns:

- service source code,
- Docker build definitions,
- local release/build/push scripts,
- test/lint/build quality gates,
- versioning and tags.

Suggested structure to add:

```text
backend/
  scripts/release/
    release-dev.sh
    release-prod.sh
    common.sh
  docker/
    core-backend.Dockerfile
    ai-service.Dockerfile
  .dockerignore
```

## Deploy Repository (separate)

Owns:

- runtime compose manifests,
- encrypted envs,
- bootstrap/deploy/rollback/backup/restore scripts,
- operational runbooks and compatibility matrix.

Required structure:

```text
deploy-repo/
  compose/
    base.yml
    dev.yml
    prod.yml
  env/
    dev.env.enc
    prod.env.enc
    dev.env.example
    prod.env.example
  scripts/
    bootstrap-server.sh
    deploy.sh
    rollback.sh
    backup.sh
    restore-test.sh
  compatibility.yaml
  runbooks/
    cutover-checklist.md
    incident-rollback.md
```

---

## 5) Tagging and Artifact Policy

## Production

- Build source: checked-out release tag only.
- Tag format:
  - `core-backend:vX.Y.Z`
  - `ai-service:vX.Y.Z`
  - optional convenience: `latest`

## Development

- Build source: merge commit on `dev`.
- Tag format:
  - `core-backend:dev-<shortsha>`
  - `ai-service:dev-<shortsha>`
  - optional convenience: `dev`

## Required Provenance

- Image labels include:
  - `org.opencontainers.image.revision=<sha>`
  - `org.opencontainers.image.version=<tag>`
  - `org.opencontainers.image.created=<timestamp>`
- Cosign signature required.
- SBOM + scan reports generated and archived.

---

## 6) Deployment Flows

## 6.1 Dev Flow (one-command local release)

```text
run release-dev
  -> validate clean git + branch
  -> run quality gates per service
  -> build amd64 images
  -> sign + generate SBOM + scan
  -> push dev-<sha> images to Docker Hub
  -> ssh to dev server
  -> run deploy.sh --env dev --core dev-<sha> --ai dev-<sha>
  -> smoke checks + Slack + deploy log
```

## 6.2 Prod Flow (manual controlled release)

```text
run release-prod --core vX.Y.Z --ai vA.B.C
  -> verify checked-out git tags exist and match requested versions
  -> run quality gates
  -> build amd64 images from tags
  -> sign + SBOM + scan
  -> push versioned images
  -> ssh to prod server
  -> deploy.sh --env prod --core vX.Y.Z --ai vA.B.C
       1) lock
       2) validate signatures/tags/compatibility.yaml
       3) pre-deploy backup
       4) core migrations
       5) ai migrations
       6) pull images + compose up -d
       7) smoke checks
       8) rollback automatically on failure
       9) unlock + Slack + history log
```

## 6.3 Rollback Flow

```text
rollback.sh --env prod --core <previous-tag> --ai <previous-tag>
  -> lock
  -> validate requested tags + compatibility
  -> pull old images
  -> compose up -d with previous tags
  -> smoke checks
  -> Slack + history log
  -> unlock
```

---

## 7) Script Contracts

All scripts must support:

- `--dry-run`
- idempotent re-run behavior
- non-zero exit codes on failure
- structured log output for parsing

## 7.1 App Repo Scripts

## `scripts/release/release-dev.sh`

Inputs:

- optional: `--services core,ai|core|ai`

Behavior:

1. Verify clean git state and correct branch.
2. Determine short SHA.
3. Run quality gates.
4. Build/push/sign selected services with `dev-<sha>` tags.
5. Trigger dev deploy via SSH.

## `scripts/release/release-prod.sh`

Inputs:

- `--core vX.Y.Z` (optional when deploying ai-only)
- `--ai vA.B.C` (optional when deploying core-only)
- optional: `--allow-breaking-migration`

Behavior:

1. Validate requested tags and checked-out source.
2. Run quality gates.
3. Build/push/sign selected service images.
4. Trigger prod deploy via SSH.

## 7.2 Deploy Repo Scripts

## `scripts/bootstrap-server.sh`

Behavior:

1. Install Docker/Compose/Caddy and baseline security tools.
2. Create standard directories (`/opt/adisu/...`).
3. Configure deploy user permissions.
4. Validate host hardening and firewall rules.

## `scripts/deploy.sh`

Inputs:

- `--env dev|prod`
- `--core <tag>` optional
- `--ai <tag>` optional
- `--allow-breaking-migration` optional

Behavior:

1. Acquire lock.
2. Validate compatibility matrix and signed artifacts.
3. Decrypt env using `sops`.
4. Run backup and migrations.
5. Pull and recreate compose services.
6. Run smoke checks.
7. Rollback on failure.
8. Append deploy-history entry + Slack notification.
9. Release lock.

## `scripts/backup.sh`

- Runs `pg_dump` for core and ai DB.
- Ships backup to object storage via `restic`.
- Applies retention policy.

## `scripts/restore-test.sh`

- Restores latest backup to temporary DB on dev server.
- Runs integrity queries.
- Emits pass/fail and Slack notification.

---

## 8) Compose and Runtime Standards

## Service Separation

- `ai-api` and `ai-worker` are separate services in compose.
- Same image, different command.

## Port Exposure

- Public: Caddy only (`80`, `443`).
- Private: all service internals (DB/cache/broker/ai gRPC).

## Volumes

- Persistent host-mounted paths under `/opt/adisu/data/...`.
- No anonymous volumes for stateful services.

## Logging

- Docker logging options:
  - `max-size: 10m`
  - `max-file: 10`
- Host retention:
  - app/access logs: 14 days
  - deploy/rollback logs: 90 days
  - backup logs: 90 days

---

## 9) Compatibility Guard

`compatibility.yaml` is authoritative for allowed core/ai version pairs.

Deploy behavior:

- If pair is not allowed, deploy fails before migrations.
- Independent deploy is allowed only when resulting pair remains compatible.

Example shape:

```yaml
core-backend:
  v0.7.x:
    ai-service: ">=v0.5.0 <v0.6.0"
  v0.8.x:
    ai-service: ">=v0.6.0 <v0.7.0"
```

---

## 10) Rollout Plan (Phased)

## Phase 0: Security and Preconditions

1. Rotate all exposed credentials and invalidate old ones.
2. Remove secrets from repository-tracked env files.
3. Prepare DNS for `api.<domain>` and `api-dev.<domain>`.

## Phase 1: Build and Artifact Pipeline (Local)

1. Add Dockerfiles + `.dockerignore`.
2. Implement `release-dev.sh` and `release-prod.sh`.
3. Add signing, SBOM, vulnerability scan steps.
4. Enforce clean git and branch/tag rules.

## Phase 2: Deploy Repo and Server Bootstrap

1. Create deploy repo structure.
2. Implement compose base/dev/prod.
3. Implement bootstrap, deploy, rollback scripts.
4. Add Caddy and private networking.
5. Add lock, history logging, Slack notifications.

## Phase 3: Data Safety and DR

1. Implement `backup.sh` with `restic`.
2. Configure cron/systemd timers for nightly backups.
3. Implement weekly restore test and Slack reporting.

## Phase 4: Cutover

1. Execute first production release window checklist.
2. Perform supervised deploy and smoke checks.
3. Validate rollback command in staging/dev.
4. Post-cutover review and runbook updates.

---

## 11) First Production Cutover Checklist

Pre-cutover:

- [ ] DNS points to prod server IP.
- [ ] TLS certificates valid via Caddy.
- [ ] Secrets rotated and verified.
- [ ] Backup run successful in last 24h.
- [ ] Rollback tags selected and tested in dry-run.
- [ ] One operator + one reviewer confirmed.

Cutover:

- [ ] Start deploy in scheduled low-traffic window.
- [ ] Monitor logs and health checks live.
- [ ] Confirm API health and key user paths.
- [ ] Confirm Slack deploy event success message.

Post-cutover:

- [ ] Confirm deploy-history entry.
- [ ] Confirm backup after deploy.
- [ ] Record lessons learned and update runbook.

---

## 12) Required Immediate Fixes in Current Repo

1. Add missing Dockerfiles for both services.
2. Add health endpoint contract for `core-backend` suitable for smoke checks.
3. Fix runtime version reporting in `ai-service` (`APP_VERSION` must reflect deployed tag/env).
4. Remove and rotate currently exposed credentials from `.env` files.

---

## 13) Out of Scope (Phase 1)

- Kubernetes orchestration.
- Blue/green or multi-node zero-downtime strategy.
- Automated file-storage backup for SeaweedFS object data.
- Full managed-infra migration.

These are planned for later phases once traffic and operational maturity increase.
