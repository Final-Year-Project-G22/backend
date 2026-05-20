#!/usr/bin/env bash
set -euo pipefail

SERVICE=$1
TAG=$2
ACTION=${3:-push}
COMPOSE_ID=${4:-}

OWNER=${GITHUB_REPOSITORY_OWNER}
REGISTRY="ghcr.io/${OWNER}/${SERVICE}"

CACHE_FROM="type=gha"
CACHE_TO="type=gha,mode=max,ignore-error=true"

build_and_push() {
  local moving_tag
  if [[ "${TAG}" == dev-* ]]; then
    moving_tag="dev"
  else
    moving_tag="latest"
  fi

  echo "::group::Building and pushing ${REGISTRY}:${TAG} + ${REGISTRY}:${moving_tag}"
  docker buildx build \
    --platform linux/amd64 \
    --label "org.opencontainers.image.revision=${GITHUB_SHA:-unknown}" \
    --label "org.opencontainers.image.version=${TAG}" \
    --label "org.opencontainers.image.created=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    --cache-from "${CACHE_FROM}" \
    --cache-to "${CACHE_TO}" \
    -f "${SERVICE}/Dockerfile" \
    -t "${REGISTRY}:${TAG}" \
    -t "${REGISTRY}:${moving_tag}" \
    --push \
    .
  echo "::endgroup::"
}

deploy() {
  local compose_id=$1
  if [[ -z "${compose_id}" ]]; then
    echo "Error: compose-id is required for deploy action"
    exit 1
  fi
  echo "::group::Triggering Dokploy deploy for ${compose_id}"
  curl -s -X POST "${DOKPLOY_URL}/api/compose.deploy" \
    -H "x-api-key: ${DOKPLOY_TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{\"composeId\": \"${compose_id}\"}"
  echo "::endgroup::"
}

case "${ACTION}" in
  build)
    build_and_push
    ;;
  push)
    build_and_push
    ;;
  deploy)
    deploy "${COMPOSE_ID}"
    ;;
  *)
    echo "Usage: $0 <service> <tag> [build|push|deploy <compose-id>]"
    exit 1
    ;;
esac
