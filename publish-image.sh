#!/usr/bin/env bash
# Build and publish your new-api fork image to GHCR.
#
# Usage:
#   ./publish-image.sh            build + tag + push (default)
#   ./publish-image.sh --no-push  build + tag only, do not push
#
# Override the target image (e.g. to push to Docker Hub instead):
#   IMAGE=docker.io/cinnamon449/new-api ./publish-image.sh
#
# One-time auth on this dev machine (token needs write:packages):
#   docker login ghcr.io -u cinnamon449 -p <TOKEN>
# One-time auth on the server for a PRIVATE image (token needs read:packages):
#   docker login ghcr.io -u cinnamon449 -p <TOKEN>
#
# Tags produced:
#   ghcr.io/cinnamon449/new-api:latest        always newest (what docker-compose.yml pulls)
#   ghcr.io/cinnamon449/new-api:<git-sha>     immutable, for traceability / rollback

set -euo pipefail

IMAGE="${IMAGE:-ghcr.io/cinnamon449/new-api}"
GIT_SHA="$(git rev-parse --short HEAD)"
PUSH=1

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-push) PUSH=0 ;;
    *) echo "Unknown argument: $1" >&2; exit 1 ;;
  esac
  shift
done

if ! grep -q '"ghcr.io"' ~/.docker/config.json 2>/dev/null; then
  echo "WARN: ghcr.io not found in ~/.docker/config.json." >&2
  echo "      If the push fails, run: docker login ghcr.io -u cinnamon449" >&2
fi

echo "==> Building ${IMAGE}:${GIT_SHA} and ${IMAGE}:latest ..."
docker build -t "${IMAGE}:latest" -t "${IMAGE}:${GIT_SHA}" .

if [[ "${PUSH}" -eq 1 ]]; then
  echo "==> Pushing ..."
  docker push "${IMAGE}:latest"
  docker push "${IMAGE}:${GIT_SHA}"
  echo "==> Done. On the server run:"
  echo "    docker compose pull && docker compose down && docker compose up -d"
else
  echo "==> Built locally (--no-push); tags were not pushed."
fi
