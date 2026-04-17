#!/bin/sh

set -eu

IMAGE_NAME="long"
CONTAINER_NAME="long"
ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

cd "$ROOT_DIR"

echo "Stopping old container if it exists..."
docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true

echo "Removing old image if it exists..."
docker rmi -f "$IMAGE_NAME" >/dev/null 2>&1 || true

echo "Building image: $IMAGE_NAME"
docker build -t "$IMAGE_NAME" .

echo "Starting container: $CONTAINER_NAME"
docker run -d \
  --name "$CONTAINER_NAME" \
  --restart unless-stopped \
  --network host \
  "$IMAGE_NAME"

echo "Container is up with host networking on port 2333."
