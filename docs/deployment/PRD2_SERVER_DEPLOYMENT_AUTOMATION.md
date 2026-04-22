# PRD2: Server Deployment Automation (Compose, Encrypted Secrets, Safe Rollout)

## Problem Statement

As the team operating two services, we need repeatable deployments on fresh or existing servers without ad-hoc manual steps. Today there is no end-to-end deployment system that guarantees environment parity, deterministic rollout order, compatibility checks, safe migrations, rollback behavior, and secure runtime configuration handling across dev and prod.

## Solution

Implement a deployment automation stack in a dedicated deployment repository with standardized compose topology, encrypted environment management, bootstrap scripts, deploy/rollback scripts, compatibility guards, and smoke-check-driven rollout safety. Keep dev/prod isolated with consistent runtime patterns and minimal public attack surface.

## User Stories

1. As an operator, I want to bootstrap a new server in minutes, so that infra replacement and scaling are fast.
2. As an operator, I want one deployment command that accepts exact image tags, so that rollout is deterministic and auditable.
3. As an operator, I want separate dev and prod servers with separate configs, so that testing mistakes cannot impact production.
4. As an operator, I want deploy scripts to run migrations explicitly before app recreation, so that startup race conditions are eliminated.
5. As an operator, I want deploy scripts to lock execution, so that concurrent deploy/rollback actions cannot corrupt runtime state.
6. As an operator, I want automatic rollback when smoke checks fail, so that failed releases recover quickly.
7. As an operator, I want compatibility checks between core and ai versions, so that independent service deploys remain safe.
8. As a security engineer, I want encrypted secrets with server-side decrypt only, so that plaintext credentials are never committed.
9. As a security engineer, I want only ingress proxy ports publicly exposed, so that internal services remain private.
10. As a backend engineer, I want ai API and ai worker as separate runtime units, so that scaling and failure isolation are clean.
11. As an operator, I want dry-run mode for deploy and rollback scripts, so that risky operations can be validated ahead of time.
12. As an incident responder, I want deploy history logs with actor, tags, and status, so that root-cause timelines are easy to reconstruct.
13. As a team lead, I want Slack notifications for deploy lifecycle events, so that release visibility is immediate.
14. As a maintainer, I want prod deploys manual and permission-restricted, so that accidental production changes are prevented.
15. As a maintainer, I want a strict policy for breaking migrations, so that non-reversible changes require explicit intent.
16. As a platform maintainer, I want runtime service version exposed from deployed tag, so that health endpoints and logs reflect actual release state.

## Implementation Decisions

- Introduce a deep deployment orchestrator module handling lock acquisition, validation gates, migration order, rollout, health verification, and rollback.
- Introduce a bootstrap module for host provisioning and baseline hardening (package installation, user setup, firewall policy checks).
- Introduce encrypted configuration module using `sops` + `age` with per-environment secret boundaries.
- Introduce compatibility policy module as an explicit contract between core and ai version ranges.
- Introduce runtime topology module with shared compose base and environment-specific overlays.
- Introduce notification and audit module for structured event emission to Slack and local deploy-history records.
- Introduce migration control policy requiring explicit operator flag for breaking migration execution.
- Keep deploy operations immutable-tag driven; server never builds runtime images.

## Testing Decisions

- A good test validates deployment behavior at module boundaries: precondition failures, happy paths, and rollback semantics, not command implementation internals.
- Test deploy orchestrator sequencing with mocked external dependencies to guarantee correct order and abort behavior.
- Test compatibility policy evaluation independently with valid and invalid version pairs.
- Test lock behavior and idempotent reruns under repeated invocation scenarios.
- Test smoke-check gate outcomes and rollback trigger paths.
- Test secrets-decryption and missing-secret failure handling at the interface boundary.
- Prior art should follow existing behavior-driven tests in the project (service orchestration, message processing, and error mapping tests).

## Out of Scope

- Kubernetes orchestration.
- Blue/green traffic switching in this phase.
- Public exposure of ai-service endpoints.
- Automated production deploys without operator action.

## Further Notes

- This PRD intentionally optimizes for reliability and control on a single-node runtime.
- The deployment contracts are designed to remain stable when infrastructure later migrates to managed services.
