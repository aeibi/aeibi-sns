#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

IMAGE_NAME="${IMAGE_NAME:-aeibi-api:latest}"
OUT_TAR="${OUT_TAR:-release/aeibi-api-image.tar}"
GOOS_TARGET="${GOOS_TARGET:-linux}"
GOARCH_TARGET="${GOARCH_TARGET:-amd64}"
RUNTIME_DOCKERFILE="${RUNTIME_DOCKERFILE:-docker/runtime/Dockerfile}"

echo "[1/3] Build binary (${GOOS_TARGET}/${GOARCH_TARGET})"
mkdir -p release/bin
CGO_ENABLED=0 GOOS="$GOOS_TARGET" GOARCH="$GOARCH_TARGET" \
  go build -trimpath -ldflags="-s -w" -o release/bin/aeibi ./cmd

echo "[2/3] Build runtime image (${IMAGE_NAME})"
docker build -f "$RUNTIME_DOCKERFILE" -t "$IMAGE_NAME" .

echo "[3/3] Export image tar (${OUT_TAR})"
mkdir -p "$(dirname "$OUT_TAR")"
docker save -o "$OUT_TAR" "$IMAGE_NAME"

echo "Done: $OUT_TAR"
