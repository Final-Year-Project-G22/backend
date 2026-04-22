# PRD3: Operations, Disaster Recovery, and Production Cutover Governance

## Problem Statement

As the team preparing for production, we need operational safeguards beyond deployment commands: backup integrity, restore confidence, retention controls, host hardening, alerting, and supervised cutover procedure. Without these controls, a successful deploy pipeline can still fail under incidents, data loss, or security events.

## Solution

Implement an operations layer that defines backup/restore standards, recovery targets, host and container hardening baselines, logging/retention policy, observability alerts, and first-cutover governance. Treat operational readiness as a release gate, not a follow-up task.

## User Stories

1. As a product owner, I want clear RPO/RTO targets, so that reliability expectations are explicit.
2. As an operator, I want nightly and pre-deploy backups, so that recent recovery points are always available.
3. As an operator, I want backups stored off-server, so that server loss does not eliminate recovery capability.
4. As an operator, I want retention policies for daily/weekly/monthly backup sets, so that storage and recoverability are balanced.
5. As an operator, I want weekly automated restore validation in dev, so that backup usability is continuously proven.
6. As an incident responder, I want restore test pass/fail results in Slack, so that operational drift is detected immediately.
7. As an SRE, I want explicit log rotation and retention, so that disk exhaustion incidents are prevented.
8. As an SRE, I want health checks and uptime monitors, so that outages are discovered quickly.
9. As a security engineer, I want strict host hardening defaults, so that common server attacks are reduced.
10. As a security engineer, I want token-based registry authentication and rotation policy, so that credential blast radius is minimized.
11. As a security engineer, I want non-root runtime and least privilege defaults in containers, so that compromise impact is limited.
12. As an engineering manager, I want supervised first production cutover with defined roles, so that launch risk is controlled.
13. As an engineering manager, I want mandatory pre-cutover checks and rollback readiness, so that release confidence is objective.
14. As an operator, I want operational runbooks for deploy, rollback, incident response, and restore, so that emergency actions are repeatable.
15. As a maintainer, I want clear phase boundaries for what is not yet implemented, so that roadmap communication remains realistic.

## Implementation Decisions

- Define recovery objectives for phase 1 as RPO 24h and RTO 60m.
- Define backup subsystem using `pg_dump` and `restic` with encrypted off-site storage and retention policies.
- Define restore-validation subsystem that performs scheduled restore drills into temporary databases with integrity checks.
- Define observability baseline with health probes, core metrics, container/host resource monitoring, and Slack alerts.
- Define logging policy module with explicit retention windows for application logs, deploy logs, and backup logs.
- Define host hardening baseline covering firewall posture, SSH restrictions, fail2ban, and unattended security updates.
- Define production cutover governance model requiring operator + reviewer and low-traffic execution window.
- Defer file-storage backup automation while explicitly documenting it as future scope.

## Testing Decisions

- A good test validates operational outcomes (backup created, backup restorable, alerts emitted, retention applied), not the internal shell implementation.
- Test backup module for successful snapshot creation, upload, retention pruning, and failure reporting behavior.
- Test restore-test module for both pass and fail integrity outcomes and notification behavior.
- Test log-retention policy enforcement behavior and protection against unbounded growth.
- Test operational readiness checklist gates as explicit pass/fail criteria before production cutover.
- Prior art should follow existing project patterns for deterministic integration-style tests around infrastructure boundaries and error mapping.

## Out of Scope

- Automated backup/restore of file-object storage in phase 1.
- Multi-region disaster recovery architecture.
- Near-zero-downtime blue/green traffic management.
- Managed-cloud migration of all stateful infrastructure.

## Further Notes

- This PRD closes the gap between "deployable" and "operable".
- Operational readiness items should be tracked as release blockers for first production launch.
