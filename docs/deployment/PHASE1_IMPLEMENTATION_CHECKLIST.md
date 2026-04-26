# Phase 1 Implementation Checklist

This checklist covers execution from preconditions through completion of **Phase 1 (Build and Artifact Pipeline - Local)** from the deployment blueprint.

Use it as an operator run-sheet. Do not move to the next section until all items in the current section are complete.

---

## 0) Exit Criteria (Definition of Done)

Phase 1 is complete only when all of the following are true:

- [ ] Both services have production-ready Dockerfiles and `.dockerignore` is in place.
- [ ] Local release scripts exist for dev and prod lanes and support `--dry-run`.
- [ ] Release scripts enforce clean git state + branch/tag policy.
- [ ] Release scripts run quality gates before any build/push.
- [ ] Images are built as `linux/amd64` with immutable tags (`dev-<sha>` and `vX.Y.Z`).
- [ ] Images are pushed to Docker Hub service repos.
- [ ] Images are signed with Cosign.
- [ ] SBOM and vulnerability scan reports are generated for each built image.
- [ ] Runtime version injection contract is defined and validated against requested image tag.
- [ ] Operator policy (authorized releasers only) is documented and acknowledged.

---

## 1) Preconditions and Security Cleanup

### 1.1 Secret Exposure Response

- [ ] Inventory all currently exposed credentials (DB, JWT, OAuth, email, API keys, etc.).
- [ ] Rotate all exposed credentials.
- [ ] Invalidate old credentials and confirm they no longer work.
- [ ] Remove plaintext secrets from repository-tracked env files.
- [ ] Replace committed envs with safe templates where required.

### 1.2 Tooling Baseline (Operator Machine)

- [ ] Docker and Docker Buildx installed and functional.
- [ ] Cosign installed and available in PATH.
- [ ] SBOM scanner tooling installed (for example Syft).
- [ ] Vulnerability scanner tooling installed (for example Trivy/Grype).
- [ ] Git, Go, Python, uv, and project prerequisites available.

### 1.3 Release Operator Access

- [ ] Docker Hub account/org permissions verified for image push.
- [ ] Docker Hub access token configured (not account password).
- [ ] Authorized release operators list finalized (1-2 maintainers).
- [ ] Non-authorized users are not able to run production release flow.

---

## 2) Build Definitions

### 2.1 Dockerfiles

- [ ] Create Dockerfile for `core-backend` with reproducible multi-stage build.
- [ ] Create Dockerfile for `ai-service` with reproducible dependency installation and runtime image.
- [ ] Ensure both Dockerfiles run as non-root user where feasible.
- [ ] Add OCI labels (`revision`, `version`, `created`) in image metadata.

### 2.2 Build Context Hygiene

- [ ] Add root `.dockerignore` to exclude unnecessary files (git metadata, caches, local artifacts, test outputs, secrets).
- [ ] Validate no secrets are copied into image build context.
- [ ] Validate resulting image sizes are acceptable for push/pull efficiency.

---

## 3) Release Script Framework

### 3.1 Script Skeleton

- [ ] Create shared release helper module (input parsing, logging, errors, common validations).
- [ ] Create `release-dev` script.
- [ ] Create `release-prod` script.
- [ ] Add `--dry-run` support to both scripts.
- [ ] Add structured, parseable script logging output.

### 3.2 Policy Gates

- [ ] Enforce clean git state before build starts.
- [ ] Dev flow allowed only from intended dev branch flow.
- [ ] Prod flow allowed only from checked-out release tags.
- [ ] Validate requested prod versions are semver tags.
- [ ] Validate at least one service is selected for release.

---

## 4) Quality Gates (Before Build/Push)

### 4.1 Core Service Gates

- [ ] Define and wire core checks (tests/lint/format checks as agreed).
- [ ] Ensure release halts immediately on core gate failure.

### 4.2 AI Service Gates

- [ ] Define and wire ai checks (tests/lint/type/security checks as agreed).
- [ ] Ensure release halts immediately on ai gate failure.

### 4.3 Independent Service Release

- [ ] Support releasing core only.
- [ ] Support releasing ai only.
- [ ] Support releasing both in one run.

---

## 5) Tagging, Build, and Push

### 5.1 Dev Lane

- [ ] Generate immutable tags in format `dev-<shortsha>`.
- [ ] Optionally update moving `dev` tag after immutable push succeeds.
- [ ] Push dev artifacts to service-specific Docker Hub repositories.

### 5.2 Prod Lane

- [ ] Build from release tag source only.
- [ ] Tag as `vX.Y.Z` per service.
- [ ] Optionally update moving `latest` after immutable push succeeds.
- [ ] Push prod artifacts to service-specific Docker Hub repositories.

### 5.3 Architecture Constraint

- [ ] Enforce `linux/amd64` target builds.

---

## 6) Signing, SBOM, and Security Evidence

### 6.1 Cosign Setup

- [ ] Generate/secure Cosign key pair.
- [ ] Store private key only on authorized operator machines.
- [ ] Protect key with passphrase and maintain encrypted offline backup.
- [ ] Document key rotation procedure.

### 6.2 Signing Enforcement

- [ ] Sign each pushed image digest.
- [ ] Verify signatures as part of release completion checks.
- [ ] Fail release if signing or verification fails.

### 6.3 SBOM and Vulnerability Evidence

- [ ] Generate SBOM per built image.
- [ ] Run vulnerability scan per built image.
- [ ] Archive artifacts in predictable location with release metadata.
- [ ] Fail release on configured severity threshold.

---

## 7) Runtime Version Contract

- [ ] Define how deployed tag maps to runtime `APP_VERSION` for both services.
- [ ] Ensure release scripts pass/inject runtime version consistently.
- [ ] Add post-release validation that runtime-reported version matches requested tag.
- [ ] Confirm mismatch causes release failure.

---

## 8) Makefile / Operator UX

- [ ] Add make targets for dev release flow.
- [ ] Add make targets for prod release flow.
- [ ] Add make targets for dry-run execution.
- [ ] Add concise operator docs for common commands.

---

## 9) Validation Runs (Must Pass)

### 9.1 Dev Validation

- [ ] Run `release-dev --dry-run` and verify all planned actions/logging.
- [ ] Run real dev release for one service and verify image appears in Docker Hub with expected tags/signature/SBOM/scan.
- [ ] Run real dev release for both services and verify both artifacts.

### 9.2 Prod Validation (Non-Production Dry Validation)

- [ ] Run `release-prod --dry-run` with valid version tags and verify policy checks.
- [ ] Validate prod flow rejects branch-head builds.
- [ ] Validate prod flow rejects dirty git state.
- [ ] Validate prod flow rejects invalid semver tags.

---

## 10) Documentation and Handover

- [ ] Update deployment docs with final release commands and examples.
- [ ] Document operator policy and approval flow.
- [ ] Document failure modes and expected operator actions.
- [ ] Document where SBOM/scan/signature outputs are stored.

---

## 11) Phase 1 Sign-Off Checklist

- [ ] Technical sign-off by maintainer responsible for core service.
- [ ] Technical sign-off by maintainer responsible for ai service.
- [ ] Security sign-off for signing + evidence flow.
- [ ] Operator sign-off that runbook is executable end-to-end.
- [ ] Final confirmation that Phase 1 exit criteria (Section 0) are all checked.
