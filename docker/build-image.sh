#!/usr/bin/env bash
# Builds the appliance image and packs it as a tar.gz, so it can be loaded on a
# server without a registry:
#
#   ./docker/build-image.sh            # rustdesk-api-server-pro:latest
#   ./docker/build-image.sh myrepo/rustdesk-api-server-pro:1.0
#
# Environment:
#   IMAGE     image name:tag (default: rustdesk-api-server-pro:latest)
#   OUTPUT    output archive (default: build/<name>-<tag>-<arch>.tar.gz)
#   PLATFORM  docker build --platform value, e.g. "linux/arm64" (default: docker host)
set -euo pipefail

cd "$(dirname "$0")/.."

IMAGE="${1:-${IMAGE:-rustdesk-api-server-pro:latest}}"
case "${IMAGE}" in
    *:*) ;;
    *) IMAGE="${IMAGE}:latest" ;;
esac
NAME="${IMAGE%:*}"
TAG="${IMAGE##*:}"
ARCH="$(uname -m)"
if [ -n "${PLATFORM:-}" ]; then
    ARCH="${PLATFORM##*/}"
fi
OUTPUT="${OUTPUT:-build/${NAME}-${TAG}-${ARCH}.tar.gz}"

if ! command -v docker >/dev/null 2>&1; then
    echo "docker is not installed" >&2
    exit 1
fi

if [ ! -f rustdesk/build_web.sh ]; then
    echo "rustdesk submodule is not checked out, getting it..."
    git submodule update --init rustdesk
fi
if [ ! -f rustdesk/build_web.sh ]; then
    echo "rustdesk/build_web.sh not found, run 'git submodule update --init rustdesk'" >&2
    exit 1
fi

echo "==> building ${IMAGE}${PLATFORM:+ for ${PLATFORM}}"
docker build ${PLATFORM:+--platform "${PLATFORM}"} --tag "${IMAGE}" .

mkdir -p "$(dirname "${OUTPUT}")"
echo "==> saving ${IMAGE} to ${OUTPUT}"
docker save "${IMAGE}" | gzip > "${OUTPUT}"

echo
echo "image:  ${IMAGE}"
echo "bundle: ${OUTPUT} ($(du -h "${OUTPUT}" | cut -f1))"
echo "load:   docker load -i $(basename "${OUTPUT}")"
