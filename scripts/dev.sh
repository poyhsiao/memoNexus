#!/usr/bin/env bash
# Development environment launcher for MemoNexus
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Check for required tools
command -v go >/dev/null 2>&1 || { echo "Go is required but not installed. Aborting." >&2; exit 1; }
command -v flutter >/dev/null 2>&1 || { echo "Flutter is required but not installed. Aborting." >&2; exit 1; }

echo "Starting MemoNexus development environment..."

# Terminal 1: Start Go Core / PocketBase
echo "Starting Go Core backend..."
cd "$PROJECT_ROOT/packages/backend"
go run cmd/desktop/main.go &
BACKEND_PID=$!
echo "Backend started (PID: $BACKEND_PID)"

# Wait for backend to be ready
sleep 2

# Terminal 2: Start Flutter app
echo "Starting Flutter frontend..."
cd "$PROJECT_ROOT/apps/frontend"
flutter run &
FRONTEND_PID=$!
echo "Frontend started (PID: $FRONTEND_PID)"

# Handle cleanup
trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true" EXIT

echo "Development environment started!"
echo "Press Ctrl+C to stop all services"

# Wait for any process to exit
wait
