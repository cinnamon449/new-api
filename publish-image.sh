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

# Colors only on a terminal, and never when NO_COLOR is set.
result='' heading='' command='' reset='' err_warning='' err_reset=''
if [[ -z "${NO_COLOR:-}" ]]; then
  if [[ -t 1 ]]; then
    result=$'\e[1;96m' heading=$'\e[1;97m' command=$'\e[92m' reset=$'\e[0m'
  fi
  if [[ -t 2 ]]; then
    err_warning=$'\e[91m' err_reset=$'\e[0m'
  fi
fi

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
  echo "${err_warning}WARN: ghcr.io not found in ~/.docker/config.json.${err_reset}" >&2
  echo "${err_warning}      If the push fails, run: docker login ghcr.io -u cinnamon449${err_reset}" >&2
fi

echo "${heading}==> Building ${IMAGE}:${GIT_SHA} and ${IMAGE}:latest ...${reset}"
docker build -t "${IMAGE}:latest" -t "${IMAGE}:${GIT_SHA}" .

if [[ "${PUSH}" -eq 1 ]]; then
  echo "${heading}==> Pushing ...${reset}"
  docker push "${IMAGE}:latest"
  docker push "${IMAGE}:${GIT_SHA}"
  echo
  echo "${result}==> Done. Published ${IMAGE}:${GIT_SHA} and :latest${reset}"
  echo
  echo "${heading}On the server run:${reset}"
  echo "  ${command}cd /opt/interapi && sudo docker compose up -d --pull always --no-deps new-api${reset}"
else
  echo "${result}==> Built locally (--no-push); tags were not pushed.${reset}"
fi
