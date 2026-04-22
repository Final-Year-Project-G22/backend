# PRD1: Release Artifact Pipeline (Local Build, Signed Images, Immutable Tags)

## Problem Statement

As the backend team, we need a reliable way to build and publish deployable artifacts for two independent services (`core-backend` and `ai-service`) without wasting GitHub Actions minutes. Today there is service versioning, but there is no production-grade release artifact pipeline that guarantees traceable, reproducible, signed images for both dev and prod release lanes.

## Solution

Introduce a local-operator-driven release pipeline that enforces quality gates before build, produces immutable Docker Hub tags for both services, signs all artifacts, and emits SBOM/security evidence. Dev releases are commit-tagged and can be deployed quickly, while prod releases are built strictly from release tags.

## User Stories

1. As a release operator, I want a single command to build and push both services for dev, so that I can ship changes quickly without manual step drift.
2. As a release operator, I want dev image tags based on commit SHA, so that every deployed dev artifact is immutable and traceable.
3. As a release operator, I want prod image builds restricted to release tags only, so that production artifacts map exactly to approved versioned source.
4. As an engineering lead, I want build scripts to fail on dirty git state, so that untracked local changes never leak into release artifacts.
5. As a backend engineer, I want independent release of core and ai services, so that each service can ship on its own cadence.
6. As an SRE, I want mandatory test/lint gates before image push, so that broken builds are blocked early.
7. As a security reviewer, I want signed images, so that deployment can verify artifact authenticity.
8. As a security reviewer, I want SBOM and vulnerability scan outputs for every release, so that supply-chain posture is auditable.
9. As an incident responder, I want image labels containing version, SHA, and build timestamp, so that runtime diagnostics can tie behavior to exact artifacts.
10. As a platform maintainer, I want standardized naming in Docker Hub repos, so that repository permissions and lifecycle management stay predictable.
11. As a team member, I want a convenience moving tag in each lane, so that local/manual checks remain easy while immutable tags remain the source of truth.
12. As a maintainer, I want release scripts to support dry-run mode, so that operator confidence is high before making external changes.
13. As a maintainer, I want the prod release command to support deploying only one service when needed, so that urgent fixes can ship with minimal risk.
14. As a maintainer, I want explicit controls for breaking migration deployments, so that dangerous releases require deliberate approval.

## Implementation Decisions

- Create a release orchestration module that validates git state, branch/tag rules, and operator inputs.
- Create service build modules with shared conventions for tags, OCI labels, and architecture target (`linux/amd64`).
- Create quality-gate adapters that run service-appropriate checks before build and push.
- Create registry publish module with lane-specific tagging rules (`dev-<sha>` and `vX.Y.Z`).
- Create signing module that performs Cosign key-based signing for each pushed artifact.
- Create software-composition module that generates SBOM and vulnerability scan evidence per artifact.
- Define version-injection contract so runtime service version reports match deployed image tag.
- Define release operator policy restricting who can run prod release commands.

## Testing Decisions

- A good test validates externally observable behavior of release orchestration (inputs, outputs, exit codes, side effects), not internal shell command implementation details.
- Test the release orchestration module with success/failure paths for dirty repo, invalid tag, missing credentials, and failed quality gates.
- Test tagging policy logic for dev and prod lane correctness and independent-service release cases.
- Test signing/verification workflow at the artifact boundary to ensure unsigned artifacts are detectable.
- Test version-injection contract by asserting runtime-reported version equals requested release tag.
- Prior art should follow existing service-level unit/integration test style used in the codebase for deterministic command orchestration and behavior-first assertions.

## Out of Scope

- Moving builds to CI-hosted runners in this phase.
- Multi-architecture image publishing.
- Fully automated prod release triggering without operator involvement.
- Kubernetes-native artifact pipelines.

## Further Notes

- This PRD intentionally optimizes for low-cost operation while preserving production controls.
- The design keeps a future migration path to server-side/CI builds by preserving stable script contracts.
