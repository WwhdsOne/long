#!/bin/sh

set -eu

if [ "$#" -lt 1 ]; then
  echo "Usage: bash ./scripts/build-save-upload.sh <user@host:/remote/path/> [image_name] [archive_name]"
  exit 1
fi

REMOTE_TARGET="$1"
IMAGE_NAME="${2:-long}"
ARCHIVE_NAME="${3:-${IMAGE_NAME}.tar.gz}"
ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

cd "$ROOT_DIR"

echo "Removing old local image if it exists..."
docker rmi -f "$IMAGE_NAME" >/dev/null 2>&1 || true

echo "Removing old local archive if it exists..."
rm -f "$ARCHIVE_NAME"

echo "Building image: $IMAGE_NAME"
docker buildx build --platform linux/amd64 -t "$IMAGE_NAME" . --load

echo "Saving archive: $ARCHIVE_NAME"
docker save "$IMAGE_NAME" | gzip > "$ARCHIVE_NAME"

echo "Uploading archive to: $REMOTE_TARGET"
scp "$ARCHIVE_NAME" "$REMOTE_TARGET"

echo "Done."
