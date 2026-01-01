#!/usr/bin/env bash
# Unified build script for MemoNexus monorepo
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Building MemoNexus..."

# Parse arguments
BUILD_TYPE=${1:-debug}
PLATFORMS=${2:-all}

# Build Go Core (shared library for mobile, binary for desktop)
# Use novidcodec tag to skip optional ffmpeg dependencies
echo "Building Go Core..."
cd "$PROJECT_ROOT/packages/backend"

if [[ "$BUILD_TYPE" == "release" ]]; then
  go build -tags=novidcodec -ldflags="-s -w" -o "$PROJECT_ROOT/build/bin/core" ./cmd/core
else
  go build -tags=novidcodec -o "$PROJECT_ROOT/build/bin/core" ./cmd/core
fi

# Build Flutter app
echo "Building Flutter app..."
cd "$PROJECT_ROOT/apps/frontend"

if [[ "$PLATFORMS" == "all" ]] || [[ "$PLATFORMS" == "desktop" ]]; then
  if [[ "$BUILD_TYPE" == "release" ]]; then
    flutter build desktop --release
  else
    flutter build desktop --debug
  fi
fi

# Package desktop wrapper (optional)
if [[ "$PLATFORMS" == "all" ]] || [[ "$PLATFORMS" == "desktop" ]]; then
  cd "$PROJECT_ROOT/apps/desktop"
  if [[ "$BUILD_TYPE" == "release" ]]; then
    pnpm build
  else
    pnpm dev
  fi
fi

echo "Build complete: $PROJECT_ROOT/build/dist/"
